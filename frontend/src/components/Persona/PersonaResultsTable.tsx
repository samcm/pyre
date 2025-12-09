import { usePersonaResults } from '@/hooks/usePersonaResults';
import { Card } from '@/components/Common/Card';
import { Badge } from '@/components/Common/Badge';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatDateTime } from '@/utils/formatters';
import { ArrowTopRightOnSquareIcon, CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/solid';
import { Link } from '@tanstack/react-router';

interface PersonaResultsTableProps {
  slug: string;
}

export function PersonaResultsTable({ slug }: PersonaResultsTableProps) {
  const { data, isLoading, error, refetch } = usePersonaResults(slug);

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
        <ErrorState message="Failed to load results" retry={refetch} />
      </Card>
    );
  }

  if (!data || !data.results || data.results.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary font-display py-12 text-center">No resolved positions</div>
      </Card>
    );
  }

  const { results, total } = data;

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-text-bright text-xl font-bold">Results</h2>
        <p className="text-text-muted mt-1 text-sm/6">{total} resolved positions across all accounts</p>
      </div>
      <Card padding={false}>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border-subtle bg-bg-deep/50 border-b">
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Result
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Account
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Market
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  Outcome
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Invested
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Realized PnL
                </th>
                <th className="text-text-muted px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase">
                  Resolved
                </th>
              </tr>
            </thead>
            <tbody className="divide-border-subtle divide-y">
              {results.map((result, index) => {
                const isWin = result.realizedPnl > 0;
                const isLoss = result.realizedPnl < 0;

                return (
                  <tr
                    key={result.id}
                    className="table-row-ember group"
                    style={{ animationDelay: `${index * 30}ms` }}
                  >
                    <td className="px-6 py-4">
                      {isWin && (
                        <div className="flex items-center gap-2">
                          <CheckCircleIcon className="size-5 text-success-400" />
                          <span className="text-text-bright text-sm/6 font-medium">Win</span>
                        </div>
                      )}
                      {isLoss && (
                        <div className="flex items-center gap-2">
                          <XCircleIcon className="size-5 text-error-400" />
                          <span className="text-text-bright text-sm/6 font-medium">Loss</span>
                        </div>
                      )}
                      {!isWin && !isLoss && (
                        <div className="flex items-center gap-2">
                          <span className="text-text-muted text-sm/6 font-medium">Break Even</span>
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <Link
                        to="/users/$username"
                        params={{ username: result.username }}
                        className="text-ember-400 hover:text-ember-300 text-sm/6 font-medium transition-colors"
                      >
                        {result.username}
                      </Link>
                    </td>
                    <td className="text-text-primary px-6 py-4 text-sm/6">
                      <div className="flex items-center gap-2">
                        <span className="line-clamp-2 max-w-md">{result.marketTitle}</span>
                        {result.marketSlug && (
                          <a
                            href={`https://polymarket.com/event/${result.marketSlug}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
                            title="View on Polymarket"
                          >
                            <ArrowTopRightOnSquareIcon className="size-4" />
                          </a>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <Badge variant={result.outcome === 'Yes' ? 'success' : 'error'}>{result.outcome}</Badge>
                    </td>
                    <td className="text-text-secondary px-6 py-4 text-right text-sm/6 tabular-nums">
                      {result.initialValue !== null && result.initialValue !== undefined
                        ? formatCurrency(result.initialValue)
                        : '-'}
                    </td>
                    <td
                      className={`font-display px-6 py-4 text-right text-sm/6 font-semibold tabular-nums ${
                        isWin ? 'text-success-400' : isLoss ? 'text-error-400' : 'text-text-bright'
                      }`}
                    >
                      {formatCurrency(result.realizedPnl)}
                    </td>
                    <td className="text-text-secondary px-6 py-4 text-right text-sm/6 tabular-nums">
                      {result.resolutionDate ? formatDateTime(result.resolutionDate) : '-'}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
