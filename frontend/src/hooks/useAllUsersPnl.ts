import { useQuery, useQueries } from '@tanstack/react-query';

export interface PnlDataPoint {
  timestamp: string;
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
}

export interface UserPnlData {
  username: string;
  dataPoints: PnlDataPoint[];
}

interface LeaderboardUser {
  username: string;
}

async function fetchLeaderboard(): Promise<LeaderboardUser[]> {
  const response = await fetch('/api/v1/leaderboard');
  if (!response.ok) {
    throw new Error('Failed to fetch leaderboard');
  }
  return response.json();
}

async function fetchUserPnl(username: string): Promise<UserPnlData> {
  const response = await fetch(`/api/v1/users/${encodeURIComponent(username)}/pnl`);
  if (!response.ok) {
    throw new Error(`Failed to fetch PnL for ${username}`);
  }
  return response.json();
}

export function useAllUsersPnl() {
  // First, fetch the leaderboard to get all usernames
  const leaderboardQuery = useQuery({
    queryKey: ['leaderboard-users'],
    queryFn: fetchLeaderboard,
    staleTime: 60000,
  });

  const usernames = leaderboardQuery.data?.map(u => u.username) ?? [];

  // Then, fetch PnL data for each user
  const pnlQueries = useQueries({
    queries: usernames.map(username => ({
      queryKey: ['user-pnl', username],
      queryFn: () => fetchUserPnl(username),
      enabled: usernames.length > 0,
      staleTime: 60000,
    })),
  });

  const isLoading = leaderboardQuery.isLoading || pnlQueries.some(q => q.isLoading);
  const error = leaderboardQuery.error || pnlQueries.find(q => q.error)?.error;

  const data: UserPnlData[] = pnlQueries.filter(q => q.data).map(q => q.data as UserPnlData);

  const refetch = () => {
    leaderboardQuery.refetch();
    pnlQueries.forEach(q => q.refetch());
  };

  return {
    data,
    isLoading,
    error,
    refetch,
  };
}
