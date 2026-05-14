package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/Nicknamezz00/gofoundation/discovery"
)

func newDef(id, name string) discovery.ServiceDefinition {
	return discovery.ServiceDefinition{
		ID:      id,
		Name:    name,
		Address: "127.0.0.1",
		Port:    8080,
		TTL:     30 * time.Second,
	}
}

func TestFakeProvider_RegisterAndResolve(t *testing.T) {
	p := &discovery.FakeProvider{}
	ctx := context.Background()

	if err := p.Register(ctx, newDef("svc-1", "kb")); err != nil {
		t.Fatalf("Register: %v", err)
	}

	instances, err := p.Resolve(ctx, "kb")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("want 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "svc-1" {
		t.Errorf("want ID svc-1, got %s", instances[0].ID)
	}
}

func TestFakeProvider_Deregister(t *testing.T) {
	p := &discovery.FakeProvider{}
	ctx := context.Background()

	if err := p.Register(ctx, newDef("svc-1", "kb")); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := p.Deregister(ctx, "svc-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}

	instances, err := p.Resolve(ctx, "kb")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("want 0 instances after deregister, got %d", len(instances))
	}
}

func TestFakeProvider_Watch(t *testing.T) {
	p := &discovery.FakeProvider{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := p.Watch(ctx, "kb")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	// Register triggers a snapshot on the watch channel.
	if err := p.Register(ctx, newDef("svc-1", "kb")); err != nil {
		t.Fatalf("Register: %v", err)
	}

	select {
	case snap := <-ch:
		if len(snap) != 1 {
			t.Errorf("want 1 instance in snapshot, got %d", len(snap))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Watch snapshot")
	}

	// Cancel context closes the channel.
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after context cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestFakeProvider_ResolveUnknownService(t *testing.T) {
	p := &discovery.FakeProvider{}
	instances, err := p.Resolve(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Resolve unknown: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("want 0 instances for unknown service, got %d", len(instances))
	}
}

func TestFakeProvider_Close(t *testing.T) {
	p := &discovery.FakeProvider{}
	if err := p.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
