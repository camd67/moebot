package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/camd67/moebot/moebot_bot/bot"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// read in configuration information
	configPath := os.Getenv("MOEBOT_CONFIG_PATH")
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("Error reading config from file. Path: " + configPath)
	}

	configText := util.NormalizeNewlines(string(configFile))
	for _, line := range strings.Split(configText, "\n") {
		splitLine := strings.Split(line, "~")
		if len(splitLine) == 0 || splitLine[0] == "" {
			continue
		}
		bot.Config[splitLine[0]] = splitLine[1]
	}
	bot.ComPrefix = bot.Config["prefix"]
	// setup discord with that information
	discord, err := discordgo.New("Bot " + bot.Config["secret"])
	if err != nil {
		log.Fatal("Error starting discord...", err)
	}

	bot.SetupMoebot(discord)

	// start up a connection with discord
	err = discord.Open()
	if err != nil {
		log.Fatal("Error starting discord: ", err)
	}

	// halt until we get a SIGTERM or similar
	log.Println("Moebot's up and running! Press CTRL + C to exit...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Clean up any connections to outside resources
	discord.Close()
	db.DisconnectAll()

	fmt.Println("Exited moebot! Seeya later!")
}
