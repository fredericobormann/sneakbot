# sneakbot
This is a very small Telegram bot to randomly choose people that decide which film to watch during a weekly movie night.
It uses a sqlite database to store the people who agreed to help.
Since a truly random choosing method could take a lot of time to produce a fair result, the bot keeps track of how often a person was drawn and prefers people that have been chosen rarely.

## Installation
1. Copy the `config.yml.sample` and rename it to `config.yml`.
1. Enter your Telegram bot API token and the URL where your bot is listening for updates. (By default sneakbot listens on port 8443 on your local machine.)
1. Run the `main.go`

## Known problems
* Automatic draw does not work at the moment due to a bug in the gocron package.
