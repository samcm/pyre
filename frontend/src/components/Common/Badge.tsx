import clsx from 'clsx';

type BadgeVariant = 'default' | 'success' | 'error' | 'warning' | 'ember';

interface BadgeProps {
  children: React.ReactNode;
  variant?: BadgeVariant;
  className?: string;
  pulse?: boolean;
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-ash-700/50 text-ash-200 border-ash-600/50',
  success: 'bg-profit/15 text-profit border-profit/30',
  error: 'bg-loss/15 text-loss border-loss/30',
  warning: 'bg-chart-gold/15 text-chart-gold border-chart-gold/30',
  ember: 'bg-ember-500/15 text-ember-400 border-ember-500/30',
};

export function Badge({ children, variant = 'default', className, pulse = false }: BadgeProps) {
  return (
    <span
      className={clsx(
        'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-semibold tracking-wide uppercase',
        variantClasses[variant],
        {
          'animate-pulse': pulse,
        },
        className,
      )}
    >
      {children}
    </span>
  );
}
