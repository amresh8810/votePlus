import { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';
import { Input } from '../components/Input';
import { Button } from '../components/Button';
import { ErrorMessage } from '../components/ErrorMessage';
import { KeyRound, ArrowLeft, CheckCircle2 } from 'lucide-react';

export const ForgotPassword = () => {
  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [devResetToken, setDevResetToken] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setDevResetToken('');

    if (!email.trim()) {
      setError('Please enter your email address.');
      return;
    }

    setLoading(true);
    try {
      const resp = await api.post('/auth/forgot-password', { email: email.trim() });
      if (resp?.dev_reset_token) {
        setDevResetToken(resp.dev_reset_token);
      }
      setSubmitted(true);
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.');
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
          <h1 className="auth-title">Forgot Password</h1>
          <p className="auth-subtitle">Enter your registered email to generate a password reset token</p>
        </div>

        <ErrorMessage message={error} onClose={() => setError('')} />

        {submitted ? (
          <div className="success-banner">
            <CheckCircle2 size={24} className="text-green-400 mb-2" />
            <p className="font-semibold text-sm mb-2">
              {devResetToken ? 'Password reset token generated.' : 'Reset instructions are available.'}
            </p>
            <p className="text-xs text-slate-300">
              {devResetToken
                ? 'Use the development reset token below to continue.'
                : 'If an account exists for this email, reset instructions are available.'}
            </p>
            {devResetToken && (
              <div className="dev-info-banner mt-4">
                <p className="font-semibold text-xs text-amber-400">[DEV ONLY] Reset Token:</p>
                <code className="text-xs break-all select-all">{devResetToken}</code>
                <div className="mt-2">
                  <Link to={`/reset-password?token=${devResetToken}`} className="text-link text-xs">
                    Go to Reset Password page with token →
                  </Link>
                </div>
              </div>
            )}
            <Link to="/login" className="btn btn-outline btn-md w-full mt-4">
              Return to Login
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="auth-form">
            <Input
              label="Email Address"
              id="email"
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
            />

            <Button type="submit" variant="primary" size="lg" loading={loading} className="w-full">
              Generate Reset Token
            </Button>
          </form>
        )}

        <div className="auth-footer">
          <Link to="/login" className="text-link text-sm inline-flex items-center gap-1">
            <ArrowLeft size={14} />
            <span>Back to Login</span>
          </Link>
        </div>
      </div>
    </div>
  );
};
