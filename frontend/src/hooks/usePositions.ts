import { useQuery } from '@tanstack/react-query';

export interface Position {
  id: string;
  marketTitle: string;
  marketSlug: string;
  outcome: string;
  side: 'YES' | 'NO';
  size: number;
  avgPrice: number;
  currentPrice: number;
  unrealizedPnl: number;
  unrealizedPnlPercent: number;
  value: number;
}

export function usePositions(username: string) {
  return useQuery({
    queryKey: ['positions', username],
    queryFn: async (): Promise<Position[]> => {
      const response = await fetch(`/api/v1/users/${encodeURIComponent(username)}/positions`);
      if (!response.ok) {
        throw new Error(`Failed to fetch positions for ${username}: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
