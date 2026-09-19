import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Vote, PlusCircle, LayoutDashboard, LogOut, LogIn, UserPlus, ChevronDown, ShieldCheck, Settings } from 'lucide-react';

export const Navbar = () => {
  const { user, isAuthenticated, loading, logout } = useAuth();
  const navigate = useNavigate();
  const [profileOpen, setProfileOpen] = useState(false);
  const profileRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (event) => {
      if (!profileRef.current?.contains(event.target)) setProfileOpen(false);
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, []);

  const handleLogout = async () => {
    setProfileOpen(false);
    await logout();
    navigate('/login');
  };

  const displayName = user?.name || 'Creator';
  const initials = displayName.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0].toUpperCase()).join('') || 'U';

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <Link to="/" className="navbar-brand">
          <div className="brand-icon">
            <Vote size={22} className="text-indigo-400" />
          </div>
          <span className="brand-name">
            Vote<span className="brand-accent">Pulse</span>
          </span>
        </Link>

        <div className="navbar-menu">
          {loading ? (
            <div className="navbar-profile-skeleton" aria-label="Loading profile" />
          ) : isAuthenticated ? (
            <>
              <Link to="/dashboard" className="nav-link" aria-label="Dashboard">
                <LayoutDashboard size={18} />
                <span>Dashboard</span>
              </Link>
              <Link to="/polls/create" className="nav-link nav-link-primary" aria-label="Create Poll">
                <PlusCircle size={18} />
                <span>Create Poll</span>
              </Link>
              <div className={`user-profile-menu ${profileOpen ? 'user-profile-menu-open' : ''}`} ref={profileRef}>
                <button
                  type="button"
                  className="user-badge"
                  onClick={() => setProfileOpen((open) => !open)}
                  aria-expanded={profileOpen}
                  aria-haspopup="menu"
                  aria-label="Open profile menu"
                >
                  <span className="user-avatar" aria-hidden="true">{initials}</span>
                  <span className="user-name">{displayName}</span>
                  <ChevronDown size={15} aria-hidden="true" />
                </button>
                {profileOpen && (
                  <div className="profile-dropdown" role="menu">
                    <div className="profile-dropdown-header">
                      <span className="profile-avatar-large" aria-hidden="true">{initials}</span>
                      <div>
                        <strong>{displayName}</strong>
                        <span>{user?.email}</span>
                      </div>
                    </div>
                    <div className="profile-dropdown-divider" />
                    <div className="profile-account-status">
                      <ShieldCheck size={15} aria-hidden="true" />
                      <span>Session</span>
                      <strong>Active</strong>
                    </div>
                    {typeof user?.email_verified === 'boolean' && (
                      <div className="profile-account-status">
                        <ShieldCheck size={15} aria-hidden="true" />
                        <span>Email</span>
                        <strong>{user.email_verified ? 'Verified' : 'Unverified'}</strong>
                      </div>
                    )}
                    <Link to="/account" className="profile-dropdown-action" role="menuitem" onClick={() => setProfileOpen(false)}>
                      <Settings size={15} />
                      <span>Account settings</span>
                    </Link>
                    <button type="button" onClick={handleLogout} className="profile-dropdown-action profile-dropdown-logout" role="menuitem">
                      <LogOut size={15} />
                      <span>Logout</span>
                    </button>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="auth-buttons">
              <Link to="/login" className="nav-link">
                <LogIn size={18} />
                <span>Log In</span>
              </Link>
              <Link to="/signup" className="btn-signup">
                <UserPlus size={18} />
                <span>Sign Up</span>
              </Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};
