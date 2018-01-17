# moebot
The bot for discord, but with moe!

## Setup (discord bot)
* Install docker
* Duplicate `config/mb_config.example.txt` and rename it to `mb_config.secret`
    * Fill in `secret` with your discord bot's secret
    * Choose your bot's prefix (This is what you use to trigger bot commands)
    * `dbPass` is the root database password and `moeDbPass` is what moebot will login with
    * `masterId` is the discord User ID associated with the bot's master. __NOTE:__ this user can perform any command on any server that this bot is a part of!
    * `debugChannel` is the channel to send __all__ moebot related error messages to.
* Duplicate `config/pg_pass.example.txt` and rename it to `pg_pass.secret`
    * Replace the whole file with your default postrges password. This must match `dbPass` in `mb_config.secret`
    * __Note:__ this file should have exactly 1 line! Any trailing newlines will break the login process 
* Create a docker volume: `docker volume create moebot-data`
* Run `docker-compose up --build -d` to run moebot in the background
* Invite moebot to your server!

## Setup (website)
* Follow setup for discord bot
* Edit your hosts file to include `127.0.0.1 local.moebot.moe`
* Go into your browser and go to `local.moebot.moe`