package commands

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/reddit"
)

var (
	whitelistedSubreddits = map[string]string{
		"random": "awwnime",
		"meme":   "animemes",
		"irl":    "anime_irl",
	}
	defaultSubreddit = "awwnime"
)

type SubCommand struct {
	RedditHandle *reddit.Handle
}

func (sc *SubCommand) Execute(pack *CommPackage) {
	subreddit, err := getSubredditFromParams(pack.params)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... I can't manage that `type`. Sorry! Here's something you might like instead!")
		log.Printf("Error getting subreddit from params: %v", pack.params)
	}

	send, err := sc.RedditHandle.GetRandomImage(subreddit)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... Looks like this command isn't working right now. Sorry!")
		log.Println("Error getting image from reddit")
		return
	}

	pack.session.ChannelMessageSendComplex(pack.channel.ID, send)
}

func getSubredditFromParams(params []string) (string, error) {
	if len(params) == 0 { // choose one at random
		return getRandomWhitelistedSubreddit(), nil
	} else if v, ok := whitelistedSubreddits[params[0]]; ok {
		return v, nil
	}
	return defaultSubreddit, fmt.Errorf("couldn't get subreddit from params, sending %s as default", defaultSubreddit)
}

func getRandomWhitelistedSubreddit() string {
	r := rand.Intn(len(whitelistedSubreddits) - 1)
	i := 0
	for _, v := range whitelistedSubreddits {
		if i == r {
			return v
		}
		i++
	}
	return defaultSubreddit // should only be reached in circumstance where the whitelistedSubreddits map is empty
}

func (sc *SubCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
func (sc *SubCommand) GetCommandKeys() []string {
	return []string{"SUB"}
}
func (sc *SubCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s sub [type]` - Posts a random image. `type` is optionally one of: `random`, `irl`, `meme`", commPrefix)
}
