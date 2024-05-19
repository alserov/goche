package goche

import (
	"context"
)

type Cache interface {
	Get(ctx context.Context, key string) (any, bool)
	Set(ctx context.Context, key string, val any)

	Close()
}

// New - cache constructor, requires cache type, additionally modifiers
func New(t Type, mods ...ModifierFunc) Cache {
	switch t {
	case LFU:
		return NewLFU(mods...)
	case LRU:
		return NewLRU(mods...)
	default:
		panic("invalid constructor parameter")
	}
}

type Type int

const (
	LRU Type = iota
	LFU

	DefaultLimit = 1000
)

type ModifierFunc func(cache any)
