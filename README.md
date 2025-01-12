# ESTU Grade Checker

A Golang Telegram bot designed to monitor and notify users of grade updates from the Eskişehir Technical University (ESTU) student information system. It offers a convenient way for students to stay updated with their academic progress directly through Telegram.

## Features

- Fetches grades from the ESTU student information system
- Notifies users of grade updates via Telegram
- Supports multiple users

## Prerequisites

- Docker and Docker Compose
- A Telegram bot token from [BotFather](https://t.me/botfather)

## Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/enharukalo/estu-grade-checker.git
   cd estu-grade-checker
   ```

2. **Configure environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your Telegram bot token and database password
   ```

3. **Launch the bot**:
   ```bash
   docker compose up -d
   ```

## Bot Commands

- `/start` - Display welcome message and instructions
- `/cookie <cookie>` - Set your ESTU website session cookie
- `/donemid <semester_id>` - Set your semester ID
- `/alarm <true/false>` - Toggle grade notifications
- `/get` - View all course grades
- `/get <course_name>` - View specific course grades

## Development

To run locally for development:

1. Install PostgreSQL
2. Set up environment variables
3. Run:
   ```bash
   go mod download
   go run .
   ```

## Contributing

Your contributions make the open-source community a better place! Feel free to fork the project, submit pull requests, report bugs, or suggest new features.

## License

This project is licensed under the MIT License. For more information, please refer to the [LICENSE](LICENSE) file.

## Acknowledgment

This project is not affiliated with Eskişehir Technical University.
