package handlers

import (
	"context"
	"fmt"
	"start-feishubot/services"
	"start-feishubot/utils"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type MsgInfo struct {
	handlerType       HandlerType
	msgType           string
	msgId             *string
	chatId            *string
	qParsed           string
	sessionId         *string
	mention           []*larkim.MentionEvent
	isImage           bool
	imgbs64           string
	isLifeStreaming   bool
	lifeStreamingType string
	name              string
}
type ActionInfo struct {
	handler *MessageHandler
	ctx     *context.Context
	info    *MsgInfo
}

type Action interface {
	Execute(a *ActionInfo) bool
}

type ProcessedUnique struct { //消息唯一性
}

func (*ProcessedUnique) Execute(a *ActionInfo) bool {
	if a.handler.msgCache.IfProcessed(*a.info.msgId) {
		return false
	}
	a.handler.msgCache.TagProcessed(*a.info.msgId)
	return true
}

type ProcessMention struct { //是否机器人应该处理
}

func (*ProcessMention) Execute(a *ActionInfo) bool {
	// 私聊直接过
	if a.info.handlerType == UserHandler {
		return true
	}
	// 群聊判断是否提到机器人
	if a.info.handlerType == GroupHandler {
		if a.handler.judgeIfMentionMe(a.info.mention) {
			return true
		}
		return false
	}
	return false
}

type EmptyAction struct { /*空消息*/
}

func (*EmptyAction) Execute(a *ActionInfo) bool {
	if !a.info.isImage {

		if len(a.info.qParsed) == 0 {
			sendMsg(*a.ctx, "🤖️：你想知道什么呢~", a.info.chatId)
			fmt.Println("msgId", *a.info.msgId,
				"message.text is empty")
			return false
		}
	}
	return true
}

type ClearAction struct { /*清除消息*/
}

func (*ClearAction) Execute(a *ActionInfo) bool {
	if _, foundClear := utils.EitherTrimEqual(a.info.qParsed,
		"/clear", "清除"); foundClear {
		sendClearCacheCheckCard(*a.ctx, a.info.sessionId,
			a.info.msgId)
		return false
	}
	return true
}

type RolePlayAction struct { /*角色扮演*/
}

func (*RolePlayAction) Execute(a *ActionInfo) bool {
	if system, foundSystem := utils.EitherCutPrefix(a.info.qParsed,
		"/system ", "角色扮演 "); foundSystem {
		a.handler.sessionCache.Clear(*a.info.sessionId)
		systemMsg := append([]services.Messages{}, services.Messages{
			Role: "system", Content: system,
		})
		a.handler.sessionCache.SetMsg(*a.info.sessionId, systemMsg)
		sendSystemInstructionCard(*a.ctx, a.info.sessionId,
			a.info.msgId, system)
		return false
	}
	return true
}

type HelpAction struct { /*帮助*/
}

func (*HelpAction) Execute(a *ActionInfo) bool {
	if _, foundHelp := utils.EitherTrimEqual(a.info.qParsed, "/help",
		"帮助"); foundHelp {
		sendHelpCard(*a.ctx, a.info.sessionId, a.info.msgId)
		return false
	}
	return true
}

type PicAction struct { /*图片*/
}

func (*PicAction) Execute(a *ActionInfo) bool {

	// 开启图片创作模式
	if !a.info.isLifeStreaming {
		if _, foundPic := utils.EitherTrimEqual(a.info.qParsed,
			"/picture", "图片创作"); foundPic {
			a.handler.sessionCache.Clear(*a.info.sessionId)
			a.handler.sessionCache.SetMode(*a.info.sessionId,
				services.ModePicCreate)
			sendPicCreateInstructionCard(*a.ctx, a.info.sessionId,
				a.info.msgId)
			return false
		}

		// 生成图片
		mode := a.handler.sessionCache.GetMode(*a.info.sessionId)
		if mode == services.ModePicCreate || a.info.isImage {
			// bs64, err := a.handler.gpt.GenerateOneImage(a.info.qParsed,
			// 	"256x256")
			var bs64 string
			var err error
			if a.info.isImage {
				if a.info.qParsed == "" {
					clipResp, err := services.TryCLIPINFO(a.info.imgbs64)
					err = replyMsg(*a.ctx, clipResp, a.info.msgId)
					if err != nil {
						replyMsg(*a.ctx, fmt.Sprintf(
							"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
						return false
					}
					return false
				} else {
					bs64, err = services.TrySDI2I(a.info.imgbs64, a.info.qParsed)
				}

			} else if a.info.msgType == "text" {
				bs64, err = services.TrySDT2I(a.info.qParsed)
			}

			if err != nil {
				replyMsg(*a.ctx, fmt.Sprintf(
					"🤖️：图片生成失败，请稍后再试～\n错误信息: %v", err), a.info.msgId)
				return false
			}
			replayImageByBase64(*a.ctx, bs64, a.info.msgId)
			return false
		}
	}
	return true
}

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	if !a.info.isLifeStreaming {
		msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
		msg = append(msg, services.Messages{
			Role: "user", Content: a.info.qParsed,
		})
		completions, err := a.handler.gpt.Completions(msg)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
			return false
		}
		msg = append(msg, completions)
		a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
		//if new topic
		if len(msg) == 2 {
			fmt.Println("new topic", msg[1].Content)
			sendNewTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
				completions.Content)
			return false
		}
		err = replyMsg(*a.ctx, completions.Content, a.info.msgId)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
			return false
		}
	}
	return true
}

type LifeStreamAction struct { /*消息*/
}

func (*LifeStreamAction) Execute(a *ActionInfo) bool {

	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)

	if a.info.isLifeStreaming {
		fmt.Println("LifeStreamAction")
		// bs64, err := a.handler.gpt.GenerateOneImage(a.info.qParsed,
		// 	"256x256")

		var bs64, propmt string
		var imageKey *string
		var err error

		promptTemplateGolden := "金句是指在一篇文章、演讲或对话中，具有特别重要或独特意义，并能够简洁表达观点的句子。通常被引用和传播，具有较高的价值和影响力。下面是一个电影片段的画面和独白，请你为这一片段的宣传海报写出一句中文金句。请直接写出金句，不要在前面增加`金句`\n：画面：%s \n, 独白：%s"

		promptTemplateStory := `下面是一本小说的设定。\n - 主题：%s \n - 作者：余秋雨 \n - 主角：%s \n  按照设定，写一段有强烈剧情反转的小说片段。这是小说的第一幕场景：“%s”。最后，你需要根据小说的内容，生成一个主角需要完成的任务，用“任务：”这样的格式写出来。`

		// msg = append(msg, services.Messages{
		// 	Role: "user", Content: "use words below to make up a interesting story",
		// })
		msg = append(msg, services.Messages{
			Role: "user", Content: a.info.qParsed,
		})
		if a.info.isImage {
			clipResp, err := services.TryCLIPINFO(a.info.imgbs64)
			fmt.Println("clip is :", clipResp)

			if err == nil {
				if a.info.lifeStreamingType == "story" {
					propmt = fmt.Sprintf(promptTemplateStory, a.info.qParsed, a.info.name, strings.Split(clipResp, ",")[0])
				}
				if a.info.lifeStreamingType == "golden" {
					propmt = fmt.Sprintf(promptTemplateGolden, strings.Split(clipResp, ",")[0], a.info.qParsed)
				}

				msg = append(msg, services.Messages{
					Role: "user", Content: propmt,
				})
				bs64, err = services.TrySDI2I(a.info.imgbs64, a.info.qParsed+clipResp)
				if a.info.lifeStreamingType == "story" {
					imageKey, err = uploadImage(bs64)
				}
			}
		}

		completions, err := a.handler.gpt.Completions(msg)

		if a.info.lifeStreamingType == "story" {
			sendLifeStreamCard(*a.ctx, completions.Content, imageKey, a.info.msgId)
		} else {
			err = replayImageByBase64WithLabel(*a.ctx, bs64, completions.Content, a.info.msgId)
		}

		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️：生活流生成错误，请稍后再试～\n错误信息: %v", err), a.info.msgId)
			return false
		}

	}

	return true
}
