package ws

import (
	"testing"
	"time"
)

func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("New hub should have 0 clients, got %d", hub.ClientCount())
	}

	if hub.UserCount() != 0 {
		t.Errorf("New hub should have 0 users, got %d", hub.UserCount())
	}
}

func TestHubIsUserOnline(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	if hub.IsUserOnline(123) {
		t.Error("User 123 should not be online")
	}
}

func TestHubGetUserClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	clients := hub.GetUserClients(123)
	if clients != nil {
		t.Errorf("Expected nil for non-existent user, got %v", clients)
	}
}

func TestHubOptions(t *testing.T) {
	hub := NewHub(
		WithReadTimeout(30*time.Second),
		WithWriteTimeout(5*time.Second),
		WithPingInterval(15*time.Second),
		WithMaxMessageSize(2048),
		WithSendBuffer(128),
	)

	if hub.opts.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want 30s", hub.opts.ReadTimeout)
	}
	if hub.opts.WriteTimeout != 5*time.Second {
		t.Errorf("WriteTimeout = %v, want 5s", hub.opts.WriteTimeout)
	}
	if hub.opts.PingInterval != 15*time.Second {
		t.Errorf("PingInterval = %v, want 15s", hub.opts.PingInterval)
	}
	if hub.opts.MaxMessageSize != 2048 {
		t.Errorf("MaxMessageSize = %d, want 2048", hub.opts.MaxMessageSize)
	}
	if hub.opts.SendBuffer != 128 {
		t.Errorf("SendBuffer = %d, want 128", hub.opts.SendBuffer)
	}
}

func TestHubDefaultOptions(t *testing.T) {
	hub := NewHub()

	if hub.opts.ReadTimeout != 60*time.Second {
		t.Errorf("Default ReadTimeout = %v, want 60s", hub.opts.ReadTimeout)
	}
	if hub.opts.WriteTimeout != 10*time.Second {
		t.Errorf("Default WriteTimeout = %v, want 10s", hub.opts.WriteTimeout)
	}
	if hub.opts.PingInterval != 30*time.Second {
		t.Errorf("Default PingInterval = %v, want 30s", hub.opts.PingInterval)
	}
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// Broadcast to empty hub should not panic
	msg := NewBroadcast("test", "hello")
	hub.Broadcast(msg)

	// Give time for broadcast to process
	time.Sleep(10 * time.Millisecond)
}

func TestHubSendToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// Send to non-existent user should return false
	msg := NewMessage(123, "test", "hello")
	if hub.SendToUser(123, msg) {
		t.Error("SendToUser should return false for non-existent user")
	}
}

func TestHubDeliverLocal(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// DeliverLocal should not panic on empty hub
	hub.DeliverLocal(NewBroadcast("test", "hello"))
	hub.DeliverLocal(NewMessage(123, "test", "hello"))

	time.Sleep(10 * time.Millisecond)
}
