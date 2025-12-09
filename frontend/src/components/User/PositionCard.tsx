import { type Position } from '@/hooks/usePositions';
import { Card } from '@/components/Common/Card';
import { Badge } from '@/components/Common/Badge';
import { formatCurrency, formatNumber, pnlColor, pnlSign } from '@/utils/formatters';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

interface PositionCardProps {
  position: Position;
}

export function PositionCard({ position }: PositionCardProps) {
  const pnlVariant = position.unrealizedPnl > 0 ? 'success' : position.unrealizedPnl < 0 ? 'error' : 'default';
  const isProfit = position.unrealizedPnl > 0;

  return (
    <Card className="group hover:border-ember-500/30 transition-all duration-200">
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2">
            <h3 className="font-display text-text-bright text-base font-semibold">{position.marketTitle}</h3>
            <a
              href={`https://polymarket.com/event/${position.marketSlug}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-text-muted hover:text-ember-400 opacity-0 transition-all group-hover:opacity-100"
              title="View on Polymarket"
            >
              <ArrowTopRightOnSquareIcon className="size-4" />
            </a>
          </div>
          <Badge variant={position.side === 'YES' ? 'success' : 'error'}>{position.side}</Badge>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <span className="text-text-muted text-xs font-medium tracking-wider uppercase">Outcome</span>
            <p className="text-text-primary mt-1 text-sm/6 font-medium">{position.outcome}</p>
          </div>
          <div>
            <span className="text-text-muted text-xs font-medium tracking-wider uppercase">Size</span>
            <p className="text-text-primary mt-1 text-sm/6 font-medium tabular-nums">
              {formatNumber(position.size, 0)} shares
            </p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <span className="text-text-muted text-xs font-medium tracking-wider uppercase">Avg Price</span>
            <p className="text-text-primary mt-1 text-sm/6 font-medium tabular-nums">
              ${formatNumber(position.avgPrice, 2)}
            </p>
          </div>
          <div>
            <span className="text-text-muted text-xs font-medium tracking-wider uppercase">Current</span>
            <p className="text-text-primary mt-1 text-sm/6 font-medium tabular-nums">
              ${formatNumber(position.currentPrice, 2)}
            </p>
          </div>
        </div>

        <div className={`border-border-subtle border-t pt-4 ${isProfit ? 'glow-profit' : ''}`}>
          <div className="flex items-end justify-between">
            <div>
              <span className="text-text-muted text-xs font-medium tracking-wider uppercase">Unrealized PNL</span>
              <p className={`font-display mt-1 text-xl font-bold tabular-nums ${pnlColor(position.unrealizedPnl)}`}>
                {pnlSign(position.unrealizedPnl)}
                {formatCurrency(position.unrealizedPnl)}
              </p>
            </div>
            <Badge variant={pnlVariant}>
              {pnlSign(position.unrealizedPnlPercent)}
              {formatNumber(position.unrealizedPnlPercent, 2)}%
            </Badge>
          </div>
        </div>
      </div>
    </Card>
  );
}
