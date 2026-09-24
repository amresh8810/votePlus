import { useState } from 'react';
import { Link } from 'react-router-dom';
import { CheckCircle2, XCircle, Copy, Check, ExternalLink, Settings, Trash2, BarChart3 } from 'lucide-react';

export const PollCard = ({ poll, votes = 0, canManage = false, onClosePoll, onDeletePoll, onViewResults }) => {
  const [copied, setCopied] = useState(false);

  const publicUrl = `${window.location.origin}/p/${poll.id}`;

  const handleCopyLink = (e) => {
    e.preventDefault();
    e.stopPropagation();
    navigator.clipboard.writeText(publicUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const isActive = poll.status === 'active';

  return (
    <div className={`poll-card ${!isActive ? 'poll-card-closed' : ''}`}>
      <div className="poll-card-header">
        <span className={`status-badge ${isActive ? 'status-active' : 'status-closed'}`}>
          {isActive ? (
            <>
              <CheckCircle2 size={14} /> Active
            </>
          ) : (
            <>
              <XCircle size={14} /> Closed
            </>
          )}
        </span>
        <span className="poll-date">{formatDate(poll.created_at)}</span>
      </div>

      <h3 className="poll-question">{poll.question}</h3>

      <div className="poll-meta">
        <span>{votes.toLocaleString()} Votes</span>
        <span className="poll-options-count">{poll.options?.length || 0} Options</span>
      </div>

      <div className="poll-card-actions">
        {canManage && (
          <Link to={`/polls/${poll.id}`} className="btn btn-secondary btn-sm">
            <Settings size={14} />
            <span>Manage</span>
          </Link>
        )}
        {onViewResults && (
          <button type="button" onClick={() => onViewResults(poll)} className="btn btn-primary btn-sm">
            <BarChart3 size={14} />
            <span>View Live Results</span>
          </button>
        )}

        <a href={`/p/${poll.id}`} target="_blank" rel="noopener noreferrer" className="btn btn-outline btn-sm">
          <ExternalLink size={14} />
          <span>Open Public</span>
        </a>

        <button type="button" onClick={handleCopyLink} className="btn btn-outline btn-sm">
          {copied ? <Check size={14} className="text-green-400" /> : <Copy size={14} />}
          <span>{copied ? 'Copied!' : 'Copy Link'}</span>
        </button>

        {canManage && isActive && onClosePoll && (
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              onClosePoll(poll.id);
            }}
            className="btn btn-danger btn-sm"
          >
            <span>Close Poll</span>
          </button>
        )}

        {canManage && onDeletePoll && (
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              onDeletePoll(poll.id);
            }}
            className="btn btn-danger btn-sm"
          >
            <Trash2 size={14} />
            <span>Delete</span>
          </button>
        )}
      </div>
    </div>
  );
};
