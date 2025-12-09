import clsx from 'clsx';

interface LoadingProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function Loading({ className, size = 'md' }: LoadingProps) {
  const sizeClasses = {
    sm: 'size-4 border-2',
    md: 'size-8 border-2',
    lg: 'size-12 border-[3px]',
  };

  return (
    <div className={clsx('flex items-center justify-center', className)}>
      <div
        className={clsx('border-ember-500 animate-spin rounded-full border-t-transparent', sizeClasses[size])}
        role="status"
        aria-label="Loading"
      >
        <span className="sr-only">Loading...</span>
      </div>
    </div>
  );
}

export function LoadingOverlay() {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center gap-4">
      <Loading size="lg" />
      <span className="font-display text-text-muted animate-pulse text-sm tracking-wide">Loading...</span>
    </div>
  );
}
