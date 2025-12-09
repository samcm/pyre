import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { PersonaSummary } from '@/components/Persona/PersonaSummary';
import { AccountsGrid } from '@/components/Persona/AccountsGrid';
import { PersonaResultsTable } from '@/components/Persona/PersonaResultsTable';

export const Route = createFileRoute('/people/$slug')({
  component: PersonaPage,
});

function PersonaPage() {
  const { slug } = Route.useParams();

  return (
    <div className="space-y-10">
      <Link
        to="/"
        className="text-text-secondary hover:text-ember-400 inline-flex items-center gap-2 text-sm/6 font-medium transition-colors"
      >
        <ArrowLeftIcon className="size-4" />
        Back to Leaderboard
      </Link>

      <PersonaSummary slug={slug} />
      <AccountsGrid slug={slug} />
      <PersonaResultsTable slug={slug} />
    </div>
  );
}
