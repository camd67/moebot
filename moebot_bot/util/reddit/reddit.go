package reddit

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jzelinskie/geddit"
)

const (
	tokenTimeLimit    = time.Minute * 59 // Tokens last an hour, refresh them every almost-hour
	imageLimit        = 100              // 100 is max allowed by reddit listing apis
	randImgRetryLimit = 5                // Maximum tries to find random image from listing that will work
)

var (
	whitelistedContentTypes = map[string]string{"image/png": ".png", "image/jpeg": ".jpg"}
)

type tokenTimer struct {
	sync.Mutex
	startTime time.Time
}

type Handle struct {
	session *geddit.OAuthSession
	tTimer  tokenTimer

	clientID     string
	clientSecret string
	username     string
	password     string
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
		return &Handle{}, err
	}

	err = session.LoginAuth(username, password)
	if err != nil {
		log.Println("Error logging into reddit", err)
		return &Handle{}, err
	}

	return &Handle{session: session, tTimer: tokenTimer{startTime: time.Now()}, clientID: clientID, clientSecret: clientSecret, username: username, password: password}, err
}

func (handle *Handle) GetRandomImage(subreddit string) (*discordgo.MessageSend, error) {
	if handle.session == nil {
		return nil, errors.New("Handle's session was not setup")
	}

	err := handle.renewTokenIfNecessary()
	if err != nil {
		return nil, err
	}

	posts, err := handle.getListing(subreddit)
	if err != nil {
		log.Println("Error getting listing from subreddit %s", subreddit)
		return nil, err
	}

	var contentType string
	var body []byte
	var fileExtension string
	tryCount := 0

	// Keep looking until you find an acceptable image
	for {
		i := rand.Intn(len(posts) - 1)
		randPost := posts[i]

		resp, err := http.Get(randPost.URL)
		if err != nil {
			log.Println("Error requesting image: " + randPost.URL)
			return nil, err
		}
		contentType = resp.Header.Get("Content-Type")

		if whitelistedExt, ok := whitelistedContentTypes[contentType]; ok {
			// We found a valid content type, get it's body and file type extension and bail out
			fileExtension = whitelistedExt

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error preparing repsonse body")
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			break
		}

		// Didn't find a valid post type. Close our request and try again
		resp.Body.Close()

		tryCount++
		if tryCount > randImgRetryLimit {
			log.Printf("Couldn't find a usable image in subreddit " + subreddit)
			return nil, errors.New("Couldn't find a usable image in subreddit " + subreddit)
		}
		removeBadSubmission(posts, i)
	}

	log.Printf("Sending image %s with content type %s", subreddit+fileExtension, contentType)
	return &discordgo.MessageSend{
		File: &discordgo.File{
			Name:        subreddit + fileExtension,
			ContentType: contentType,
			Reader:      bytes.NewReader(body),
		},
	}, err
}

func (handle *Handle) getListing(subreddit string) ([]*geddit.Submission, error) {
	return handle.session.SubredditSubmissions(subreddit, geddit.HotSubmissions, geddit.ListingOptions{Limit: imageLimit})
}

func (handle *Handle) renewTokenIfNecessary() error {
	handle.tTimer.Lock()
	defer handle.tTimer.Unlock()
	if handle.tTimer.startTime.Add(tokenTimeLimit).Before(time.Now()) {
		log.Println("Reddit token expired, getting new token")
		err := handle.session.LoginAuth(handle.username, handle.password)
		if err != nil {
			return errors.New("Couldn't renew token")
		}
		handle.tTimer.startTime = time.Now()
	}
	return nil
}

func removeBadSubmission(s []*geddit.Submission, i int) []*geddit.Submission {
	s[i] = s[0]
	s[0] = nil
	return s[1:]
}
