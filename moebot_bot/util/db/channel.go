package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

func ChannelQueryOrInsert(channelUid string, server *models.Server) (c *models.Channel, e error) {
	c, err := models.Channels(qm.Where("channel_uid = ?", channelUid)).One(context.Background(), moeDb)
	if err == sql.ErrNoRows {
		c = &models.Channel{
			ChannelUID: channelUid,
			ServerID:   server.ID,
		}
		c.Insert(context.Background(), moeDb, boil.Whitelist("channel_uid", "server_id"))
		if e != nil {
			log.Println("Error inserting channel to db ", e)
			return nil, e
		}
		return c, nil
	}
	return c, nil
}

func ChannelUpdate(channel *models.Channel) (err error) {
	channel.Update(context.Background(), moeDb, boil.Whitelist("bot_allowed", "move_pins", "move_text_pins", "delete_pin", "move_channel_uid"))
	if err != nil {
		log.Println("Error update channel table", err)
		return
	}
	return
}

func ChannelQueryByServer(server *models.Server) (channels models.ChannelSlice, err error) {
	channels, err = models.Channels(qm.Where("server_id = ?", server.ID)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for channels", err)
	}
	return
}

func ChannelQueryById(channelId int) (c *models.Channel, e error) {
	c, e = models.FindChannel(context.Background(), moeDb, channelId)
	if e != nil {
		log.Println("Error querying channel", e)
		return nil, e
	}
	return c, nil
}
