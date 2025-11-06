## 🎯 4 in a Row - Real-Time Multiplayer Game
A full-stack, real-time implementation of the classic Connect Four game with competitive AI, built with Go backend and React frontend.

## 📋 Table of Contents
Features
Tech Stack
Architecture
Installation & Setup
API Documentation
Game Rules
Deployment
Project Structure
Contributing

## 🚀 Features
# Core Gameplay

✅ Real-time multiplayer using WebSockets

✅ Smart competitive bot with strategic AI

✅ 10-second matchmaking with bot fallback

✅ Complete game logic with win/draw detection

✅ Game state persistence with PostgreSQL

✅ Leaderboard system with player statistics

# User Experience

✅ Modern dark theme UI with smooth animations

✅ Responsive design for mobile and desktop

✅ Real-time game updates without page refresh

✅ Game reconnection support within 30 seconds

✅ Win/loss/draw tracking with detailed analytics

# Advanced Features

✅ Kafka integration for game analytics

✅ Docker containerization for easy deployment

✅ Health checks and monitoring endpoints

✅ Comprehensive logging and error handling

## 🛠 Tech Stack
# Backend
Language: Go 1.19+

Web Framework: Native HTTP with Gorilla WebSocket

Database: PostgreSQL with lib/pq driver

Message Queue: Apache Kafka with kafka-go

Containerization: Docker & Docker Compose

# Frontend
Framework: React 18 with TypeScript

Build Tool: Vite

Styling: Modern CSS with CSS Variables

WebSocket: Native WebSocket API

# Infrastructure
Container Runtime: Docker Engine

Orchestration: Docker Compose

Database: PostgreSQL 15

Message Broker: Apache Kafka with Zookeeper

## 🏗 Architecture
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React         │    │   Go Backend     │    │   PostgreSQL    │
│   Frontend      │◄──►│   WebSocket Hub  │◄──►│   Database      │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐             │
         └──────────────►│   Apache Kafka  │◄────────────┘
                         │   Analytics     │
                         └─────────────────┘

# Data Flow
Client Connection: WebSocket establishes real-time connection

Game Management: Hub manages game states and player matching

Bot Intelligence: Strategic AI makes competitive moves

Data Persistence: Game results stored in PostgreSQL

Analytics: Game events streamed to Kafka for metrics

## ⚙️ Installation & Setup
# Prerequisites
Docker & Docker Compose
Git

# Quick Start (Recommended)
# Clone the repository
```
git clone <repository-url>
cd connect-four

# Start all services
docker-compose up --build

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
```

# Manual Setup (Development)
Backend Setup
```
cd backend

# Install dependencies
go mod download

# Set environment variables
export DATABASE_URL=postgres://user:password@localhost:5432/connectfour?sslmode=disable
export KAFKA_BROKERS=localhost:9092

# Run the server
go run cmd/server/main.go
```

Frontend Setup
```
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Access at: http://localhost:5173
```

Database Setup
```
# Access PostgreSQL container
docker-compose exec postgres psql -U user -d connectfour

# Manual table creation (auto-created by app)
\dt
```
📚 API Documentation
WebSocket Endpoints
Connect to Game
```
const ws = new WebSocket('ws://localhost:8080/ws?gameId=<gameId>&username=<username>');
```
WebSocket Messages
Make Move
```
{
  "type": "make_move",
  "content": {
    "gameId": "game_123",
    "playerId": "player_456", 
    "column": 3
  }
}
```
Game Update (Server → Client)
```
{
  "type": "game_update",
  "content": {
    "id": "game_123",
    "board": [[...]],
    "players": [...],
    "currentPlayer": 0,
    "status": "playing",
    "winner": -1
  }
}
```

# REST API Endpoints
Create Game
```
POST /game/create
Content-Type: application/json

{
  "username": "player1"
}
```
Response:
```
{
  "id": "game_123",
  "board": [...],
  "players": [...],
  "status": "waiting"
}
```
Get Leaderboard
```
GET /leaderboard
```
Response:
```
[
  {
    "username": "player1",
    "wins": 5,
    "losses": 2,
    "draws": 1
  }
]
```
Health Check
```
GET /health
```
Response:
```
{
  "status": "healthy"
}
```

## 🎮 Game Rules
Basic Rules
Board: 7 columns × 6 rows grid

Objective: Connect 4 discs vertically, horizontally, or diagonally

Turns: Players alternate dropping discs into columns

Win Conditions: First to 4 in a row wins

Draw: Board fills completely with no winner

Bot Strategy
The competitive bot implements:

Immediate Win: Plays winning move if available

Block Opponent: Prevents player from winning

Create Threats: Builds multiple winning opportunities

Center Control: Prefers center columns for better positioning

Random Variation: Adds unpredictability to moves

Matchmaking Flow
Player enters username and creates game

System waits 10 seconds for opponent

If no opponent joins, competitive bot joins automatically

Random player starts first

Game continues until win/draw

## 🚀 Deployment
Production Deployment
Environment Variables
```
# Backend Environment
DATABASE_URL=postgres://user:password@postgres:5432/connectfour?sslmode=disable
KAFKA_BROKERS=kafka:29092
PORT=8080

# Frontend Environment
VITE_API_URL=http://your-domain.com:8080
```         
Docker Compose for Production
```
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: connectfour
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://user:password@postgres:5432/connectfour?sslmode=disable
      KAFKA_BROKERS: kafka:29092
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend
```
Cloud Deployment Options
AWS ECS
```
# ecs-task-definition.yml
# Define ECS task with all services
```
Kubernetes
```
# kubernetes-deployment.yml
# K8s manifests for microservices
```

## 📁 Project Structure

connect-four/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                 # Application entry point
│   ├── internal/
│   │   ├── game/                       # Game logic and rules
│   │   ├── bot/                        # AI bot implementation
│   │   ├── websockethub/               # WebSocket connection management
│   │   ├── database/                   # PostgreSQL operations
│   │   └── kafka/                      # Analytics event streaming
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── GameBoard.tsx           # Game board component
│   │   │   └── LeaderBoard.tsx         # Leaderboard component
│   │   ├── App.tsx                     # Main application component
│   │   └── App.css                     # Modern CSS styles
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml                  # Multi-container setup
└── README.md

🔧 Development
Running Tests
```
# Backend tests
cd backend
go test ./...

# Frontend tests  
cd frontend
npm test
```
Code Quality
```
# Backend linting
gofmt -w .
go vet ./...

# Frontend linting
npm run lint
```
Database Migrations
```
-- Manual schema creation (auto-handled by application)
CREATE TABLE games (
    id VARCHAR(50) PRIMARY KEY,
    player1 VARCHAR(100) NOT NULL,
    player2 VARCHAR(100),
    winner VARCHAR(100),
    status VARCHAR(20) NOT NULL,
    board_state TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP
);

CREATE TABLE leaderboard (
    username VARCHAR(100) PRIMARY KEY,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);
```

## 📊 Analytics & Monitoring
Kafka Events
The application streams game events to Kafka for analytics:
```
type GameEvent struct {
    Type      string    `json:"type"`    // game_start, move, game_end
    GameID    string    `json:"gameId"`
    PlayerID  string    `json:"playerId"`
    Username  string    `json:"username"`
    Data      string    `json:"data"`
    Timestamp time.Time `json:"timestamp"`
}
```
Sample Analytics Queries
```
-- Most active players
SELECT username, COUNT(*) as games_played 
FROM games 
WHERE player1 = username OR player2 = username 
GROUP BY username 
ORDER BY games_played DESC;

-- Win rates
SELECT username, 
       wins, 
       losses, 
       draws,
       ROUND(wins::decimal / NULLIF(wins + losses + draws, 0) * 100, 2) as win_rate
FROM leaderboard 
ORDER BY win_rate DESC;
```

## 🤝 Contributing
Development Workflow
Fork the repository

Create a feature branch (git checkout -b feature/amazing-feature)

Commit changes (git commit -m 'Add amazing feature')

Push to branch (git push origin feature/amazing-feature)

Open a Pull Request

Code Standards
Backend: Follow Go standard formatting and conventions

Frontend: Use TypeScript with strict type checking

Commits: Conventional commits format

Documentation: Update README for new features

🙏 Acknowledgments
Connect Four game rules and mechanics

Go standard library and Gorilla WebSocket team

React and Vite communities

Docker and containerization technologies

🎯 Live Demo
Access the live application: http://localhost:3000
