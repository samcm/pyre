import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowUpIcon, ArrowDownIcon, ArrowTopRightOnSquareIcon, UserIcon, UsersIcon } from '@heroicons/react/24/solid';
import { useLeaderboard, type LeaderboardEntry } from '@/hooks/useLeaderboard';
import { usePersonaLeaderboard, type PersonaLeaderboardEntry } from '@/hooks/usePersonaLeaderboard';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatPercent, pnlColor, pnlSign } from '@/utils/formatters';

const POLYMARKET_PROFILE_URL = 'https://polymarket.com/profile/@';

type ViewMode = 'username' | 'person';
type SortField = 'rank' | 'totalPnl' | 'realizedPnl' | 'unrealizedPnl' | 'winRate' | 'totalTrades';
type SortDirection = 'asc' | 'desc';

interface SortIconProps {
  field: SortField;
  currentSortField: SortField;
  sortDirection: SortDirection;
}

function SortIcon({ field, currentSortField, sortDirection }: SortIconProps) {
  if (currentSortField !== field) return null;
  return sortDirection === 'asc' ? (
    <ArrowUpIcon className="text-ember-400 size-3.5" />
  ) : (
    <ArrowDownIcon className="text-ember-400 size-3.5" />
  );
}

function RankBadge({ rank }: { rank: number }) {
  if (rank === 1) {
    return (
      <div className="from-ember-400 to-ember-600 font-display shadow-ember-500/20 flex size-8 items-center justify-center rounded-lg bg-linear-to-br text-sm font-bold text-white shadow-lg">
        1
      </div>
    );
  }
  if (rank === 2) {
    return (
      <div className="from-ash-300 to-ash-500 font-display text-ash-900 flex size-8 items-center justify-center rounded-lg bg-linear-to-br text-sm font-bold">
        2
      </div>
    );
  }
  if (rank === 3) {
    return (
      <div className="from-ember-700 to-ember-900 font-display text-ember-200 flex size-8 items-center justify-center rounded-lg bg-linear-to-br text-sm font-bold">
        3
      </div>
    );
  }
  return (
    <div className="bg-bg-elevated font-display text-text-secondary flex size-8 items-center justify-center rounded-lg text-sm font-semibold">
      {rank}
    </div>
  );
}

interface ViewToggleProps {
  viewMode: ViewMode;
  setViewMode: (mode: ViewMode) => void;
}

function ViewToggle({ viewMode, setViewMode }: ViewToggleProps) {
  return (
    <div className="bg-bg-elevated inline-flex rounded-lg p-1">
      <button
        onClick={() => setViewMode('person')}
        className={`inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all ${
          viewMode === 'person'
            ? 'bg-ember-600 text-white shadow-sm'
            : 'text-text-muted hover:text-text-secondary'
        }`}
      >
        <UsersIcon className="size-4" />
        By Person
      </button>
      <button
        onClick={() => setViewMode('username')}
        className={`inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all ${
          viewMode === 'username'
            ? 'bg-ember-600 text-white shadow-sm'
            : 'text-text-muted hover:text-text-secondary'
        }`}
      >
        <UserIcon className="size-4" />
        By Username
      </button>
    </div>
  );
}

export function LeaderboardTable() {
  const [viewMode, setViewMode] = useState<ViewMode>('person');
  const { data: userLeaderboard, isLoading: userLoading, error: userError, refetch: userRefetch } = useLeaderboard();
  const { data: personaLeaderboard, isLoading: personaLoading, error: personaError, refetch: personaRefetch } = usePersonaLeaderboard();

  const [sortField, setSortField] = useState<SortField>('rank');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  const isLoading = viewMode === 'username' ? userLoading : personaLoading;
  const error = viewMode === 'username' ? userError : personaError;
  const refetch = viewMode === 'username' ? userRefetch : personaRefetch;

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <ViewToggle viewMode={viewMode} setViewMode={setViewMode} />
        </div>
        <Card>
          <LoadingOverlay />
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <ViewToggle viewMode={viewMode} setViewMode={setViewMode} />
        </div>
        <Card>
          <ErrorState message="Failed to load leaderboard" retry={refetch} />
        </Card>
      </div>
    );
  }

  const leaderboard = viewMode === 'username' ? userLeaderboard : personaLeaderboard;

  if (!leaderboard || leaderboard.length === 0) {
    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <ViewToggle viewMode={viewMode} setViewMode={setViewMode} />
        </div>
        <Card>
          <div className="text-text-secondary py-16 text-center">
            <p className="font-display text-lg">No data available</p>
            <p className="text-text-muted mt-2 text-sm">Add users to start tracking</p>
          </div>
        </Card>
      </div>
    );
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedData = [...leaderboard].sort((a, b) => {
    const aValue = a[sortField as keyof typeof a] ?? 0;
    const bValue = b[sortField as keyof typeof b] ?? 0;
    const multiplier = sortDirection === 'asc' ? 1 : -1;
    return aValue > bValue ? multiplier : -multiplier;
  });

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <ViewToggle viewMode={viewMode} setViewMode={setViewMode} />
      </div>
      <Card padding={false}>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border-subtle bg-bg-deep/50 border-b">
                <th
                  onClick={() => handleSort('rank')}
                  className="text-text-muted hover:text-text-secondary cursor-pointer px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase transition-colors"
                >
                  <div className="flex items-center gap-2">
                    Rank
                    <SortIcon field="rank" currentSortField={sortField} sortDirection={sortDirection} />
                  </div>
                </th>
                <th className="text-text-muted px-6 py-4 text-left text-xs font-semibold tracking-wider uppercase">
                  {viewMode === 'username' ? 'Trader' : 'Person'}
                </th>
                <th
                  onClick={() => handleSort('totalPnl')}
                  className="text-text-muted hover:text-text-secondary cursor-pointer px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase transition-colors"
                >
                  <div className="flex items-center justify-end gap-2">
                    Total PnL
                    <SortIcon field="totalPnl" currentSortField={sortField} sortDirection={sortDirection} />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('realizedPnl')}
                  className="text-text-muted hover:text-text-secondary hidden cursor-pointer px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase transition-colors md:table-cell"
                >
                  <div className="flex items-center justify-end gap-2">
                    Realized
                    <SortIcon field="realizedPnl" currentSortField={sortField} sortDirection={sortDirection} />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('unrealizedPnl')}
                  className="text-text-muted hover:text-text-secondary hidden cursor-pointer px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase transition-colors lg:table-cell"
                >
                  <div className="flex items-center justify-end gap-2">
                    Unrealized
                    <SortIcon field="unrealizedPnl" currentSortField={sortField} sortDirection={sortDirection} />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('winRate')}
                  className="text-text-muted hover:text-text-secondary hidden cursor-pointer px-6 py-4 text-right text-xs font-semibold tracking-wider uppercase transition-colors sm:table-cell"
                >
                  <div className="flex items-center justify-end gap-2">
                    Win Rate
                    <SortIcon field="winRate" currentSortField={sortField} sortDirection={sortDirection} />
                  </div>
                </th>
              </tr>
            </thead>
            <tbody className="divide-border-subtle divide-y">
              {viewMode === 'username' ? (
                (sortedData as LeaderboardEntry[]).map((entry, index) => (
                  <tr
                    key={entry.username}
                    className="table-row-ember group transition-all duration-200"
                    style={{ animationDelay: `${index * 50}ms` }}
                  >
                    <td className="px-6 py-4">
                      <RankBadge rank={entry.rank} />
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        {entry.profileImage ? (
                          <img
                            src={entry.profileImage}
                            alt={entry.username}
                            className="size-8 rounded-full object-cover ring-2 ring-bg-deep"
                          />
                        ) : (
                          <div className="bg-bg-elevated text-text-muted flex size-8 items-center justify-center rounded-full ring-2 ring-bg-deep">
                            <UserIcon className="size-4" />
                          </div>
                        )}
                        <div className="flex items-center gap-2">
                          <Link
                            to="/users/$username"
                            params={{ username: entry.username }}
                            className="font-display text-text-primary hover:text-ember-400 text-base font-semibold transition-colors"
                          >
                            {entry.username}
                          </Link>
                          <a
                            href={`${POLYMARKET_PROFILE_URL}${entry.username}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
                            title="View on Polymarket"
                          >
                            <ArrowTopRightOnSquareIcon className="size-4" />
                          </a>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span className={`font-display text-lg font-bold tabular-nums ${pnlColor(entry.totalPnl)}`}>
                        {pnlSign(entry.totalPnl)}
                        {formatCurrency(entry.totalPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right md:table-cell">
                      <span className={`text-sm/6 font-medium tabular-nums ${pnlColor(entry.realizedPnl)}`}>
                        {pnlSign(entry.realizedPnl)}
                        {formatCurrency(entry.realizedPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right lg:table-cell">
                      <span className={`text-sm/6 font-medium tabular-nums ${pnlColor(entry.unrealizedPnl)}`}>
                        {pnlSign(entry.unrealizedPnl)}
                        {formatCurrency(entry.unrealizedPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right sm:table-cell">
                      <span className="text-text-secondary text-sm/6 tabular-nums">{formatPercent(entry.winRate)}</span>
                    </td>
                  </tr>
                ))
              ) : (
                (sortedData as PersonaLeaderboardEntry[]).map((entry, index) => (
                  <tr
                    key={entry.slug}
                    className="table-row-ember group transition-all duration-200"
                    style={{ animationDelay: `${index * 50}ms` }}
                  >
                    <td className="px-6 py-4">
                      <RankBadge rank={entry.rank} />
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        {entry.image ? (
                          <img
                            src={entry.image}
                            alt={entry.displayName}
                            className="size-8 rounded-full object-cover ring-2 ring-bg-deep"
                          />
                        ) : (
                          <div className="bg-bg-elevated text-text-muted flex size-8 items-center justify-center rounded-full ring-2 ring-bg-deep">
                            <UsersIcon className="size-4" />
                          </div>
                        )}
                        <div className="flex flex-col gap-1">
                          <Link
                            to="/people/$slug"
                            params={{ slug: entry.slug }}
                            className="font-display text-text-primary hover:text-ember-400 text-base font-semibold transition-colors"
                          >
                            {entry.displayName}
                          </Link>
                          <span className="text-text-muted text-xs">
                            {entry.usernames.length} account{entry.usernames.length !== 1 ? 's' : ''}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span className={`font-display text-lg font-bold tabular-nums ${pnlColor(entry.totalPnl)}`}>
                        {pnlSign(entry.totalPnl)}
                        {formatCurrency(entry.totalPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right md:table-cell">
                      <span className={`text-sm/6 font-medium tabular-nums ${pnlColor(entry.realizedPnl)}`}>
                        {pnlSign(entry.realizedPnl)}
                        {formatCurrency(entry.realizedPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right lg:table-cell">
                      <span className={`text-sm/6 font-medium tabular-nums ${pnlColor(entry.unrealizedPnl)}`}>
                        {pnlSign(entry.unrealizedPnl)}
                        {formatCurrency(entry.unrealizedPnl)}
                      </span>
                    </td>
                    <td className="hidden px-6 py-4 text-right sm:table-cell">
                      <span className="text-text-secondary text-sm/6 tabular-nums">
                        {formatPercent((entry.winRate ?? 0) * 100)}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
