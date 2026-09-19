import { useMemo, useState } from 'react';
import { Activity, BarChart3, CalendarDays, TrendingUp, Table2 } from 'lucide-react';
import { AnalyticsExportActions } from './PollExportActions';

const chartWidth = 720;
const chartHeight = 236;
const padding = { top: 18, right: 18, bottom: 36, left: 42 };

const parsePointDate = (date) => {
  const parsed = new Date(date);
  return Number.isNaN(parsed.getTime()) ? new Date(`${date}T00:00:00`) : parsed;
};

const formatDate = (date, options) => parsePointDate(date).toLocaleDateString(undefined, options);

const formatPointLabel = (date, period) => {
  if (period === 'today') return parsePointDate(date).toLocaleTimeString(undefined, { hour: 'numeric' });
  if (period === '7d') return formatDate(date, { weekday: 'short' });
  return formatDate(date, { month: 'short', day: 'numeric' });
};

const formatTooltipDate = (date) => formatDate(date, { weekday: 'short', month: 'short', day: 'numeric' });

export const VoteTimelineChart = ({
  points,
  period,
  onPeriodChange,
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  previousDayTotal = null,
}) => {
  const [activePoint, setActivePoint] = useState(null);
  const [showDataTable, setShowDataTable] = useState(false);
  const plotWidth = chartWidth - padding.left - padding.right;
  const plotHeight = chartHeight - padding.top - padding.bottom;
  const maxCount = Math.max(...points.map((point) => point.count), 1);
  const total = points.reduce((sum, point) => sum + point.count, 0);
  const activeDays = points.filter((point) => point.count > 0).length;
  const peakPoint = points.reduce((peak, point) => (point.count > (peak?.count || 0) ? point : peak), null);

  const coordinates = useMemo(() => points.map((point, index) => ({
    ...point,
    x: padding.left + (points.length <= 1 ? plotWidth / 2 : (index / (points.length - 1)) * plotWidth),
    y: padding.top + plotHeight - (point.count / maxCount) * plotHeight,
    label: formatPointLabel(point.date, period),
  })), [points, maxCount, period, plotWidth, plotHeight]);

  const line = coordinates.length > 1 ? coordinates.map((point) => `${point.x},${point.y}`).join(' ') : '';
  const area = coordinates.length > 1
    ? `${coordinates[0].x},${padding.top + plotHeight} ${line} ${coordinates.at(-1).x},${padding.top + plotHeight}`
    : '';
  const yAxisLabels = [...new Set([0, Math.ceil(maxCount / 2), maxCount])];
  const averagePerActiveDay = activeDays ? (total / activeDays).toFixed(1) : '—';
  const comparisonDelta = previousDayTotal == null || previousDayTotal === 0
    ? null
    : ((total - previousDayTotal) / previousDayTotal) * 100;

  return (
    <section className="analytics-card votes-timeline-card" aria-labelledby="votes-timeline-title">
      <div className="analytics-card-header">
        <div className="analytics-title-group">
          <div className="analytics-title-icon" aria-hidden="true"><Activity size={16} /></div>
          <div>
            <p className="analytics-eyebrow">Activity</p>
            <h2 id="votes-timeline-title">Votes Over Time</h2>
            <p className="analytics-subtitle">{total.toLocaleString()} votes in selected period</p>
          </div>
        </div>
        <div className="analytics-filters">
          <AnalyticsExportActions points={points} period={period} />
          <button type="button" className="btn btn-outline btn-sm analytics-table-toggle" onClick={() => setShowDataTable((visible) => !visible)} aria-expanded={showDataTable}>
            <Table2 size={14} /><span>{showDataTable ? 'Hide data' : 'View data'}</span>
          </button>
          <select value={period} onChange={(event) => onPeriodChange(event.target.value)} aria-label="Vote timeline period">
            <option value="today">Today</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="custom">Custom range</option>
          </select>
          {period === 'custom' && (
            <>
              <input type="date" value={startDate} onChange={(event) => onStartDateChange(event.target.value)} aria-label="Start date" />
              <input type="date" value={endDate} onChange={(event) => onEndDateChange(event.target.value)} aria-label="End date" />
            </>
          )}
        </div>
      </div>

      {total === 0 ? (
        <div className="analytics-chart-empty" role="status">
          <strong>No votes yet</strong>
          <span>Share your poll to start collecting responses.</span>
        </div>
      ) : (
        <div className="analytics-chart-wrap">
          <svg viewBox={`0 0 ${chartWidth} ${chartHeight}`} role="img" aria-label="Votes over time chart">
            <defs>
              <linearGradient id="votes-area-gradient" x1="0" x2="0" y1="0" y2="1">
                <stop offset="0%" stopColor="var(--color-primary)" stopOpacity="0.16" />
                <stop offset="100%" stopColor="var(--color-primary)" stopOpacity="0.015" />
              </linearGradient>
            </defs>
            {yAxisLabels.map((value) => {
              const y = padding.top + plotHeight - (value / maxCount) * plotHeight;
              return (
                <g key={value}>
                  <line x1={padding.left} x2={chartWidth - padding.right} y1={y} y2={y} className="chart-grid-line" />
                  <text x={padding.left - 10} y={y + 3} textAnchor="end" className="chart-y-axis-label">{value}</text>
                </g>
              );
            })}
            {activePoint && <line x1={activePoint.x} x2={activePoint.x} y1={padding.top} y2={padding.top + plotHeight} className="chart-hover-line" />}
            {area && <polygon points={area} className="chart-area" />}
            {line && <polyline points={line} className="chart-line" />}
            {coordinates.map((point) => (
              <g key={point.date} className="chart-data-point" onMouseEnter={() => setActivePoint(point)} onMouseLeave={() => setActivePoint(null)}>
                <circle cx={point.x} cy={point.y} r="12" className="chart-point-hit-area" tabIndex="0" role="button" aria-label={`${formatTooltipDate(point.date)}, ${point.count} votes. Press Enter to view details`} onFocus={() => setActivePoint(point)} onBlur={() => setActivePoint(null)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') setActivePoint(point); }} />
                <circle cx={point.x} cy={point.y} r={activePoint?.date === point.date ? '5' : '3.5'} className="chart-point" />
                <text x={point.x} y={chartHeight - 9} textAnchor="middle" className="chart-axis-label">{point.label}</text>
              </g>
            ))}
            {activePoint && (
              <g className="chart-tooltip" transform={`translate(${Math.min(Math.max(activePoint.x - 54, 4), chartWidth - 112)} ${Math.max(activePoint.y - 48, 4)})`}>
                <rect width="108" height="36" rx="6" />
                <text x="10" y="15">{formatTooltipDate(activePoint.date)}</text>
                <text x="10" y="29" className="chart-tooltip-value">{activePoint.count} {activePoint.count === 1 ? 'vote' : 'votes'}</text>
              </g>
            )}

            {showDataTable && (
              <div className="analytics-data-table-wrap">
                <table className="analytics-data-table">
                  <caption>Votes over time data</caption>
                  <thead><tr><th scope="col">Period</th><th scope="col">Votes</th></tr></thead>
                  <tbody>{points.map((point) => <tr key={point.date}><th scope="row">{formatTooltipDate(point.date)}</th><td>{point.count.toLocaleString()}</td></tr>)}</tbody>
                </table>
              </div>
            )}
          </svg>
        </div>
      )}

      {total > 0 && (
        <div className="analytics-summary" aria-label="Vote timeline summary">
          <div className="analytics-summary-item">
            <BarChart3 size={15} aria-hidden="true" />
            <span>Total Votes</span>
            <strong>{total.toLocaleString()}</strong>
            <small>In selected period</small>
          </div>
          {period === 'today' && comparisonDelta != null && (
            <div className="analytics-summary-item">
              <TrendingUp size={15} aria-hidden="true" />
              <span>vs Previous Day</span>
              <strong>{comparisonDelta >= 0 ? '+' : ''}{comparisonDelta.toFixed(0)}%</strong>
              <small>{previousDayTotal.toLocaleString()} previous votes</small>
            </div>
          )}
          {peakPoint && (
            <div className="analytics-summary-item">
              <TrendingUp size={15} aria-hidden="true" />
              <span>Peak Day</span>
              <strong>{formatPointLabel(peakPoint.date, period)}</strong>
              <small>{peakPoint.count} {peakPoint.count === 1 ? 'vote' : 'votes'}</small>
            </div>
          )}
          <div className="analytics-summary-item">
            <CalendarDays size={15} aria-hidden="true" />
            <span>Active Days</span>
            <strong>{activeDays}</strong>
            <small>Days with votes</small>
          </div>
          <div className="analytics-summary-item">
            <Activity size={15} aria-hidden="true" />
            <span>Average per Active Day</span>
            <strong>{averagePerActiveDay}</strong>
            <small>Real votes only</small>
          </div>
        </div>
      )}
    </section>
  );
};
