package strategy

import (
	"fmt"
	"slices"
	"sync"
)

var (
	mu       sync.RWMutex
	registry []Strategy
	byNode   map[string][]Strategy
)

func init() {
	byNode = make(map[string][]Strategy)
}

// Register registers a strategy. Called from init() in strategy packages.
// Panics if a strategy with the same name is already registered.
func Register(s Strategy) {
	mu.Lock()
	defer mu.Unlock()

	for _, existing := range registry {
		if existing.Name() == s.Name() {
			panic(fmt.Sprintf("strategy %q already registered", s.Name()))
		}
	}

	registry = append(registry, s)
	for _, nt := range s.NodeTypes() {
		byNode[nt] = append(byNode[nt], s)
	}
}

// All returns all registered strategies.
func All() []Strategy {
	mu.RLock()
	defer mu.RUnlock()

	return slices.Clone(registry)
}

// ForNodeType returns all strategies registered for the given AST node type string.
func ForNodeType(t string) []Strategy {
	mu.RLock()
	defer mu.RUnlock()

	return byNode[t]
}
