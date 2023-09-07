package storage

import (
	zslices "github.com/zSnails/plyr/slices"
	"slices"
)

type Cache[K comparable, V any] struct {
	cached map[K]*V
	values []*V
}

func NewCache[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		cached: make(map[K]*V),
		values: []*V{},
	}
}

func (c *Cache[K, V]) All() []*V {
	return c.values
}

func (c *Cache[K, V]) StoreIfNotExists(key K, data *V) *V {
	original, ok := c.cached[key]
	if !ok {
		c.Store(key, data)
		return original
	}
	return original
}

func (c *Cache[K, V]) Store(key K, data *V) {
	c.cached[key] = data
	c.values = append(c.values, data)
}

func (c *Cache[K, V]) Get(key K) *V {
	val, ok := c.cached[key]
	if !ok {
		return nil
	}
	return val
}

func (c *Cache[K, V]) Delete(key K) {
	data, ok := c.cached[key]
	if !ok {
		return
	}
	idx := slices.Index(c.values, data)

	if idx != -1 {
		slices.Delete(c.values, idx, idx)
	}
}

func (s *Cache[K, V]) Filter(ff func(v *V) bool) []*V {
	data := s.values

	return zslices.Filter(data, ff)
}
