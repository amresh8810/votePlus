import { useState, useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { api } from '../services/api';
import { CheckCircle2, AlertTriangle, Mail } from 'lucide-react';
import { Loading } from '../components/Loading';

export const VerifyEmail = () => {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState('verifying'); // 'verifying' | 'success' | 'error'
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setStatus('error');
      setErrorMessage('Verification token missing from URL.');
      return;
    }

    const verify = async () => {
      try {
        await api.post('/auth/verify-email', { token });
        setStatus('success');
      } catch (err) {
        setStatus('error');
        setErrorMessage(err.message || 'Invalid or expired verification token.');
      }
    };

    verify();
  }, [searchParams]);

  return (
    <div className="auth-page">
      <div className="auth-card text-center py-6">
        <div className="auth-brand-icon mx-auto mb-4">
          <Mail size={36} className="text-indigo-400" />
        </div>

        {status === 'verifying' && <Loading message="Verifying your email address..." />}

        {status === 'success' && (
          <div className="success-banner">
            <CheckCircle2 size={48} className="text-green-400 mx-auto mb-3" />
            <h2 className="text-xl font-bold mb-2">Email Verified!</h2>
            <p className="text-sm text-slate-300 mb-6">
              Your email address has been successfully verified.
            </p>
            <Link to="/dashboard" className="btn btn-primary btn-md w-full">
              Go to Dashboard
            </Link>
          </div>
        )}

        {status === 'error' && (
          <div className="error-card">
            <AlertTriangle size={48} className="text-red-400 mx-auto mb-3" />
            <h2 className="text-xl font-bold mb-2">Verification Failed</h2>
            <p className="text-sm text-red-300 mb-6">{errorMessage}</p>
            <Link to="/login" className="btn btn-secondary btn-md w-full">
              Return to Login
            </Link>
          </div>
        )}
      </div>
    </div>
  );
};
