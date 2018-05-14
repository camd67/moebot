package reddit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/turnage/graw/reddit"
)

type Handle struct {
	bot reddit.Bot
}

func NewHandle(filePath string) (*Handle, error) {
	handle, err := reddit.NewBotFromAgentFile(filePath, 0)
	if err != nil {
		fmt.Print("error from getting new bot")
		handle = nil
	}

	return &Handle{bot: handle}, err
}

func (handle *Handle) GetRandomImage(subreddit string) (*discordgo.MessageSend, error) {
	harvest, err := handle.getListing(subreddit)
	if err != nil {
		fmt.Printf("Error getting harvest from subreddit %s", subreddit)
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	randPost := harvest.Posts[rand.Intn(len(harvest.Posts)-1)]

	resp, err := http.Get(randPost.URL)
	if err != nil {
		fmt.Printf("Error requesting image: " + randPost.URL)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error preparing repsonse body")
		return nil, err
	}

	content := "content ;3"
	return &discordgo.MessageSend{
		Content: content,
		File: &discordgo.File{
			Name:        "Spoiler.gif",
			ContentType: "image/gif",
			Reader:      bytes.NewReader(body),
		},
	}, err
}

func (handle *Handle) getListing(subreddit string) (reddit.Harvest, error) {
	return handle.bot.Listing(subreddit, "")
}
