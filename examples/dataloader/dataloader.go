package main

import (
	"fmt"
	"strconv"

	"github.com/aldy505/cache2go"
)

func main() {
	cache := cache2go.Cache("myCache")

	// The data loader gets called automatically whenever something
	// tries to retrieve a non-existing key from the cache.
	cache.SetDataLoader(func(key string, args ...interface{}) *cache2go.CacheItem {
		// Apply some clever loading logic here, e.g. read values for
		// this key from database, network or file.
		val := "This is a test with key " + key

		// This helper method creates the cached item for us. Yay!
		item := cache2go.NewCacheItem(key, []byte(val), 0)
		return item
	})

	// Let's retrieve a few auto-generated items from the cache.
	for i := 0; i < 10; i++ {
		res, err := cache.Value("someKey_" + strconv.Itoa(i))
		if err == nil {
			fmt.Println("Found value in cache:", string(res))
		} else {
			fmt.Println("Error retrieving value from cache:", err)
		}
	}
}
