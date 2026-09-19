import { forwardRef } from 'react';

export const Input = forwardRef(
  (
    {
      label,
      id,
      name,
      type = 'text',
      placeholder,
      value,
      onChange,
      error,
      helperText,
      required = false,
      disabled = false,
      maxLength,
      className = '',
      ...props
    },
    ref
  ) => {
    const inputId = id || name;

    return (
      <div className={`form-group ${error ? 'has-error' : ''} ${className}`.trim()}>
        {label && (
          <label htmlFor={inputId} className="form-label">
            {label} {required && <span className="required-asterisk">*</span>}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          name={name}
          type={type}
          value={value}
          onChange={onChange}
          placeholder={placeholder}
          disabled={disabled}
          maxLength={maxLength}
          required={required}
          className={`form-input ${error ? 'input-error' : ''}`}
          aria-invalid={!!error}
          aria-describedby={error ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
          {...props}
        />
        {error && (
          <p id={`${inputId}-error`} className="form-error-text" role="alert">
            {error}
          </p>
        )}
        {!error && helperText && (
          <p id={`${inputId}-helper`} className="form-helper-text">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
