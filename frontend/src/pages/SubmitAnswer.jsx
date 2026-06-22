import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { fetchTodayQuestion, submitAnswer } from '../api/client';
import './SubmitAnswer.css';

const SubmitAnswer = () => {
  const navigate = useNavigate();
  const [question, setQuestion] = useState(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [answer, setAnswer] = useState('');
  
  // Username Prompt Modal State
  const [showUserModal, setShowUserModal] = useState(false);
  const [tempUsername, setTempUsername] = useState('');
  const [usernameError, setUsernameError] = useState('');
  const [activeHints, setActiveHints] = useState({});

  // Checklist items state
  const [checklist, setChecklist] = useState({
    requirements: false,
    apis: false,
    database: false,
    caching: false,
    scaling: false,
    availability: false,
    tradeoffs: false
  });

  useEffect(() => {
    // Check if user has an identity in localStorage
    const storedUser = localStorage.getItem('userId');
    if (!storedUser || storedUser.startsWith('architect_')) {
      // Prompt for name on first visit
      setShowUserModal(true);
      if (storedUser) {
        setTempUsername(storedUser);
      }
    }

    const loadQuestion = async () => {
      try {
        const data = await fetchTodayQuestion();
        setQuestion(data);
      } catch (err) {
        logError(err);
        setError(err.message || 'Failed to fetch today\'s challenge.');
      } finally {
        setLoading(false);
      }
    };

    loadQuestion();
  }, []);

  const logError = (err) => {
    console.error('SubmitAnswer load error:', err);
  };

  const handleSaveUsername = (e) => {
    e.preventDefault();
    const usernameClean = tempUsername.trim().replace(/[^a-zA-Z0-9_]/g, '');
    if (usernameClean.length < 3) {
      setUsernameError('Display name must be at least 3 alphanumeric characters/underscores.');
      return;
    }
    localStorage.setItem('userId', usernameClean);
    setShowUserModal(false);
    
    // Dispatch a custom event so the Navbar updates its display immediately
    window.dispatchEvent(new Event('userChanged'));
  };

  const handleCheckboxChange = (key) => {
    setChecklist(prev => ({
      ...prev,
      [key]: !prev[key]
    }));
  };

  const toggleHint = (index) => {
    setActiveHints(prev => ({
      ...prev,
      [index]: !prev[index]
    }));
  };

  const handleSubmit = async () => {
    if (answer.trim().length < 100) return;
    
    setSubmitting(true);
    setError('');
    
    try {
      const uId = localStorage.getItem('userId') || 'guest_user';
      const result = await submitAnswer(uId, question.date, answer);
      // Redirect to results page
      navigate(`/result/${result.submissionId}`);
    } catch (err) {
      logError(err);
      setError(err.message || 'Submission failed. Please try again.');
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="container challenge-loading">
        <div className="spinner"></div>
        <p className="animate-pulse">Loading system design workspace...</p>
      </div>
    );
  }

  if (error && !question) {
    return (
      <div className="container page-container challenge-error-screen animate-fade-in">
        <div className="glass-card error-card">
          <h2>Error Loading Challenge</h2>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={() => navigate('/')}>Return Home</button>
        </div>
      </div>
    );
  }

  return (
    <div className="container page-container challenge-workspace animate-fade-in">
      {/* Username Modal overlay */}
      {showUserModal && (
        <div className="modal-overlay">
          <div className="glass-card modal-content animate-slide-up">
            <h3>Configure Your Display Name</h3>
            <p>Set a display name to track streaks and rank on the global daily leaderboard.</p>
            <form onSubmit={handleSaveUsername}>
              <div className="form-group">
                <input
                  type="text"
                  placeholder="e.g. cloud_master99"
                  value={tempUsername}
                  onChange={(e) => {
                    setTempUsername(e.target.value);
                    setUsernameError('');
                  }}
                  className="modal-input"
                  required
                />
                {usernameError && <span className="input-error">{usernameError}</span>}
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn btn-primary modal-save-btn">
                  Set Username
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Main Workspace split */}
      <div className="workspace-grid">
        {/* Left Side: Question Display */}
        <div className="workspace-left">
          <div className="glass-card question-pane">
            <div className="pane-header">
              <span className={`badge badge-${question.difficulty.toLowerCase()}`}>
                {question.difficulty}
              </span>
              <span className="pane-date">{question.date}</span>
            </div>
            
            <h1 className="question-title-workspace">{question.title}</h1>
            
            <div className="markdown-content">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {question.description}
              </ReactMarkdown>
            </div>

            {/* Accordion Hints */}
            {question.hints && question.hints.length > 0 && (
              <div className="hints-section">
                <h3 className="hints-header">Hints & Guides</h3>
                {question.hints.map((hint, idx) => (
                  <div key={idx} className="hint-accordion">
                    <button className="hint-trigger" onClick={() => toggleHint(idx)}>
                      <span>Hint #{idx + 1}</span>
                      <svg
                        className={`hint-chevron ${activeHints[idx] ? 'open' : ''}`}
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2.5"
                      >
                        <polyline points="6 9 12 15 18 9" />
                      </svg>
                    </button>
                    {activeHints[idx] && (
                      <div className="hint-panel animate-slide-up">
                        <p>{hint}</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right Side: Answer Input & Checklist */}
        <div className="workspace-right">
          {/* Answer card */}
          <div className="glass-card answer-pane">
            <h3 className="pane-title">Your Architectural Solution</h3>
            <textarea
              className="answer-textarea"
              placeholder="Structure your answer clearly. E.g.
1. Requirements & Scaling Limits
2. API Designs
3. Database & Cache Choices
4. Scaling & Availability Tradeoffs..."
              value={answer}
              onChange={(e) => setAnswer(e.target.value)}
              disabled={submitting}
            ></textarea>
            
            <div className="answer-footer">
              <span className={`char-counter ${answer.trim().length < 100 ? 'counter-danger' : 'counter-success'}`}>
                {answer.trim().length} chars (Min: 100)
              </span>
              
              <button
                className="btn btn-primary submit-workspace-btn"
                onClick={handleSubmit}
                disabled={submitting || answer.trim().length < 100}
              >
                {submitting ? (
                  <>
                    <span className="btn-spinner"></span>
                    Submitting...
                  </>
                ) : (
                  'Submit Solution'
                )}
              </button>
            </div>
            {error && <p className="submission-error">{error}</p>}
          </div>

          {/* Checklist card */}
          <div className="glass-card checklist-pane">
            <h3>Architecture Checklist</h3>
            <p className="checklist-sub text-muted">Tick off elements as you cover them in your design:</p>
            
            <div className="checklist-list">
              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.requirements}
                  onChange={() => handleCheckboxChange('requirements')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">Requirements & Scale Estimations</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.apis}
                  onChange={() => handleCheckboxChange('apis')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">API Endpoints & Protocols</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.database}
                  onChange={() => handleCheckboxChange('database')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">DB Schema & Storage Choices</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.caching}
                  onChange={() => handleCheckboxChange('caching')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">Caching Strategy & Eviction</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.scaling}
                  onChange={() => handleCheckboxChange('scaling')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">Horizontal Scaling & Load Balancing</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.availability}
                  onChange={() => handleCheckboxChange('availability')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">Availability & DR (Replication/Failover)</span>
              </label>

              <label className="checklist-item">
                <input
                  type="checkbox"
                  checked={checklist.tradeoffs}
                  onChange={() => handleCheckboxChange('tradeoffs')}
                />
                <span className="checkmark"></span>
                <span className="checklist-text">System Tradeoffs (CAP, Latency vs Consistency)</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SubmitAnswer;
