import { useQuery } from '@tanstack/react-query';

export interface PersonaDetail {
  slug: string;
  displayName: string;
  image?: string;
  usernames: string[];
  totalPnl: number;
  realizedPnl: number;
  unrealizedPnl: number;
  winRate?: number;
  totalTrades?: number;
  openPositions?: number;
}

export function usePersona(slug: string) {
  return useQuery({
    queryKey: ['persona', slug],
    queryFn: async (): Promise<PersonaDetail> => {
      const response = await fetch(`/api/v1/personas/${encodeURIComponent(slug)}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch persona ${slug}: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
    enabled: !!slug,
  });
}
