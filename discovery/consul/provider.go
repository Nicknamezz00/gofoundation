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
