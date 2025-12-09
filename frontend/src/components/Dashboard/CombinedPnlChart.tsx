import ReactECharts from 'echarts-for-react';
import { useAllUsersPnl } from '@/hooks/useAllUsersPnl';
import { Card } from '@/components/Common/Card';
import { LoadingOverlay } from '@/components/Common/Loading';
import { ErrorState } from '@/components/Common/ErrorState';
import { formatCurrency, formatDate } from '@/utils/formatters';

// Ember-themed color palette for different users
const USER_COLORS = [
  '#d97706', // ember/amber
  '#14b8a6', // teal
  '#8b5cf6', // violet
  '#f97316', // orange
  '#ef4444', // red
  '#eab308', // gold
  '#ec4899', // pink
  '#22c55e', // green
];

const CHART_THEME = {
  grid: '#2d2d2d',
  label: '#737373',
  tooltip: {
    bg: '#1c1917',
    border: '#44403c',
    text: '#fafaf9',
  },
};

export function CombinedPnlChart() {
  const { data: usersData, isLoading, error, refetch } = useAllUsersPnl();

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
        <ErrorState message="Failed to load PNL data" retry={refetch} />
      </Card>
    );
  }

  if (!usersData || usersData.length === 0) {
    return (
      <Card>
        <div className="text-text-secondary font-display py-12 text-center">No PNL data available</div>
      </Card>
    );
  }

  // Get all unique timestamps across all users and sort them
  const allTimestamps = [...new Set(usersData.flatMap(u => u.dataPoints.map(d => d.timestamp)))].sort();

  // Build series data for each user
  const series = usersData.map((userData, index) => {
    const color = USER_COLORS[index % USER_COLORS.length];
    const dataMap = new Map(userData.dataPoints.map(d => [d.timestamp, d.realizedPnl]));

    return {
      name: userData.username,
      type: 'line',
      data: allTimestamps.map(ts => dataMap.get(ts) ?? null),
      smooth: true,
      connectNulls: true,
      lineStyle: {
        color,
        width: 2.5,
      },
      itemStyle: {
        color,
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: `${color}40` },
            { offset: 1, color: `${color}08` },
          ],
        },
      },
    };
  });

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
      backgroundColor: CHART_THEME.tooltip.bg,
      borderColor: CHART_THEME.tooltip.border,
      borderRadius: 8,
      textStyle: {
        color: CHART_THEME.tooltip.text,
        fontFamily: 'DM Sans, system-ui, sans-serif',
      },
      formatter: (params: Array<{ axisValue: string; color: string; value: number | null; seriesName: string }>) => {
        const date = formatDate(params[0].axisValue);
        let content = `<div style="font-weight: 600; margin-bottom: 8px; font-family: Outfit, system-ui, sans-serif;">${date}</div>`;
        params
          .filter(p => p.value !== null)
          .sort((a, b) => (b.value ?? 0) - (a.value ?? 0))
          .forEach(param => {
            const color = param.color;
            const value = formatCurrency(param.value ?? 0);
            content += `<div style="display: flex; align-items: center; gap: 8px; margin-top: 4px;">
              <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: ${color};"></span>
              <span>${param.seriesName}: <strong>${value}</strong></span>
            </div>`;
          });
        return content;
      },
    },
    legend: {
      data: usersData.map(u => u.username),
      top: 10,
      textStyle: {
        color: CHART_THEME.label,
        fontFamily: 'DM Sans, system-ui, sans-serif',
      },
    },
    xAxis: {
      type: 'category',
      data: allTimestamps,
      axisLine: {
        lineStyle: {
          color: CHART_THEME.grid,
        },
      },
      axisLabel: {
        color: CHART_THEME.label,
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
          color: CHART_THEME.grid,
          type: 'dashed',
        },
      },
      axisLabel: {
        color: CHART_THEME.label,
        fontFamily: 'DM Sans, system-ui, sans-serif',
        formatter: (value: number) => {
          if (Math.abs(value) >= 1000) {
            return `$${(value / 1000).toFixed(1)}k`;
          }
          return `$${value.toFixed(0)}`;
        },
      },
    },
    series,
  };

  return (
    <Card>
      <div className="mb-6">
        <h2 className="font-display text-text-bright text-xl font-bold">Realized PNL Over Time</h2>
        <p className="text-text-muted mt-1 text-sm/6">Compare performance across all tracked traders</p>
      </div>
      <ReactECharts option={option} style={{ height: '400px' }} />
    </Card>
  );
}
