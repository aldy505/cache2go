# cache2go

[![Latest Release](https://img.shields.io/github/release/aldy505/cache2go.svg)](https://github.com/aldy505/cache2go/releases)
[![Build Status](https://github.com/aldy505/cache2go/workflows/build/badge.svg)](https://github.com/aldy505/cache2go/actions)
[![Coverage Status](https://coveralls.io/repos/github/aldy505/cache2go/badge.svg?branch=master)](https://coveralls.io/github/aldy505/cache2go?branch=master)
[![Go ReportCard](https://goreportcard.com/badge/aldy505/cache2go)](https://goreportcard.com/report/aldy505/cache2go)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/aldy505/cache2go)

Concurrency-safe golang caching library with expiration capabilities.

## Installation

Make sure you have a working Go environment (Go 1.2 or higher is required).
See the [install instructions](https://golang.org/doc/install.html).

To install cache2go, simply import:

```go
    import github.com/aldy505/cache2go
```

## Example
```go
package main

import (
	"fmt"
	"time"

	"github.com/aldy505/cache2go"
)


func main() {
	// Accessing a new cache table for the first time will create it.
	cache := cache2go.Cache("myCache")

	// We will put a new item in the cache. It will expire after
	// not being accessed via Value(key) for more than 5 seconds.
	cache.Add("someKey", []byte("This is a test!"), 5*time.Second)

	// Let's retrieve the item from the cache.
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Found value in cache:", string(res))
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
	cache.Add("someKey", []byte("Look! Another key"), 0)

	// Remove the item from the cache.
	cache.Delete("someKey")

	// And wipe the entire cache table.
	cache.Flush()
}
```

You can find a [few more examples here](https://github.com/aldy505/cache2go/tree/master/examples).
Also see our test-cases in cache_test.go for further working examples.
