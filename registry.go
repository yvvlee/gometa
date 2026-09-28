package gometa

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// Registration binds metadata to the exact type used for registration.
// Type may be a pointer type when the registered service is a pointer.
type Registration struct {
	Type     reflect.Type
	Metadata *TypeMetadata
}

type registry struct {
	mu      sync.RWMutex
	entries map[reflect.Type]Registration
}

var defaultRegistry = registry{entries: make(map[reflect.Type]Registration)}

// Register builds metadata for T, validates its bindings with reflection, and
// adds it to the process-wide registry. Register panics on duplicate or invalid
// declarations because registration happens during application initialization.
func Register[T any](parts ...any) {
	defaultRegistry.register(reflect.TypeFor[T](), parts...)
}

// MetadataOf returns metadata registered for the exact type T.
func MetadataOf[T any]() (*TypeMetadata, bool) {
	return defaultRegistry.metadata(reflect.TypeFor[T]())
}

// Registrations returns every registration ordered by the registered type name.
func Registrations() []Registration {
	return defaultRegistry.registrations()
}

func (r *registry) register(target reflect.Type, parts ...any) *TypeMetadata {
	metadata := buildMetadata(parts...)
	validateMetadata(target, metadata)

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.entries[target]; exists {
		definitionPanic("type", fmt.Sprintf("metadata for %s is already registered", target))
	}
	r.entries[target] = Registration{Type: target, Metadata: metadata}
	return metadata
}

func (r *registry) metadata(target reflect.Type) (*TypeMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	registration, ok := r.entries[target]
	return registration.Metadata, ok
}

func (r *registry) registrations() []Registration {
	r.mu.RLock()
	registrations := make([]Registration, 0, len(r.entries))
	for _, registration := range r.entries {
		registrations = append(registrations, registration)
	}
	r.mu.RUnlock()

	sort.Slice(registrations, func(i, j int) bool {
		return registrations[i].Type.String() < registrations[j].Type.String()
	})
	return registrations
}
