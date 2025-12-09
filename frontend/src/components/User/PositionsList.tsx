import { usePositions } from '@/hooks/usePositions';
import { Card } from '@/components/Common/Card';
import { PositionCard } from './PositionCard';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';

interface PositionsListProps {
  username: string;
}

export function PositionsList({ username }: PositionsListProps) {
  const { data: positions, isLoading, error, refetch } = usePositions(username);

  if (isLoading) {
    return (
      <Card>
        <LoadingOverlay />
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <ErrorState message="Failed to load positions" retry={refetch} />
      </Card>
    );
  }

  if (!positions || positions.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary font-display py-12 text-center">No open positions</div>
      </Card>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-text-bright text-xl font-bold">Open Positions</h2>
        <p className="text-text-muted mt-1 text-sm/6">{positions.length} active positions</p>
      </div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {positions.map((position, index) => (
          <div
            key={position.id}
            style={{ animationDelay: `${index * 50}ms` }}
            className="animate-[fade-in_0.5s_ease-out]"
          >
            <PositionCard position={position} />
          </div>
        ))}
      </div>
    </div>
  );
}
