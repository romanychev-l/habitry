# Habitry - Habit Tracker with TON Rewards

![Habitry Logo](https://placehold.co/600x400?text=Habitry&font=montserrat)

## 📋 About the Project

Habitry is a habit tracking application with TON blockchain integration that helps users develop beneficial habits and achieve their goals. The reward system based on TON cryptocurrency motivates users to consistently maintain their habits.

### 🌟 Key Features

- 📱 Create and track daily habits
- 🔔 Notifications to perform habits
- 🔄 Mutual accountability system with friends
- 💰 TON rewards for completing habits
- 📊 Progress statistics and analytics

## 🔧 Tech Stack

### Backend
- Golang
- PostgreSQL
- RESTful API
- Telegram Bot API

### Frontend
- Svelte
- TypeScript
- Vite
- Tailwind CSS

### Integrations
- TON Blockchain
- Telegram Bot

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go (for local development)
- Node.js and npm (for local development)

### Launch with Docker

```bash
# Clone the repository
git clone https://github.com/romanychev-l/habitry.git
cd habitry

# Create .env file with necessary settings
# Example settings can be found in backend/_.env

# Launch containers
docker-compose up -d
```

### Local Development

#### Backend
```bash
cd backend
go mod download
go run main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

#### Telegram Bot
```bash
cd py_bot
# Bot launch instructions...
```

## 🤝 Contributing

We welcome contributions to Habitry! If you want to contribute, please:

1. Fork the repository
2. Create a branch for your changes
3. Make changes and create a pull request

## 📄 License

This project is distributed under the [MIT](LICENSE) license.

## 📞 Contact

If you have questions or suggestions, please contact us:

- [Open an issue](https://github.com/romanychev-l/habitry/issues)
- Telegram: [@romanychev](https://t.me/romanychev)