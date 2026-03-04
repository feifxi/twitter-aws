import { useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/useAuthStore';
import { notificationQueryKey, unreadCountQueryKey } from './useNotifications';
import { EventSourcePolyfill } from 'event-source-polyfill';

const baseURL = process.env.NEXT_PUBLIC_API_URL;

export function useNotificationSSE() {
    const queryClient = useQueryClient();
    const { accessToken } = useAuthStore();
    const eventSourceRef = useRef<EventSourcePolyfill | null>(null);
    const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const openedOnceRef = useRef(false);

    useEffect(() => {
        if (!accessToken) {
            // If accessToken is not present, close any existing connection
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
                eventSourceRef.current = null;
            }
            if (refreshTimerRef.current) {
                clearTimeout(refreshTimerRef.current);
                refreshTimerRef.current = null;
            }
            return;
        }

        const url = `${baseURL}/notifications/stream`;
        const scheduleRefresh = () => {
            // Coalesce burst events into a single refresh.
            if (refreshTimerRef.current) return;
            refreshTimerRef.current = setTimeout(() => {
                refreshTimerRef.current = null;
                queryClient.invalidateQueries({ queryKey: unreadCountQueryKey });
                queryClient.invalidateQueries({ queryKey: notificationQueryKey });
            }, 250);
        };

        // Use EventSourcePolyfill to support headers
        const es = new EventSourcePolyfill(url, {
            headers: {
                Authorization: `Bearer ${accessToken}`,
            },
            heartbeatTimeout: 120000, // 2 minutes (default is 45s)
        });
        const sse = es as unknown as {
            addEventListener: (type: string, listener: (event: { data: string }) => void) => void;
            removeEventListener: (type: string, listener: (event: { data: string }) => void) => void;
        };

        eventSourceRef.current = es;

        es.onopen = () => {
            // Avoid extra unread-count fetch on first connect.
            // Only re-sync when this is a reconnect.
            if (openedOnceRef.current) {
                scheduleRefresh();
                return;
            }
            openedOnceRef.current = true;
        };

        const handleNotification = (event: { data: string }) => {
            try {
                const data = JSON.parse(event.data);
                if (data && data.id) {
                    scheduleRefresh();
                }
            } catch {
                // Ignore parse errors
            }
        };

        sse.addEventListener('notification', handleNotification);

        es.onerror = () => {
            // Handle error silently
        };

        return () => {
            sse.removeEventListener('notification', handleNotification);
            es.close();
            eventSourceRef.current = null;
            if (refreshTimerRef.current) {
                clearTimeout(refreshTimerRef.current);
                refreshTimerRef.current = null;
            }
        };
    }, [accessToken, queryClient]);
}
