package reddit

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path"

	"github.com/bwmarrin/discordgo"
	"github.com/jzelinskie/geddit"
)

var (
	whitelistedContentTypes = map[string]bool{"image/png": true, "image/jpeg": true}
)

type Handle struct {
	session         *geddit.OAuthSession
	isAuthenticated bool
}

func NewHandle(clientID, clientSecret, username, password string) (*Handle, error) {
	session, err := geddit.NewOAuthSession(
		clientID,
		clientSecret,
		fmt.Sprintf("Discord `moebot` by %s", username),
		"http://redirect.url",
	)
	if err != nil {
		log.Println("Error getting reddit oauth session")
		return &Handle{isAuthenticated: false}, err
	}

	err = session.LoginAuth(username, password)
	if err != nil {
		return &Handle{isAuthenticated: false}, err
	}

	return &Handle{session: session, isAuthenticated: true}, err
}

func (handle *Handle) GetRandomImage(subreddit string) (*discordgo.MessageSend, error) {
	if !handle.isAuthenticated {
		return nil, errors.New("Handle was not authenticated")
	}

	posts, err := handle.getListing(subreddit)
	if err != nil {
		log.Println("Error getting listing from subreddit %s", subreddit)
		return nil, err
	}

	var resp *http.Response
	var ext string

	// Keep looking until you find an acceptable image
	for {
		randPost := posts[rand.Intn(len(posts)-1)]
		ext = path.Ext(randPost.URL)

		resp, err = http.Get(randPost.URL)
		if err != nil {
			log.Printf("Error requesting image: " + randPost.URL)
			return nil, err
		}
		defer resp.Body.Close()

		if _, ok := whitelistedContentTypes[resp.Header.Get("Content-Type")]; ok {
			break
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error preparing repsonse body")
		return nil, err
	}

	return &discordgo.MessageSend{
		File: &discordgo.File{
			Name:        fmt.Sprintf("%s%s", subreddit, ext),
			ContentType: resp.Header.Get("Content-Type"),
			Reader:      bytes.NewReader(body),
		},
	}, err
}

func (handle *Handle) getListing(subreddit string) ([]*geddit.Submission, error) {
	return handle.session.SubredditSubmissions(subreddit, geddit.HotSubmissions, geddit.ListingOptions{Limit: 100})
}
