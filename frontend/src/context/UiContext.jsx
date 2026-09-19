import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';

const UiContext = createContext(null);

export const UiProvider = ({ children }) => {
  const [theme, setTheme] = useState(() => localStorage.getItem('votepulse-theme') || 'light');
  const [toasts, setToasts] = useState([]);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem('votepulse-theme', theme);
  }, [theme]);

  const toast = useCallback((message, tone = 'success') => {
    const id = `${Date.now()}-${Math.random()}`;
    setToasts((current) => [...current, { id, message, tone }]);
    window.setTimeout(() => setToasts((current) => current.filter((item) => item.id !== id)), 4200);
  }, []);

  const value = useMemo(() => ({
    theme,
    toggleTheme: () => setTheme((current) => current === 'dark' ? 'light' : 'dark'),
    toast,
  }), [theme, toast]);

  return (
    <UiContext.Provider value={value}>
      {children}
      <div className="toast-stack" aria-live="polite">
        {toasts.map((item) => <div className={`app-toast app-toast-${item.tone}`} key={item.id}>{item.message}</div>)}
      </div>
    </UiContext.Provider>
  );
};

export const useUi = () => {
  const context = useContext(UiContext);
  if (!context) throw new Error('useUi must be used within UiProvider');
  return context;
};
