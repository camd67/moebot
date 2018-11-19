package commands

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"

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

type veteranBuffer struct {
	sync.RWMutex
	buffCooldown int
	m            map[string]int
}

type VeteranHandler struct {
	reactionCooldownMap util.SyncCooldownMap
	messageCooldownMap  util.SyncCooldownMap
	vBuffer             veteranBuffer
	comPrefix           string
	debugChannel        string
	masterId            string
}

func NewVeteranHandler(comPrefix string, debugChannel string, masterId string) *VeteranHandler {
	result := &VeteranHandler{}
	result.reactionCooldownMap = util.SyncCooldownMap{
		M: make(map[string]int64),
	}
	result.messageCooldownMap = util.SyncCooldownMap{
		M: make(map[string]int64),
	}
	result.vBuffer = veteranBuffer{
		m:            make(map[string]int),
		buffCooldown: veteranBufferSizeMax,
	}
	result.comPrefix = comPrefix
	result.debugChannel = debugChannel
	result.masterId = masterId
	return result
}

func (vh *VeteranHandler) EventHandlers() []interface{} {
	return []interface{}{vh.veteranMessageCreate, vh.veteranReactionAdd}
}

func (vh *VeteranHandler) veteranMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// todo: Another place where we need to update to prevent network failure due to hokago-tea-time. Perhaps we could group these together somehow?
	// 1 entry point for all handlers perhaps?
	channel, err := moeDiscord.GetChannel(message.ChannelID, session)
	if err != nil {
		// missing channel
		log.Println("ERROR! Unable to get guild in messageCreate ", err, channel)
		return
	}

	server, err := db.ServerQueryOrInsert(channel.GuildID)
	if err != nil {
		return
	}
	if !server.VeteranRank.Valid || !server.VeteranRole.Valid {
		// they didn't set up any veteran config settings, bail out
		return
	}

	// ignore some common bot prefixes
	if !(strings.HasPrefix(message.Content, "->") || strings.HasPrefix(message.Content, "~") || strings.HasPrefix(message.Content, vh.comPrefix)) {
		changedUsers, err := vh.handleVeteranMessage(message.Author.ID, channel.GuildID)
		if err != nil {
			session.ChannelMessageSend(vh.debugChannel, fmt.Sprint("An error occurred when trying to update veteran users ", err))
		} else {
			for _, user := range changedUsers {
				// ignore the master from any rank related stuff. Could ignore them earlier, but this is the main "public" facing point
				if user.UserUid != vh.masterId {
					session.ChannelMessageSend(user.SendTo, "Congrats "+util.UserIdToMention(user.UserUid)+" you can become a server veteran! Type `"+
						vh.comPrefix+" role veteran` In this channel.")
				}
			}
		}
	}
	// need to clear the server buffer here, since we don't have full clear functionality yet
	db.FlushServerCache()
}

func (vh *VeteranHandler) handleVeteranMessage(userUid string, guildUid string) (users []types.UserServerRankWrapper, err error) {
	key := buildVeteranBufferKey(userUid, guildUid)
	if isCooldownReached(key, messageCooldown, &vh.messageCooldownMap) {
		return vh.handleVeteranChange(userUid, guildUid, messagePoints)
	}
	return
}

func (vh *VeteranHandler) veteranReactionAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	// should make some local caches for channels and guilds...
	channel, err := moeDiscord.GetChannel(reactionAdd.ChannelID, session)
	if err != nil {
		// already logged, and no channel to send to.
		return
	}

	server, err := db.ServerQueryOrInsert(channel.GuildID)
	if err != nil {
		return
	}
	if !server.VeteranRank.Valid || !server.VeteranRole.Valid {
		// they didn't set up any veteran config settings, bail out
		return
	}

	changedUsers, err := vh.handleVeteranReaction(reactionAdd.UserID, channel.GuildID)
	if err != nil {
		session.ChannelMessageSend(vh.debugChannel, fmt.Sprint("An error occurred when trying to update veteran users ", err))
	} else {
		for _, user := range changedUsers {
			if user.UserUid != vh.masterId {
				session.ChannelMessageSend(user.SendTo, "Congrats "+util.UserIdToMention(user.UserUid)+" you can become a server veteran! Type `"+
					vh.comPrefix+" role veteran` In this channel.")
			}
		}
	}
	db.FlushServerCache()
}

func (vh *VeteranHandler) handleVeteranReaction(userUid string, guildUid string) (users []types.UserServerRankWrapper, err error) {
	key := buildVeteranBufferKey(userUid, guildUid)
	if isCooldownReached(key, reactionCooldown, &vh.reactionCooldownMap) {
		return vh.handleVeteranChange(userUid, guildUid, reactionPoints)
	}
	return
}

func (vh *VeteranHandler) handleVeteranChange(userUid string, guildUid string, points int) (users []types.UserServerRankWrapper, err error) {
	vh.vBuffer.Lock()
	vh.vBuffer.m[buildVeteranBufferKey(userUid, guildUid)] += points
	vh.vBuffer.buffCooldown--
	buffCount := vh.vBuffer.buffCooldown
	vh.vBuffer.Unlock()

	// only actually go through and process the veterans that have been buffered if we pass our max
	if buffCount < 0 {
		var idsToUpdate []int
		// we've got to read and write this one unfortunately
		vh.vBuffer.Lock()
		defer vh.vBuffer.Unlock()
		for key, count := range vh.vBuffer.m {
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
			id, newPoint, messageSent, err := db.UserServerRankUpdateOrInsert(user.Id, server.ID, count)
			if err != nil {
				// we had an error, just don't delete the user and their points
				continue
			}
			if !messageSent && server.VeteranRank.Valid && server.BotChannel.Valid && newPoint >= server.VeteranRank.Int {
				// we haven't had an error so the user was updated
				users = append(users, types.UserServerRankWrapper{
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
		vh.vBuffer.m = make(map[string]int)
		vh.vBuffer.buffCooldown = veteranBufferSizeMax
	}
	db.FlushServerCache()
	return users, nil
}

/*
Returns true if the given key in the syncCooldownMap has passed the given cooldown duration, false otherwise
*/
func isCooldownReached(key string, cooldown time.Duration, cooldownMap *util.SyncCooldownMap) bool {
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
