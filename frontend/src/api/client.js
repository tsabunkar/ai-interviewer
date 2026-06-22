const API_BASE = import.meta.env.VITE_API_URL || '';

export const fetchTodayQuestion = async () => {
  const response = await fetch(`${API_BASE}/question`);
  if (!response.ok) {
    if (response.status === 404) {
      throw new Error('No question generated for today yet.');
    }
    throw new Error('Failed to fetch today\'s question.');
  }
  return response.json();
};

export const submitAnswer = async (userId, questionDate, answer) => {
  const response = await fetch(`${API_BASE}/answer`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ userId, questionDate, answer })
  });

  if (!response.ok) {
    const errObj = await response.json().catch(() => ({}));
    throw new Error(errObj.error || 'Failed to submit answer.');
  }

  return response.json();
};

export const getResults = async (submissionId) => {
  const response = await fetch(`${API_BASE}/results/${submissionId}`);
  if (!response.ok) {
    throw new Error('Failed to fetch evaluation results.');
  }
  return response.json();
};

export const getLeaderboard = async (date) => {
  const response = await fetch(`${API_BASE}/leaderboard?date=${encodeURIComponent(date)}`);
  if (!response.ok) {
    throw new Error('Failed to fetch leaderboard data.');
  }
  return response.json();
};

export const getUserStats = async (userId) => {
  const response = await fetch(`${API_BASE}/stats/${userId}`);
  if (!response.ok) {
    throw new Error('Failed to fetch user statistics.');
  }
  return response.json();
};
