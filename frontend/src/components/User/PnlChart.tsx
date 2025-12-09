import ReactECharts from 'echarts-for-react';
import { usePnlHistory } from '@/hooks/usePnlHistory';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatDate } from '@/utils/formatters';

// Ember-themed chart colors
const CHART_COLORS = {
  ember: '#d97706', // Ember orange for total
  teal: '#14b8a6', // Teal for realized
  violet: '#8b5cf6', // Violet for unrealized
  grid: '#2d2d2d',
  label: '#737373',
  tooltip: {
    bg: '#1c1917',
    border: '#44403c',
    text: '#fafaf9',
  },
};

interface PnlChartProps {
  username: string;
}

export function PnlChart({ username }: PnlChartProps) {
  const { data: pnlData, isLoading, error, refetch } = usePnlHistory(username);

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
        <ErrorState message="Failed to load PNL history" retry={refetch} />
      </Card>
    );
  }

  if (!pnlData || pnlData.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary font-display py-12 text-center">No PNL data available</div>
      </Card>
    );
  }

  const option = {
    backgroundColor: 'transparent',
    grid: {
      left: 70,
      right: 20,
      top: 60,
      bottom: 60,
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: CHART_COLORS.tooltip.bg,
      borderColor: CHART_COLORS.tooltip.border,
      borderRadius: 8,
      textStyle: {
        color: CHART_COLORS.tooltip.text,
        fontFamily: 'DM Sans, system-ui, sans-serif',
      },
      formatter: (params: Array<{ axisValue: string; color: string; value: number; seriesName: string }>) => {
        const date = formatDate(params[0].axisValue);
        let content = `<div style="font-weight: 600; margin-bottom: 8px; font-family: Outfit, system-ui, sans-serif;">${date}</div>`;
        params.forEach(param => {
          const color = param.color;
          const value = formatCurrency(param.value);
          content += `<div style="display: flex; align-items: center; gap: 8px; margin-top: 4px;">
            <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: ${color};"></span>
            <span>${param.seriesName}: <strong>${value}</strong></span>
          </div>`;
        });
        return content;
      },
    },
    legend: {
      data: ['Total PNL', 'Realized PNL', 'Unrealized PNL'],
      top: 10,
      textStyle: {
        color: CHART_COLORS.label,
        fontFamily: 'DM Sans, system-ui, sans-serif',
      },
    },
    xAxis: {
      type: 'category',
      data: pnlData.map(d => d.timestamp),
      axisLine: {
        lineStyle: {
          color: CHART_COLORS.grid,
        },
      },
      axisLabel: {
        color: CHART_COLORS.label,
        fontFamily: 'DM Sans, system-ui, sans-serif',
        formatter: (value: string) => {
          const date = new Date(value);
          return `${date.getMonth() + 1}/${date.getDate()}`;
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLine: {
        show: false,
      },
      splitLine: {
        lineStyle: {
          color: CHART_COLORS.grid,
          type: 'dashed',
        },
      },
      axisLabel: {
        color: CHART_COLORS.label,
        fontFamily: 'DM Sans, system-ui, sans-serif',
        formatter: (value: number) => {
          if (Math.abs(value) >= 1000) {
            return `$${(value / 1000).toFixed(1)}k`;
          }
          return `$${value.toFixed(0)}`;
        },
      },
    },
    series: [
      {
        name: 'Total PNL',
        type: 'line',
        data: pnlData.map(d => d.totalPnl),
        smooth: true,
        lineStyle: {
          color: CHART_COLORS.ember,
          width: 3,
        },
        itemStyle: {
          color: CHART_COLORS.ember,
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(217, 119, 6, 0.35)' },
              { offset: 1, color: 'rgba(217, 119, 6, 0.05)' },
            ],
          },
        },
      },
      {
        name: 'Realized PNL',
        type: 'line',
        data: pnlData.map(d => d.realizedPnl),
        smooth: true,
        lineStyle: {
          color: CHART_COLORS.teal,
          width: 2,
        },
        itemStyle: {
          color: CHART_COLORS.teal,
        },
      },
      {
        name: 'Unrealized PNL',
        type: 'line',
        data: pnlData.map(d => d.unrealizedPnl),
        smooth: true,
        lineStyle: {
          color: CHART_COLORS.violet,
          width: 2,
        },
        itemStyle: {
          color: CHART_COLORS.violet,
        },
      },
    ],
  };

  return (
    <Card>
      <div className="mb-6">
        <h2 className="font-display text-text-bright text-xl font-bold">PNL Over Time</h2>
        <p className="text-text-muted mt-1 text-sm/6">Track your profit and loss history</p>
      </div>
      <ReactECharts option={option} style={{ height: '400px' }} />
    </Card>
  );
}
