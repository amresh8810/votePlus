import { Loader2 } from 'lucide-react';

export const Button = ({
  children,
  type = 'button',
  variant = 'primary', // 'primary' | 'secondary' | 'outline' | 'danger' | 'success'
  size = 'md', // 'sm' | 'md' | 'lg'
  loading = false,
  disabled = false,
  className = '',
  onClick,
  ...props
}) => {
  const baseClass = 'btn';
  const variantClass = `btn-${variant}`;
  const sizeClass = `btn-${size}`;
  const loadingClass = loading ? 'btn-loading' : '';

  return (
    <button
      type={type}
      className={`${baseClass} ${variantClass} ${sizeClass} ${loadingClass} ${className}`.trim()}
      disabled={disabled || loading}
      onClick={onClick}
      {...props}
    >
      {loading ? (
        <>
          <Loader2 className="spinner" size={size === 'sm' ? 14 : size === 'lg' ? 22 : 18} />
          <span>Please wait...</span>
        </>
      ) : (
        children
      )}
    </button>
  );
};
