import { useQuery } from '@tanstack/react-query';

export interface PersonaAccount {
  username: string;
  addresses: string[];
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
  winRate?: number;
  totalTrades?: number;
  openPositions?: number;
}

export function usePersonaAccounts(slug: string) {
  return useQuery({
    queryKey: ['persona-accounts', slug],
    queryFn: async (): Promise<PersonaAccount[]> => {
      const response = await fetch(`/api/v1/personas/${encodeURIComponent(slug)}/accounts`);
      if (!response.ok) {
        throw new Error(`Failed to fetch accounts for persona ${slug}: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
    enabled: !!slug,
  });
}
