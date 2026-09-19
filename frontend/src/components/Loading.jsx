import { Loader2 } from 'lucide-react';

export const Loading = ({ message = 'Loading...', fullPage = false }) => {
  if (fullPage) {
    return (
      <div className="fullpage-loader">
        <div className="loader-content">
          <Loader2 className="spinner-large" size={48} />
          <p className="loader-text">{message}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="component-loader" role="status">
      <Loader2 className="spinner" size={28} />
      <span className="loader-text">{message}</span>
    </div>
  );
};

export const SkeletonCard = ({ lines = 3 }) => (
  <div className="skeleton-card" aria-label="Loading content">
    {Array.from({ length: lines }, (_, index) => <span key={index} className={`skeleton-line skeleton-line-${index}`} />)}
  </div>
);
