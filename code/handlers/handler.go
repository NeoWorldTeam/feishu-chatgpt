package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"start-feishubot/initialization"
	"start-feishubot/services"
	"strings"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	// larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type PostMessage struct {
	Title   string `json:"title"`
	Content [][]struct {
		Tag      string `json:"tag"`
		Text     string `json:"text,omitempty"`
		Href     string `json:"href,omitempty"`
		UserID   string `json:"user_id,omitempty"`
		UserName string `json:"user_name,omitempty"`
	} `json:"content"`
}

// 责任链
func chain(data *ActionInfo, actions ...Action) bool {
	for _, v := range actions {
		if !v.Execute(data) {
			return false
		}
	}
	return true
}

type MessageHandler struct {
	sessionCache services.SessionServiceCacheInterface
	msgCache     services.MsgCacheInterface
	gpt          services.ChatGPT
	config       initialization.Config
}

func (m MessageHandler) cardHandler(_ context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
	var cardMsg CardMsg
	actionValue := cardAction.Action.Value
	actionValueJson, _ := json.Marshal(actionValue)
	json.Unmarshal(actionValueJson, &cardMsg)
	if cardMsg.Kind == ClearCardKind {
		newCard, err, done := CommonProcessClearCache(cardMsg, m.sessionCache)
		if done {
			return newCard, err
		}
	}
	return nil, nil
}

func CommonProcessClearCache(cardMsg CardMsg, session services.SessionServiceCacheInterface) (
	interface{}, error, bool) {
	if cardMsg.Value == "1" {
		newCard, _ := newSendCard(
			withHeader("️🆑 机器人提醒", larkcard.TemplateRed),
			withMainMd("已删除此话题的上下文信息"),
			withNote("我们可以开始一个全新的话题，继续找我聊天吧"),
		)
		session.Clear(cardMsg.SessionId)
		return newCard, nil, true
	}
	if cardMsg.Value == "0" {
		newCard, _ := newSendCard(
			withHeader("️🆑 机器人提醒", larkcard.TemplateGreen),
			withMainMd("依旧保留此话题的上下文信息"),
			withNote("我们可以继续探讨这个话题,期待和您聊天。如果您有其他问题或者想要讨论的话题，请告诉我哦"),
		)
		return newCard, nil, true
	}
	return nil, nil, false
}

func (m MessageHandler) msgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	handlerType := judgeChatType(event)
	fmt.Println("openID:", *event.Event.Sender.SenderId.OpenId)
	name := "默认名字"
	req := larkim.NewGetChatMembersReqBuilder().ChatId(*event.Event.Message.ChatId).Build()
	resp, _ := initialization.GetLarkClient().Im.ChatMembers.Get(ctx, req)
	for _, item := range resp.Data.Items {
		if *item.MemberId == *event.Event.Sender.SenderId.OpenId {
			name = *item.Name
		}
	}
	// fmt.Println("chat members:", larkcore.Prettify(resp))
	if handlerType == "otherChat" {
		fmt.Println("unknown chat type")
		return nil
	}
	imgbs64 := ""
	var contentMap map[string]interface{}
	msgType := judgeMsgType(event)
	fmt.Println("msg type:", msgType, *event.Event.Message.Content)
	isImage := false
	prompt := ""
	isLifeStreaming := false
	lifeStreamingType := "story"
	if msgType != "text" {

		if msgType == "image" {
			image := &larkim.CreateImageRespData{}
			json.Unmarshal([]byte(*event.Event.Message.Content), image)
			req := larkim.NewGetImageReqBuilder().
				ImageKey(*image.ImageKey).
				Build()
			resp, _ := initialization.GetLarkClient().Im.Image.Get(ctx, req)
			// resp.WriteFile(*image.ImageKey)
			imgbs64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(resp.RawBody)
			// fmt.Println("get image by key", *image.ImageKey, err)
			isImage = true

			// return nil
		} else if msgType == "post" {
			json.Unmarshal([]byte(*event.Event.Message.Content), &contentMap)
			if contentMap["content"] != nil {
				fmt.Println("post content:", contentMap["content"])
				v := reflect.ValueOf(contentMap["content"])
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				if v.Kind() != reflect.Slice {
					panic(fmt.Errorf("forEachValue: expected slice type, found %q", v.Kind().String()))
				}

				for i := 0; i < v.Len(); i++ {
					val := v.Index(i).Interface()
					for i, item := range val.([]interface{}) {
						itemmap := item.(map[string]interface{})
						if itemmap["tag"] == "img" {

							imgKey := itemmap["image_key"].(string)
							fmt.Println("image", i, imgKey)
							req := larkim.NewGetImageReqBuilder().
								ImageKey(imgKey).
								Build()
							resp, _ := initialization.GetLarkClient().Im.Image.Get(ctx, req)
							// resp.WriteFile(imgKey)
							imgbs64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(resp.RawBody)
							// fmt.Println("get image by key", imgKey, larkcore.Prettify(resp), err)
							isImage = true
						}
						if itemmap["tag"] == "text" {
							text := strings.TrimSpace(itemmap["text"].(string))
							if strings.Contains(text, "故事模式") {
								isLifeStreaming = true
								lifeStreamingType = "story"
								prompt += strings.Replace(text, "故事模式", "", 0)

							} else if strings.Contains(text, "金句模式") {
								isLifeStreaming = true
								prompt += strings.Replace(text, "金句模式", "", 0)
								lifeStreamingType = "golden"
							} else {
								prompt += text
							}

						}
					}
				}
			}
			if contentMap["title"] != nil {
				title := contentMap["title"]
				fmt.Println("title:", title)
				// title := strings.TrimSpace(contentMap["title"].(string))
				if strings.HasSuffix(title.(string), "开始你的表演") || strings.HasSuffix(title.(string), "ction") {
					fmt.Println("title:", title)
					isLifeStreaming = true
				}

			}
		}

	}

	content := event.Event.Message.Content
	msgId := event.Event.Message.MessageId
	rootId := event.Event.Message.RootId
	chatId := event.Event.Message.ChatId
	mention := event.Event.Message.Mentions

	sessionId := rootId
	if sessionId == nil || *sessionId == "" {
		sessionId = msgId
	}
	var msgInfo MsgInfo

	if isImage || isLifeStreaming {
		msgInfo = MsgInfo{
			handlerType:       handlerType,
			msgType:           msgType,
			msgId:             msgId,
			chatId:            chatId,
			qParsed:           prompt,
			sessionId:         sessionId,
			mention:           mention,
			isImage:           isImage,
			imgbs64:           imgbs64,
			isLifeStreaming:   isLifeStreaming,
			lifeStreamingType: lifeStreamingType,
			name:              name,
		}
	} else {
		msgInfo = MsgInfo{
			handlerType:       handlerType,
			msgType:           msgType,
			msgId:             msgId,
			chatId:            chatId,
			qParsed:           strings.Trim(parseContent(*content), " "),
			sessionId:         sessionId,
			mention:           mention,
			isImage:           isImage,
			imgbs64:           imgbs64,
			isLifeStreaming:   isLifeStreaming,
			lifeStreamingType: lifeStreamingType,
			name:              name,
		}
	}
	//责任链重构示例
	data := &ActionInfo{
		ctx:     &ctx,
		handler: &m,
		info:    &msgInfo,
	}
	actions := []Action{
		&ProcessedUnique{},  //避免重复处理
		&ProcessMention{},   //判断机器人是否应该被调用
		&LifeStreamAction{}, //生活流处理
		&EmptyAction{},      //空消息处理
		&ClearAction{},      //清除消息处理
		&HelpAction{},       //帮助处理
		&RolePlayAction{},   //角色扮演处理
		&PicAction{},        //图片处理
		&MessageAction{},    //消息处理

	}
	chain(data, actions...)
	return nil
}

var _ MessageHandlerInterface = (*MessageHandler)(nil)

func NewMessageHandler(gpt services.ChatGPT,
	config initialization.Config) MessageHandlerInterface {
	return &MessageHandler{
		sessionCache: services.GetSessionCache(),
		msgCache:     services.GetMsgCache(),
		gpt:          gpt,
		config:       config,
	}
}

func (m MessageHandler) judgeIfMentionMe(mention []*larkim.
	MentionEvent) bool {
	if len(mention) != 1 {
		return false
	}
	return *mention[0].Name == m.config.FeishuBotName
}
