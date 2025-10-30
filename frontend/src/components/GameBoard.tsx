import React from 'react';

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

interface GameBoardProps {
  game: Game;
  onMove: (column: number) => void;
}

const GameBoard: React.FC<GameBoardProps> = ({ game, onMove }) => {
  const renderCell = (row: number, col: number) => {
    const cellValue = game.board[row][col];
    let cellClass = 'cell';
    
    if (cellValue === 1) {
      cellClass += ' player1';
    } else if (cellValue === 2) {
      cellClass += ' player2';
    }

    return <div key={`${row}-${col}`} className={cellClass}></div>;
  };

  const renderColumn = (colIndex: number) => {
    const isColumnFull = game.board[0][colIndex] !== 0;
    
    return (
      <div 
        key={colIndex} 
        className="column"
        onClick={() => !isColumnFull && onMove(colIndex)}
      >
        {Array.from({ length: 6 }).map((_, rowIndex) =>
          renderCell(rowIndex, colIndex)
        )}
      </div>
    );
  };

  const getGameStatus = () => {
    if (game.status === 'waiting') {
      return 'Waiting for opponent...';
    } else if (game.status === 'finished') {
      if (game.winner === -1) {
        return "It's a draw!";
      } else {
        return `Winner: ${game.players[game.winner].username}`;
      }
    } else {
      const currentPlayer = game.players[game.currentPlayer];
      return `Current turn: ${currentPlayer.username}`;
    }
  };

  return (
    <div className="game-board">
      <div className="game-info">
        <h2>Game {game.id}</h2>
        <div className="players">
          <div className="player-info">
            <span className="player-dot player1"></span>
            {game.players[0].username} (You)
          </div>
          <div className="player-info">
            <span className="player-dot player2"></span>
            {game.players[1]?.username || 'Waiting...'}
          </div>
        </div>
        <div className="game-status">{getGameStatus()}</div>
      </div>

      <div className="board">
        {Array.from({ length: 7 }).map((_, colIndex) =>
          renderColumn(colIndex)
        )}
      </div>

      {game.status === 'finished' && (
        <div className="game-over">
          <button onClick={() => window.location.reload()}>
            Play Again
          </button>
        </div>
      )}
    </div>
  );
};

export default GameBoard;