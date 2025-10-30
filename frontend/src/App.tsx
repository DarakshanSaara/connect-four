// docker-compose up -d
// http://localhost:3000/
// docker-compose build frontend
// docker-compose build backend
// docker-compose down

import { useState } from 'react';
import GameBoard from './components/GameBoard';
import Leaderboard from './components/LeaderBoard';
import './App.css';

// Define TypeScript interfaces
interface Player {
  id: string;
  username: string;
  isBot: boolean;
}

interface Game {
  id: string;
  board: number[][];
  players: Player[];
  currentPlayer: number;
  status: 'waiting' | 'playing' | 'finished';
  winner: number;
}

function App() {
  const [currentView, setCurrentView] = useState<'home' | 'game' | 'leaderboard'>('home');
  const [username, setUsername] = useState('');
  const [game, setGame] = useState<Game | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);

  const handleCreateGame = async () => {
    if (!username.trim()) {
      alert('Please enter a username');
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/game/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username }),
      });

      const gameData: Game = await response.json();
      setGame(gameData);
      connectWebSocket(gameData.id, username);
      setCurrentView('game');
    } catch (error) {
      console.error('Error creating game:', error);
      alert('Failed to create game');
    }
  };

  const connectWebSocket = (gameId: string, playerUsername: string) => {
    const ws = new WebSocket(`ws://localhost:8080/ws?gameId=${gameId}&username=${playerUsername}`);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      if (message.type === 'game_update') {
        setGame(message.content);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
    };

    setSocket(ws);
  };

  const makeMove = (column: number) => {
    if (socket && game && game.status === 'playing') {
      const message = {
        type: 'make_move',
        content: {
          gameId: game.id,
          playerId: game.players[0].id, // Assuming player is always first
          column: column,
        },
      };
      socket.send(JSON.stringify(message));
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>4 in a Row</h1>
        <nav>
          <button onClick={() => setCurrentView('home')}>Home</button>
          <button onClick={() => setCurrentView('leaderboard')}>Leaderboard</button>
        </nav>
      </header>

      <main className="app-main">
        {currentView === 'home' && (
          <div className="home-view">
            <div className="game-setup">
              <input
                type="text"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="username-input"
              />
              <button onClick={handleCreateGame} className="create-game-btn">
                Start New Game
              </button>
              <p className="waiting-note">
                If no opponent joins within 10 seconds, you'll play against our competitive bot!
              </p>
            </div>
          </div>
        )}

        {currentView === 'game' && game && (
          <GameBoard game={game} onMove={makeMove} />
        )}

        {currentView === 'leaderboard' && (
          <Leaderboard />
        )}
      </main>
    </div>
  );
}

export default App;