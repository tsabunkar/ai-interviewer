import React from 'react';
import './Footer.css';

const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer-container container">
        <p className="footer-text">
          Built with <span className="footer-heart">❤️</span> for System Design Mastery.
        </p>
        <p className="footer-copyright">
          &copy; {new Date().getFullYear()} Antigravity. All rights reserved.
        </p>
      </div>
    </footer>
  );
};

export default Footer;
