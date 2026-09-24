import { BarChart3 } from 'lucide-react';

export const PollPerformanceChart = ({ polls = [], votesByPoll = [] }) => {
  const rows = polls.map((poll) => ({
    ...poll,
    votes: votesByPoll[poll.id] || 0,
  }));
  const maxVotes = Math.max(...rows.map((poll) => poll.votes), 1);
  const totalVotes = rows.reduce((sum, poll) => sum + poll.votes, 0);

  return (
    <section className="analytics-card poll-performance-card" aria-labelledby="poll-performance-title">
      <div className="analytics-card-header">
        <div className="analytics-title-group">
          <div className="analytics-title-icon" aria-hidden="true"><BarChart3 size={16} /></div>
          <div>
            <p className="analytics-eyebrow">Comparison</p>
            <h2 id="poll-performance-title">Poll Performance</h2>
            <p className="analytics-subtitle">Votes across all shared polls</p>
          </div>
        </div>
        <span className="poll-performance-axis-label">{totalVotes.toLocaleString()} total votes</span>
      </div>

      {rows.length === 0 ? (
        <div className="analytics-chart-empty poll-performance-empty" role="status">
          <strong>No polls yet</strong>
          <span>Create a poll to start comparing performance.</span>
        </div>
      ) : (
        <>
          <div className="poll-performance-chart" role="img" aria-label="Horizontal bar chart comparing votes across polls">
            {rows.map((poll) => (
              <div
                className="poll-performance-row"
                key={poll.id}
                title={`${poll.question}: ${poll.votes} ${poll.votes === 1 ? 'vote' : 'votes'} (${poll.status})`}
                tabIndex="0"
                aria-label={`${poll.question}: ${poll.votes} ${poll.votes === 1 ? 'vote' : 'votes'}, ${poll.status}`}
              >
                <span className="poll-performance-label" title={poll.question}>{poll.question}</span>
                <span className="poll-performance-track">
                  <span
                    className="poll-performance-bar"
                    style={{ width: `${(poll.votes / maxVotes) * 100}%` }}
                  />
                </span>
                <strong className="poll-performance-value">
                  {poll.votes.toLocaleString()}
                  <small>{totalVotes ? `${((poll.votes / totalVotes) * 100).toFixed(0)}% share` : '0% share'}</small>
                </strong>
              </div>
            ))}
          </div>
          <div className="poll-performance-axis" aria-hidden="true">
            <span>0</span>
            <span>{Math.ceil(maxVotes / 2)}</span>
            <span>{maxVotes.toLocaleString()}</span>
          </div>
        </>
      )}
    </section>
  );
};
