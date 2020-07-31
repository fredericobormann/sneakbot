# sneakbot
This is a very small Telegram bot to randomly choose people that decide which film to watch during a weekly movie night.
It uses a sqlite database to store the people who agreed to help.

## Installation
1. Copy the `config.yml.sample` and rename it to `config.yml`.
1. Enter your Telegram bot API token and the URL where your bot is listening for updates. (By default sneakbot listens on port 8443 on your local machine.)
1. Run the `main.go`

## Known problems
* Automatic draw does not work at the moment due to a bug in the gocron package.
* Telegram does not allow to fetch usernames of users that did not interact with the bot recently, so storing usernames in the database will be necessary.
