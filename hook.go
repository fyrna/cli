package cli

import (
	"strings"
	"sync"
)

// Store interface for key-value storage with namespace support
type Store interface {
	Set(key string, value any)
	Get(key string) (any, bool)

	// Namespace support
	Namespace(prefix string) Store
	Keys(prefix ...string) []string
	Delete(key string)
	Clear(prefix ...string)
}

// MapStore is a simple thread-safe implementation of Store
type MapStore struct {
	data   map[string]any
	mu     sync.RWMutex
	prefix string
}

// NewStore creates a new store instance
func NewStore() Store {
	return &MapStore{
		data: make(map[string]any),
	}
}

func (s *MapStore) makeKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + "." + key
}

func (s *MapStore) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[s.makeKey(key)] = value
}

func (s *MapStore) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[s.makeKey(key)]
	return value, ok
}

func (s *MapStore) Namespace(prefix string) Store {
	newPrefix := prefix
	if s.prefix != "" {
		newPrefix = s.prefix + "." + prefix
	}

	return &MapStore{
		data:   s.data,
		mu:     s.mu,
		prefix: newPrefix,
	}
}

func (s *MapStore) Keys(prefix ...string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var searchPrefix string
	if len(prefix) > 0 {
		searchPrefix = s.makeKey(prefix[0])
	} else {
		searchPrefix = s.prefix
	}

	var keys []string
	for key := range s.data {
		if searchPrefix == "" || strings.HasPrefix(key, searchPrefix) {
			keys = append(keys, key)
		}
	}

	return keys
}

func (s *MapStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, s.makeKey(key))
}

func (s *MapStore) Clear(prefix ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var searchPrefix string
	if len(prefix) > 0 {
		searchPrefix = s.makeKey(prefix[0])
	} else {
		searchPrefix = s.prefix
	}

	for key := range s.data {
		if searchPrefix == "" || strings.HasPrefix(key, searchPrefix) {
			delete(s.data, key)
		}
	}
}

// HookHandler is the function signature for hook handlers
type HookHandler func(*Context, ...any) error

// HookManager manages application hooks with fluent API
type HookManager struct {
	app      *App
	handlers map[string][]HookHandler
	store    Store
}

// newHookManager creates a new hook manager instance
func newHookManager(app *App) *HookManager {
	return &HookManager{
		app:      app,
		handlers: make(map[string][]HookHandler),
		store:    NewStore(),
	}
}

// Storage methods for hook-level data
func (h *HookManager) Set(key string, value any) {
	h.store.Set(key, value)
}

func (h *HookManager) Get(key string) (any, bool) {
	return h.store.Get(key)
}

func (h *HookManager) Store() Store {
	return h.store
}

// Hook registration methods
func (h *HookManager) OnNotFound(handler func(*Context, string) error) *HookManager {
	h.register("not_found", func(ctx *Context, args ...any) error {
		return handler(ctx, args[0].(string))
	})
	return h
}

func (h *HookManager) BeforeCommand(handler func(*Context, *Command) error) *HookManager {
	h.register("before_command", func(ctx *Context, args ...any) error {
		return handler(ctx, args[0].(*Command))
	})
	return h
}

func (h *HookManager) AfterCommand(handler func(*Context, *Command) error) *HookManager {
	h.register("after_command", func(ctx *Context, args ...any) error {
		return handler(ctx, args[0].(*Command))
	})
	return h
}

func (h *HookManager) BeforeParse(handler func(*Context, []string) error) *HookManager {
	h.register("before_parse", func(ctx *Context, args ...any) error {
		return handler(ctx, args[0].([]string))
	})
	return h
}

func (h *HookManager) AfterParse(handler func(*Context, []string) error) *HookManager {
	h.register("after_parse", func(ctx *Context, args ...any) error {
		return handler(ctx, args[0].([]string))
	})
	return h
}

// Root-level hooks (applies to all command executions)
func (h *HookManager) BeforeRoot(handler func(*Context) error) *HookManager {
	h.register("before_root", func(ctx *Context, args ...any) error {
		return handler(ctx)
	})
	return h
}

func (h *HookManager) AfterRoot(handler func(*Context) error) *HookManager {
	h.register("after_root", func(ctx *Context, args ...any) error {
		return handler(ctx)
	})
	return h
}

// check if event_hook registered
func (h *HookManager) HasHook(event string) bool {
	handlers, exists := h.handlers[event]
	return exists && len(handlers) > 0
}

// Internal methods
func (h *HookManager) register(event string, handler HookHandler) {
	h.handlers[event] = append(h.handlers[event], handler)
}

func (h *HookManager) trigger(event string, ctx *Context, args ...any) error {
	handlers, exists := h.handlers[event]
	if !exists {
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, args...); err != nil {
			return err
		}
	}
	return nil
}
