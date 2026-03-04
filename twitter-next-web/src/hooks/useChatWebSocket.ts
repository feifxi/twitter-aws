import { useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/useAuthStore';
import { conversationMessagesQueryKey, conversationsQueryKey, publicRoomMessagesQueryKey } from './useMessages';
import type { MessageResponse, PublicRoomMessageResponse } from '@/types/message';

const baseURL = process.env.NEXT_PUBLIC_API_URL?.replace(/^http/, 'ws') || 'ws://localhost:8080/api/v1';

type WSEnvelope = {
    type: string;
    conversationId?: number;
    roomKey?: string;
    message?: MessageResponse | PublicRoomMessageResponse;
    data?: any;
};

export function useChatWebSocket() {
    const queryClient = useQueryClient();
    const { accessToken } = useAuthStore();
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout>(null);

    useEffect(() => {
        let isSubscribed = true;

        function connect() {
            if (!isSubscribed) return;

            // Only connect if we have an access token (backend requires authentication)
            if (!accessToken) {
                if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
                if (wsRef.current) {
                    wsRef.current.close();
                    wsRef.current = null;
                }
                return;
            }

            let url = `${baseURL}/messages/ws?access_token=${accessToken}`;

            const ws = new WebSocket(url);
            wsRef.current = ws;

            ws.onmessage = (event) => {
                try {
                    const envelope = JSON.parse(event.data) as WSEnvelope;

                    if (envelope.type === 'dm.message' && envelope.conversationId) {
                        queryClient.invalidateQueries({ queryKey: conversationsQueryKey });
                        queryClient.invalidateQueries({ queryKey: conversationMessagesQueryKey(envelope.conversationId) });
                    } else if (envelope.type === 'public.message' && envelope.roomKey) {
                        queryClient.invalidateQueries({ queryKey: publicRoomMessagesQueryKey(envelope.roomKey) });
                    }
                } catch {
                    // ignore malformed events
                }
            };

            ws.onclose = () => {
                if (!isSubscribed) return;
                // Attempt to reconnect after 3 seconds
                if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
                reconnectTimeoutRef.current = setTimeout(connect, 3000);
            };

            ws.onerror = () => {
                ws.close();
            };
        }

        // Small delay avoids React StrictMode immediate mount/unmount canceling the TCP handshake
        const initialTimeout = setTimeout(connect, 50);

        return () => {
            isSubscribed = false;
            clearTimeout(initialTimeout);
            if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
            if (wsRef.current) {
                // Ignore closing connecting websockets which triggers browser console errors
                if (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING) {
                    wsRef.current.close();
                }
                wsRef.current = null;
            }
        };
    }, [accessToken, queryClient]);
}
