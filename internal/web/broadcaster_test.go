package web

import (
	"strings"
	"testing"
	"time"
)

func TestBroadcaster_SubscribeAndBroadcast(t *testing.T) {
	b := NewBroadcaster(10)
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	b.Broadcast("test-event", "test-data")

	select {
	case msg := <-ch:
		if msg.Event != "test-event" {
			t.Errorf("expected event 'test-event', got %q", msg.Event)
		}
		if msg.Data != "test-data" {
			t.Errorf("expected data 'test-data', got %q", msg.Data)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for broadcast message")
	}
}

func TestBroadcaster_AppendLogAndHistory(t *testing.T) {
	b := NewBroadcaster(3)

	b.AppendLog("info", "Message 1")
	b.AppendLog("warn", "Message 2")
	b.AppendLog("error", "Message 3")
	b.AppendLog("info", "Message 4") // Should evict Message 1

	history := b.GetHistory()
	if len(history) != 3 {
		t.Fatalf("expected history length 3, got %d", len(history))
	}

	if history[0].Text != "Message 2" {
		t.Errorf("expected first history item to be 'Message 2', got %q", history[0].Text)
	}
	if history[2].Text != "Message 4" {
		t.Errorf("expected last history item to be 'Message 4', got %q", history[2].Text)
	}
}

func TestBroadcaster_LogWriter(t *testing.T) {
	b := NewBroadcaster(10)
	writer := b.LogWriter("info")

	msg := "Hello log streaming\nSecond line\n"
	n, err := writer.Write([]byte(msg))
	if err != nil {
		t.Fatalf("unexpected error writing to LogWriter: %v", err)
	}
	if n != len(msg) {
		t.Errorf("expected written bytes %d, got %d", len(msg), n)
	}

	history := b.GetHistory()
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if !strings.Contains(history[0].Text, "Hello log streaming") {
		t.Errorf("expected first line to contain 'Hello log streaming', got %q", history[0].Text)
	}
}
