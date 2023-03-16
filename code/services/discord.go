package services

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
)

var BotId string
var goBot *discordgo.Session
var openaiKey string

func TryDiscord(token, gptKey string) {
	goBot, err := discordgo.New("Bot " + token)
	openaiKey = gptKey

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	BotId = u.ID

	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running fine!")

}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// promptTemplateGolden := "金句是指在一篇文章、演讲或对话中，具有特别重要或独特意义，并能够简洁表达观点的句子。通常被引用和传播，具有较高的价值和影响力。下面是一个电影片段的画面和独白，请你为这一片段的宣传海报写出一句中文金句。请直接写出金句，不要在前面增加`金句`\n：画面：%s \n, 独白：%s"

	promptTemplateStory := `下面是一本小说的设定。\n - 主题：%s \n - 作者：余秋雨 \n - 主角：%s \n  按照设定，写一段有强烈剧情反转的小说片段。这是小说的第一幕场景：“%s”。最后，你需要根据小说的内容，生成一个主角需要完成的任务，用“任务：”这样的格式写出来。`

	fmt.Println("get message:", m.Content, m.Attachments)
	if m.Author.ID == BotId {
		return
	}
	var msg []Messages
	gptResp := ""
	// if strings.TrimSpace(m.Content) != "" {
	// 	msg = append(msg, Messages{
	// 		Role: "user", Content: m.Content,
	// 	})
	// 	gpt := &ChatGPT{ApiKey: openaiKey}

	// 	completions, err := gpt.Completions(msg)
	// 	if err == nil {
	// 		gptResp = completions.Content
	// 		// _, _ = s.ChannelMessageSend(m.ChannelID, completions.Content)
	// 	}
	// }
	ebs := []*discordgo.MessageEmbed{}

	if len(m.Attachments) == 1 {
		url := *&m.Attachments[0].URL
		fmt.Println("picurl:", url)
		bs4, err := GetPicByUrl(url)
		clipResp, err := TryCLIPINFO(bs4)
		fmt.Println("clip is :", clipResp)
		propmt := fmt.Sprintf(promptTemplateStory, m.Content, "Cathy", strings.Split(clipResp, ",")[0])
		msg = append(msg, Messages{
			Role: "user", Content: propmt,
		})
		gpt := &ChatGPT{ApiKey: openaiKey}

		completions, err := gpt.Completions(msg)
		gptResp = completions.Content
		if err == nil {
			sdimgbs64, err := TrySDI2I(bs4, strings.Split(clipResp, ",")[0])
			imageBytes, err := base64.StdEncoding.DecodeString(sdimgbs64)
			fmt.Println(len(sdimgbs64), err)
			UploadFile(imageBytes, sdimgbs64[20:26]+".png")
			img := discordgo.MessageEmbedImage{
				URL: fmt.Sprintf("https://sagemaker-us-west-2-887392381071.s3.us-west-2.amazonaws.com/images/%s", sdimgbs64[0:6]+".png"),
			}
			ebs = append(ebs, &discordgo.MessageEmbed{
				URL:         url,
				Type:        discordgo.EmbedTypeImage,
				Description: "",
				Image:       &img,
			})
			_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content: gptResp,
				Embeds:  ebs,
			})
			fmt.Println("send discord complex err:", err)
		}

	}

}

func GetPicByUrl(url string) (string, error) {
	// Create a Resty Client
	client := resty.New()
	// POST JSON string
	// No need to set content type, if you have client level setting
	// r := &SDClipINFOResptBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "image/png").
		// SetBody(SDPNGINFORequestBody{Image: bs64, Model: "clip"}).
		// SetResult(r).
		Get(url)
	if err != nil {
		fmt.Println(err)
		return "", err
	} else {
		bs64 := base64.StdEncoding.EncodeToString(resp.Body())
		fmt.Println("bs64:", len(bs64))
		return bs64, err

	}

}
