import clsx from 'clsx';

interface CardProps {
  children: React.ReactNode;
  className?: string;
  padding?: boolean;
  glow?: boolean;
}

export function Card({ children, className, padding = true, glow = false }: CardProps) {
  return (
    <div
      className={clsx(
        'card-ember border-border-subtle relative overflow-hidden rounded-xl border',
        'from-bg-card to-bg-deep bg-linear-to-br',
        {
          'p-6': padding,
          'animate-[glow-pulse_4s_ease-in-out_infinite]': glow,
        },
        className,
      )}
    >
      {/* Top edge highlight */}
      <div className="via-ash-400/20 pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent to-transparent" />

      {/* Content */}
      <div className="relative">{children}</div>
    </div>
  );
}
