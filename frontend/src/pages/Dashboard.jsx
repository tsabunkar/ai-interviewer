import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getUserStats } from '../api/client';
import StreakCalendar from '../components/StreakCalendar';
import './Dashboard.css';

const Dashboard = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Custom user switcher state
  const [userId, setUserId] = useState('');
  const [editUserId, setEditUserId] = useState(false);
  const [newUserId, setNewUserId] = useState('');

  useEffect(() => {
    const storedUser = localStorage.getItem('userId') || 'guest_user';
    setUserId(storedUser);
    loadStats(storedUser);
  }, []);

  const loadStats = async (user) => {
    setLoading(true);
    setError('');
    try {
      const data = await getUserStats(user);
      setStats(data);
    } catch (err) {
      logError(err);
      setError(err.message || 'Failed to retrieve stats.');
    } finally {
      setLoading(false);
    }
  };

  const logError = (err) => {
    console.error('Dashboard load error:', err);
  };

  const handleUserSwitchSubmit = (e) => {
    e.preventDefault();
    const cleanUser = newUserId.trim().replace(/[^a-zA-Z0-9_]/g, '');
    if (cleanUser.length < 3) return;
    
    localStorage.setItem('userId', cleanUser);
    setUserId(cleanUser);
    setEditUserId(false);
    
    // Dispatch custom event to notify Navbar immediately
    window.dispatchEvent(new Event('userChanged'));
    
    loadStats(cleanUser);
  };

  const getScoreColorClass = (score) => {
    if (score >= 85) return 'score-green';
    if (score >= 70) return 'score-yellow';
    return 'score-red';
  };

  if (loading && !stats) {
    return (
      <div className="container dashboard-loading">
        <div className="spinner"></div>
        <p className="animate-pulse">Loading architectural dashboard...</p>
      </div>
    );
  }

  return (
    <div className="container page-container dashboard-screen animate-fade-in">
      <div className="dashboard-header">
        <div>
          <h1 className="page-title">Candidate Profile</h1>
          <div className="user-selector-row">
            {editUserId ? (
              <form onSubmit={handleUserSwitchSubmit} className="user-switch-form">
                <input
                  type="text"
                  value={newUserId}
                  onChange={(e) => setNewUserId(e.target.value)}
                  placeholder="Enter User ID..."
                  className="switch-input"
                  required
                />
                <button type="submit" className="btn btn-primary btn-switch-save">Save</button>
                <button type="button" className="btn btn-secondary btn-switch-cancel" onClick={() => setEditUserId(false)}>Cancel</button>
              </form>
            ) : (
              <div className="user-id-display">
                <span>Architect ID: <strong>{userId}</strong></span>
                <button
                  className="edit-user-btn"
                  onClick={() => {
                    setNewUserId(userId);
                    setEditUserId(true);
                  }}
                >
                  <svg className="edit-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                    <path d="M18.5 2.5a2.121 2.121 0 1 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                  </svg>
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {error ? (
        <div className="glass-card error-card">
          <h2>Failed to Load Stats</h2>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={() => loadStats(userId)}>Retry</button>
        </div>
      ) : stats ? (
        <div className="dashboard-content">
          {/* Stats Cards Row */}
          <div className="dashboard-stats-grid">
            <div className="glass-card d-stat-card">
              <span className="d-stat-label">TOTAL SUBMISSIONS</span>
              <span className="d-stat-value">{stats.totalSubmissions}</span>
            </div>
            <div className="glass-card d-stat-card">
              <span className="d-stat-label">AVERAGE SCORE</span>
              <span className="d-stat-value">{stats.averageScore.toFixed(1)}</span>
            </div>
            <div className="glass-card d-stat-card">
              <span className="d-stat-label">CURRENT STREAK</span>
              <span className="d-stat-value">{stats.currentStreak} 🔥</span>
            </div>
            <div className="glass-card d-stat-card">
              <span className="d-stat-label">LONGEST STREAK</span>
              <span className="d-stat-value">{stats.longestStreak} 🏆</span>
            </div>
          </div>

          {/* Activity Heatmap Grid */}
          <div className="dashboard-heatmap-section">
            <StreakCalendar submissions={stats.recentSubmissions} />
          </div>

          {/* Bottom details splits */}
          <div className="dashboard-bottom-grid">
            {/* Weak Areas List */}
            <div className="glass-card weak-areas-card">
              <h3>Strength & Weakness Breakdown</h3>
              <p className="weak-card-sub text-muted">Your performance sorted from weakest to strongest categories:</p>
              
              <div className="weak-areas-list">
                {stats.weakAreas.map((area, idx) => (
                  <div key={idx} className="weak-area-item">
                    <div className="weak-area-header">
                      <span className="weak-area-name">{area.category}</span>
                      <span className="weak-area-score">{area.averageScore.toFixed(0)}/100</span>
                    </div>
                    <div className="progress-track">
                      <div
                        className="progress-fill"
                        style={{
                          width: `${area.averageScore}%`,
                          background: area.averageScore >= 80 ? 'var(--success)' : area.averageScore >= 65 ? 'var(--accent-blue)' : 'var(--warning)'
                        }}
                      ></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Submission History Table */}
            <div className="glass-card history-card">
              <h3>Submission History</h3>
              <div className="table-wrapper">
                <table className="history-table">
                  <thead>
                    <tr>
                      <th>Challenge Date</th>
                      <th>Status</th>
                      <th>Score</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.recentSubmissions.length === 0 ? (
                      <tr>
                        <td colSpan="4" className="empty-table-row">No architecture solutions submitted yet.</td>
                      </tr>
                    ) : (
                      stats.recentSubmissions.map((sub, idx) => (
                        <tr key={idx}>
                          <td>{sub.questionDate}</td>
                          <td>
                            <span className={`status-tag status-${sub.status}`}>
                              {sub.status}
                            </span>
                          </td>
                          <td>
                            {sub.status === 'evaluated' ? (
                              <span className={`history-score ${getScoreColorClass(sub.score)}`}>
                                {sub.score}
                              </span>
                            ) : (
                              <span className="history-score-pending">-</span>
                            )}
                          </td>
                          <td>
                            <button
                              className="btn btn-secondary btn-history-action"
                              onClick={() => navigate(`/result/${sub.submissionId}`)}
                            >
                              View Details
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default Dashboard;
