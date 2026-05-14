package consul_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/Nicknamezz00/gofoundation/discovery"
	"github.com/Nicknamezz00/gofoundation/discovery/consul"
)

// newTestProvider creates a ConsulProvider pointed at a mock HTTP server.
func newTestProvider(t *testing.T, mux *http.ServeMux) *consul.ConsulProvider {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	p, err := consul.NewConsulProvider(consul.Config{Address: strings.TrimPrefix(srv.URL, "http://")})
	if err != nil {
		t.Fatalf("NewConsulProvider: %v", err)
	}
	return p
}

func TestConsulProvider_Register(t *testing.T) {
	var registered, ttlPassed atomic.Bool

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		registered.Store(true)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		ttlPassed.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	p := newTestProvider(t, mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := p.Register(ctx, discovery.ServiceDefinition{
		ID:      "svc-1",
		Name:    "kb",
		Address: "127.0.0.1",
		Port:    9000,
		TTL:     30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if !registered.Load() {
		t.Error("expected service registration request")
	}
	if !ttlPassed.Load() {
		t.Error("expected initial TTL pass request")
	}
}

func TestConsulProvider_Deregister(t *testing.T) {
	var deregistered atomic.Bool

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/deregister/svc-1", func(w http.ResponseWriter, r *http.Request) {
		deregistered.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	p := newTestProvider(t, mux)
	if err := p.Deregister(context.Background(), "svc-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}
	if !deregistered.Load() {
		t.Error("expected deregister request")
	}
}

func TestConsulProvider_Resolve(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health/service/kb", func(w http.ResponseWriter, r *http.Request) {
		entries := []consulapi.ServiceEntry{
			{
				Service: &consulapi.AgentService{
					ID:      "svc-1",
					Service: "kb",
					Address: "10.0.0.1",
					Port:    9000,
				},
			},
		}
		w.Header().Set("X-Consul-Index", "5")
		_ = json.NewEncoder(w).Encode(entries)
	})

	p := newTestProvider(t, mux)
	instances, err := p.Resolve(context.Background(), "kb")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("want 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "svc-1" {
		t.Errorf("want ID svc-1, got %s", instances[0].ID)
	}
	if instances[0].Address != "10.0.0.1" {
		t.Errorf("want address 10.0.0.1, got %s", instances[0].Address)
	}
}

func TestConsulProvider_Watch_ContextCancel(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health/service/kb", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Simulate a blocking long-poll that respects client disconnect.
		select {
		case <-r.Context().Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
		w.Header().Set("X-Consul-Index", "1")
		_ = json.NewEncoder(w).Encode([]consulapi.ServiceEntry{})
	})

	p := newTestProvider(t, mux)
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := p.Watch(ctx, "kb")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	// Let one poll cycle complete.
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Channel must close after cancel.
	select {
	case _, ok := <-ch:
		if ok {
			// Drain any pending snapshot.
			for range ch {
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Watch channel to close")
	}
}
