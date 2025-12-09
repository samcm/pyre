import { useState } from 'react';
import { useTrades, type Trade } from '@/hooks/useTrades';
import { Card } from '@/components/Common/Card';
import { Badge } from '@/components/Common/Badge';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatDateTime, formatNumber } from '@/utils/formatters';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

interface TradesTableProps {
  username: string;
}

export function TradesTable({ username }: TradesTableProps) {
  const [page, setPage] = useState(1);
  const { data, isLoading, error, refetch } = useTrades(username, page, 10);

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
        <ErrorState message="Failed to load trades" retry={refetch} />
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

  const { trades, total, hasMore } = data;

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-text-bright text-xl font-bold">Trade History</h2>
        <p className="text-text-muted mt-1 text-sm/6">{total} total trades</p>
      </div>
      <Card padding={false}>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border-subtle bg-bg-deep/50 border-b">
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Time
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Market
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Outcome
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Side
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Price
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Size
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Value
                </th>
              </tr>
            </thead>
            <tbody className="divide-border-subtle divide-y">
              {trades.map((trade: Trade, index: number) => (
                <tr
                  key={trade.conditionId}
                  className="table-row-ember group"
                  style={{ animationDelay: `${index * 30}ms` }}
                >
                  <td className="text-text-secondary px-6 py-4 text-sm/6">{formatDateTime(trade.timestamp)}</td>
                  <td className="text-text-primary px-6 py-4 text-sm/6">
                    <div className="flex items-center gap-2">
                      <span>{trade.marketTitle}</span>
                      <a
                        href={`https://polymarket.com/event/${trade.marketSlug}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
                        title="View on Polymarket"
                      >
                        <ArrowTopRightOnSquareIcon className="size-4" />
                      </a>
                    </div>
                  </td>
                  <td className="text-text-secondary px-6 py-4 text-sm/6">{trade.outcome}</td>
                  <td className="px-6 py-4">
                    <Badge variant={trade.side === 'BUY' ? 'success' : 'error'}>{trade.side}</Badge>
                  </td>
                  <td className="text-text-primary px-6 py-4 text-right text-sm/6 tabular-nums">
                    ${formatNumber(trade.price, 2)}
                  </td>
                  <td className="text-text-secondary px-6 py-4 text-right text-sm/6 tabular-nums">
                    {formatNumber(trade.size, 0)}
                  </td>
                  <td className="font-display text-text-bright px-6 py-4 text-right text-sm/6 font-semibold tabular-nums">
                    {formatCurrency(trade.value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {total > 10 && (
          <div className="border-border-subtle flex items-center justify-between border-t px-6 py-4">
            <div className="text-text-muted text-sm/6">
              Showing {(page - 1) * 10 + 1} to {Math.min(page * 10, total)} of {total}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="bg-bg-elevated text-text-primary hover:bg-bg-hover border-border-subtle rounded-lg border px-4 py-2 text-sm/6 font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setPage(p => p + 1)}
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
