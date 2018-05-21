# moebot
The bot for discord, but with moe!

## Package rules per sub-project
### api
This project/package is responsible for all API requests. Mostly for the website. Not currently in use.
### moebot_bot
Everything for the discord part of moebot
#### bot
This package is for any commands and processing related directly to discord itself.
Functions in this package should directly do things related to discord such as sending messages,
receiving messages, and changing discord related information
#### util
This package should be for general functions that are independent of discord (but may relate to discord, such as converting to a discord mention or searching for a role)
No state should be persisted in the `util` package. When considering adding a function to `util` consider if the function would
be used in more than one command. If it's tied to one specific feature, it doesn't belong in `util`.
#### db
This package is for any database tables and database connections/commands. Each database table should have it's own file, with querying/insert/deletion commands.

## Setup (discord bot)
* Install docker
* Duplicate `config/mb_config.example.txt` and rename it to `mb_config.secret`
    * Fill in `secret` with your discord bot's secret
    * Choose your bot's prefix (This is what you use to trigger bot commands)
    * `dbPass` is the root database password and `moeDbPass` is what moebot will login with
    * `masterId` is the discord User ID associated with the bot's master. __NOTE:__ this user can perform any command on any server that this bot is a part of!
    * `debugChannel` is the channel to send __all__ moebot related error messages to.
    * `loadPings` determines if pins are loaded. 0 = don't load pins, 1 = load pins
    * To use commands that make use of reddit commands, you must have a registered script app here: https://www.reddit.com/prefs/apps
    * `redditClientID` is the client ID of your app
    * `redditClientSecret` is the secret of your app
    * `redditUserName` is the login username for a reddit account
    * `redditPassword` is the login password for a reddit account
* Duplicate `config/pg_pass.example.txt` and rename it to `pg_pass.secret`
    * Replace the whole file with your default postgres password. This must match `dbPass` in `mb_config.secret`
    * __Note:__ this file should have exactly 1 line! Any trailing newlines will break the login process
* Create a docker volume: `docker volume create moebot-data`
* Run `docker-compose up --build -d` to run moebot in the background
* Invite moebot to your server!

## Setup (website)
* Follow setup for discord bot
* Edit your hosts file to include `127.0.0.1 local.moebot.moe`
* Go into your browser and go to `local.moebot.moe`
