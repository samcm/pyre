import { useQuery } from '@tanstack/react-query';

export interface PnlDataPoint {
  timestamp: string;
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
}

interface PnlHistoryResponse {
  username: string;
  dataPoints: PnlDataPoint[];
}

export function usePnlHistory(username: string, days: number = 30) {
  return useQuery({
    queryKey: ['pnl-history', username, days],
    queryFn: async (): Promise<PnlDataPoint[]> => {
      const response = await fetch(`/api/v1/users/${encodeURIComponent(username)}/pnl`);
      if (!response.ok) {
        throw new Error(`Failed to fetch PNL history for ${username}: ${response.statusText}`);
      }
      const data: PnlHistoryResponse = await response.json();
      return data.dataPoints;
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
