import React from 'react';
import {
  Chart as ChartJS,
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend,
} from 'chart.js';
import { Radar } from 'react-chartjs-2';

ChartJS.register(
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend
);

const ScoreRadar = ({ categories }) => {
  // categories maps category name (e.g. "Requirements") -> { score: number, feedback: string }
  // or it could be an array. We can accommodate both or convert to array.
  
  const categoryNames = [
    'Requirements',
    'API Design',
    'Database',
    'Cache',
    'Scalability',
    'Availability',
    'Tradeoffs'
  ];

  const scores = categoryNames.map(name => {
    if (categories && categories[name]) {
      return categories[name].score;
    }
    // Also support alternative flat maps if categories has just category -> score
    if (categories && typeof categories[name] === 'number') {
      return categories[name];
    }
    // Check if it's the weakAreas format
    if (Array.isArray(categories)) {
      const match = categories.find(item => item.category === name);
      if (match) return match.averageScore;
    }
    return 0;
  });

  const data = {
    labels: categoryNames,
    datasets: [
      {
        label: 'Candidate Score',
        data: scores,
        backgroundColor: 'rgba(99, 102, 241, 0.25)', // semitransparent indigo
        borderColor: '#6366f1', // solid indigo
        borderWidth: 2,
        pointBackgroundColor: '#8b5cf6', // purple points
        pointBorderColor: '#fff',
        pointHoverBackgroundColor: '#fff',
        pointHoverBorderColor: '#8b5cf6',
        pointRadius: 4,
        pointHoverRadius: 6,
      },
    ],
  };

  const options = {
    scales: {
      r: {
        angleLines: {
          color: 'rgba(255, 255, 255, 0.08)',
        },
        grid: {
          color: 'rgba(255, 255, 255, 0.08)',
        },
        pointLabels: {
          color: '#94a3b8',
          font: {
            family: 'Inter, sans-serif',
            size: 11,
            weight: '600'
          }
        },
        ticks: {
          backdropColor: 'transparent',
          color: '#64748b',
          font: {
            size: 9
          },
          stepSize: 20
        },
        min: 0,
        max: 100,
      },
    },
    plugins: {
      legend: {
        display: false // hide legend since there is only 1 dataset
      },
      tooltip: {
        backgroundColor: 'rgba(10, 14, 26, 0.9)',
        titleColor: '#fff',
        bodyColor: '#f8fafc',
        borderColor: 'rgba(255, 255, 255, 0.1)',
        borderWidth: 1,
        titleFont: {
          family: 'Inter, sans-serif',
          weight: 'bold'
        },
        bodyFont: {
          family: 'Inter, sans-serif'
        },
        padding: 10,
        cornerRadius: 8
      }
    },
    maintainAspectRatio: false
  };

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      <Radar data={data} options={options} />
    </div>
  );
};

export default ScoreRadar;
