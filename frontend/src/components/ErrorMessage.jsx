import { AlertCircle, X } from 'lucide-react';

export const ErrorMessage = ({ message, onClose, className = '' }) => {
  if (!message) return null;

  return (
    <div className={`error-banner ${className}`.trim()} role="alert">
      <div className="error-banner-content">
        <AlertCircle size={20} className="error-icon" />
        <span className="error-message-text">{message}</span>
      </div>
      {onClose && (
        <button type="button" onClick={onClose} className="btn-close-error" aria-label="Close error message">
          <X size={16} />
        </button>
      )}
    </div>
  );
};
