# ESTU Grade Checker

A Golang Telegram bot designed to monitor and notify users of grade updates from the Eskişehir Technical University (ESTU) student information system. It offers a convenient way for students to stay updated with their academic progress directly through Telegram.

## Features

- Fetches grades from the ESTU student information system
- Notifies users of grade updates via Telegram
- Supports multiple users

## Getting Started

### Prerequisites

- Golang installed on your machine
- A Telegram bot token (obtained by creating a bot with [BotFather](https://t.me/botfather))

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/enharukalo/estu-grade-checker.git
   ```

2. **Install dependencies**:

   ```bash
   go mod tidy
   ```

3. **Configure environment variables**:

   Create a `.env` file in the root directory and add your Telegram bot token:

   ```bash
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

4. **Launch the bot**:

   ```bash
   go run .
   ```

## How to Use

After setting up the bot, interact with it on Telegram to manage your grade notifications:

- **`/start`**: Displays a welcome message and basic instructions.
- **`/cookie <your_cookie>`**: Sets your ESTU website session cookie.
- **`/donemid <your_semester_id>`**: Sets your semester ID for grade tracking.
- **`/alarm <true/false>`**: Toggles grade update notifications on or off.
- **`/get`**: Retrieves your current grades for all courses.
- **`/get <course_name>`**: Fetches the grades for a specific course.

## Contributing

Your contributions make the open-source community a better place! Feel free to fork the project, submit pull requests, report bugs, or suggest new features.

## License

This project is licensed under the MIT License. For more information, please refer to the [LICENSE](LICENSE) file.

## Acknowledgment

This project is not affiliated with Eskişehir Technical University.
