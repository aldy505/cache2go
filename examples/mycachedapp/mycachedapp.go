package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aldy505/cache2go"
)

// Keys & values in cache2go can be of arbitrary types, e.g. a struct.
type myStruct struct {
	text     string
	moreData []byte
}

func main() {
	// Accessing a new cache table for the first time will create it.
	cache := cache2go.Cache("myCache")

	// We will put a new item in the cache. It will expire after
	// not being accessed via Value(key) for more than 5 seconds.
	val, _ := json.Marshal(myStruct{"This is a test!", []byte{}})

	cache.Add("someKey", val, 5*time.Second)

	// Let's retrieve the item from the cache.
	res, err := cache.Value("someKey")
	if err == nil {
		var m myStruct
		_ = json.Unmarshal(res, &m)
		fmt.Println("Found value in cache:", m.text)
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}

	// Wait for the item to expire in cache.
	time.Sleep(6 * time.Second)
	res, err = cache.Value("someKey")
	if err != nil {
		fmt.Println("Item is not cached (anymore).")
	}

	// Add another item that never expires.
	cache.Add("someKey", val, 0)

	// Remove the item from the cache.
	cache.Delete("someKey")

	// And wipe the entire cache table.
	cache.Flush()
}
