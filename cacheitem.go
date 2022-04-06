/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2022, Reinaldy Rafli <aldy505@tutanota.com>
 *
 *   For license see LICENSE.txt
 */

package cache2go

import (
	"sync"
	"time"
)

// CacheItem is an individual cache item
// Parameter data contains the user-set value in the cache.
type CacheItem struct {
	sync.RWMutex

	// The item's key.
	key string
	// The item's data.
	data []byte
	// How long will the item live in the cache when not being accessed/kept alive.
	lifeSpan time.Duration

	// Creation timestamp.
	createdOn time.Time
	// Last access timestamp.
	accessedOn time.Time

	// Callback method triggered right before removing the item from the cache
	aboutToExpire []func(key interface{})
}

// NewCacheItem returns a newly created CacheItem.
// Parameter key is the item's cache-key.
// Parameter ttl determines after which time period without an access the item
// will get removed from the cache.
// Parameter data is the item's value.
func NewCacheItem(key string, data []byte, ttl time.Duration) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:           key,
		lifeSpan:      ttl,
		createdOn:     t,
		accessedOn:    t,
		aboutToExpire: nil,
		data:          data,
	}
}

// KeepAlive marks an item to be kept for another expireDuration period.
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
}

// LifeSpan returns this item's expiration duration.
func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// AccessedOn returns when this item was last accessed.
func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

// CreatedOn returns when this item was added to the cache.
func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

// Key returns the key of this cached item.
func (item *CacheItem) Key() string {
	// immutable
	return item.key
}

// Data returns the value of this cached item.
func (item *CacheItem) Data() []byte {
	// immutable
	return item.data
}

// SetAboutToExpireCallback configures a callback, which will be called right
// before the item is about to be removed from the cache.
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	if len(item.aboutToExpire) > 0 {
		item.RemoveAboutToExpireCallback()
	}
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = append(item.aboutToExpire, f)
}

// AddAboutToExpireCallback appends a new callback to the AboutToExpire queue
func (item *CacheItem) AddAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = append(item.aboutToExpire, f)
}

// RemoveAboutToExpireCallback empties the about to expire callback queue
func (item *CacheItem) RemoveAboutToExpireCallback() {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = nil
}
