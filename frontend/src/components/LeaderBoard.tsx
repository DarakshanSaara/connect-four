import React, { useState, useEffect } from 'react';

interface LeaderboardEntry {
  username: string;
  wins: number;
  losses: number;
  draws: number;
}

const Leaderboard: React.FC = () => {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
  try {
    console.log('Fetching leaderboard...');
    const response = await fetch('http://localhost:8080/leaderboard');
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Leaderboard data:', data);
    
    // Convert null to empty array
    const safeData = data === null ? [] : data;
    setLeaderboard(safeData);
  } catch (error) {
    console.error('Error fetching leaderboard:', error);
    setError('Failed to load leaderboard');
    setLeaderboard([]); // Set to empty array on error
  } finally {
    setLoading(false);
  }
};

  if (loading) {
    return (
      <div className="leaderboard">
        <h2>Leaderboard</h2>
        <div className="loading">Loading leaderboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="leaderboard">
        <h2>Leaderboard</h2>
        <div className="error-message">{error}</div>
        <button onClick={fetchLeaderboard} className="retry-btn">
          Retry
        </button>
      </div>
    );
  }

  // Handle null case explicitly
  const displayData = leaderboard || [];

  return (
    <div className="leaderboard">
      <h2>Leaderboard</h2>
      {displayData.length === 0 ? (
        <div className="no-data">No leaderboard data available. Play some games to see rankings!</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Rank</th>
              <th>Player</th>
              <th>Wins</th>
              <th>Losses</th>
              <th>Draws</th>
              <th>Total Games</th>
            </tr>
          </thead>
          <tbody>
            {displayData.map((entry, index) => (
              <tr key={entry.username || index}>
                <td>{index + 1}</td>
                <td>{entry.username || 'Unknown Player'}</td>
                <td>{entry.wins || 0}</td>
                <td>{entry.losses || 0}</td>
                <td>{entry.draws || 0}</td>
                <td>{(entry.wins || 0) + (entry.losses || 0) + (entry.draws || 0)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default Leaderboard;