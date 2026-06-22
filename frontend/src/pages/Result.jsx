import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getResults } from '../api/client';
import ScoreRadar from '../components/ScoreRadar';
import './Result.css';

const Result = () => {
  const { submissionId } = useParams();
  const navigate = useNavigate();
  const [submission, setSubmission] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Score countup animation state
  const [animatedScore, setAnimatedScore] = useState(0);

  useEffect(() => {
    let timer;
    
    const pollResults = async () => {
      try {
        const data = await getResults(submissionId);
        setSubmission(data);
        
        if (data.status === 'evaluated') {
          setLoading(false);
          // Trigger count up animation
          triggerCountUp(data.score);
        } else {
          // Poll again after 3 seconds
          timer = setTimeout(pollResults, 3000);
        }
      } catch (err) {
        logError(err);
        setError(err.message || 'Failed to fetch evaluation results.');
        setLoading(false);
      }
    };

    pollResults();

    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [submissionId]);

  const logError = (err) => {
    console.error('Result load error:', err);
  };

  const triggerCountUp = (targetScore) => {
    let start = 0;
    const duration = 1500; // 1.5s
    const stepTime = Math.max(Math.floor(duration / targetScore), 15);
    
    const counterTimer = setInterval(() => {
      start += 1;
      setAnimatedScore(start);
      if (start >= targetScore) {
        clearInterval(counterTimer);
        setAnimatedScore(targetScore); // set exact final
      }
    }, stepTime);
  };

  const getScoreRatingLabel = (score) => {
    if (score >= 90) return { label: 'Distinguished Architect', color: 'rating-gold' };
    if (score >= 80) return { label: 'Principal Tier', color: 'rating-purple' };
    if (score >= 70) return { label: 'Senior System Designer', color: 'rating-blue' };
    return { label: 'System Designer', color: 'rating-gray' };
  };

  if (loading && (!submission || submission.status === 'pending')) {
    return (
      <div className="container page-container result-loading animate-fade-in">
        <div className="glass-card evaluation-loader-card">
          <div className="radar-scanner">
            <div className="scanner-line"></div>
            <svg className="scanner-grid" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="45" fill="none" stroke="rgba(255,255,255,0.05)" strokeWidth="1" />
              <circle cx="50" cy="50" r="30" fill="none" stroke="rgba(255,255,255,0.05)" strokeWidth="1" />
              <circle cx="50" cy="50" r="15" fill="none" stroke="rgba(255,255,255,0.05)" strokeWidth="1" />
              <line x1="50" y1="5" x2="50" y2="95" stroke="rgba(255,255,255,0.05)" strokeWidth="1" />
              <line x1="5" y1="50" x2="95" y2="50" stroke="rgba(255,255,255,0.05)" strokeWidth="1" />
            </svg>
          </div>
          <h2>AI Architect is Evaluating...</h2>
          <p>We are analyzing your system across Requirements, API Design, Data Models, Scalability, Availability, and Tradeoffs. This may take up to 30 seconds.</p>
          <div className="loading-bar">
            <div className="bar-fill"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container page-container result-error-screen animate-fade-in">
        <div className="glass-card error-card">
          <h2>Evaluation Retrieval Failed</h2>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={() => navigate('/')}>Back to Home</button>
        </div>
      </div>
    );
  }

  const { score, evaluation } = submission;
  const rating = getScoreRatingLabel(score);

  // SVG parameters for score circle
  const radius = 70;
  const strokeWidth = 8;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (animatedScore / 100) * circumference;

  return (
    <div className="container page-container result-screen animate-fade-in">
      <div className="result-header">
        <h1 className="page-title">Evaluation Report</h1>
        <p className="page-subtitle">Submitted for question: <strong>{submission.questionDate}</strong></p>
      </div>

      <div className="result-layout">
        {/* Left Side: Score display & Radar Chart */}
        <div className="result-left-col">
          {/* Score Reveal Card */}
          <div className="glass-card score-reveal-card">
            <div className="score-circle-wrapper">
              <svg className="score-svg" width="160" height="160" viewBox="0 0 160 160">
                <defs>
                  <linearGradient id="scoreGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stopColor="#3b82f6" />
                    <stop offset="100%" stopColor="#8b5cf6" />
                  </linearGradient>
                </defs>
                {/* Background Ring */}
                <circle
                  cx="80"
                  cy="80"
                  r={radius}
                  fill="none"
                  stroke="rgba(255,255,255,0.03)"
                  strokeWidth={strokeWidth}
                />
                {/* Foreground Filled Ring */}
                <circle
                  cx="80"
                  cy="80"
                  r={radius}
                  fill="none"
                  stroke="url(#scoreGrad)"
                  strokeWidth={strokeWidth}
                  strokeDasharray={circumference}
                  strokeDashoffset={strokeDashoffset}
                  strokeLinecap="round"
                  transform="rotate(-90 80 80)"
                  style={{ transition: 'stroke-dashoffset 0.1s ease-out' }}
                />
              </svg>
              <div className="score-text-overlay">
                <span className="score-num">{animatedScore}</span>
                <span className="score-max">/100</span>
              </div>
            </div>
            
            <div className="score-tier-details">
              <span className={`rating-badge ${rating.color}`}>{rating.label}</span>
            </div>
          </div>

          {/* Radar Chart */}
          <div className="glass-card radar-card">
            <h3>Architectural Footprint</h3>
            <div className="radar-container">
              <ScoreRadar categories={evaluation.categories} />
            </div>
          </div>
        </div>

        {/* Right Side: Feedback categories & summary */}
        <div className="result-right-col">
          {/* Overall summary */}
          <div className="glass-card summary-card">
            <h3 className="section-title-res">Executive Summary</h3>
            <p className="summary-text">{evaluation.summary}</p>
          </div>

          {/* Strengths & Weaknesses row */}
          <div className="strengths-weaknesses-row">
            <div className="glass-card strengths-card">
              <h3 className="section-title-res success-text">Strengths</h3>
              <ul>
                {evaluation.strengths.map((str, idx) => (
                  <li key={idx}>
                    <span className="bullet success-bg">✓</span>
                    <p>{str}</p>
                  </li>
                ))}
              </ul>
            </div>

            <div className="glass-card weaknesses-card">
              <h3 className="section-title-res warning-text">Weaknesses</h3>
              <ul>
                {evaluation.weaknesses.map((weak, idx) => (
                  <li key={idx}>
                    <span className="bullet warning-bg">!</span>
                    <p>{weak}</p>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          {/* Category Details */}
          <div className="glass-card category-details-card">
            <h3 className="section-title-res">Category Breakdown</h3>
            <div className="category-details-list">
              {Object.entries(evaluation.categories).map(([catName, data], idx) => (
                <div key={idx} className="category-detail-item">
                  <div className="category-detail-header">
                    <span className="cat-name">{catName}</span>
                    <span className="cat-score">{data.score}/100</span>
                  </div>
                  
                  {/* Custom progress bar */}
                  <div className="progress-track">
                    <div
                      className="progress-fill"
                      style={{
                        width: `${data.score}%`,
                        background: data.score >= 80 ? 'var(--success)' : data.score >= 65 ? 'var(--accent-blue)' : 'var(--warning)'
                      }}
                    ></div>
                  </div>
                  <p className="cat-feedback">{data.feedback}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Navigation CTA Actions */}
          <div className="result-ctas">
            <button className="btn btn-secondary" onClick={() => navigate('/dashboard')}>
              Go to Dashboard
            </button>
            <button className="btn btn-primary" onClick={() => navigate('/leaderboard')}>
              View Leaderboard
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Result;
