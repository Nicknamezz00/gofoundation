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
