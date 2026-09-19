import { useState } from 'react';
import { Input } from './Input';
import { Button } from './Button';
import { Plus, Trash2 } from 'lucide-react';

export const PollForm = ({ initialData, onSubmit, loading, buttonText = 'Create Poll' }) => {
  const [question, setQuestion] = useState(initialData?.question || '');
  const [options, setOptions] = useState(
    initialData?.options ? initialData.options.map((o) => o.text) : ['', '']
  );
  const [errors, setErrors] = useState({});

  const handleAddOption = () => {
    if (options.length < 6) {
      setOptions([...options, '']);
    }
  };

  const handleRemoveOption = (index) => {
    if (options.length > 2) {
      const newOpts = options.filter((_, i) => i !== index);
      setOptions(newOpts);
    }
  };

  const handleOptionChange = (index, value) => {
    const newOpts = [...options];
    newOpts[index] = value;
    setOptions(newOpts);
  };

  const validate = () => {
    const errs = {};
    const trimmedQuestion = question.trim();

    if (!trimmedQuestion) {
      errs.question = 'Poll question is required.';
    } else if (trimmedQuestion.length < 5) {
      errs.question = 'Question must be at least 5 characters long.';
    } else if (trimmedQuestion.length > 200) {
      errs.question = 'Question cannot exceed 200 characters.';
    }

    const trimmedOpts = options.map((o) => o.trim());
    if (trimmedOpts.length < 2) {
      errs.options = 'At least 2 options are required.';
    } else if (trimmedOpts.length > 6) {
      errs.options = 'Maximum 6 options allowed.';
    }

    const emptyIdx = trimmedOpts.findIndex((o) => o === '');
    if (emptyIdx !== -1) {
      errs.options = `Option ${emptyIdx + 1} cannot be empty.`;
    }

    // Check for duplicates
    const uniqueOpts = new Set(trimmedOpts.map((o) => o.toLowerCase()));
    if (uniqueOpts.size !== trimmedOpts.length) {
      errs.options = 'Duplicate options are not allowed.';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!validate()) return;

    onSubmit({
      question: question.trim(),
      options: options.map((o) => o.trim()),
    });
  };

  return (
    <form onSubmit={handleSubmit} className="poll-form poll-form-with-preview">
      <div className="poll-form-fields">
        <div className="form-section">
        <Input
          label="Poll Question"
          id="question"
          name="question"
          placeholder="e.g. What is your favorite programming language?"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          error={errors.question}
          required
          maxLength={200}
          helperText={`${question.trim().length}/200 characters (min 5)`}
        />
        </div>

        <div className="form-section">
        <div className="options-header">
          <label className="form-label">
            Poll Options <span className="required-asterisk">*</span>
          </label>
          <span className="options-count-badge">
            {options.length} / 6 Options
          </span>
        </div>

        {errors.options && <p className="form-error-text mb-3">{errors.options}</p>}

        <div className="options-list">
          {options.map((opt, idx) => (
            <div key={idx} className="option-input-row">
              <Input
                id={`option-${idx}`}
                placeholder={`Option ${idx + 1}`}
                value={opt}
                onChange={(e) => handleOptionChange(idx, e.target.value)}
                maxLength={100}
                required
                className="option-input-field"
              />
              {options.length > 2 && (
                <button
                  type="button"
                  onClick={() => handleRemoveOption(idx)}
                  className="btn-remove-option"
                  title="Remove Option"
                  aria-label={`Remove option ${idx + 1}`}
                >
                  <Trash2 size={16} />
                </button>
              )}
            </div>
          ))}
        </div>

        {options.length < 6 && (
          <button type="button" onClick={handleAddOption} className="btn-add-option">
            <Plus size={16} />
            <span>Add Option</span>
          </button>
        )}
        </div>

        <div className="form-actions">
          <Button type="submit" variant="primary" size="lg" loading={loading} className="w-full">
            {buttonText}
          </Button>
        </div>
      </div>
      <aside className="poll-live-preview" aria-label="Live poll preview">
        <div className="poll-live-preview-heading">
          <span className="dashboard-live-indicator"><span aria-hidden="true">●</span> Live preview</span>
        </div>
        <div className="poll-preview-card">
          <h3>{question.trim() || 'Your poll question appears here'}</h3>
          <div className="poll-preview-options">
            {options.filter((option) => option.trim()).map((option, index) => (
              <div className="poll-preview-option" key={`${option}-${index}`}>
                <span className="poll-preview-radio" aria-hidden="true" />
                <span>{option.trim()}</span>
              </div>
            ))}
            {!options.some((option) => option.trim()) && <span className="poll-preview-placeholder">Add options to preview them here.</span>}
          </div>
          <div className="poll-preview-footer">Vote</div>
        </div>
      </aside>
    </form>
  );
};
