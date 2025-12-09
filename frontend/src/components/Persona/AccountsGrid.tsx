import { Link } from '@tanstack/react-router';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import { usePersonaAccounts } from '@/hooks/usePersonaAccounts';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatPercent, pnlColor, pnlSign } from '@/utils/formatters';

const POLYMARKET_PROFILE_URL = 'https://polymarket.com/profile/@';

interface AccountsGridProps {
  slug: string;
}

export function AccountsGrid({ slug }: AccountsGridProps) {
  const { data: accounts, isLoading, error, refetch } = usePersonaAccounts(slug);

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
        <ErrorState message="Failed to load accounts" retry={refetch} />
      </Card>
    );
  }

  if (!accounts || accounts.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary py-12 text-center">No accounts found</div>
      </Card>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-end justify-between">
        <div>
          <h2 className="font-display text-text-bright text-2xl font-bold">Accounts</h2>
          <p className="text-text-muted mt-1 text-sm/6">Individual account performance</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {accounts.map((account, index) => (
          <Card key={account.username}>
            <div style={{ animationDelay: `${index * 50}ms` }}>
              {/* Account Header */}
              <div className="mb-4 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="from-ember-500/20 to-ember-600/20 font-display text-ember-400 flex size-10 items-center justify-center rounded-lg bg-linear-to-br text-lg font-bold">
                    {account.username.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <Link
                      to="/users/$username"
                      params={{ username: account.username }}
                      className="font-display text-text-bright hover:text-ember-400 text-lg font-semibold transition-colors"
                    >
                      @{account.username}
                    </Link>
                    <p className="text-text-muted text-xs">{(account.openPositions ?? 0)} open positions</p>
                  </div>
                </div>

                <a
                  href={`${POLYMARKET_PROFILE_URL}${account.username}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-text-muted hover:text-ember-400 transition-colors"
                  title="View on Polymarket"
                >
                  <ArrowTopRightOnSquareIcon className="size-5" />
                </a>
              </div>

              {/* Account Stats */}
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <span className="text-text-muted block text-xs font-medium uppercase">Total PnL</span>
                  <span className={`font-display text-lg font-bold tabular-nums ${pnlColor(account.totalPnl)}`}>
                    {pnlSign(account.totalPnl)}
                    {formatCurrency(account.totalPnl)}
                  </span>
                </div>
                <div>
                  <span className="text-text-muted block text-xs font-medium uppercase">Realized</span>
                  <span className={`text-sm/6 font-medium tabular-nums ${pnlColor(account.realizedPnl)}`}>
                    {pnlSign(account.realizedPnl)}
                    {formatCurrency(account.realizedPnl)}
                  </span>
                </div>
                <div>
                  <span className="text-text-muted block text-xs font-medium uppercase">Win Rate</span>
                  <span className="text-text-secondary text-sm/6 tabular-nums">
                    {formatPercent((account.winRate ?? 0) * 100)}
                  </span>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
