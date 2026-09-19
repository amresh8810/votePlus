import { useMemo, useState } from 'react';

export const VotesByOptionChart = ({ options = [], counts = {}, totalVotes = 0 }) => {
  const [activeOption, setActiveOption] = useState(null);
  const rows = useMemo(
    () => options.map((option) => ({
      ...option,
      votes: counts[option.id] ?? option.vote_count ?? 0,
    })),
    [options, counts],
  );
  const calculatedTotal = rows.reduce((sum, option) => sum + option.votes, 0);
  const displayedTotal = totalVotes || calculatedTotal;
  const maxVotes = Math.max(...rows.map((option) => option.votes), 1);

  return (
    <section className="votes-by-option" aria-labelledby="votes-by-option-title">
      <div className="votes-by-option-header">
        <h3 id="votes-by-option-title">Votes by Option</h3>
        <span>Vote count</span>
      </div>
      {displayedTotal === 0 ? (
        <p className="votes-by-option-empty" role="status">No votes yet</p>
      ) : (
        <>
          <div className="votes-by-option-chart" role="img" aria-label="Horizontal bar chart showing votes by option">
            {rows.map((option) => {
              const percentage = (option.votes / maxVotes) * 100;
              const isActive = activeOption === option.id;
              return (
                <div
                  className={`votes-by-option-row ${isActive ? 'votes-by-option-row-active' : ''}`}
                  key={option.id}
                  title={`${option.text}: ${option.votes} ${option.votes === 1 ? 'vote' : 'votes'}`}
                  tabIndex="0"
                  aria-label={`${option.text}: ${option.votes} ${option.votes === 1 ? 'vote' : 'votes'}`}
                  onMouseEnter={() => setActiveOption(option.id)}
                  onMouseLeave={() => setActiveOption(null)}
                  onFocus={() => setActiveOption(option.id)}
                  onBlur={() => setActiveOption(null)}
                >
                  <span className="votes-by-option-label" title={option.text}>{option.text}</span>
                  <span className="votes-by-option-track">
                    <span className="votes-by-option-bar" style={{ width: `${percentage}%` }} />
                  </span>
                  <strong className="votes-by-option-value">{option.votes.toLocaleString()}</strong>
                </div>
              );
            })}
          </div>
          <div className="votes-by-option-axis" aria-hidden="true">
            <span>0</span>
            <span>{Math.ceil(maxVotes / 2)}</span>
            <span>{maxVotes.toLocaleString()}</span>
          </div>
        </>
      )}
    </section>
  );
};
