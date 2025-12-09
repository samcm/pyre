import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { UserSummary } from '@/components/User/UserSummary';
import { PnlChart } from '@/components/User/PnlChart';
import { PositionsList } from '@/components/User/PositionsList';
import { ResultsTable } from '@/components/User/ResultsTable';
import { TradesTable } from '@/components/User/TradesTable';

export const Route = createFileRoute('/users/$username')({
  component: UserPage,
});

function UserPage() {
  const { username } = Route.useParams();

  return (
    <div className="space-y-10">
      <Link
        to="/"
        className="text-text-secondary hover:text-ember-400 inline-flex items-center gap-2 text-sm/6 font-medium transition-colors"
      >
        <ArrowLeftIcon className="size-4" />
        Back to Leaderboard
      </Link>

      <UserSummary username={username} />
      <PnlChart username={username} />
      <PositionsList username={username} />
      <ResultsTable username={username} />
      <TradesTable username={username} />
    </div>
  );
}
