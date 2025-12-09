import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import { useUser } from '@/hooks/useUser';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatPercent, pnlColor, pnlSign } from '@/utils/formatters';

const POLYMARKET_PROFILE_URL = 'https://polymarket.com/profile/@';

interface UserSummaryProps {
  username: string;
}

export function UserSummary({ username }: UserSummaryProps) {
  const { data: user, isLoading, error, refetch } = useUser(username);

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
        <ErrorState message={`Failed to load user ${username}`} retry={refetch} />
      </Card>
    );
  }

  if (!user) {
    return (
      <Card>
        <div className="text-text-secondary py-12 text-center">User not found</div>
      </Card>
    );
  }

  const stats = [
    {
      label: 'Total PnL',
      value: formatCurrency(user.totalPnl),
      color: pnlColor(user.totalPnl),
      sign: pnlSign(user.totalPnl),
      highlight: true,
    },
    {
      label: 'Realized PnL',
      value: formatCurrency(user.realizedPnl),
      color: pnlColor(user.realizedPnl),
      sign: pnlSign(user.realizedPnl),
      highlight: false,
    },
    {
      label: 'Unrealized PnL',
      value: formatCurrency(user.unrealizedPnl),
      color: pnlColor(user.unrealizedPnl),
      sign: pnlSign(user.unrealizedPnl),
      highlight: false,
    },
    {
      label: 'Win Rate',
      value: formatPercent(user.winRate),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
    {
      label: 'Total Trades',
      value: user.totalTrades.toString(),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
    {
      label: 'Open Positions',
      value: user.openPositions.toString(),
      color: 'text-text-bright',
      sign: '',
      highlight: false,
    },
  ];

  return (
    <div>
      {/* User Header */}
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-4">
          {/* Avatar placeholder with fire gradient */}
          <div className="from-ember-500 to-ember-700 font-display shadow-ember-500/20 flex size-16 items-center justify-center rounded-xl bg-linear-to-br text-2xl font-bold text-white shadow-lg">
            {user.username.charAt(0).toUpperCase()}
          </div>
          <div>
            <h1 className="font-display text-text-bright text-3xl font-bold">{user.username}</h1>
            <p className="text-text-muted mt-0.5 text-sm/6">Polymarket Trader</p>
          </div>
        </div>

        <a
          href={`${POLYMARKET_PROFILE_URL}${user.username}`}
          target="_blank"
          rel="noopener noreferrer"
          className="from-ember-600 to-ember-700 shadow-ember-500/20 hover:from-ember-500 hover:to-ember-600 inline-flex items-center gap-2 rounded-lg bg-linear-to-r px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-200 hover:shadow-xl"
        >
          <ArrowTopRightOnSquareIcon className="size-4" />
          View on Polymarket
        </a>
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
