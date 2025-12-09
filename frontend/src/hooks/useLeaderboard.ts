import { useQuery } from '@tanstack/react-query';

export interface LeaderboardEntry {
  rank: number;
  username: string;
  profileImage?: string;
  personaDisplayName?: string;
  personaSlug?: string;
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
  winRate: number;
  totalTrades: number;
  openPositions: number;
}

export function useLeaderboard() {
  return useQuery({
    queryKey: ['leaderboard'],
    queryFn: async (): Promise<LeaderboardEntry[]> => {
      const response = await fetch('/api/v1/leaderboard');
      if (!response.ok) {
        throw new Error(`Failed to fetch leaderboard: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
