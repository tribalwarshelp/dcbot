# TWHelp DC Bot

Discord bot for the online game Tribal Wars.

Features:

1. Coords translation
   ![Screenshot](/screenshots/coordstranslation.png?raw=true)
2. Near real-time notifications about conquers
   ![Screenshot](/screenshots/notifications.png?raw=true)
3. Tribe members ordered by OD/ODA/ODD/ODS/points

[You can check all available commands here.](https://dcbot.tribalwarshelp.com/commands/)

## Development

**Required env variables to run this bot** (you can set them directly in your system or create .env.development file):

```
DB_USER=your_pgdb_user
DB_NAME=your_pgdb_name
DB_PORT=your_pgdb_port
DB_HOST=your_pgdb_host
DB_PASSWORD=your_pgdb_password
API_URL=your_api_url
BOT_TOKEN=your_bot_token
```

### Prerequisites

1. Golang
2. PostgreSQL database
3. Configured [API](https://github.com/tribalwarshelp/api)

### Installing

1. Clone this repo.
2. Navigate to the directory where you have cloned this repo.
3. Set the required env variables directly in your system or create .env.development file.
4. go run main.go
