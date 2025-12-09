import { useQuery } from '@tanstack/react-query';

export interface Trade {
  conditionId: string;
  timestamp: string;
  marketTitle: string;
  marketSlug: string;
  outcome: string;
  side: 'BUY' | 'SELL';
  price: number;
  size: number;
  value: number;
}

export interface TradesResponse {
  trades: Trade[];
  total: number;
  limit: number;
}

export function useTrades(username: string, page: number = 1, limit: number = 20) {
  return useQuery({
    queryKey: ['trades', username, page, limit],
    queryFn: async (): Promise<{ trades: Trade[]; total: number; hasMore: boolean }> => {
      const offset = (page - 1) * limit;
      const response = await fetch(
        `/api/v1/users/${encodeURIComponent(username)}/trades?limit=${limit}&offset=${offset}`,
      );
      if (!response.ok) {
        throw new Error(`Failed to fetch trades for ${username}: ${response.statusText}`);
      }
      const data: TradesResponse = await response.json();
      return {
        trades: data.trades,
        total: data.total,
        hasMore: offset + limit < data.total,
      };
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
