package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Listen for receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
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

	reg, err := regexp.Compile(`https:\/\/vm\.tiktok\.com\/(.*)\/`)
	if err != nil {
		log.Fatalf("Error compiling regular expression %v", err)
		return
	}

	// see if url matches regex pattern
	match := reg.FindString(m.Content)

	if match != "" {

		req, err := http.NewRequest("GET", match, nil)

		if err != nil {
			log.Fatalf("Error creating HTTP Request %v", err)
			return
		}

		// convert shortened request to longer tiktok url for scraping
		headers := map[string]string{"Cookie": "69tikiman69", "User-Agent": "BOBKILL", "Referer": match}

		req.Header.Set("User-Agent", headers["User-Agent"])
		req.Header.Set("Cookie", headers["Cookie"])
		req.Header.Set("Referer", headers["Referer"])

		res, err := client.Do(req)

		if err != nil {
			log.Fatalf("Error doing the HTTP request %v", err)
			return
		}

		longUrl := strings.Split(res.Request.URL.String(), "?")[0]
		req, err = http.NewRequest("GET", longUrl, nil)

		if err != nil {
			log.Fatalf("Error creating longUrl request %v", err)
			return
		}

		req.Header.Set("User-Agent", headers["User-Agent"])
		req.Header.Set("Cookie", headers["Cookie"])
		req.Header.Set("Referer", headers["Referer"])

		res, err = client.Do(req)

		if err != nil {
			log.Fatalf("Error doing the longUrl HTTP request %v", err)
			return
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			return
		}

		bodyReplaced := strings.ReplaceAll(string(body), "amp;", "")

		videoUrlRegEx, err := regexp.Compile(`content="(https://[^w-w]{3}(.*?)tiktok\.com/video/tos(.*?)&vr=)`)
		if err != nil {
			log.Fatalf("Error compiling regular expression 2: %v", err)
			return
		}

		videoUrlMatches := videoUrlRegEx.FindAllStringSubmatch(bodyReplaced, 1)
		if len(videoUrlMatches) == 0 {
			log.Fatalf("No matches for videoUrlRegex: %v", videoUrlMatches)
		}

		videoUrl := videoUrlMatches[0][1]

		fileBytes, err := GetVideo(videoUrl, headers, client)
		if err != nil {
			log.Fatalf("Error downloading video: %v", err)
		}
		fmt.Println("Sending the video!")

		s.ChannelFileSend(m.ChannelID, "video.mp4", bytes.NewReader(fileBytes))
	}

}

func GetVideo(url string, headers map[string]string, client *http.Client) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", headers["User-Agent"])
	req.Header.Set("Cookie", headers["Cookie"])
	req.Header.Set("Referer", headers["Referer"])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fileBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return fileBytes, err
}
