import { createRootRoute, Outlet } from '@tanstack/react-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Header } from '@/components/Layout/Header';
import { Container } from '@/components/Layout/Container';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000,
      refetchInterval: 60000,
    },
  },
});

export const Route = createRootRoute({
  component: () => (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-dvh bg-bg-primary">
        <Header />
        <Container>
          <Outlet />
        </Container>
      </div>
    </QueryClientProvider>
  ),
});
