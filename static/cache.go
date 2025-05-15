package static

import (
	"sync"
)

type Cache struct {
	dict map[string]*File
	lock sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		dict: make(map[string]*File),
	}
}

func (c *Cache) Get(key string) *File {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.dict[key]
}

func (c *Cache) Set(key string, file *File) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dict[key] = file
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.dict, key)
}

func (c *Cache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k := range c.dict {
		delete(c.dict, k)
	}
}

func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.dict)
}

func (c *Cache) Keys() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	keys := make([]string, 0, len(c.dict))
	for k := range c.dict {
		keys = append(keys, k)
	}
	return keys
}
