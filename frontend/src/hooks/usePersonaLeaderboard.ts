import { useQuery } from '@tanstack/react-query';

export interface PersonaLeaderboardEntry {
  rank: number;
  slug: string;
  displayName: string;
  image?: string;
  usernames: string[];
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
  winRate?: number;
  openPositions?: number;
}

export function usePersonaLeaderboard() {
  return useQuery({
    queryKey: ['persona-leaderboard'],
    queryFn: async (): Promise<PersonaLeaderboardEntry[]> => {
      const response = await fetch('/api/v1/personas/leaderboard');
      if (!response.ok) {
        throw new Error(`Failed to fetch persona leaderboard: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
