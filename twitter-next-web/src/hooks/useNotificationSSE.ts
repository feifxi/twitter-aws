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

    useEffect(() => {
        if (!accessToken) {
            // If accessToken is not present, close any existing connection
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
                eventSourceRef.current = null;
            }
            return;
        }

        const url = `${baseURL}/notifications/stream`;

        // Use EventSourcePolyfill to support headers
        const es = new EventSourcePolyfill(url, {
            headers: {
                Authorization: `Bearer ${accessToken}`,
            },
            heartbeatTimeout: 120000, // 2 minutes (default is 45s)
        });

        eventSourceRef.current = es;

        es.onopen = () => {
            // Connected
        };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const handleMessage = (event: any) => {
            if (event.data === 'ping') return;

            try {
                const data = JSON.parse(event.data);
                // We're listening to 'notification' events, so any valid data is a notification
                if (data && data.id) {
                    // Optimistic update for unread count
                    queryClient.setQueryData<number>(unreadCountQueryKey, (old) => (old ?? 0) + 1);

                    // Invalidate notification list to show new item
                    queryClient.invalidateQueries({ queryKey: notificationQueryKey });
                }
            } catch (e) {
                // Ignore parse errors
            }
        };

        es.onmessage = handleMessage;
        es.addEventListener('notification', handleMessage);

        es.onerror = () => {
            // Handle error silently
        };

        return () => {
            es.removeEventListener('notification', handleMessage);
            es.close();
            eventSourceRef.current = null;
        };
    }, [accessToken, queryClient]);
}
