import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getLeaderboard } from '../api/client';
import './Leaderboard.css';

const Leaderboard = () => {
  const navigate = useNavigate();
  
  // Default to today's date in IST
  const getTodayIST = () => {
    const now = new Date();
    // Offset for IST (UTC + 5:30)
    const istTime = new Date(now.getTime() + (5.5 * 3600 * 1000));
    const year = istTime.getUTCFullYear();
    const month = String(istTime.getUTCMonth() + 1).padStart(2, '0');
    const date = String(istTime.getUTCDate()).padStart(2, '0');
    return `${year}-${month}-${date}`;
  };

  const [selectedDate, setSelectedDate] = useState(getTodayIST());
  const [leaderboardData, setLeaderboardData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadLeaderboardData(selectedDate);
  }, [selectedDate]);

  const loadLeaderboardData = async (date) => {
    setLoading(true);
    setError('');
    try {
      const data = await getLeaderboard(date);
      setLeaderboardData(data);
    } catch (err) {
      logError(err);
      setError(err.message || 'Failed to retrieve leaderboard data.');
    } finally {
      setLoading(false);
    }
  };

  const logError = (err) => {
    console.error('Leaderboard load error:', err);
  };

  const handleDateChange = (e) => {
    setSelectedDate(e.target.value);
  };

  const getRankBadge = (rank) => {
    if (rank === 1) return <span className="rank-badge gold-medal">🥇</span>;
    if (rank === 2) return <span className="rank-badge silver-medal">🥈</span>;
    if (rank === 3) return <span className="rank-badge bronze-medal">🥉</span>;
    return <span className="rank-badge-number">{rank}</span>;
  };

  if (loading && !leaderboardData) {
    return (
      <div className="container leaderboard-loading">
        <div className="spinner"></div>
        <p className="animate-pulse">Loading daily leaderboard...</p>
      </div>
    );
  }

  const hasSubmissions = leaderboardData && leaderboardData.topSubmissions && leaderboardData.topSubmissions.length > 0;
  const topEntry = hasSubmissions ? leaderboardData.topSubmissions[0] : null;

  return (
    <div className="container page-container leaderboard-screen animate-fade-in">
      <div className="leaderboard-header-row">
        <div>
          <h1 className="page-title font-extrabold">Global Leaderboard</h1>
          <p className="page-subtitle">Compare solutions with other system design architects.</p>
        </div>
        
        {/* Date Selector input */}
        <div className="date-picker-wrapper">
          <label htmlFor="leaderboard-date">CHALLENGE DATE</label>
          <input
            type="date"
            id="leaderboard-date"
            value={selectedDate}
            onChange={handleDateChange}
            className="leaderboard-date-input"
            max={getTodayIST()}
          />
        </div>
      </div>

      {error ? (
        <div className="glass-card error-card">
          <h2>Failed to Load Leaderboard</h2>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={() => loadLeaderboardData(selectedDate)}>Retry</button>
        </div>
      ) : leaderboardData ? (
        <div className="leaderboard-content-wrapper">
          {/* Challenge summary details */}
          <div className="glass-card challenge-info-card">
            <span className="card-lbl">CHALLENGE SYSTEM</span>
            <h2>{leaderboardData.leaderboard.questionTitle || 'Architecture Challenge'}</h2>
            <div className="challenge-aggregates">
              <div className="aggregate-box">
                <span className="agg-label">TOTAL SUBMISSIONS</span>
                <span className="agg-val">{leaderboardData.leaderboard.totalSubmissions}</span>
              </div>
              <div className="aggregate-box">
                <span className="agg-label">HIGHEST SCORE</span>
                <span className="agg-val">{leaderboardData.leaderboard.highestScore}</span>
              </div>
              <div className="aggregate-box">
                <span className="agg-label">TOP SCORER</span>
                <span className="agg-val text-truncate" title={leaderboardData.leaderboard.topScorer}>
                  {leaderboardData.leaderboard.topScorer}
                </span>
              </div>
            </div>
          </div>

          {/* Top 1 Highlight Card */}
          {topEntry && (
            <div className="glass-card top-performer-highlight animate-pulse">
              <div className="crown-icon-wrapper">👑</div>
              <div className="top-performer-details">
                <span className="highlight-tag">TOP ARCHITECT</span>
                <h3>{topEntry.userId}</h3>
                <p>Achieved a score of <strong>{topEntry.score}/100</strong> with a highly polished distributed design.</p>
              </div>
              <div className="highlight-score-badge">
                <span>{topEntry.score}</span>
              </div>
            </div>
          )}

          {/* Leaderboard Rankings list */}
          <div className="glass-card ranking-list-card">
            <h3>Architect Rankings</h3>
            <div className="table-wrapper">
              <table className="rankings-table">
                <thead>
                  <tr>
                    <th className="th-center">Rank</th>
                    <th>Architect ID</th>
                    <th>Architecture Score</th>
                    <th>Submitted At</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {!hasSubmissions ? (
                    <tr>
                      <td colSpan="5" className="empty-rankings-row">No architecture solutions evaluated for this date yet.</td>
                    </tr>
                  ) : (
                    leaderboardData.topSubmissions.map((sub, idx) => {
                      const rank = idx + 1;
                      return (
                        <tr key={idx} className={rank === 1 ? 'rank-row-top' : ''}>
                          <td className="td-center">{getRankBadge(rank)}</td>
                          <td className="font-semibold">{sub.userId}</td>
                          <td>
                            <div className="score-cell-rank">
                              <span className="score-num-rank">{sub.score}</span>
                              <div className="score-bar-rank">
                                <div
                                  className="score-bar-fill"
                                  style={{
                                    width: `${sub.score}%`,
                                    background: sub.score >= 85 ? 'var(--success)' : sub.score >= 70 ? 'var(--accent-blue)' : 'var(--warning)'
                                  }}
                                ></div>
                              </div>
                            </div>
                          </td>
                          <td className="time-col">
                            {new Date(sub.submittedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </td>
                          <td>
                            <button
                              className="btn btn-secondary btn-rank-view"
                              onClick={() => navigate(`/result/${sub.submissionId}`)}
                            >
                              View Design
                            </button>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default Leaderboard;
