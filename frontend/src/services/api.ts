const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token');
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(error.message || 'Request failed');
  }
  return res.json();
}

export const api = {
  register: (username: string, email: string, password: string) =>
    request<{ token: string }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    }),

  login: (email: string, password: string) =>
    request<{ token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  getProfile: () =>
    request<{ id: string; username: string; elo: number; wins: number; losses: number }>('/user/profile'),

  getLeaderboard: () =>
    request<Array<{ username: string; elo: number; rank: number }>>('/leaderboard'),
};
