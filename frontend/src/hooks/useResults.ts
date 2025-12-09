import { useQuery } from '@tanstack/react-query';
import { getUserResultsOptions } from '@/api/@tanstack/react-query.gen';

export function useResults(username: string) {
  return useQuery(
    getUserResultsOptions({
      path: {
        username,
      },
      query: {
        limit: 50,
        offset: 0,
      },
    })
  );
}
