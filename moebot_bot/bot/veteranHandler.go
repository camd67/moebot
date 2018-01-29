package bot

import (
	"log"
	"sync"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	messagePoints        = 5
	reactionPoints       = 1
	reactionCooldown     = time.Second
	veteranBufferSizeMax = 3
)

var (
	// each of our reactions/messages are done in goroutines so we need to sync these
	reactionCooldownMap = struct {
		sync.RWMutex
		m map[string]int64
	}{m: make(map[string]int64)}

	veteranBuffer = struct {
		sync.RWMutex
		buffCooldown int
		m            map[string]int
	}{m: make(map[string]int), buffCooldown: veteranBufferSizeMax}
)

func handleVeteranMessage(m *discordgo.Member, guildUid string) (users []db.UserServerRankWrapper, err error) {
	return handleVeteranChange(m.User.ID, guildUid, messagePoints)
}

func handleVeteranReaction(userUid string, guildUid string) (users []db.UserServerRankWrapper, err error) {
	reactionCooldownMap.RWMutex.RLock()
	lastTime, present := reactionCooldownMap.m[userUid]
	reactionCooldownMap.RUnlock()
	if present {
		if lastTime+reactionCooldown.Nanoseconds() > time.Now().UnixNano() {
			return
		}
	}
	// if they don't have a time yet, or the cooldown was passed, give them a new time
	reactionCooldownMap.Lock()
	reactionCooldownMap.m[userUid] = time.Now().UnixNano()
	reactionCooldownMap.Unlock()
	return handleVeteranChange(userUid, guildUid, reactionPoints)
}

func handleVeteranChange(userUid string, guildUid string, points int) (users []db.UserServerRankWrapper, err error) {
	veteranBuffer.Lock()
	veteranBuffer.m[userUid] += points
	veteranBuffer.buffCooldown--
	buffCount := veteranBuffer.buffCooldown
	veteranBuffer.Unlock()

	// only actually go through and process the veterans that have been buffered if we pass our max
	if buffCount < 0 {
		// TODO: This only works when the entire map contains only 1 guild. Should store guild information in the map
		server, err := db.ServerQueryOrInsert(guildUid)
		if err != nil {
			log.Println("Error getting server during veteran change", err)
			return nil, err
		}

		// we've got to read and write this one unfortunately
		veteranBuffer.Lock()
		defer veteranBuffer.Unlock()
		for uid, count := range veteranBuffer.m {
			user, err := db.UserQueryOrInsert(uid)
			log.Println("PROCESSING {", user.Id, "}", user.UserUid)
			if err != nil {
				log.Println("Error getting user during veteran change", err)
				return nil, err
			}
			newPoint, err := db.UserServerRankUpdateOrInsert(user.Id, server.Id, count)
			if err != nil {
				// we had an error, just don't delete the user and their points
				continue
			}
			// we haven't had an error so the user was updated
			users = append(users, db.UserServerRankWrapper{
				UserUid:   uid,
				ServerUid: guildUid,
				Rank:      newPoint,
			})
		}
		// clear the whole map
		veteranBuffer.m = make(map[string]int)
		veteranBuffer.buffCooldown = veteranBufferSizeMax
	}
	return users, nil
}
