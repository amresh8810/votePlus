import { useState, useEffect, useMemo, useRef } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../services/api';
import { PollCard } from '../components/PollCard';
import { Loading } from '../components/Loading';
import { ErrorMessage } from '../components/ErrorMessage';
import { DashboardStats } from '../components/DashboardStats';
import { VoteTimelineChart } from '../components/VoteTimelineChart';
import { PollPerformanceChart } from '../components/PollPerformanceChart';
import { PollTable } from '../components/PollTable';
import { AdvancedFeatures } from '../components/AdvancedFeatures';
import { LiveResultsModal } from '../components/LiveResultsModal';
import { useDashboardRealtimeStats } from '../hooks/useDashboardRealtimeStats';
import { useUi } from '../context/UiContext';
import { PlusCircle, Inbox, Search, SlidersHorizontal, X } from 'lucide-react';

export const Dashboard = () => {
  const { user } = useAuth();
  const { toast } = useUi();
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [timeline, setTimeline] = useState([]);
  const [timelinePeriod, setTimelinePeriod] = useState('today');
  const [timelineStart, setTimelineStart] = useState('');
  const [timelineEnd, setTimelineEnd] = useState('');
  const [previousDayTotal, setPreviousDayTotal] = useState(null);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [dateFilter, setDateFilter] = useState('all');
  const [customDate, setCustomDate] = useState('');
  const [sortBy, setSortBy] = useState('newest');
  const [selectedResultsPoll, setSelectedResultsPoll] = useState(null);
  const [liveToast, setLiveToast] = useState('');
  const [pendingDelete, setPendingDelete] = useState(null);
  const realtimeReady = useRef(false);
  const stats = useDashboardRealtimeStats(polls);

  useEffect(() => {
    let cancelled = false;

    const fetchPolls = async () => {
      if (!user?.id) {
        setPolls([]);
        setLoading(false);
        return;
      }

      setLoading(true);
      setError('');
      setPolls([]);
      setTimeline([]);
      setSelectedResultsPoll(null);
      setSearch('');
      setStatusFilter('all');
      setDateFilter('all');
      setCustomDate('');
      setSortBy('newest');

      try {
        const data = await api.get('/polls');
        if (!cancelled) {
          setPolls(data?.polls || []);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err.message || 'Failed to load polls.');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    fetchPolls();
    return () => {
      cancelled = true;
    };
  }, [user?.id]);

  useEffect(() => {
    let cancelled = false;
    const fetchTimeline = async () => {
      if (!user?.id) {
        setTimeline([]);
        return;
      }
      if (timelinePeriod === 'custom' && (!timelineStart || !timelineEnd)) return;
      const params = new URLSearchParams({ period: timelinePeriod });
      if (timelinePeriod === 'custom') {
        params.set('start', timelineStart);
        params.set('end', timelineEnd);
      }
      try {
        const data = await api.get(`/polls/analytics/votes-over-time?${params.toString()}`);
        if (!cancelled) setTimeline(data?.points || []);
      } catch (err) {
        if (!cancelled) setError(err.message || 'Failed to load vote analytics.');
      }
    };
    fetchTimeline();
    const refreshTimer = window.setInterval(fetchTimeline, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(refreshTimer);
    };
  }, [user?.id, timelinePeriod, timelineStart, timelineEnd, polls, stats.realtimeVersion]);

  useEffect(() => {
    let cancelled = false;
    if (!user?.id || timelinePeriod !== 'today') {
      setPreviousDayTotal(null);
      return undefined;
    }
    const previousDay = new Date();
    previousDay.setDate(previousDay.getDate() - 1);
    const date = previousDay.toISOString().slice(0, 10);
    api.get(`/polls/analytics/votes-over-time?period=custom&start=${date}&end=${date}`)
      .then((data) => {
        if (!cancelled) setPreviousDayTotal((data?.points || []).reduce((sum, point) => sum + (point.count || 0), 0));
      })
      .catch(() => {
        if (!cancelled) setPreviousDayTotal(null);
      });
    return () => { cancelled = true; };
  }, [user?.id, timelinePeriod, stats.realtimeVersion]);

  useEffect(() => {
    if (!realtimeReady.current) {
      realtimeReady.current = true;
      return undefined;
    }
    const update = stats.lastRealtimeUpdate;
    const poll = polls.find((item) => item.id === update?.pollId);
    setLiveToast(`${poll?.question || 'Poll'} · ${update?.totalVotes ?? 0} ${update?.totalVotes === 1 ? 'vote' : 'votes'}`);
    const timeoutId = window.setTimeout(() => setLiveToast(''), 2800);
    return () => window.clearTimeout(timeoutId);
  }, [stats.realtimeVersion, stats.lastRealtimeUpdate, polls]);

  useEffect(() => {
    const handleShortcut = (event) => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        document.querySelector('.poll-search-field input')?.focus();
      }
      if (event.key === '?' && !['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName)) {
        toast('Shortcuts: Ctrl/Cmd + K focuses poll search');
      }
    };
    window.addEventListener('keydown', handleShortcut);
    return () => window.removeEventListener('keydown', handleShortcut);
  }, [toast]);

  const connectionValues = Object.values(stats.connectionStatusByPoll);
  const realtimeStatus = connectionValues.length === 0 || connectionValues.every((status) => status === 'connected')
    ? 'Live updates connected'
    : connectionValues.some((status) => status === 'reconnecting')
      ? 'Reconnecting to live updates'
      : 'Connecting to live updates';

  const handleClosePoll = async (pollId) => {
    if (!window.confirm('Are you sure you want to close this poll? Audiences will no longer be able to submit votes.')) {
      return;
    }

    try {
      await api.patch(`/polls/${pollId}/close`);
      // Update local state
      setPolls((prev) =>
        prev.map((p) => (p.id === pollId ? { ...p, status: 'closed' } : p))
      );
    } catch (err) {
      alert(err.message || 'Failed to close poll.');
    }
  };

  const handleDeletePoll = async (pollId) => {
    if (!window.confirm('Delete this poll permanently? Its votes and public link will no longer be available.')) {
      return;
    }

    const deletedPoll = polls.find((poll) => poll.id === pollId);
    setPolls((prev) => prev.filter((poll) => poll.id !== pollId));
    setPendingDelete({ poll: deletedPoll, timer: window.setTimeout(async () => {
      try {
        await api.delete(`/polls/${pollId}`);
        setPendingDelete(null);
        toast('Poll deleted');
      } catch (err) {
        setPolls((prev) => [...prev, deletedPoll]);
        setPendingDelete(null);
        setError(err.message || 'Failed to delete poll.');
      }
    }, 5000) });
    toast('Poll removed. Undo within 5 seconds.', 'info');
  };

  const undoDelete = () => {
    if (!pendingDelete) return;
    window.clearTimeout(pendingDelete.timer);
    setPolls((prev) => [...prev, pendingDelete.poll]);
    setPendingDelete(null);
    toast('Poll restored', 'success');
  };

  const filteredPolls = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();
    const now = new Date();
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const dateStart = {
      today: startOfToday,
      '7d': new Date(startOfToday.getTime() - 6 * 24 * 60 * 60 * 1000),
      '30d': new Date(startOfToday.getTime() - 29 * 24 * 60 * 60 * 1000),
    }[dateFilter];

    const result = polls.filter((poll) => {
      const createdAt = poll.created_at ? new Date(poll.created_at) : null;
      const matchesSearch = !normalizedSearch || poll.question.toLowerCase().includes(normalizedSearch);
      const matchesStatus = statusFilter === 'all' || poll.status === statusFilter;
      let matchesDate = true;
      if (dateFilter === 'custom') {
        matchesDate = Boolean(customDate && createdAt && createdAt.toISOString().slice(0, 10) === customDate);
      } else if (dateStart) {
        matchesDate = Boolean(createdAt && createdAt >= dateStart);
      }
      return matchesSearch && matchesStatus && matchesDate;
    });

    return result.sort((a, b) => {
      const votesA = stats.votesByPoll[a.id] || 0;
      const votesB = stats.votesByPoll[b.id] || 0;
      if (sortBy === 'votes') return votesB - votesA;
      if (sortBy === 'active') {
        if (a.status !== b.status) return a.status === 'active' ? -1 : 1;
        return new Date(b.updated_at || b.created_at) - new Date(a.updated_at || a.created_at);
      }
      if (sortBy === 'alphabetical') return a.question.localeCompare(b.question, undefined, { sensitivity: 'base' });
      return new Date(b.created_at) - new Date(a.created_at);
    });
  }, [polls, search, statusFilter, dateFilter, customDate, sortBy, stats.votesByPoll]);

  const hasActiveFilters = Boolean(search || statusFilter !== 'all' || dateFilter !== 'all' || sortBy !== 'newest');
  const clearFilters = () => {
    setSearch('');
    setStatusFilter('all');
    setDateFilter('all');
    setCustomDate('');
    setSortBy('newest');
  };

  return (
    <div className="dashboard-container page-layout">
      {liveToast && <div className="live-vote-toast" role="status">✓ {liveToast}</div>}
      {pendingDelete && <div className="undo-delete-toast" role="status">Poll removed <button type="button" onClick={undoDelete}>Undo</button></div>}
      <header className="dashboard-header">
        <div>
          <h1 className="page-title">Creator Dashboard</h1>
          <p className="page-subtitle">Welcome back, {user?.name || user?.email}!</p>
        </div>
        <Link to="/polls/create" className="btn btn-primary btn-md">
          <PlusCircle size={18} />
          <span>Create New Poll</span>
        </Link>
      </header>

      <ErrorMessage message={error} onClose={() => setError('')} />
      <DashboardStats values={stats} />
      <section className="dashboard-analytics" aria-labelledby="dashboard-analytics-title">
        <div className="dashboard-section-heading">
          <div>
            <p className="analytics-eyebrow">Insights</p>
            <h2 id="dashboard-analytics-title">Analytics</h2>
            <p>Track voting activity and compare your polls.</p>
            <span className={`dashboard-realtime-status ${realtimeStatus === 'Live updates connected' ? 'dashboard-realtime-status-connected' : ''}`}>
              <span aria-hidden="true" />{realtimeStatus}
            </span>
          </div>
        </div>
        <VoteTimelineChart
          points={timeline}
          period={timelinePeriod}
          onPeriodChange={setTimelinePeriod}
          startDate={timelineStart}
          endDate={timelineEnd}
          onStartDateChange={setTimelineStart}
          onEndDateChange={setTimelineEnd}
          previousDayTotal={previousDayTotal}
        />
        <PollPerformanceChart polls={filteredPolls} votesByPoll={stats.votesByPoll} />
      </section>

      <section className="poll-management-toolbar" aria-labelledby="polls-heading">
        <div className="poll-management-heading">
          <div>
            <h2 id="polls-heading">Polls</h2>
            <p>Manage and monitor your polls <span className="dashboard-live-indicator"><span aria-hidden="true">●</span> Live</span></p>
          </div>
          {polls.length > 0 && (
            <span className="poll-result-count">
              {filteredPolls.length} {filteredPolls.length === 1 ? 'poll' : 'polls'}
            </span>
          )}
        </div>
        <div className="poll-search-row">
          <label className="poll-search-field">
            <span className="sr-only">Search polls by question</span>
            <Search size={17} aria-hidden="true" />
            <input
              type="search"
              placeholder="Search polls..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </label>
          <div className="poll-filter-group" aria-label="Poll status filter">
            <SlidersHorizontal size={16} aria-hidden="true" />
            {['all', 'active', 'closed'].map((status) => (
              <button
                key={status}
                type="button"
                className={`filter-segment ${statusFilter === status ? 'filter-segment-selected' : ''}`}
                aria-pressed={statusFilter === status}
                onClick={() => setStatusFilter(status)}
              >
                {status[0].toUpperCase() + status.slice(1)}
              </button>
            ))}
          </div>
          <label className="poll-select-field">
            <span>Date</span>
            <select value={dateFilter} onChange={(event) => setDateFilter(event.target.value)}>
              <option value="all">All time</option>
              <option value="today">Today</option>
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
              <option value="custom">Custom date</option>
            </select>
          </label>
          {dateFilter === 'custom' && (
            <label className="poll-select-field">
              <span>Date</span>
              <input type="date" value={customDate} onChange={(event) => setCustomDate(event.target.value)} />
            </label>
          )}
          <label className="poll-select-field">
            <span>Sort</span>
            <select value={sortBy} onChange={(event) => setSortBy(event.target.value)}>
              <option value="newest">Newest</option>
              <option value="votes">Most votes</option>
              <option value="active">Most active</option>
              <option value="alphabetical">A-Z</option>
            </select>
          </label>
          {hasActiveFilters && (
            <button type="button" className="clear-filters-button" onClick={clearFilters}>
              <X size={15} />
              Clear filters
            </button>
          )}
        </div>
      </section>

      {loading ? (
        <Loading message="Loading your polls..." />
      ) : polls.length === 0 ? (
        <div className="empty-state-card">
          <div className="empty-state-icon">
            <Inbox size={48} className="text-slate-500" />
          </div>
          <h3 className="empty-state-title">No polls yet</h3>
          <p className="empty-state-description">
            Create your first poll to get started.
          </p>
          <Link to="/polls/create" className="btn btn-primary btn-md mt-4">
            <PlusCircle size={18} />
            <span>Create Your First Poll</span>
          </Link>
        </div>
      ) : filteredPolls.length === 0 ? (
        <div className="empty-state-card filtered-empty-state">
          <div className="empty-state-icon"><Search size={42} /></div>
          <h3 className="empty-state-title">No matching polls</h3>
          <p className="empty-state-description">Try changing your search or filters.</p>
          <button type="button" className="btn btn-primary btn-md mt-4" onClick={clearFilters}>
            Clear filters
          </button>
        </div>
      ) : (
        <>
          <PollTable
            polls={filteredPolls}
            votesByPoll={stats.votesByPoll}
            lastActivityByPoll={stats.lastActivityByPoll}
            onClosePoll={handleClosePoll}
            onDeletePoll={handleDeletePoll}
            onViewResults={setSelectedResultsPoll}
          />
          <div className="polls-grid">
            {filteredPolls.map((poll) => (
              <PollCard
                key={poll.id}
                poll={poll}
                votes={stats.votesByPoll[poll.id] || 0}
                counts={stats.countsByPoll[poll.id] || {}}
                onClosePoll={handleClosePoll}
                onDeletePoll={handleDeletePoll}
                onViewResults={setSelectedResultsPoll}
              />
            ))}
          </div>
        </>
      )}
      <AdvancedFeatures />
      {selectedResultsPoll && (
        <LiveResultsModal
          poll={selectedResultsPoll}
          counts={stats.countsByPoll[selectedResultsPoll.id] || {}}
          totalVotes={stats.votesByPoll[selectedResultsPoll.id] || 0}
          connectionStatus={stats.connectionStatusByPoll[selectedResultsPoll.id] || 'connecting'}
          onClose={() => setSelectedResultsPoll(null)}
        />
      )}
    </div>
  );
};
