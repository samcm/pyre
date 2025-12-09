import { useQuery } from '@tanstack/react-query';

export interface UserDetail {
  username: string;
  addresses: string[];
  lastSynced: string;
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
  winRate: number;
  totalTrades: number;
  openPositions: number;
}

export function useUser(username: string) {
  return useQuery({
    queryKey: ['user', username],
    queryFn: async (): Promise<UserDetail> => {
      const response = await fetch(`/api/v1/users/${encodeURIComponent(username)}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch user ${username}: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
