import { Activity, ListChecks, Vote } from 'lucide-react';

const stats = [
  { key: 'totalPolls', label: 'Total Polls', icon: ListChecks, tone: 'indigo' },
  { key: 'activePolls', label: 'Active Polls', icon: Activity, tone: 'green' },
  { key: 'totalVotes', label: 'Total Votes', icon: Vote, tone: 'slate', supporting: 'Across polls' },
];

export const DashboardStats = ({ values }) => (
  <section className="dashboard-stats" aria-label="Live poll statistics">
    {stats.map(({ key, label, icon: Icon, tone, supporting }) => (
      <article className={`dashboard-stat-card dashboard-stat-${tone}`} key={key}>
        <div className="dashboard-stat-icon" aria-hidden="true">
          <Icon size={20} />
        </div>
        <div>
          <p className="dashboard-stat-label">{label}</p>
          <p className="dashboard-stat-value">{(values[key] || 0).toLocaleString()}</p>
          <p className="dashboard-stat-live">{supporting || (key === 'activePolls' ? '● Live' : 'Live overview')}</p>
        </div>
      </article>
    ))}
  </section>
);
