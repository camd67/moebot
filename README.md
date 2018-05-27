# moebot [![CodeFactor](https://www.codefactor.io/repository/github/camd67/moebot)](https://www.codefactor.io/repository/github/camd67/moebot)
The bot for discord, but with moe!

## Setup (discord bot)
* Install docker
* Duplicate `config/mb_config.example.txt` and rename it to `mb_config.secret`
    * Remove all the example text and replace with your information in the new file
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
