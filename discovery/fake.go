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
