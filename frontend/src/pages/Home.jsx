import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchTodayQuestion, getUserStats } from '../api/client';
import './Home.css';

const Home = () => {
  const navigate = useNavigate();
  const [question, setQuestion] = useState(null);
  const [stats, setStats] = useState({ currentStreak: 0, totalSubmissions: 0, averageScore: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState({ hours: 0, minutes: 0, seconds: 0 });

  // Get current userId or set default
  const getUserId = () => {
    let uId = localStorage.getItem('userId');
    if (!uId) {
      uId = `architect_${Math.floor(Math.random() * 9000) + 1000}`;
      localStorage.setItem('userId', uId);
    }
    return uId;
  };

  useEffect(() => {
    const loadHomeData = async () => {
      try {
        const uId = getUserId();
        const qData = await fetchTodayQuestion();
        setQuestion(qData);
      } catch (err) {
        logError(err);
        setError(err.message || 'Could not load today\'s challenge.');
      }

      try {
        const uId = getUserId();
        const sData = await getUserStats(uId);
        setStats(sData);
      } catch (err) {
        logError(err);
      } finally {
        setLoading(false);
      }
    };

    loadHomeData();
  }, []);

  const logError = (err) => {
    console.error('Home load error:', err);
  };

  // Countdown timer logic to next challenge (tomorrow 06:30 IST = 01:00 UTC)
  useEffect(() => {
    const timer = setInterval(() => {
      const now = new Date();
      // Target is today or tomorrow at 01:00 UTC (06:30 IST)
      const target = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate(), 1, 0, 0));
      
      if (now > target) {
        target.setUTCDate(target.getUTCDate() + 1);
      }
      
      const diffMs = target - now;
      
      const hours = Math.floor(diffMs / (1000 * 60 * 60));
      const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diffMs % (1000 * 60)) / 1000);
      
      setCountdown({ hours, minutes, seconds });
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const formatCountdownUnit = (val) => {
    return String(val).padStart(2, '0');
  };

  if (loading) {
    return (
      <div className="container home-loading">
        <div className="spinner"></div>
        <p className="animate-pulse">Retrieving architecture challenge...</p>
      </div>
    );
  }

  return (
    <div className="container page-container animate-fade-in">
      {/* Hero section */}
      <section className="home-hero">
        <div className="hero-text">
          <span className="hero-badge">DAILY CHALLENGE PLATFORM</span>
          <h1 className="hero-title">
            Master System Design <span className="gradient-text">One Day at a Time</span>
          </h1>
          <p className="hero-subtitle">
            Every day at 06:30 IST, tackle a new production-scale architecture challenge. Get instant, detailed AI feedback across 7 core categories.
          </p>
        </div>
      </section>

      {/* Main dashboard view */}
      <div className="home-grid">
        {/* Left column: today's question card */}
        <div className="home-left">
          <div className="glass-card today-card">
            <div className="card-header">
              <span className="today-label">TODAY'S CHALLENGE</span>
              {question && (
                <span className={`badge badge-${question.difficulty.toLowerCase()}`}>
                  {question.difficulty}
                </span>
              )}
            </div>
            {error ? (
              <div className="question-error">
                <svg className="error-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                <p>{error}</p>
              </div>
            ) : question ? (
              <div className="question-details">
                <h2 className="question-title-text">{question.title}</h2>
                <p className="question-summary">
                  {question.description.split('\n')[2] || 'Tackle today\'s production-scale system design and earn architectural credentials.'}
                </p>
                <div className="question-categories">
                  {question.categories.slice(0, 4).map((cat, idx) => (
                    <span key={idx} className="category-pill">{cat}</span>
                  ))}
                  {question.categories.length > 4 && (
                    <span className="category-pill-more">+{question.categories.length - 4} more</span>
                  )}
                </div>
                <button className="btn btn-primary start-btn" onClick={() => navigate('/challenge')}>
                  Start Challenge
                  <svg className="btn-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <line x1="5" y1="12" x2="19" y2="12" />
                    <polyline points="12 5 19 12 12 19" />
                  </svg>
                </button>
              </div>
            ) : (
              <p>Loading question...</p>
            )}
          </div>
        </div>

        {/* Right column: user stats & countdown */}
        <div className="home-right-cols">
          {/* Stats grid */}
          <div className="stats-row">
            <div className="glass-card stat-item">
              <span className="stat-label">STREAK</span>
              <span className="stat-value">{stats.currentStreak} 🔥</span>
            </div>
            <div className="glass-card stat-item">
              <span className="stat-label">SUBMISSIONS</span>
              <span className="stat-value">{stats.totalSubmissions} 📊</span>
            </div>
            <div className="glass-card stat-item">
              <span className="stat-label">AVG SCORE</span>
              <span className="stat-value">{stats.averageScore.toFixed(0)} 🏆</span>
            </div>
          </div>

          {/* Countdown card */}
          <div className="glass-card countdown-card">
            <span className="countdown-label">NEXT CHALLENGE IN</span>
            <div className="countdown-time">
              <div className="time-block">
                <span className="time-num">{formatCountdownUnit(countdown.hours)}</span>
                <span className="time-lbl">hours</span>
              </div>
              <span className="time-separator">:</span>
              <div className="time-block">
                <span className="time-num">{formatCountdownUnit(countdown.minutes)}</span>
                <span className="time-lbl">mins</span>
              </div>
              <span className="time-separator">:</span>
              <div className="time-block">
                <span className="time-num">{formatCountdownUnit(countdown.seconds)}</span>
                <span className="time-lbl">secs</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Feature explanations */}
      <section className="features-section">
        <h2 className="section-title">How It Works</h2>
        <div className="features-grid">
          <div className="glass-card feature-item">
            <div className="feature-icon-wrapper blue">
              <svg className="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <h3>1. Solve Daily Challenges</h3>
            <p>Read the problem statement, requirements, traffic estimates, and compose your detailed solution.</p>
          </div>
          <div className="glass-card feature-item">
            <div className="feature-icon-wrapper purple">
              <svg className="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <polygon points="12 2 2 7 12 12 22 7 12 2" />
                <polyline points="2 17 12 22 22 17" />
                <polyline points="2 12 12 17 22 12" />
              </svg>
            </div>
            <h3>2. AI-Powered Evaluation</h3>
            <p>Amazon Bedrock evaluates your design across Requirements, APIs, DB, Cache, Scaling, Availability, and Tradeoffs.</p>
          </div>
          <div className="glass-card feature-item">
            <div className="feature-icon-wrapper emerald">
              <svg className="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
            </div>
            <h3>3. Identify Weak Areas</h3>
            <p>Track your score history and see where you need to improve with personalized candidate metrics.</p>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Home;
