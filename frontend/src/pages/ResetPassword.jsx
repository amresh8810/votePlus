import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { api } from '../services/api';
import { Input } from '../components/Input';
import { Button } from '../components/Button';
import { ErrorMessage } from '../components/ErrorMessage';
import { KeyRound, CheckCircle2 } from 'lucide-react';

export const ResetPassword = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [token, setToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const tokenFromUrl = searchParams.get('token');
    if (tokenFromUrl) {
      setToken(tokenFromUrl);
    }
  }, [searchParams]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (!token.trim()) {
      setError('Reset token is required.');
      return;
    }
    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters long.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      await api.post('/auth/reset-password', {
        token: token.trim(),
        new_password: newPassword,
      });
      setSuccess(true);
      setTimeout(() => {
        navigate('/login');
      }, 3000);
    } catch (err) {
      setError(err.message || 'Invalid, expired, or already used reset token.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-header">
          <div className="auth-brand-icon">
            <KeyRound size={32} className="text-indigo-400" />
          </div>
          <h1 className="auth-title">Reset Password</h1>
          <p className="auth-subtitle">Set a new secure password for your account</p>
        </div>

        <ErrorMessage message={error} onClose={() => setError('')} />

        {success ? (
          <div className="success-banner text-center py-4">
            <CheckCircle2 size={40} className="text-green-400 mx-auto mb-3" />
            <h3 className="font-semibold text-lg mb-1">Password Reset Successful</h3>
            <p className="text-sm text-slate-300 mb-4">
              All active sessions have been revoked. Redirecting to login page...
            </p>
            <Link to="/login" className="btn btn-primary btn-md w-full">
              Go to Login Now
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="auth-form">
            <Input
              label="Reset Token"
              id="token"
              placeholder="Paste token from email/console"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              required
            />

            <Input
              label="New Password"
              id="newPassword"
              type="password"
              placeholder="At least 8 characters"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              helperText="Minimum 8 characters"
            />

            <Input
              label="Confirm New Password"
              id="confirmPassword"
              type="password"
              placeholder="Repeat new password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />

            <Button type="submit" variant="primary" size="lg" loading={loading} className="w-full">
              Reset Password
            </Button>
          </form>
        )}
      </div>
    </div>
  );
};
