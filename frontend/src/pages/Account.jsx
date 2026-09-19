import { useAuth } from '../context/AuthContext';
import { useUi } from '../context/UiContext';
import { ShieldCheck, Moon, Sun, History } from 'lucide-react';

export const Account = () => {
  const { user } = useAuth();
  const { theme, toggleTheme } = useUi();
  const history = JSON.parse(localStorage.getItem('votepulse-activity') || '[]');

  return (
    <div className="page-layout account-page">
      <header className="page-section-heading">
        <p className="analytics-eyebrow">Workspace</p>
        <h1 className="page-title">Account settings</h1>
        <p className="page-subtitle">Manage your profile and dashboard preferences.</p>
      </header>
      <div className="account-grid">
        <section className="account-card">
          <div className="account-avatar">{(user?.name || user?.email || 'U').slice(0, 1).toUpperCase()}</div>
          <h2>{user?.name || 'Creator'}</h2>
          <p>{user?.email}</p>
          <span className="account-status"><ShieldCheck size={15} /> Session active</span>
        </section>
        <section className="account-card">
          <h2>Appearance</h2>
          <button type="button" className="account-setting-button" onClick={toggleTheme}>
            {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
            <span>{theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}</span>
          </button>
        </section>
        <section className="account-card account-history-card">
          <h2><History size={18} /> Recent activity</h2>
          {history.length ? history.slice(0, 8).map((item) => <p key={item.id}>{item.message}<small>{new Date(item.at).toLocaleString()}</small></p>) : <p>No recent activity recorded on this device.</p>}
        </section>
      </div>
    </div>
  );
};
