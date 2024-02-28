package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	discordToken := os.Getenv("discordBotToken")
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	dg.AddHandler(processMessage)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session:", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-sc

	dg.Close()
}

func processMessage(sesssion *discordgo.Session, message *discordgo.MessageCreate) {
	fmt.Println("|")
	fmt.Println(message.Activity)
	fmt.Println(message.Application)
	fmt.Println(message.Attachments)
	fmt.Println(message.Author)
	fmt.Println(message.ChannelID)
	fmt.Println(message.Components)
	fmt.Println(message.Content)
	fmt.Println(message.EditedTimestamp)
	fmt.Println(message.Embeds)
	fmt.Println(message.Flags)
	fmt.Println(message.GuildID)
	fmt.Println(message.ID)
	fmt.Println("|")
}
