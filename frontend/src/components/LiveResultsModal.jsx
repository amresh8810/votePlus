import { useEffect, useMemo, useRef } from 'react';
import { CheckCircle2, CircleDot, LoaderCircle, Radio, X } from 'lucide-react';
import { VotesByOptionChart } from './VotesByOptionChart';

const colors = ['#4f46e5', '#0891b2', '#16a34a', '#d97706', '#db2777', '#7c3aed'];
const donutCircumference = 2 * Math.PI * 54;

const getInitialCount = (option, counts) => counts[option.id] ?? option.vote_count ?? 0;

export const LiveResultsModal = ({ poll, counts = {}, totalVotes = 0, connectionStatus = 'connecting', onClose }) => {
  const closeButtonRef = useRef(null);
  const results = useMemo(() => {
    const options = (poll.options || []).map((option, index) => ({
      ...option,
      votes: getInitialCount(option, counts),
      color: colors[index % colors.length],
    }));
    const currentTotal = options.reduce((sum, option) => sum + option.votes, 0);
    return { options, total: totalVotes || currentTotal };
  }, [poll.options, counts, totalVotes]);

  useEffect(() => {
    closeButtonRef.current?.focus();
    const handleKeyDown = (event) => {
      if (event.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const segments = results.options.reduce((items, option) => {
    const start = items.length ? items[items.length - 1].end : 0;
    const end = start + (results.total ? (option.votes / results.total) * donutCircumference : 0);
    items.push({ option, start, end });
    return items;
  }, []);
  const hasKnownCounts = Object.keys(counts).length > 0 || results.options.some((option) => option.vote_count != null);
  const isLoading = connectionStatus === 'connecting' && !hasKnownCounts;
  const isReconnecting = connectionStatus !== 'connected';

  return (
    <div className="live-results-backdrop" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget) onClose();
    }}>
      <section
        className="live-results-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="live-results-title"
      >
        <header className="live-results-header">
          <div>
            <div className={`live-results-status ${isReconnecting ? 'live-results-status-reconnecting' : ''}`}>
              {isReconnecting ? <LoaderCircle size={13} /> : <Radio size={13} />}
              {isReconnecting ? 'Reconnecting...' : 'LIVE'}
            </div>
            <h2 id="live-results-title">Live Results</h2>
            <p>{poll.question}</p>
          </div>
          <button ref={closeButtonRef} type="button" className="live-results-close" onClick={onClose} aria-label="Close live results">
            <X size={18} />
          </button>
        </header>

        {isLoading ? (
          <div className="live-results-loading" role="status">
            <div className="live-results-skeleton live-results-skeleton-chart" />
            <div className="live-results-skeleton-list">
              <div className="live-results-skeleton" />
              <div className="live-results-skeleton" />
              <div className="live-results-skeleton" />
            </div>
          </div>
        ) : (
        <>
        <div className="live-results-analytics-grid">
          <section className="live-results-panel" aria-labelledby="live-results-panel-title">
            <div className="live-results-panel-heading">
              <div>
                <p className="analytics-eyebrow">Distribution</p>
                <h3 id="live-results-panel-title">Live Results</h3>
              </div>
              <span className="live-results-panel-total">{results.total.toLocaleString()} votes</span>
            </div>

            <div className="live-results-body">
              <div className="live-results-chart" aria-label={`${results.total} total votes`}>
                <svg viewBox="0 0 160 160" role="img" aria-label="Vote distribution donut chart">
                  <circle cx="80" cy="80" r="54" className="live-results-donut-track" />
                  {segments.map(({ option, start, end }) => (
                    <circle
                      key={option.id}
                      cx="80"
                      cy="80"
                      r="54"
                      className="live-results-donut-segment"
                      stroke={option.color}
                      strokeDasharray={`${Math.max(end - start, 0)} ${donutCircumference - Math.max(end - start, 0)}`}
                      strokeDashoffset={-start}
                      transform="rotate(-90 80 80)"
                    />
                  ))}
                </svg>
                <div className="live-results-chart-total">
                  <strong>{results.total.toLocaleString()}</strong>
                  <span>Total votes</span>
                </div>
              </div>

              <div className="live-results-list">
                {results.total === 0 && <p className="live-results-empty">No votes yet</p>}
                {results.options.map((option) => {
                  const percentage = results.total ? ((option.votes / results.total) * 100).toFixed(1) : '0.0';
                  return (
                    <div className="live-results-row" key={option.id}>
                      <span className="live-results-option">
                        <CircleDot size={14} color={option.color} aria-hidden="true" />
                        <span title={option.text}>{option.text}</span>
                      </span>
                      <span className="live-results-votes">{option.votes.toLocaleString()} votes</span>
                      <strong>{percentage}%</strong>
                    </div>
                  );
                })}
                {results.options.length === 0 && <p className="live-results-empty">No options available.</p>}
              </div>
            </div>
          </section>

          <section className="live-results-panel live-results-bar-panel" aria-label="Votes by Option">
            <VotesByOptionChart
              options={results.options}
              counts={counts}
              totalVotes={results.total}
            />
          </section>
        </div>
        </>
        )}

        <footer className="live-results-footer">
          <span><CheckCircle2 size={14} /> {isReconnecting ? 'Last known results shown' : 'Results update automatically'}</span>
          <button type="button" className="btn btn-secondary btn-sm" onClick={onClose}>Close</button>
        </footer>
      </section>
    </div>
  );
};
