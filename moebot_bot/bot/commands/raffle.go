package commands

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RaffleCommand struct {
	MasterId     string
	DebugChannel string
}

const ticketCooldown = int64(time.Hour * 24)

func (rc *RaffleCommand) Execute(pack *CommPackage) {
	// Previous servers
	if pack.guild.ID == "378336255030722570" {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, the raffle has ended!")
		return
	}
	// Salt
	if pack.guild.ID != "93799773856862208" {
		pack.session.ChannelMessageSend(pack.channel.ID, "Raffles are not enabled in this server! Speak to Salt to get your server added to the raffle!")
		return
	}

	if len(pack.params) > 0 && pack.message.Author.ID == rc.MasterId {
		// special master only commands
		if pack.params[0] == "vote" {
			// delete original message
			pack.session.ChannelMessageDelete(pack.channel.ID, pack.message.ID)
			// post all the raffle entries
			allRaffles, err := db.RaffleEntryQueryAny(pack.guild.ID)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, an error occured when fetching raffles!")
				return
			}
			const sleepTime = time.Second

			for _, r := range allRaffles {
				// make sure to pause between message sends so we don't get discord throttled
				submissions := strings.Split(r.RaffleData, db.RaffleDataSeparator)
				if submissions[0] != "NONE" {
					sent, _ := pack.session.ChannelMessageSend("392528129412956170", "-----------------------\n"+
						util.UserIdToMention(r.UserUid)+"'s submitted art: "+submissions[0])
					if sent != nil {
						pack.session.MessageReactionAdd(sent.ChannelID, sent.ID, "👍")
					}
					time.Sleep(sleepTime)
				}
				if submissions[1] != "NONE" {
					sent, _ := pack.session.ChannelMessageSend("392528172245319680", "-----------------------\n"+
						util.UserIdToMention(r.UserUid)+"'s submitted relic: "+submissions[1])
					if sent != nil {
						pack.session.MessageReactionAdd(sent.ChannelID, sent.ID, "👍")
					}
					time.Sleep(sleepTime)
				}
			}
		} else if pack.params[0] == "count" {
			// count up any reactions to the images and award bonus tickets
			pack.session.ChannelMessageDelete(pack.message.ChannelID, pack.message.ID)
			messages, err := pack.session.ChannelMessages(pack.message.ChannelID, 100, pack.message.ID, "", "")
			if err != nil {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue fetching historical messages")
				return
			}
			// loop over every message and count up reactions per ID
			userReactCounts := make(map[string]int)
			userSubmissionVotes := make(map[string]int)
			const waitTime = time.Millisecond * 50
			for _, m := range messages {
				// only process bot messages (presumably by moebot) since that was how submissions were sent in
				if m.Author.Bot && strings.HasPrefix(m.Content, "-----------------------") {
					userReacts, err := pack.session.MessageReactions(pack.message.ChannelID, m.ID, "👍", 100)
					if len(m.Mentions) != 1 {
						pack.session.ChannelMessageSend(rc.DebugChannel, "Error processing raffle submission count: "+fmt.Sprintf("%+v", m))
						continue
					}
					if err != nil {
						pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, unable to get reactions for one of the messages!")
						return
					}
					// add up all the reactions
					userSubmissionVotes[m.Mentions[0].ID] = len(userReacts) - 1
					for _, ur := range userReacts {
						if !ur.Bot {
							userReactCounts[ur.ID] = userReactCounts[ur.ID] + 1
						}
					}
					// just pause for a bit so we don't hit the ratelimit
					time.Sleep(waitTime)
				}
			}
			// go through each of the user react counts, and give a bonus ticket for everyone who got >3 votes
			const minVotes = 3
			raffles, err := db.RaffleEntryQueryAny(pack.guild.ID)
			if err != nil {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue fetching raffle entries for this server")
				return
			}
			rafflesToUpdate := make([]db.RaffleEntry, 0)
			for user, count := range userReactCounts {
				for _, r := range raffles {
					// only those with valid raffle entries and the min number of votes get updated
					if r.UserUid == user && count >= minVotes {
						rafflesToUpdate = append(rafflesToUpdate, r)
						break
					}
				}
			}
			if len(rafflesToUpdate) > 0 {
				db.RaffleEntryUpdateMany(rafflesToUpdate, 1)
			}
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Top 3 submissions:")
			// find the top 3 votes (probably a better way than this, but it works...)
			for i := 1; i <= 3; i++ {
				maxVoteKey := ""
				for user, count := range userSubmissionVotes {
					if maxVoteKey == "" {
						// always grab the first one
						maxVoteKey = user
					} else if count > userSubmissionVotes[maxVoteKey] {
						// found a larger value, use it
						maxVoteKey = user
					}
				}
				// grab that user, reset their votes, and say they won
				pack.session.ChannelMessageSend(pack.message.ChannelID, strconv.Itoa(i)+": "+util.UserIdToMention(maxVoteKey)+" with "+
					strconv.Itoa(userSubmissionVotes[maxVoteKey])+" votes!")
				userSubmissionVotes[maxVoteKey] = 0
			}
		} else if pack.params[0] == "winner" {
			raffles, err := db.RaffleEntryQueryAny(pack.guild.ID)
			if err != nil {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue fetching raffle entries")
				return
			}
			// go through and add users based on how many tickets they got (if a user had 5 tickets they'd have 5 entries in the array
			users := make([]string, 0)
			for _, r := range raffles {
				for i := 0; i < r.TicketCount; i++ {
					users = append(users, r.UserUid)
				}
			}
			// now that we have a list of users and their ticket values ["123", "123", "123", "456", 456", "789" ...] figure out who won
			selected := rand.Int() % len(users)
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Congrats "+util.UserIdToMention(users[selected])+" you've won the raffle!")
		}
	} else {
		const startTickets = 5
		raffleEntries, err := db.RaffleEntryQuery(pack.message.Author.ID, pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information!")
			return
		}
		// there should only be 1 of each raffle entry for every user + guild combo
		if len(raffleEntries) > 1 {
			log.Println("Queried for more than one raffle entry: userUid-", raffleEntries[0].UserUid)
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information!")
			return
		}
		if len(raffleEntries) == 0 {
			// haven't joined the raffle yet, make an entry + add tickets
			newRaffle := db.RaffleEntry{
				GuildUid:    pack.guild.ID,
				UserUid:     pack.message.Author.ID,
				RaffleType:  db.RaffleMIA,
				TicketCount: startTickets,
				RaffleData:  "NONE" + db.RaffleDataSeparator + "NONE",
			}
			err := db.RaffleEntryAdd(newRaffle)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue adding your raffle entry!")
				return
			}
			pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+", welcome to the raffle! You get "+
				strconv.Itoa(startTickets)+" tickets for joining!")
		} else {
			// already joined the raffle, let them know their ticket count and other information
			raffleData := strings.Split(raffleEntries[0].RaffleData, db.RaffleDataSeparator)
			timeLeft := time.Duration(raffleEntries[0].LastTicketUpdate + ticketCooldown)
			messageTime, _ := pack.message.Timestamp.Parse()
			// get the difference between the time left and the message time
			timeLeft = timeLeft - time.Duration(messageTime.UnixNano())
			if timeLeft > 0 {
				pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+" you're already in the raffle! Your ticket count is: "+
					strconv.Itoa(raffleEntries[0].TicketCount)+". Your art submission is: `"+raffleData[0]+"`. Your relic submission is: `"+raffleData[1]+"`."+
					" Time until new ticket drop: "+timeLeft.String())
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+" you're already in the raffle! Your ticket count is: "+
					strconv.Itoa(raffleEntries[0].TicketCount)+". Your art submission is: `"+raffleData[0]+"`. Your relic submission is: `"+raffleData[1]+"`."+
					" A ticket could drop at any time now!")
			}
		}
	}
}

func (rc *RaffleCommand) Setup(session *discordgo.Session) {}

func (rc *RaffleCommand) EventHandlers() []interface{} {
	// distribute tickets
	// temporarily disable ticket distribution
	return []interface{}{ /*rc.distributeTickets*/ }
}

func (rc *RaffleCommand) distributeTickets(guild *discordgo.Guild, message *discordgo.MessageCreate, session *discordgo.Session, messageTime time.Time) {
	if false {
		const maxChance = 100
		const ticketChance = 5
		if rand.Int()%maxChance <= ticketChance {
			raffles, err := db.RaffleEntryQuery(message.Author.ID, guild.ID)
			if err != nil {
				session.ChannelMessageSend(rc.DebugChannel, "Error loading raffle information during ticket distribution"+fmt.Sprintf("%+v | %+v", guild, message))
				return
			}

			if len(raffles) != 1 {
				// if they're not in the raffle, bail out
				return
			}
			r := raffles[0]
			// check to see if their last ticket time is more than the cooldown
			if r.LastTicketUpdate+ticketCooldown > messageTime.UnixNano() {
				// they haven't waited enough, bail
				return
			}
			// they've won a ticket and passed the timestamp check, let them know and update db
			r.LastTicketUpdate = messageTime.UnixNano()
			db.RaffleEntryUpdate(r, 1)
			currTickets := r.TicketCount + 1
			session.ChannelMessageSend("378680855339728918", message.Author.Mention()+", congrats! You just earned another ticket! Your current tickets are: "+strconv.Itoa(currTickets))
		}
	}
}

func (rc *RaffleCommand) GetPermLevel() db.Permission {
	return db.PermNone
}
