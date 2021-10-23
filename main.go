package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Listen for receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	client := &http.Client{}
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if tiktokUrl, err := isTiktokUrl(m.Content, client); tiktokUrl != "" {
		// Matched a tiktok url
		log.Debugf("Matched a tiktok url: %v", tiktokUrl)

		if err != nil {
			log.Fatalf("Error retrieving tiktok URL: %v", nil)
		}
		videoBytes, err := getTiktokVideo(tiktokUrl, client)
		if err != nil {
			log.Fatalf("Error getting tiktok video: %v", err)
		}
		fmt.Println("Sending the video!")
		s.ChannelFileSend(m.ChannelID, "video.mp4", bytes.NewReader(videoBytes))
	}
}
