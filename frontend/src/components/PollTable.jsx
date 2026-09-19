import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Check, Copy, ExternalLink, MoreHorizontal, Settings, Trash2, XCircle, BarChart3 } from 'lucide-react';

const formatDate = (date) => {
  if (!date) return '—';
  return new Date(date).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

export const PollTable = ({ polls, votesByPoll, lastActivityByPoll, onClosePoll, onDeletePoll, onViewResults }) => {
  const [copiedPollId, setCopiedPollId] = useState('');
  const [openMenuId, setOpenMenuId] = useState('');

  const copyPollLink = async (pollId) => {
    await navigator.clipboard.writeText(`${window.location.origin}/p/${pollId}`);
    setCopiedPollId(pollId);
    window.setTimeout(() => setCopiedPollId(''), 2000);
  };

  return (
    <section className="poll-table-card" aria-labelledby="poll-table-title">
      <div className="poll-table-header">
        <div>
          <p className="analytics-eyebrow">Manage</p>
          <h2 id="poll-table-title">All Polls</h2>
        </div>
        <span className="poll-table-count">{polls.length} total</span>
      </div>
      <div className="poll-table-scroll">
        <table className="poll-table">
          <thead>
            <tr>
              <th scope="col">Poll</th>
              <th scope="col">Status</th>
              <th scope="col">Votes</th>
              <th scope="col">Created</th>
              <th scope="col">Last Activity</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {polls.map((poll) => {
              const isActive = poll.status === 'active';
              return (
                <tr key={poll.id}>
                  <td className="poll-table-question">{poll.question}</td>
                  <td>
                    <span className={`status-badge ${isActive ? 'status-active' : 'status-closed'}`}>
                      {isActive ? <Check size={13} /> : <XCircle size={13} />}
                      {isActive ? 'Active' : 'Closed'}
                    </span>
                  </td>
                  <td className="poll-table-number">{(votesByPoll[poll.id] || 0).toLocaleString()}</td>
                  <td>{formatDate(poll.created_at)}</td>
                  <td>{formatDate(lastActivityByPoll[poll.id] || poll.updated_at || poll.created_at)}</td>
                  <td>
                    <div className="poll-table-actions">
                      <button type="button" className="table-action table-action-primary" onClick={() => onViewResults?.(poll)} title="View live results" aria-label={`View live results for ${poll.question}`}>
                        <BarChart3 size={15} />
                        <span>Live Results</span>
                      </button>
                      <Link to={`/polls/${poll.id}`} className="table-action table-action-secondary" title="Manage poll" aria-label={`Manage ${poll.question}`}>
                        <Settings size={15} />
                        <span>Manage</span>
                      </Link>
                      <div className="poll-more-menu">
                        <button type="button" className="table-action" title="More poll actions" aria-label={`More actions for ${poll.question}`} aria-expanded={openMenuId === poll.id} onClick={() => setOpenMenuId(openMenuId === poll.id ? '' : poll.id)}>
                          <MoreHorizontal size={17} />
                        </button>
                        {openMenuId === poll.id && (
                          <div className="poll-more-menu-items">
                            <a href={`/p/${poll.id}`} target="_blank" rel="noopener noreferrer" onClick={() => setOpenMenuId('')}><ExternalLink size={14} /> Open Public</a>
                            <button type="button" onClick={() => { copyPollLink(poll.id); setOpenMenuId(''); }}><Copy size={14} /> {copiedPollId === poll.id ? 'Copied' : 'Copy Link'}</button>
                            {isActive && <button type="button" className="menu-warning" onClick={() => { onClosePoll(poll.id); setOpenMenuId(''); }}><XCircle size={14} /> Close Poll</button>}
                            <button type="button" className="menu-danger" onClick={() => { onDeletePoll(poll.id); setOpenMenuId(''); }}><Trash2 size={14} /> Delete</button>
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
};
