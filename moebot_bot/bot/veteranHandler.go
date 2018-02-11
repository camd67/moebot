package bot

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/camd67/moebot/moebot_bot/util"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	messagePoints        = 5
	reactionPoints       = 1
	reactionCooldown     = 45 * time.Second
	messageCooldown      = 30 * time.Second
	veteranBufferSizeMax = 30
)

var (
	// each of our reactions/messages are done in goroutines so we need to sync these
	reactionCooldownMap = util.SyncCooldownMap{
		sync.RWMutex{},
		make(map[string]int64),
	}

	messageCooldownMap = util.SyncCooldownMap{
		sync.RWMutex{},
		make(map[string]int64),
	}

	veteranBuffer = struct {
		sync.RWMutex
		buffCooldown int
		m            map[string]int
	}{m: make(map[string]int), buffCooldown: veteranBufferSizeMax}
)

func handleVeteranMessage(m *discordgo.Member, guildUid string) (users []db.UserServerRankWrapper, err error) {
	key := buildVeteranBufferKey(m.User.ID, guildUid)
	if isCooldownReached(key, messageCooldown, messageCooldownMap) {
		return handleVeteranChange(m.User.ID, guildUid, messagePoints)
	}
	return
}

func handleVeteranReaction(userUid string, guildUid string) (users []db.UserServerRankWrapper, err error) {
	key := buildVeteranBufferKey(userUid, guildUid)
	if isCooldownReached(key, reactionCooldown, reactionCooldownMap) {
		return handleVeteranChange(userUid, guildUid, messagePoints)
	}
	return handleVeteranChange(userUid, guildUid, reactionPoints)
}

func handleVeteranChange(userUid string, guildUid string, points int) (users []db.UserServerRankWrapper, err error) {
	veteranBuffer.Lock()
	veteranBuffer.m[buildVeteranBufferKey(userUid, guildUid)] += points
	veteranBuffer.buffCooldown--
	buffCount := veteranBuffer.buffCooldown
	veteranBuffer.Unlock()

	// only actually go through and process the veterans that have been buffered if we pass our max
	if buffCount < 0 {
		var idsToUpdate []int
		// we've got to read and write this one unfortunately
		veteranBuffer.Lock()
		defer veteranBuffer.Unlock()
		for key, count := range veteranBuffer.m {
			uid, gid := splitVeteranBufferKey(key)
			server, err := db.ServerQueryOrInsert(gid)
			if err != nil {
				log.Println("Error getting server during veteran change", err)
				return nil, err
			}
			user, err := db.UserQueryOrInsert(uid)
			if err != nil {
				log.Println("Error getting user during veteran change", err)
				return nil, err
			}
			id, newPoint, messageSent, err := db.UserServerRankUpdateOrInsert(user.Id, server.Id, count)
			if err != nil {
				// we had an error, just don't delete the user and their points
				continue
			}
			if !messageSent && server.VeteranRank.Valid && server.BotChannel.Valid && int64(newPoint) >= server.VeteranRank.Int64 {
				// we haven't had an error so the user was updated
				users = append(users, db.UserServerRankWrapper{
					UserUid:   uid,
					ServerUid: gid,
					Rank:      newPoint,
					SendTo:    server.BotChannel.String,
				})
				idsToUpdate = append(idsToUpdate, id)
			}
		}
		if len(idsToUpdate) > 0 {
			db.UserServerRankSetMessageSent(idsToUpdate)
		}
		// clear the whole map
		veteranBuffer.m = make(map[string]int)
		veteranBuffer.buffCooldown = veteranBufferSizeMax
	}
	db.FlushServerCache()
	return users, nil
}

/**
Returns true if the given key in the syncCooldownMap has passed the given cooldown duration, false otherwise
*/
func isCooldownReached(key string, cooldown time.Duration, cooldownMap util.SyncCooldownMap) bool {
	cooldownMap.RWMutex.RLock()
	lastTime, present := cooldownMap.M[key]
	cooldownMap.RUnlock()
	if present {
		if lastTime+cooldown.Nanoseconds() > time.Now().UnixNano() {
			return false
		}
	}
	cooldownMap.Lock()
	cooldownMap.M[key] = time.Now().UnixNano()
	cooldownMap.Unlock()
	return true
}

func buildVeteranBufferKey(u string, g string) string {
	return u + ":" + g
}

func splitVeteranBufferKey(key string) (u string, g string) {
	split := strings.Split(key, ":")
	return split[0], split[1]
}
