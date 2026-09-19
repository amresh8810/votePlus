import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../services/api';
import { PollForm } from '../components/PollForm';
import { ErrorMessage } from '../components/ErrorMessage';
import { ArrowLeft, PlusCircle } from 'lucide-react';

export const CreatePoll = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleCreatePoll = async (formData) => {
    setLoading(true);
    setError('');

    try {
      const response = await api.post('/polls', formData);
      const poll = response?.poll;
      if (poll?.id) {
        navigate(`/polls/${poll.id}`);
      } else {
        navigate('/dashboard');
      }
    } catch (err) {
      setError(err.message || 'Failed to create poll. Please check your inputs.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-poll-container page-layout">
      <div className="mb-6">
        <Link to="/dashboard" className="text-link text-sm inline-flex items-center gap-1 mb-2">
          <ArrowLeft size={16} />
          <span>Back to Dashboard</span>
        </Link>
        <h1 className="page-title flex items-center gap-2">
          <PlusCircle size={28} className="text-indigo-400" />
          <span>Create New Live Poll</span>
        </h1>
        <p className="page-subtitle">Ask a question and define 2 to 6 choices for your audience</p>
      </div>

      <ErrorMessage message={error} onClose={() => setError('')} />

      <div className="form-card">
        <PollForm onSubmit={handleCreatePoll} loading={loading} buttonText="Publish & Create Poll" />
      </div>
    </div>
  );
};
