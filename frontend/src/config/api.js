const configuredBaseUrl = import.meta.env.VITE_API_BASE_URL
  || import.meta.env.VITE_API_URL
  || 'http://localhost:8080';

export const API_BASE_URL = configuredBaseUrl
  .replace(/\/+$/, '')
  .replace(/\/api\/v1$/, '');

export const API_ROOT = `${API_BASE_URL}/api/v1`;

export const getWebSocketUrl = (pollId) => {
  const websocketBaseUrl = API_BASE_URL.replace(/^http/, 'ws');
  return `${websocketBaseUrl}/api/v1/public/polls/${pollId}/ws`;
};
