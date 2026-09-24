import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { PollResults } from '../components/PollResults';
import { Button } from '../components/Button';
import { Loading } from '../components/Loading';
import { ErrorMessage } from '../components/ErrorMessage';
import { Vote, Radio, CheckCircle2, AlertCircle, XCircle } from 'lucide-react';

export const PublicPoll = () => {
  const { id } = useParams();
  const [poll, setPoll] = useState(null);
  const [counts, setCounts] = useState({});
  const [totalVotes, setTotalVotes] = useState(0);
  const [selectedOptionId, setSelectedOptionId] = useState('');
  const [hasVoted, setHasVoted] = useState(false);
  const [votedOptionId, setVotedOptionId] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const fetchPublicPoll = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.get(`/public/polls/${id}`);
      const pollData = data?.poll;
      setPoll(pollData);

      if (pollData?.voter_has_voted) {
        setHasVoted(true);
        setVotedOptionId(pollData.voter_option_id || '');
      }

      if (pollData?.options) {
        const initialCounts = {};
        let total = 0;
        pollData.options.forEach((opt) => {
          initialCounts[opt.id] = opt.vote_count || 0;
          total += opt.vote_count || 0;
        });
        setCounts(initialCounts);
        setTotalVotes(total);
      }
    } catch (err) {
      if (err.status === 404) {
        setError('Poll not found. Please check the link and try again.');
      } else {
        setError(err.message || 'Failed to load poll.');
      }
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchPublicPoll();
  }, [fetchPublicPoll]);

  useEffect(() => {
    if (!poll) return undefined;
    const previousTitle = document.title;
    const description = document.querySelector('meta[name="description"]');
    const previousDescription = description?.getAttribute('content');
    const ogTitle = document.querySelector('meta[property="og:title"]');
    const ogDescription = document.querySelector('meta[property="og:description"]');
    const previousOgTitle = ogTitle?.getAttribute('content');
    const previousOgDescription = ogDescription?.getAttribute('content');
    document.title = `${poll.question} | VotePulse`;
    if (description) description.setAttribute('content', `Vote in ${poll.question} on VotePulse.`);
    if (ogTitle) ogTitle.setAttribute('content', poll.question);
    if (ogDescription) ogDescription.setAttribute('content', `Vote in ${poll.question} on VotePulse.`);
    return () => {
      document.title = previousTitle;
      if (description && previousDescription != null) description.setAttribute('content', previousDescription);
      if (ogTitle && previousOgTitle != null) ogTitle.setAttribute('content', previousOgTitle);
      if (ogDescription && previousOgDescription != null) ogDescription.setAttribute('content', previousOgDescription);
    };
  }, [poll]);

  // WebSocket real-time update handler
  const handleWsUpdate = useCallback((update) => {
    if (update && update.counts) {
      setCounts(update.counts);
      setTotalVotes(update.totalVotes || 0);
    }
  }, []);

  const { status: wsStatus } = useWebSocket(id, handleWsUpdate);

  const handleVoteSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (!selectedOptionId) {
      setError('Please select an option to vote.');
      return;
    }

    setSubmitting(true);
    try {
      const resp = await api.post(`/public/polls/${id}/vote`, {
        option_id: selectedOptionId,
      });

      setHasVoted(true);
      setVotedOptionId(selectedOptionId);
      setSuccessMessage('Thank you! Your vote has been recorded.');

      if (resp?.counts) {
        setCounts(resp.counts);
        setTotalVotes(resp.totalVotes || 0);
      }
    } catch (err) {
      if (err.status === 409) {
        setHasVoted(true);
        setError('You have already voted in this poll.');
      } else if (err.status === 401) {
        setError('Please log in to vote. Each account can vote once per poll.');
      } else {
        setError(err.message || 'Failed to submit vote. Please try again.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <Loading message="Loading live poll..." fullPage />;

  const isActive = poll?.status === 'active';

  return (
    <div className="public-poll-page page-layout">
      <div className="public-poll-card">
        {/* Header Badges */}
        <div className="public-poll-header">
          <span className={`status-badge ${isActive ? 'status-active' : 'status-closed'}`}>
            {isActive ? <CheckCircle2 size={14} /> : <XCircle size={14} />}
            <span>{isActive ? 'Active Poll' : 'Poll Closed'}</span>
          </span>

          <span className={`ws-status-badge ws-${wsStatus}`}>
            <Radio size={12} className={wsStatus === 'connected' ? 'animate-pulse' : ''} />
            <span>{wsStatus === 'connected' ? 'Live' : 'Reconnecting...'}</span>
          </span>
        </div>

        {/* Question Title */}
        <h1 className="public-poll-title">{poll?.question}</h1>

        <ErrorMessage message={error} onClose={() => setError('')} />

        {successMessage && (
          <div className="success-banner mb-6">
            <CheckCircle2 size={20} className="text-green-400" />
            <span>{successMessage}</span>
          </div>
        )}

        {/* Voting Options Section */}
        {isActive && !hasVoted ? (
          <form onSubmit={handleVoteSubmit} className="voting-form">
            <div className="voting-options-group" role="radiogroup" aria-label="Poll choices">
              {poll?.options?.map((opt) => (
                <label
                  key={opt.id}
                  className={`option-choice-card ${selectedOptionId === opt.id ? 'option-choice-selected' : ''}`}
                >
                  <input
                    type="radio"
                    name="poll-option"
                    value={opt.id}
                    checked={selectedOptionId === opt.id}
                    onChange={() => setSelectedOptionId(opt.id)}
                    className="option-radio-input"
                  />
                  <span className="option-choice-text">{opt.text}</span>
                </label>
              ))}
            </div>

            <Button
              type="submit"
              variant="primary"
              size="lg"
              loading={submitting}
              disabled={!selectedOptionId}
              className="w-full mt-6"
            >
              <Vote size={20} />
              <span>Submit Vote</span>
            </Button>
          </form>
        ) : (
          <div className="voted-state-info mb-6">
            {!isActive ? (
              <div className="closed-banner">
                <AlertCircle size={18} />
                <span>This poll is closed and no longer accepting votes.</span>
              </div>
            ) : (
              <div className="already-voted-banner">
                <CheckCircle2 size={18} className="text-green-400" />
                <span>You have already voted in this poll.</span>
              </div>
            )}
          </div>
        )}

        {/* Results Section */}
        <div className="mt-8 pt-6 border-t border-slate-700/60">
          <PollResults poll={poll} counts={counts} totalVotes={totalVotes} />
        </div>
      </div>
    </div>
  );
};
