import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api, setAccessToken, setAuthCallbacks } from '../services/api';

const AuthContext = createContext(null);
let sessionRestorePromise = null;

const restoreSession = () => {
  if (!sessionRestorePromise) {
    sessionRestorePromise = api.post('/auth/refresh').finally(() => {
      sessionRestorePromise = null;
    });
  }
  return sessionRestorePromise;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [accessTokenState, setAccessTokenState] = useState(null);
  const [loading, setLoading] = useState(true);

  const handleUpdateToken = useCallback((newToken, userData = null) => {
    setAccessTokenState(newToken);
    setAccessToken(newToken);
    if (userData) {
      setUser(userData);
    }
  }, []);

  const handleAuthFailed = useCallback(() => {
    setAccessTokenState(null);
    setAccessToken(null);
    setUser(null);
  }, []);

  // Set up API interceptor callbacks
  useEffect(() => {
    setAuthCallbacks({
      onTokenRefreshed: (token, userData) => {
        handleUpdateToken(token, userData);
      },
      onAuthFailed: () => {
        handleAuthFailed();
      },
    });
  }, [handleUpdateToken, handleAuthFailed]);

  // Restore authentication state on application load via HttpOnly refresh cookie
  useEffect(() => {
    let isMounted = true;
    const restoreAuth = async () => {
      try {
        const refreshData = await restoreSession();
        if (isMounted && refreshData?.token) {
          handleUpdateToken(refreshData.token, refreshData.user);
        }
      } catch {
        // Silent catch: user is simply unauthenticated on cold start
        if (isMounted) {
          handleAuthFailed();
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    restoreAuth();
    return () => {
      isMounted = false;
    };
  }, [handleUpdateToken, handleAuthFailed]);

  const login = async (email, password) => {
    const data = await api.post('/auth/login', { email, password });
    if (data?.token) {
      handleUpdateToken(data.token, data.user);
    }
    return data;
  };

  const signup = async (name, email, password) => {
    return await api.post('/auth/signup', { name, email, password });
  };

  const logout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (err) {
      console.warn('Logout request failed:', err);
    } finally {
      handleAuthFailed();
    }
  };

  const value = {
    user,
    accessToken: accessTokenState,
    loading,
    isAuthenticated: !!user,
    login,
    signup,
    logout,
    setUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
