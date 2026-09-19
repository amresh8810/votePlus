import { Users } from 'lucide-react';

export const PollResults = ({ poll, counts = {}, totalVotes = 0 }) => {
  if (!poll || !poll.options) return null;

  // Calculate percentages safely to avoid NaN/Infinity
  const calculatePercentage = (count) => {
    if (!totalVotes || totalVotes <= 0 || !count) return 0;
    const pct = (count / totalVotes) * 100;
    return Math.round(pct * 10) / 10; // Round to 1 decimal place
  };

  return (
    <div className="poll-results-container">
      <div className="results-header">
        <h3 className="results-title">Live Results</h3>
        <div className="total-votes-badge">
          <Users size={16} />
          <span>{totalVotes} Total {totalVotes === 1 ? 'Vote' : 'Votes'}</span>
        </div>
      </div>

      <div className="results-list">
        {poll.options.map((opt) => {
          const voteCount = counts[opt.id] || 0;
          const percentage = calculatePercentage(voteCount);

          return (
            <div key={opt.id} className="result-item">
              <div className="result-label-row">
                <span className="result-option-name">{opt.text}</span>
                <div className="result-stats">
                  <span className="result-percentage">{percentage}%</span>
                  <span className="result-count">
                    ({voteCount} {voteCount === 1 ? 'vote' : 'votes'})
                  </span>
                </div>
              </div>

              <div className="progress-track" role="progressbar" aria-valuenow={percentage} aria-valuemin="0" aria-valuemax="100">
                <div
                  className="progress-fill"
                  style={{ width: `${percentage}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
