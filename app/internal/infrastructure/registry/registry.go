package registry

import (
	"sync"
)

var (
	mu       sync.RWMutex
	services = make(map[string]interface{})
)

func Register(name string, svc interface{}) {
	mu.Lock()
	defer mu.Unlock()
	services[name] = svc
}

func Get[T any](name string) (T, bool) {
	mu.RLock()
	defer mu.RUnlock()
	svc, ok := services[name]
	if !ok {
		var zero T
		return zero, false
	}
	return svc.(T), true
}
