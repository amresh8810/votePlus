import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Input } from '../components/Input';
import { Button } from '../components/Button';
import { ErrorMessage } from '../components/ErrorMessage';
import { Vote, UserPlus } from 'lucide-react';

export const Signup = () => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [devVerificationToken, setDevVerificationToken] = useState('');
  const [loading, setLoading] = useState(false);

  const { signup, login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setDevVerificationToken('');

    if (!name.trim()) {
      setError('Name is required.');
      return;
    }
    if (!email.trim()) {
      setError('Email address is required.');
      return;
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters long.');
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      const resp = await signup(name.trim(), email.trim(), password);
      
      // If dev verification token is returned, store it for easy testing
      if (resp?.dev_verification_token) {
        setDevVerificationToken(resp.dev_verification_token);
      }

      // Auto login after successful signup
      await login(email.trim(), password);
      navigate('/dashboard');
    } catch (err) {
      setError(err.message || 'Failed to create account. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-header">
          <div className="auth-brand-icon">
            <Vote size={32} className="text-indigo-400" />
          </div>
          <h1 className="auth-title">Create Account</h1>
          <p className="auth-subtitle">Start creating live real-time polls in seconds</p>
        </div>

        <ErrorMessage message={error} onClose={() => setError('')} />

        {devVerificationToken && (
          <div className="dev-info-banner">
            <p className="font-semibold text-xs text-amber-400">[DEV ONLY] Email Verification Token:</p>
            <code className="text-xs break-all select-all">{devVerificationToken}</code>
          </div>
        )}

        <form onSubmit={handleSubmit} className="auth-form">
          <Input
            label="Full Name"
            id="name"
            placeholder="Jane Doe"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoComplete="name"
          />

          <Input
            label="Email Address"
            id="email"
            type="email"
            placeholder="jane@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />

          <Input
            label="Password"
            id="password"
            type="password"
            placeholder="At least 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="new-password"
            helperText="Minimum 8 characters"
          />

          <Input
            label="Confirm Password"
            id="confirmPassword"
            type="password"
            placeholder="Repeat password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            autoComplete="new-password"
          />

          <Button type="submit" variant="primary" size="lg" loading={loading} className="w-full">
            <UserPlus size={18} />
            <span>Create Account</span>
          </Button>
        </form>

        <div className="auth-footer">
          <p>
            Already have an account?{' '}
            <Link to="/login" className="text-link font-semibold">
              Log In
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};
