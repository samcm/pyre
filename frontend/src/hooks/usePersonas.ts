import { useQuery } from '@tanstack/react-query';

export interface PersonaSummary {
  slug: string;
  displayName: string;
  image?: string;
  usernames: string[];
}

export function usePersonas() {
  return useQuery({
    queryKey: ['personas'],
    queryFn: async (): Promise<PersonaSummary[]> => {
      const response = await fetch('/api/v1/personas');
      if (!response.ok) {
        throw new Error(`Failed to fetch personas: ${response.statusText}`);
      }
      return response.json();
    },
    staleTime: 30000,
    refetchInterval: 60000,
  });
}
