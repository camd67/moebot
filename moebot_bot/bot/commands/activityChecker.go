package commands

import (
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type activityChecker struct {
	channelUID         string
	roleUID            string
	permissionDefaults map[int]bool
	locked             bool
	messageRead        chan *messageReader
	mutex              *sync.Mutex
	stopReader         chan int
	removeHandlerFunc  func()
	removing           bool
}

type messageReader struct {
	stopReading chan int
}

func newActivityChecker(session *discordgo.Session, channelUID string, roleUID string, basePermissions map[int]bool, activityLimit int) *activityChecker {
	a := &activityChecker{
		channelUID:         channelUID,
		roleUID:            roleUID,
		permissionDefaults: make(map[int]bool),
		locked:             false,
		messageRead:        make(chan *messageReader, activityLimit),
		mutex:              &sync.Mutex{},
	}
	a.permissionDefaults = basePermissions
	a.removeHandlerFunc = session.AddHandler(a.checkActivity)
	return a
}

func (a *activityChecker) checkActivity(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.ChannelID != a.channelUID {
		return
	}
	r := &messageReader{make(chan int)}
	select {
	case a.messageRead <- r:
		go a.setReader(r, session)
	default:
		go func() {
			a.messageRead <- r
			go a.setReader(r, session)
		}() //queue a message write to fill the channel again after we read the first queued reader do stop it
		r = <-a.messageRead
		r.stopReading <- 1
		go a.lockChannel(session)
	}
}

func (a *activityChecker) removeChecker(session *discordgo.Session) {
	a.mutex.Lock()
	a.removing = true
	a.removeHandlerFunc()
	a.terminateAllReaders()
	a.unlockChannelInternal(session)
	a.mutex.Unlock()
}

func (a *activityChecker) terminateAllReaders() {
	for {
		select {
		case r := <-a.messageRead:
			r.stopReading <- 1
		default:
			return
		}
	}
}

func (a *activityChecker) setReader(r *messageReader, session *discordgo.Session) {
	select {
	case <-time.After(1 * time.Minute): //if this succeeds, the channel is below the required message threshold
		<-a.messageRead
		go a.unlockChannel(session)
	case <-r.stopReading: //new message above the set threshold or reader being closed on purpose, removing planned read
	}
}

func (a *activityChecker) lockChannel(session *discordgo.Session) {
	a.mutex.Lock()
	if !a.locked && !a.removing {
		p, err := a.getCurrentPermissions(session)
		if err != nil {
			a.mutex.Unlock()
			return
		}
		allowed := p.Allow &^ discordgo.PermissionAttachFiles &^ discordgo.PermissionEmbedLinks
		denied := p.Deny | discordgo.PermissionAttachFiles | discordgo.PermissionEmbedLinks
		log.Println("Locking channel " + a.channelUID)
		err = session.ChannelPermissionSet(a.channelUID, a.roleUID, "role", allowed, denied)
		if err == nil {
			a.locked = true
		} else {
			log.Println("Error while setting channel permissions:", err)
		}
	}
	a.mutex.Unlock()
}

func (a *activityChecker) unlockChannel(session *discordgo.Session) {
	a.mutex.Lock()
	if a.locked {
		a.unlockChannelInternal(session)
	}
	a.mutex.Unlock()
}

func (a *activityChecker) unlockChannelInternal(session *discordgo.Session) {
	p, err := a.getCurrentPermissions(session)
	if err != nil {
		a.mutex.Unlock()
		return
	}
	allowed := p.Allow
	denied := p.Deny
	if a.permissionDefaults[discordgo.PermissionAttachFiles] {
		allowed = allowed | discordgo.PermissionAttachFiles
		denied = denied &^ discordgo.PermissionAttachFiles
	}
	if a.permissionDefaults[discordgo.PermissionEmbedLinks] {
		allowed = allowed | discordgo.PermissionEmbedLinks
		denied = denied &^ discordgo.PermissionEmbedLinks
	}
	log.Println("Unlocking channel " + a.channelUID)
	err = session.ChannelPermissionSet(a.channelUID, a.roleUID, "role", allowed, denied)
	if err == nil {
		a.locked = false
	} else {
		log.Println("Error while setting channel permissions:", err)
	}
}

func (a *activityChecker) getCurrentPermissions(session *discordgo.Session) (*discordgo.PermissionOverwrite, error) {
	channel, err := session.Channel(a.channelUID)
	if err != nil {
		return nil, err
	}
	if p, ok := moeDiscord.FindPermissionByRoleID(channel.PermissionOverwrites, a.roleUID); !ok {
		return &discordgo.PermissionOverwrite{
			ID:   a.roleUID,
			Type: "role",
		}, nil
	} else {
		return p, nil
	}
}
