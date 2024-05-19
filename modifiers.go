package goche

import "time"

// Modifiers

func WithLimit(lim uint64) ModifierFunc {
	return func(cache any) {
		switch cache.(type) {
		case *lru:
			cache.(*lru).limit = lim
		case *lfu:
			cache.(*lfu).limit = lim
		default:
			panic("not compatible modifier: WithLimit")
		}
	}
}

func WithClear(period time.Duration) ModifierFunc {
	return func(cache any) {
		switch cache.(type) {
		case *lfu:
			cache.(*lfu).clearPeriod = &period
		default:
			panic("not compatible modifier: WithClear")
		}
	}
}
