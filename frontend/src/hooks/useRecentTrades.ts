import { useQuery } from '@tanstack/react-query';

export interface RecentTrade {
  id: string;
  username: string;
  personaDisplayName?: string;
  personaSlug?: string;
  timestamp: string;
  conditionId: string;
  marketTitle: string;
  marketSlug: string;
  outcome: string;
  side: 'BUY' | 'SELL';
  price: number;
  size: number;
  value: number;
}

export interface RecentTradesFilters {
  limit?: number;
  offset?: number;
  username?: string;
  side?: 'BUY' | 'SELL';
  minValue?: number;
  sortBy?: 'timestamp' | 'value' | 'size';
  sortDirection?: 'asc' | 'desc';
}

interface RecentTradesResponse {
  trades: RecentTrade[];
  total: number;
  limit: number;
  offset: number;
}

async function fetchRecentTrades(filters: RecentTradesFilters): Promise<RecentTradesResponse> {
  const params = new URLSearchParams();

  if (filters.limit) params.set('limit', filters.limit.toString());
  if (filters.offset) params.set('offset', filters.offset.toString());
  if (filters.username) params.set('username', filters.username);
  if (filters.side) params.set('side', filters.side);
  if (filters.minValue) params.set('minValue', filters.minValue.toString());
  if (filters.sortBy) params.set('sortBy', filters.sortBy);
  if (filters.sortDirection) params.set('sortDirection', filters.sortDirection);

  const response = await fetch(`/api/v1/trades?${params.toString()}`);

  if (!response.ok) {
    throw new Error('Failed to fetch recent trades');
  }

  return response.json();
}

export function useRecentTrades(filters: RecentTradesFilters = {}) {
  const defaultFilters: RecentTradesFilters = {
    limit: 20,
    offset: 0,
    sortBy: 'timestamp',
    sortDirection: 'desc',
    ...filters,
  };

  return useQuery({
    queryKey: ['recent-trades', defaultFilters],
    queryFn: () => fetchRecentTrades(defaultFilters),
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
