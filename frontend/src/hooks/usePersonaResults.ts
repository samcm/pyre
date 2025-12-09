import { useQuery } from '@tanstack/react-query';
import { getPersonaResultsOptions } from '@/api/@tanstack/react-query.gen';

export function usePersonaResults(slug: string) {
  return useQuery(
    getPersonaResultsOptions({
      path: {
        slug,
      },
      query: {
        limit: 50,
        offset: 0,
      },
    })
  );
}
