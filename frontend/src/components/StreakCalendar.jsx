import React from 'react';
import './StreakCalendar.css';

const StreakCalendar = ({ submissions }) => {
  // submissions is an array of { submissionId, questionDate, score, status, submittedAt }
  
  // Create a map of date -> score
  const submissionMap = {};
  if (Array.isArray(submissions)) {
    submissions.forEach(sub => {
      // Map it if it's evaluated
      if (sub.status === 'evaluated') {
        submissionMap[sub.questionDate] = sub.score;
      }
    });
  }

  // Generate last 90 days
  const cells = [];
  const today = new Date();
  
  for (let i = 89; i >= 0; i--) {
    const day = new Date();
    day.setDate(today.getDate() - i);
    
    // Format date string as YYYY-MM-DD in local time
    const year = day.getFullYear();
    const month = String(day.getMonth() + 1).padStart(2, '0');
    const date = String(day.getDate()).padStart(2, '0');
    const dateStr = `${year}-${month}-${date}`;
    
    const score = submissionMap[dateStr];
    
    let colorClass = 'cell-empty';
    if (score !== undefined) {
      if (score < 70) {
        colorClass = 'cell-low';
      } else if (score < 85) {
        colorClass = 'cell-medium';
      } else {
        colorClass = 'cell-high';
      }
    }

    cells.push({
      date: dateStr,
      score,
      colorClass
    });
  }

  // Helper for displaying tooltips
  const getTooltipText = (cell) => {
    if (cell.score === undefined) {
      return `No submission on ${cell.date}`;
    }
    return `Score: ${cell.score}/100 on ${cell.date}`;
  };

  return (
    <div className="streak-calendar-wrapper">
      <h3 className="streak-title">Submission Activity (Last 90 Days)</h3>
      <div className="streak-grid">
        {cells.map((cell, idx) => (
          <div
            key={idx}
            className={`streak-cell ${cell.colorClass}`}
            data-date={cell.date}
          >
            <span className="streak-tooltip">{getTooltipText(cell)}</span>
          </div>
        ))}
      </div>
      <div className="streak-legend">
        <span>Less</span>
        <div className="streak-cell cell-empty"></div>
        <div className="streak-cell cell-low"></div>
        <div className="streak-cell cell-medium"></div>
        <div className="streak-cell cell-high"></div>
        <span>More</span>
      </div>
    </div>
  );
};

export default StreakCalendar;
