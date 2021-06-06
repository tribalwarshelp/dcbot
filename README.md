# tribalwarshelp.com DC Bot

A Discord bot for the online game Tribal Wars.

Features:
1. Coords translation

   ![Screenshot](/screenshots/coordstranslation.png?raw=true)

2. Live conquer notifications

   ![Screenshot](/screenshots/notifications2.png?raw=true)

3. List of tribe members ordered by OD/ODA/ODD/ODS/points
4. Translated into 4 languages: Polish, Czech, English, Dutch

[You can check all available commands here.](https://dcbot.tribalwarshelp.com/commands/)

## Development



### Prerequisites

1. Golang
2. PostgreSQL database
3. Configured [API](https://github.com/tribalwarshelp/api)

### Installation
**Required ENV variables:**

```
DB_USER=your_pgdb_user
DB_NAME=your_pgdb_name
DB_PORT=your_pgdb_port
DB_HOST=your_pgdb_host
DB_PASSWORD=your_pgdb_password
API_URL=your_api_url
BOT_TOKEN=your_bot_token
LOG_DB_QUERIES=true
```

1. Clone this repo.
```
git clone git@github.com:tribalwarshelp/dcbot.git
```
2. Navigate to the directory where you have cloned this repo.
3. Set the required env variables directly in your system or create .env.local file.
4. Run the app.
```
go run main.go
```

## License

Distributed under the MIT License. See ``LICENSE`` for more information.

## Contact

Dawid Wysokiński - [contact@dwysokinski.me](mailto:contact@dwysokinski.me)
