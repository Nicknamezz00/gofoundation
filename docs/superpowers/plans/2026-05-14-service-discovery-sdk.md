# Service Discovery SDK Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `discovery` package to `gofoundation` providing a pluggable service discovery interface with a Consul backend and a FakeProvider for testing.

**Architecture:** A `discovery/` package defines the `Provider` interface and core types (`ServiceDefinition`, `ServiceInstance`). A `discovery/consul/` sub-package implements the interface against the Consul agent HTTP API using TTL health checks and a long-poll Watch. A `FakeProvider` in `discovery/fake.go` is an in-memory test double.

**Tech Stack:** Go 1.24, `github.com/hashicorp/consul/api` v1.x

---

## File Map

| File | Responsibility |
|------|---------------|
| `discovery/types.go` | `Provider` interface, `ServiceDefinition`, `ServiceInstance` |
| `discovery/fake.go` | `FakeProvider` — thread-safe in-memory test double |
| `discovery/fake_test.go` | Unit tests for `FakeProvider` |
| `discovery/consul/provider.go` | `ConsulProvider` — Register, Deregister, Resolve, Watch, Close |
| `discovery/consul/provider_test.go` | Unit tests for `ConsulProvider` using mock HTTP |

---

## Task 1: Core types

**Files:**
- Create: `discovery/types.go`

- [ ] **Step 1: Write `discovery/types.go`**

```go
package discovery

import (
	"context"
	"time"
)

// ServiceDefinition describes a service instance to register.
type ServiceDefinition struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	TTL     time.Duration
}

// ServiceInstance is a resolved, healthy instance of a service.
type ServiceInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

// Provider is the interface for service registration and discovery.
// Implementations must be safe for concurrent use.
type Provider interface {
	// Register registers the service with the discovery backend.
	// The registration is tied to ctx: when ctx is cancelled the
	// health-renewal goroutine (if any) stops, but the service entry
	// is NOT automatically deregistered — call Deregister explicitly.
	Register(ctx context.Context, svc ServiceDefinition) error

	// Deregister removes the service entry identified by id.
	Deregister(ctx context.Context, id string) error

	// Resolve returns all currently healthy instances of the named service.
	Resolve(ctx context.Context, name string) ([]ServiceInstance, error)

	// Watch returns a channel that emits a full snapshot of healthy instances
	// whenever the set changes. The channel is closed when ctx is cancelled.
	Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error)

	// Close releases any resources held by the provider.
	Close() error
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/didi/Documents/personal/gofoundation && go build ./discovery/...
```

Expected: no output (clean build).

- [ ] **Step 3: Commit**

```bash
cd /Users/didi/Documents/personal/gofoundation && git add discovery/types.go && git commit -m "feat(discovery): add Provider interface and core types"
```

---

## Task 2: FakeProvider

**Files:**
- Create: `discovery/fake.go`
- Create: `discovery/fake_test.go`

- [ ] **Step 1: Write the failing tests first**

Create `discovery/fake_test.go`:

```go
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

	_ = p.Register(ctx, newDef("svc-1", "kb"))
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
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/didi/Documents/personal/gofoundation && go test ./discovery/... 2>&1 | head -20
```

Expected: compile error — `discovery.FakeProvider` undefined.

- [ ] **Step 3: Implement `FakeProvider`**

Create `discovery/fake.go`:

```go
package discovery

import (
	"context"
	"sync"
)

// FakeProvider is an in-memory Provider for use in tests.
// Its zero value is ready to use.
type FakeProvider struct {
	mu        sync.RWMutex
	instances map[string]ServiceInstance // keyed by instance ID
	watchers  map[string][]chan []ServiceInstance
}

func (f *FakeProvider) init() {
	if f.instances == nil {
		f.instances = make(map[string]ServiceInstance)
	}
	if f.watchers == nil {
		f.watchers = make(map[string][]chan []ServiceInstance)
	}
}

// Register stores the service definition and notifies watchers.
func (f *FakeProvider) Register(_ context.Context, svc ServiceDefinition) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.init()
	f.instances[svc.ID] = ServiceInstance{
		ID:      svc.ID,
		Name:    svc.Name,
		Address: svc.Address,
		Port:    svc.Port,
		Tags:    svc.Tags,
	}
	f.notify(svc.Name)
	return nil
}

// Deregister removes the instance and notifies watchers.
func (f *FakeProvider) Deregister(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.init()
	inst, ok := f.instances[id]
	if !ok {
		return nil
	}
	name := inst.Name
	delete(f.instances, id)
	f.notify(name)
	return nil
}

// Resolve returns all healthy instances for the named service.
func (f *FakeProvider) Resolve(_ context.Context, name string) ([]ServiceInstance, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var out []ServiceInstance
	for _, inst := range f.instances {
		if inst.Name == name {
			out = append(out, inst)
		}
	}
	return out, nil
}

// Watch returns a channel that receives a full snapshot on each change.
// The channel is closed when ctx is cancelled.
func (f *FakeProvider) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	ch := make(chan []ServiceInstance, 8)
	f.mu.Lock()
	f.init()
	f.watchers[name] = append(f.watchers[name], ch)
	f.mu.Unlock()

	go func() {
		<-ctx.Done()
		f.mu.Lock()
		chans := f.watchers[name]
		filtered := chans[:0]
		for _, c := range chans {
			if c != ch {
				filtered = append(filtered, c)
			}
		}
		f.watchers[name] = filtered
		f.mu.Unlock()
		close(ch)
	}()

	return ch, nil
}

// Close is a no-op for FakeProvider.
func (f *FakeProvider) Close() error { return nil }

// notify sends a snapshot to all watchers of name. Must be called with f.mu held (write).
func (f *FakeProvider) notify(name string) {
	var snap []ServiceInstance
	for _, inst := range f.instances {
		if inst.Name == name {
			snap = append(snap, inst)
		}
	}
	for _, ch := range f.watchers[name] {
		select {
		case ch <- snap:
		default:
		}
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/didi/Documents/personal/gofoundation && go test ./discovery/... -v -race
```

Expected: all tests PASS, no race conditions.

- [ ] **Step 5: Commit**

```bash
cd /Users/didi/Documents/personal/gofoundation && git add discovery/fake.go discovery/fake_test.go && git commit -m "feat(discovery): add FakeProvider with Watch support"
```

---

## Task 3: Consul dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add consul/api dependency**

```bash
cd /Users/didi/Documents/personal/gofoundation && go get github.com/hashicorp/consul/api@latest
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 2: Tidy**

```bash
cd /Users/didi/Documents/personal/gofoundation && go mod tidy
```

- [ ] **Step 3: Commit**

```bash
cd /Users/didi/Documents/personal/gofoundation && git add go.mod go.sum && git commit -m "chore(deps): add hashicorp/consul/api"
```

---

## Task 4: ConsulProvider — Register and Deregister

**Files:**
- Create: `discovery/consul/provider.go`

- [ ] **Step 1: Write `discovery/consul/provider.go` (struct + Register + Deregister)**

```go
package consul

import (
	"context"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/Nicknamezz00/gofoundation/discovery"
)

// Config holds the configuration for ConsulProvider.
type Config struct {
	// Address is the Consul agent address, e.g. "127.0.0.1:8500".
	// Defaults to the CONSUL_HTTP_ADDR env var or "127.0.0.1:8500".
	Address string
	// Scheme is "http" or "https". Defaults to "http".
	Scheme string
}

// ConsulProvider implements discovery.Provider backed by a Consul agent.
type ConsulProvider struct {
	client *consulapi.Client
}

// NewConsulProvider creates a ConsulProvider connected to the given agent.
func NewConsulProvider(cfg Config) (*ConsulProvider, error) {
	c := consulapi.DefaultConfig()
	if cfg.Address != "" {
		c.Address = cfg.Address
	}
	if cfg.Scheme != "" {
		c.Scheme = cfg.Scheme
	}
	client, err := consulapi.NewClient(c)
	if err != nil {
		return nil, fmt.Errorf("consul: create client: %w", err)
	}
	return &ConsulProvider{client: client}, nil
}

// Register registers svc with a TTL health check and starts a renewal goroutine.
// The renewal goroutine runs until ctx is cancelled.
func (p *ConsulProvider) Register(ctx context.Context, svc discovery.ServiceDefinition) error {
	ttl := svc.TTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	reg := &consulapi.AgentServiceRegistration{
		ID:      svc.ID,
		Name:    svc.Name,
		Address: svc.Address,
		Port:    svc.Port,
		Tags:    svc.Tags,
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        "service:" + svc.ID,
			TTL:                            ttl.String(),
			DeregisterCriticalServiceAfter: (ttl * 3).String(),
		},
	}

	if err := p.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("consul: register %s: %w", svc.ID, err)
	}

	// Pass initial TTL check.
	if err := p.client.Agent().UpdateTTL("service:"+svc.ID, "started", consulapi.HealthPassing); err != nil {
		return fmt.Errorf("consul: initial ttl pass %s: %w", svc.ID, err)
	}

	// Renewal goroutine.
	go func() {
		ticker := time.NewTicker(ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = p.client.Agent().UpdateTTL("service:"+svc.ID, "alive", consulapi.HealthPassing)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Deregister removes the service entry from the Consul agent.
func (p *ConsulProvider) Deregister(_ context.Context, id string) error {
	if err := p.client.Agent().ServiceDeregister(id); err != nil {
		return fmt.Errorf("consul: deregister %s: %w", id, err)
	}
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/didi/Documents/personal/gofoundation && go build ./discovery/...
```

Expected: clean build.

---

## Task 5: ConsulProvider — Resolve and Watch

**Files:**
- Modify: `discovery/consul/provider.go`

- [ ] **Step 1: Append Resolve, Watch, and Close to `discovery/consul/provider.go`**

Add the following methods to the file (append after the existing Deregister method):

```go
// Resolve returns all healthy instances of the named service from the catalog.
func (p *ConsulProvider) Resolve(_ context.Context, name string) ([]discovery.ServiceInstance, error) {
	entries, _, err := p.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("consul: resolve %s: %w", name, err)
	}
	out := make([]discovery.ServiceInstance, 0, len(entries))
	for _, e := range entries {
		out = append(out, discovery.ServiceInstance{
			ID:      e.Service.ID,
			Name:    e.Service.Service,
			Address: e.Service.Address,
			Port:    e.Service.Port,
			Tags:    e.Service.Tags,
		})
	}
	return out, nil
}

// Watch long-polls Consul for healthy instance changes and emits full snapshots.
// The channel is closed when ctx is cancelled.
func (p *ConsulProvider) Watch(ctx context.Context, name string) (<-chan []discovery.ServiceInstance, error) {
	ch := make(chan []discovery.ServiceInstance, 8)

	go func() {
		defer close(ch)
		var lastIndex uint64
		for {
			opts := &consulapi.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			}
			opts = opts.WithContext(ctx)

			entries, meta, err := p.client.Health().Service(name, "", true, opts)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				// Transient error — retry after a short pause.
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}

			if meta.LastIndex == lastIndex {
				continue
			}
			lastIndex = meta.LastIndex

			snap := make([]discovery.ServiceInstance, 0, len(entries))
			for _, e := range entries {
				snap = append(snap, discovery.ServiceInstance{
					ID:      e.Service.ID,
					Name:    e.Service.Service,
					Address: e.Service.Address,
					Port:    e.Service.Port,
					Tags:    e.Service.Tags,
				})
			}

			select {
			case ch <- snap:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// Close is a no-op for ConsulProvider (the underlying HTTP client has no persistent connection).
func (p *ConsulProvider) Close() error { return nil }
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/didi/Documents/personal/gofoundation && go build ./discovery/...
```

Expected: clean build.

- [ ] **Step 3: Verify interface satisfaction**

Add a blank import check. Create `discovery/consul/check.go`:

```go
package consul

import "github.com/Nicknamezz00/gofoundation/discovery"

var _ discovery.Provider = (*ConsulProvider)(nil)
```

```bash
cd /Users/didi/Documents/personal/gofoundation && go build ./discovery/...
```

Expected: clean build (compile-time interface check passes).

- [ ] **Step 4: Commit**

```bash
cd /Users/didi/Documents/personal/gofoundation && git add discovery/consul/ && git commit -m "feat(discovery/consul): implement ConsulProvider"
```

---

## Task 6: ConsulProvider tests

**Files:**
- Create: `discovery/consul/provider_test.go`

- [ ] **Step 1: Write tests using httptest to mock the Consul HTTP API**

Create `discovery/consul/provider_test.go`:

```go
package consul_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	registered := false
	ttlPassed := false

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		registered = true
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		ttlPassed = true
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
	if !registered {
		t.Error("expected service registration request")
	}
	if !ttlPassed {
		t.Error("expected initial TTL pass request")
	}
}

func TestConsulProvider_Deregister(t *testing.T) {
	deregistered := false

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/deregister/svc-1", func(w http.ResponseWriter, r *http.Request) {
		deregistered = true
		w.WriteHeader(http.StatusOK)
	})

	p := newTestProvider(t, mux)
	if err := p.Deregister(context.Background(), "svc-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}
	if !deregistered {
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
```

- [ ] **Step 2: Run tests**

```bash
cd /Users/didi/Documents/personal/gofoundation && go test ./discovery/... -v -race -timeout 30s
```

Expected: all tests PASS.

- [ ] **Step 3: Commit**

```bash
cd /Users/didi/Documents/personal/gofoundation && git add discovery/consul/provider_test.go && git commit -m "test(discovery/consul): add ConsulProvider unit tests with mock HTTP"
```

---

## Task 7: Final build verification

- [ ] **Step 1: Full build**

```bash
cd /Users/didi/Documents/personal/gofoundation && go build ./...
```

Expected: clean build.

- [ ] **Step 2: Full test suite with race detector**

```bash
cd /Users/didi/Documents/personal/gofoundation && go test ./... -race -timeout 60s
```

Expected: all tests PASS, no race conditions.

- [ ] **Step 3: Vet**

```bash
cd /Users/didi/Documents/personal/gofoundation && go vet ./discovery/...
```

Expected: no output.
