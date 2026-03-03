package server

import (
	"testing"
	"time"
)

func TestSendNotificationToUserDeliversToRegisteredClient(t *testing.T) {
	t.Parallel()

	s := &Server{
		sseClients: make(map[int64][]*sseClient),
	}
	client := &sseClient{channel: make(chan notificationResponse, 1)}
	s.sseClients[7] = []*sseClient{client}

	s.sendNotificationToUser(7, notificationResponse{ID: 99})

	select {
	case msg := <-client.channel:
		if msg.ID != 99 {
			t.Fatalf("unexpected notification id: got %d want %d", msg.ID, 99)
		}
	case <-time.After(time.Second):
		t.Fatal("expected notification to be delivered")
	}
}

func TestSendNotificationToUserWithoutClientsNoop(t *testing.T) {
	t.Parallel()

	s := &Server{
		sseClients: make(map[int64][]*sseClient),
	}

	s.sendNotificationToUser(42, notificationResponse{ID: 1})
}
