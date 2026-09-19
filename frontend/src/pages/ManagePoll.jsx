import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { PollResults } from '../components/PollResults';
import { PollExportActions } from '../components/PollExportActions';
import { Loading } from '../components/Loading';
import { ErrorMessage } from '../components/ErrorMessage';
import { ArrowLeft, Copy, Check, ExternalLink, XCircle, CheckCircle2, Radio } from 'lucide-react';

export const ManagePoll = () => {
  const { id } = useParams();
  const [poll, setPoll] = useState(null);
  const [counts, setCounts] = useState({});
  const [totalVotes, setTotalVotes] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  const fetchPoll = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.get(`/polls/${id}`);
      const pollData = data?.poll;
      setPoll(pollData);

      // Initialize vote counts from returned poll object
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
      setError(err.message || 'Failed to load poll details.');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchPoll();
  }, [fetchPoll]);

  // WebSocket real-time updates handler
  const handleWsUpdate = useCallback((update) => {
    if (update && update.counts) {
      setCounts(update.counts);
      setTotalVotes(update.totalVotes || 0);
    }
  }, []);

  const { status: wsStatus } = useWebSocket(id, handleWsUpdate);

  const handleClosePoll = async () => {
    if (!window.confirm('Are you sure you want to close this poll? Users will no longer be able to vote.')) {
      return;
    }

    try {
      await api.patch(`/polls/${id}/close`);
      setPoll((prev) => (prev ? { ...prev, status: 'closed' } : prev));
    } catch (err) {
      alert(err.message || 'Failed to close poll.');
    }
  };

  const publicUrl = `${window.location.origin}/p/${id}`;

  const handleCopyLink = () => {
    navigator.clipboard.writeText(publicUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (loading) return <Loading message="Loading poll management data..." fullPage />;

  return (
    <div className="manage-poll-container page-layout">
      <div className="mb-6">
        <Link to="/dashboard" className="text-link text-sm inline-flex items-center gap-1 mb-2">
          <ArrowLeft size={16} />
          <span>Back to Dashboard</span>
        </Link>

        <div className="manage-header-row">
          <div>
            <div className="poll-status-row mb-2">
              <span className={`status-badge ${poll?.is_active ? 'status-active' : 'status-closed'}`}>
                {poll?.is_active ? (
                  <>
                    <CheckCircle2 size={14} /> Active Poll
                  </>
                ) : (
                  <>
                    <XCircle size={14} /> Closed Poll
                  </>
                )}
              </span>

              <span className={`ws-status-badge ws-${wsStatus}`}>
                <Radio size={12} className={wsStatus === 'connected' ? 'animate-pulse' : ''} />
                <span>{wsStatus === 'connected' ? 'Live' : 'Reconnecting...'}</span>
              </span>
            </div>

            <h1 className="page-title">{poll?.question}</h1>
          </div>

          <div className="manage-actions-row">
            <button type="button" onClick={handleCopyLink} className="btn btn-outline btn-md">
              {copied ? <Check size={16} className="text-green-400" /> : <Copy size={16} />}
              <span>{copied ? 'Link Copied!' : 'Copy Share Link'}</span>
            </button>

            <a href={`/p/${id}`} target="_blank" rel="noopener noreferrer" className="btn btn-secondary btn-md">
              <ExternalLink size={16} />
              <span>Open Public View</span>
            </a>
            {poll && <PollExportActions poll={poll} counts={counts} totalVotes={totalVotes} />}

            {poll?.is_active && (
              <button type="button" onClick={handleClosePoll} className="btn btn-danger btn-md">
                <XCircle size={16} />
                <span>Close Poll</span>
              </button>
            )}
          </div>
        </div>
      </div>

      <ErrorMessage message={error} onClose={() => setError('')} />

      <div className="manage-grid">
        <div className="manage-main-card">
          <PollResults poll={poll} counts={counts} totalVotes={totalVotes} />
        </div>

        <div className="manage-sidebar-card">
          <h4 className="sidebar-title">Poll Info</h4>
          <div className="info-list">
            <div className="info-item">
              <span className="info-label">Public Share Link</span>
              <div className="share-link-box">
                <input type="text" readOnly value={publicUrl} className="share-link-input" />
                <button type="button" onClick={handleCopyLink} className="btn-copy-icon">
                  {copied ? <Check size={14} className="text-green-400" /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            <div className="info-item">
              <span className="info-label">Created At</span>
              <span className="info-value">
                {poll?.created_at ? new Date(poll.created_at).toLocaleString() : 'N/A'}
              </span>
            </div>

            <div className="info-item">
              <span className="info-label">Total Choices</span>
              <span className="info-value">{poll?.options?.length || 0} Options</span>
            </div>
            <div className="info-item">
              <span className="info-label">Public views</span>
              <span className="info-value">{(poll?.view_count || 0).toLocaleString()}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
