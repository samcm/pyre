import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowTopRightOnSquareIcon, FunnelIcon } from '@heroicons/react/24/solid';
import { useRecentTrades, type RecentTradesFilters } from '@/hooks/useRecentTrades';
import { Card } from '@/components/Common/Card';
import { Badge } from '@/components/Common/Badge';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatDateTime, formatNumber } from '@/utils/formatters';

const POLYMARKET_PROFILE_URL = 'https://polymarket.com/profile/@';
const POLYMARKET_MARKET_URL = 'https://polymarket.com/event/';

export function RecentTradesTable() {
  const [filters, setFilters] = useState<RecentTradesFilters>({
    limit: 20,
    offset: 0,
    sortBy: 'timestamp',
    sortDirection: 'desc',
  });
  const [showFilters, setShowFilters] = useState(false);

  const { data, isLoading, error, refetch } = useRecentTrades(filters);

  const handleFilterChange = (key: keyof RecentTradesFilters, value: string | number | undefined) => {
    setFilters(prev => ({
      ...prev,
      [key]: value === '' ? undefined : value,
      offset: 0, // Reset to first page when filters change
    }));
  };

  const handlePageChange = (newOffset: number) => {
    setFilters(prev => ({ ...prev, offset: newOffset }));
  };

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
        <ErrorState message="Failed to load recent trades" retry={refetch} />
      </Card>
    );
  }

  if (!data || data.trades.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary font-display py-12 text-center">No trades found</div>
      </Card>
    );
  }

  const { trades, total, limit, offset } = data;
  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(total / limit);
  const hasMore = offset + limit < total;

  return (
    <div>
      <div className="mb-6 flex items-end justify-between">
        <div>
          <h2 className="font-display text-text-bright text-2xl font-bold">Recent Trades</h2>
          <p className="text-text-muted mt-1 text-sm/6">{total} trades across all tracked users</p>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`inline-flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm/6 font-medium transition-all duration-200 ${
            showFilters
              ? 'from-ember-600 to-ember-700 bg-linear-to-r text-white shadow-lg'
              : 'bg-bg-elevated text-text-primary hover:bg-bg-hover border-border-subtle border'
          }`}
        >
          <FunnelIcon className="size-4" />
          Filters
        </button>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <Card className="mb-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <label className="text-text-secondary mb-2 block text-xs font-medium tracking-wider uppercase">
                Username
              </label>
              <input
                type="text"
                value={filters.username ?? ''}
                onChange={e => handleFilterChange('username', e.target.value)}
                placeholder="Filter by username"
                className="border-border-subtle bg-bg-deep text-text-primary placeholder:text-text-muted focus:border-ember-500 w-full rounded-lg border px-3 py-2 text-sm/6 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-text-secondary mb-2 block text-xs font-medium tracking-wider uppercase">
                Side
              </label>
              <select
                value={filters.side ?? ''}
                onChange={e => handleFilterChange('side', e.target.value as 'BUY' | 'SELL' | '')}
                className="border-border-subtle bg-bg-deep text-text-primary focus:border-ember-500 w-full rounded-lg border px-3 py-2 text-sm/6 focus:outline-none"
              >
                <option value="">All</option>
                <option value="BUY">Buy</option>
                <option value="SELL">Sell</option>
              </select>
            </div>
            <div>
              <label className="text-text-secondary mb-2 block text-xs font-medium tracking-wider uppercase">
                Min Value ($)
              </label>
              <input
                type="number"
                value={filters.minValue ?? ''}
                onChange={e => handleFilterChange('minValue', e.target.value ? parseFloat(e.target.value) : undefined)}
                placeholder="0"
                min="0"
                step="0.01"
                className="border-border-subtle bg-bg-deep text-text-primary placeholder:text-text-muted focus:border-ember-500 w-full rounded-lg border px-3 py-2 text-sm/6 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-text-secondary mb-2 block text-xs font-medium tracking-wider uppercase">
                Sort By
              </label>
              <div className="flex gap-2">
                <select
                  value={filters.sortBy ?? 'timestamp'}
                  onChange={e => handleFilterChange('sortBy', e.target.value as 'timestamp' | 'value' | 'size')}
                  className="border-border-subtle bg-bg-deep text-text-primary focus:border-ember-500 flex-1 rounded-lg border px-3 py-2 text-sm/6 focus:outline-none"
                >
                  <option value="timestamp">Time</option>
                  <option value="value">Value</option>
                  <option value="size">Size</option>
                </select>
                <select
                  value={filters.sortDirection ?? 'desc'}
                  onChange={e => handleFilterChange('sortDirection', e.target.value as 'asc' | 'desc')}
                  className="border-border-subtle bg-bg-deep text-text-primary focus:border-ember-500 w-20 rounded-lg border px-3 py-2 text-sm/6 focus:outline-none"
                >
                  <option value="desc">Desc</option>
                  <option value="asc">Asc</option>
                </select>
              </div>
            </div>
          </div>
        </Card>
      )}

      <Card padding={false}>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border-subtle bg-bg-deep/50 border-b">
                <th className="text-text-muted px-4 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Time
                </th>
                <th className="text-text-muted px-4 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  User
                </th>
                <th className="text-text-muted px-4 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Market
                </th>
                <th className="text-text-muted px-4 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Outcome
                </th>
                <th className="text-text-muted px-4 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Side
                </th>
                <th className="text-text-muted px-4 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Price
                </th>
                <th className="text-text-muted px-4 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Size
                </th>
                <th className="text-text-muted px-4 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Value
                </th>
              </tr>
            </thead>
            <tbody className="divide-border-subtle divide-y">
              {trades.map((trade, index) => (
                <tr
                  key={`${trade.username}-${trade.timestamp}-${trade.conditionId}`}
                  className="table-row-ember group"
                  style={{ animationDelay: `${index * 30}ms` }}
                >
                  <td className="text-text-secondary px-4 py-3 text-sm/6">{formatDateTime(trade.timestamp)}</td>
                  <td className="px-4 py-3">
                    {trade.personaDisplayName && trade.personaSlug ? (
                      <div className="flex flex-col gap-0.5">
                        <Link
                          to="/people/$slug"
                          params={{ slug: trade.personaSlug }}
                          className="font-display text-text-primary hover:text-ember-400 text-sm/6 font-semibold transition-colors"
                        >
                          {trade.personaDisplayName}
                        </Link>
                        <div className="flex items-center gap-1.5">
                          <Link
                            to="/users/$username"
                            params={{ username: trade.username }}
                            className="text-text-muted hover:text-text-secondary text-xs transition-colors"
                          >
                            @{trade.username}
                          </Link>
                          <a
                            href={`${POLYMARKET_PROFILE_URL}${trade.username}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
                            title="View on Polymarket"
                          >
                            <ArrowTopRightOnSquareIcon className="size-3" />
                          </a>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2">
                        <Link
                          to="/users/$username"
                          params={{ username: trade.username }}
                          className="font-display text-text-primary hover:text-ember-400 text-sm/6 font-semibold transition-colors"
                        >
                          {trade.username}
                        </Link>
                        <a
                          href={`${POLYMARKET_PROFILE_URL}${trade.username}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
                          title="View on Polymarket"
                        >
                          <ArrowTopRightOnSquareIcon className="size-3.5" />
                        </a>
                      </div>
                    )}
                  </td>
                  <td className="max-w-xs px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="text-text-primary truncate text-sm/6" title={trade.marketTitle}>
                        {trade.marketTitle}
                      </span>
                      {trade.marketSlug && (
                        <a
                          href={`${POLYMARKET_MARKET_URL}${trade.marketSlug}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-text-muted hover:text-ember-400 shrink-0 opacity-0 transition-all group-hover:opacity-100"
                          title="View market on Polymarket"
                        >
                          <ArrowTopRightOnSquareIcon className="size-3.5" />
                        </a>
                      )}
                    </div>
                  </td>
                  <td className="text-text-secondary px-4 py-3 text-sm/6">{trade.outcome}</td>
                  <td className="px-4 py-3">
                    <Badge variant={trade.side === 'BUY' ? 'success' : 'error'}>{trade.side}</Badge>
                  </td>
                  <td className="text-text-primary px-4 py-3 text-right text-sm/6 tabular-nums">
                    ${formatNumber(trade.price, 2)}
                  </td>
                  <td className="text-text-secondary px-4 py-3 text-right text-sm/6 tabular-nums">
                    {formatNumber(trade.size, 0)}
                  </td>
                  <td className="font-display text-text-bright px-4 py-3 text-right text-sm/6 font-semibold tabular-nums">
                    {formatCurrency(trade.value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {total > limit && (
          <div className="border-border-subtle flex items-center justify-between border-t px-4 py-3">
            <div className="text-text-muted text-sm/6">
              Page {currentPage} of {totalPages}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => handlePageChange(Math.max(0, offset - limit))}
                disabled={offset === 0}
                className="bg-bg-elevated text-text-primary hover:bg-bg-hover border-border-subtle rounded-lg border px-4 py-2 text-sm/6 font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => handlePageChange(offset + limit)}
                disabled={!hasMore}
                className="bg-bg-elevated text-text-primary hover:bg-bg-hover border-border-subtle rounded-lg border px-4 py-2 text-sm/6 font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
