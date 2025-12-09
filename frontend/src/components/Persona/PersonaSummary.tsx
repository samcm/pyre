import { usePersona } from '@/hooks/usePersona';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatPercent, pnlColor, pnlSign } from '@/utils/formatters';
import { UsersIcon } from '@heroicons/react/24/solid';

interface PersonaSummaryProps {
  slug: string;
}

export function PersonaSummary({ slug }: PersonaSummaryProps) {
  const { data: persona, isLoading, error, refetch } = usePersona(slug);

  if (isLoading) {
    return (
      <Card>
        <LoadingOverlay />
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <ErrorState message={`Failed to load ${slug}`} retry={refetch} />
      </Card>
    );
  }

  if (!persona) {
    return (
      <Card>
        <div className="text-text-secondary py-12 text-center">Person not found</div>
      </Card>
    );
  }

  const stats = [
    {
      label: 'Total PnL',
      value: formatCurrency(persona.totalPnl),
      color: pnlColor(persona.totalPnl),
      sign: pnlSign(persona.totalPnl),
      highlight: true,
    },
    {
      label: 'Realized PnL',
      value: formatCurrency(persona.realizedPnl),
      color: pnlColor(persona.realizedPnl),
      sign: pnlSign(persona.realizedPnl),
      highlight: false,
    },
    {
      label: 'Unrealized PnL',
      value: formatCurrency(persona.unrealizedPnl),
      color: pnlColor(persona.unrealizedPnl),
      sign: pnlSign(persona.unrealizedPnl),
      highlight: false,
    },
    {
      label: 'Win Rate',
      value: formatPercent((persona.winRate ?? 0) * 100),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
    {
      label: 'Total Trades',
      value: (persona.totalTrades ?? 0).toString(),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
    {
      label: 'Open Positions',
      value: (persona.openPositions ?? 0).toString(),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
  ];

  return (
    <div>
      {/* Header */}
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-4">
          {/* Avatar - persona image or fallback */}
          {persona.image ? (
            <img
              src={persona.image}
              alt={persona.displayName}
              className="size-16 rounded-xl object-cover ring-4 ring-bg-deep shadow-lg"
            />
          ) : (
            <div className="from-ember-500 to-ember-700 font-display shadow-ember-500/20 flex size-16 items-center justify-center rounded-xl bg-linear-to-br shadow-lg">
              <UsersIcon className="size-8 text-white" />
            </div>
          )}
          <div>
            <h1 className="font-display text-text-bright text-3xl font-bold">{persona.displayName}</h1>
            <p className="text-text-muted mt-0.5 text-sm/6">
              {persona.usernames.length} account{persona.usernames.length !== 1 ? 's' : ''} tracked
            </p>
          </div>
        </div>

        <div className="text-text-muted flex flex-wrap gap-2 text-sm">
          {persona.usernames.map((username) => (
            <span key={username} className="bg-bg-elevated rounded-md px-2 py-1">
              @{username}
            </span>
          ))}
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
        {stats.map((stat, index) => (
          <Card key={stat.label} glow={stat.highlight} className={stat.highlight ? 'col-span-2 sm:col-span-1' : ''}>
            <div className="flex flex-col gap-1" style={{ animationDelay: `${index * 50}ms` }}>
              <span className="text-text-muted text-xs font-medium tracking-wider uppercase">{stat.label}</span>
              <span className={`font-display text-2xl font-bold tabular-nums ${stat.color}`}>
                {stat.sign}
                {stat.value}
              </span>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
