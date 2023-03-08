package services

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var BotId string
var goBot *discordgo.Session

func TryDiscord() {
	goBot, err := discordgo.New("Bot " + "MTA3MDIyMjUyODU0ODcwNDM2Ng.Gc5RHY.HLBTUOgxD0QTIJ7r7_YuAC7X6oZuc0Mc0psjJ8")

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
	fmt.Println("get message:", m.Content)
	if m.Author.ID == BotId {
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "啊对对对")

}
