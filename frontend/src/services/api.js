const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

let currentAccessToken = null;
let onTokenRefreshedCallback = null;
let onAuthFailedCallback = null;
let isRefreshing = false;
let failedQueue = [];

export const setAccessToken = (token) => {
  currentAccessToken = token;
};

export const getAccessToken = () => currentAccessToken;

export const setAuthCallbacks = ({ onTokenRefreshed, onAuthFailed }) => {
  onTokenRefreshedCallback = onTokenRefreshed;
  onAuthFailedCallback = onAuthFailed;
};

const processQueue = (error, token = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

export async function apiRequest(endpoint, options = {}) {
  const url = `${BASE_URL}${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (currentAccessToken) {
    headers['Authorization'] = `Bearer ${currentAccessToken}`;
  }

  const fetchOptions = {
    ...options,
    headers,
    credentials: 'include', // Automatically send and receive HttpOnly cookies
  };

  try {
    const response = await fetch(url, fetchOptions);

    // 401 Unauthorized handling for token refresh
    if (response.status === 401 && !options._isRetry && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/refresh')) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then((token) => {
          options._isRetry = true;
          options.headers = {
            ...options.headers,
            Authorization: `Bearer ${token}`,
          };
          return apiRequest(endpoint, options);
        });
      }

      options._isRetry = true;
      isRefreshing = true;

      try {
        const refreshResponse = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        });

        if (!refreshResponse.ok) {
          throw new Error('Refresh token expired or invalid');
        }

        const refreshData = await refreshResponse.json();
        const newAccessToken = refreshData.token;

        setAccessToken(newAccessToken);
        if (onTokenRefreshedCallback) {
          onTokenRefreshedCallback(newAccessToken, refreshData.user);
        }

        processQueue(null, newAccessToken);
        isRefreshing = false;

        // Retry original request with new access token
        return apiRequest(endpoint, options);
      } catch (refreshErr) {
        processQueue(refreshErr, null);
        isRefreshing = false;
        setAccessToken(null);
        if (onAuthFailedCallback) {
          onAuthFailedCallback();
        }
        throw new Error('Session expired. Please log in again.');
      }
    }

    let data;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    if (!response.ok) {
      const errorMsg = data?.error?.message || data?.message || `Request failed with status ${response.status}`;
      const error = new Error(errorMsg);
      error.status = response.status;
      error.code = data?.error?.code || 'API_ERROR';
      error.data = data;
      throw error;
    }

    return data;
  } catch (err) {
    if (err.name === 'TypeError' && err.message === 'Failed to fetch') {
      const networkError = new Error('Unable to connect to the backend server. Please make sure the backend is running.');
      networkError.status = 0;
      networkError.code = 'NETWORK_ERROR';
      throw networkError;
    }
    throw err;
  }
}

export const api = {
  get: (endpoint, options) => apiRequest(endpoint, { ...options, method: 'GET' }),
  post: (endpoint, body, options) => apiRequest(endpoint, { ...options, method: 'POST', body: JSON.stringify(body) }),
  put: (endpoint, body, options) => apiRequest(endpoint, { ...options, method: 'PUT', body: JSON.stringify(body) }),
  patch: (endpoint, body, options) => apiRequest(endpoint, { ...options, method: 'PATCH', body: JSON.stringify(body) }),
  delete: (endpoint, options) => apiRequest(endpoint, { ...options, method: 'DELETE' }),
};
