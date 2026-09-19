import { useEffect, useRef, useState, useCallback } from 'react';

const getWebSocketUrl = (pollId) => {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
  // Replace http(s) protocol with ws(s)
  let wsUrl = apiUrl.replace(/^http/, 'ws');
  if (!wsUrl.startsWith('ws://') && !wsUrl.startsWith('wss://')) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    wsUrl = `${protocol}//${window.location.host}${apiUrl}`;
  }
  return `${wsUrl}/public/polls/${pollId}/ws`;
};

export const useWebSocket = (pollId, onUpdate) => {
  const [status, setStatus] = useState('connecting'); // 'connecting' | 'connected' | 'reconnecting' | 'disconnected'
  const socketRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const retryCountRef = useRef(0);
  const maxRetries = 10;
  const onUpdateRef = useRef(onUpdate);

  // Keep latest onUpdate reference without triggering re-connects
  useEffect(() => {
    onUpdateRef.current = onUpdate;
  }, [onUpdate]);

  const connect = useCallback(() => {
    if (!pollId) return;

    // Close existing socket if any
    if (socketRef.current) {
      socketRef.current.close();
    }

    try {
      const url = getWebSocketUrl(pollId);
      const ws = new WebSocket(url);
      socketRef.current = ws;

      ws.onopen = () => {
        setStatus('connected');
        retryCountRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data && data.type === 'poll_update' && onUpdateRef.current) {
            onUpdateRef.current(data);
          }
        } catch (err) {
          console.warn('Failed to parse WebSocket message:', err);
        }
      };

      ws.onerror = (err) => {
        console.warn('WebSocket connection error:', err);
      };

      ws.onclose = (event) => {
        // Don't reconnect if unmounted or cleanly closed
        if (event.wasClean) {
          setStatus('disconnected');
          return;
        }

        if (retryCountRef.current < maxRetries) {
          setStatus('reconnecting');
          const delay = Math.min(1000 * Math.pow(1.5, retryCountRef.current), 10000);
          retryCountRef.current += 1;
          reconnectTimeoutRef.current = setTimeout(() => {
            if (socketRef.current) {
              socketRef.current = null;
            }
            // Trigger connection attempt
            setStatus((prev) => (prev === 'disconnected' ? 'disconnected' : 'reconnecting'));
          }, delay);
        } else {
          setStatus('disconnected');
        }
      };
    } catch (err) {
      console.warn('WebSocket initialization error:', err);
      setStatus('disconnected');
    }
  }, [pollId]);

  useEffect(() => {
    if (pollId) {
      connect();
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (socketRef.current) {
        socketRef.current.close(1000, 'Component unmounted');
      }
    };
  }, [pollId, connect]);

  return { status };
};
