import { useEffect, useMemo, useState } from 'react';
import { getWebSocketUrl } from '../config/api';

export const useDashboardRealtimeStats = (polls) => {
  const [votesByPoll, setVotesByPoll] = useState({});
  const [countsByPoll, setCountsByPoll] = useState({});
  const [lastActivityByPoll, setLastActivityByPoll] = useState({});
  const [connectionStatusByPoll, setConnectionStatusByPoll] = useState({});
  const [lastRealtimeUpdate, setLastRealtimeUpdate] = useState(null);
  const [realtimeVersion, setRealtimeVersion] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const sockets = new Map();
    const reconnectTimers = new Map();

    const connect = (pollId) => {
      if (cancelled) return;
      setConnectionStatusByPoll((current) => ({ ...current, [pollId]: 'connecting' }));
      const socket = new WebSocket(getWebSocketUrl(pollId));
      sockets.set(pollId, socket);

      socket.onopen = () => setConnectionStatusByPoll((current) => ({ ...current, [pollId]: 'connected' }));
      socket.onclose = () => {
        if (cancelled) return;
        sockets.delete(pollId);
        setConnectionStatusByPoll((current) => ({ ...current, [pollId]: 'reconnecting' }));
        const timer = window.setTimeout(() => {
          reconnectTimers.delete(pollId);
          connect(pollId);
        }, 1000);
        reconnectTimers.set(pollId, timer);
      };
      socket.onerror = () => setConnectionStatusByPoll((current) => ({ ...current, [pollId]: 'reconnecting' }));
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          if (message?.type === 'poll_update') {
            setVotesByPoll((current) => ({
              ...current,
              [pollId]: message.totalVotes || 0,
            }));
            setCountsByPoll((current) => ({
              ...current,
              [pollId]: message.counts || {},
            }));
            if (message.lastActivity) {
              setLastActivityByPoll((current) => ({
                ...current,
                [pollId]: message.lastActivity,
              }));
            }
            setLastRealtimeUpdate({ pollId, totalVotes: message.totalVotes || 0 });
            setRealtimeVersion((version) => version + 1);
          }
        } catch {
          // Ignore malformed frames from a disconnected socket.
        }
      };
    };

    polls.filter((poll) => poll?.id).forEach((poll) => connect(poll.id));

    return () => {
      cancelled = true;
      reconnectTimers.forEach((timer) => window.clearTimeout(timer));
      sockets.forEach((socket) => socket.close(1000, 'Dashboard unmounted'));
    };
  }, [polls]);

  return useMemo(() => {
    const activePolls = polls.filter((poll) => poll.status === 'active');
    const closedPolls = polls.filter((poll) => poll.status === 'closed');
    const totalVotes = polls.reduce((total, poll) => total + (votesByPoll[poll.id] || 0), 0);

    return {
      totalPolls: polls.length,
      activePolls: activePolls.length,
      closedPolls: closedPolls.length,
      totalVotes,
      votesByPoll,
      countsByPoll,
      lastActivityByPoll,
      connectionStatusByPoll,
      lastRealtimeUpdate,
      realtimeVersion,
    };
  }, [polls, votesByPoll, countsByPoll, lastActivityByPoll, connectionStatusByPoll, lastRealtimeUpdate, realtimeVersion]);
};
