import React, { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import './Navbar.css';

const Navbar = () => {
  const [userId, setUserId] = useState('guest_user');

  useEffect(() => {
    const storedUser = localStorage.getItem('userId');
    if (storedUser) {
      setUserId(storedUser);
    }

    // Set up a listener for storage events (e.g. if username changes on Submit page)
    const handleStorageChange = () => {
      const updatedUser = localStorage.getItem('userId');
      if (updatedUser) {
        setUserId(updatedUser);
      }
    };

    window.addEventListener('storage', handleStorageChange);
    // Custom event dispatch for local updates in the same window
    window.addEventListener('userChanged', handleStorageChange);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      window.removeEventListener('userChanged', handleStorageChange);
    };
  }, []);

  return (
    <nav className="navbar">
      <div className="navbar-container container">
        <NavLink to="/" className="navbar-brand">
          <svg className="navbar-logo" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
          </svg>
          <span className="gradient-text brand-name">Antigravity</span>
          <span className="brand-suffix">Interviewer</span>
        </NavLink>

        <div className="navbar-links">
          <NavLink to="/" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'} end>
            Home
          </NavLink>
          <NavLink to="/dashboard" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
            Dashboard
          </NavLink>
          <NavLink to="/leaderboard" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
            Leaderboard
          </NavLink>
        </div>

        <div className="navbar-user">
          <div className="user-avatar">
            {userId.substring(0, 2).toUpperCase()}
          </div>
          <span className="user-name" title={userId}>{userId}</span>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
