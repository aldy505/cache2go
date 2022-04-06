/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2022, Reinaldy Rafli <aldy505@tutanota.com>
 *
 *   For license see LICENSE.txt
 */

package cache2go_test

import (
	"bytes"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aldy505/cache2go"
)

var (
	k = "testkey"
	v = "testvalue"
)

func TestCache(t *testing.T) {
	// add an expiring item after a non-expiring one to
	// trigger expirationCheck iterating over non-expiring items
	table := cache2go.Cache("testCache")
	table.Add(k+"_1", []byte(v), 0*time.Second)
	table.Add(k+"_2", []byte(v), 1*time.Second)

	// check if both items are still there
	p, err := table.Value(k + "_1")
	if err != nil {
		t.Error("Error retrieving non expiring data from cache", err)
	}

	if string(p) != v {
		t.Errorf("Error retrieving non expiring data from cache: %s", p)
	}

	p, err = table.Value(k + "_2")
	if err != nil {
		t.Error("Error retrieving data from cache", err)
	}

	if string(p) != v {
		t.Errorf("Error retrieving data from cache: %s", p)
	}
}

func TestCacheExpire(t *testing.T) {
	table := cache2go.Cache("testCache")

	table.Add(k+"_1", []byte(v+"_1"), 250*time.Millisecond)
	table.Add(k+"_2", []byte(v+"_2"), 200*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	// check key `1` is still alive
	_, err := table.Value(k + "_1")
	if err != nil {
		t.Error("Error retrieving value from cache:", err)
	}

	time.Sleep(150 * time.Millisecond)

	// check key `1` again, it should still be alive since we just accessed it
	_, err = table.Value(k + "_1")
	if err != nil {
		t.Error("Error retrieving value from cache:", err)
	}

	// check key `2`, it should have been removed by now
	_, err = table.Value(k + "_2")
	if err == nil {
		t.Error("Found key which should have been expired by now")
	}
}

func TestExists(t *testing.T) {
	// add an expiring item
	table := cache2go.Cache("testExists")
	table.Add(k, []byte(v), 0)
	// check if it exists
	if !table.Exists(k) {
		t.Error("Error verifying existing data in cache")
	}
}

func TestNotFoundAdd(t *testing.T) {
	table := cache2go.Cache("testNotFoundAdd")

	if !table.NotFoundAdd(k, []byte(v), 0) {
		t.Error("Error verifying NotFoundAdd, data not in cache")
	}

	if table.NotFoundAdd(k, []byte(v), 0) {
		t.Error("Error verifying NotFoundAdd data in cache")
	}
}

func TestNotFoundAddConcurrency(t *testing.T) {
	table := cache2go.Cache("testNotFoundAdd")

	var finish sync.WaitGroup
	var added int32
	var idle int32

	fn := func(id int) {
		for i := 0; i < 100; i++ {
			if table.NotFoundAdd(strconv.Itoa(i), []byte(strconv.Itoa(i+id)), 0) {
				atomic.AddInt32(&added, 1)
			} else {
				atomic.AddInt32(&idle, 1)
			}
			time.Sleep(0)
		}
		finish.Done()
	}

	finish.Add(10)
	go fn(0x0000)
	go fn(0x1100)
	go fn(0x2200)
	go fn(0x3300)
	go fn(0x4400)
	go fn(0x5500)
	go fn(0x6600)
	go fn(0x7700)
	go fn(0x8800)
	go fn(0x9900)
	finish.Wait()

	t.Log(added, idle)

	table.Foreach(func(key string, item *cache2go.CacheItem) {
		v := item.Data()
		k := key
		t.Logf("%02x  %04x\n", k, v)
	})
}

func TestCacheKeepAlive(t *testing.T) {
	// add an expiring item
	table := cache2go.Cache("testKeepAlive")
	p := table.Add(k, []byte(v), 250*time.Millisecond)

	// keep it alive before it expires
	time.Sleep(100 * time.Millisecond)
	p.KeepAlive()

	// check it's still alive after it was initially supposed to expire
	time.Sleep(150 * time.Millisecond)
	if !table.Exists(k) {
		t.Error("Error keeping item alive")
	}

	// check it expires eventually
	time.Sleep(300 * time.Millisecond)
	if table.Exists(k) {
		t.Error("Error expiring item after keeping it alive")
	}
}

func TestDelete(t *testing.T) {
	// add an item to the cache
	table := cache2go.Cache("testDelete")
	table.Add(k, []byte(v), 0)
	// check it's really cached
	p, err := table.Value(k)
	if err != nil {
		t.Error("Error retrieving data from cache", err)
	}

	if string(p) != v {
		t.Errorf("Error retrieving data from cache: %s", p)
	}

	// try to delete it
	table.Delete(k)
	// verify it has been deleted
	p, err = table.Value(k)
	if err == nil || p != nil {
		t.Error("Error deleting data")
	}

	// test error handling
	_, err = table.Delete(k)
	if err == nil {
		t.Error("Expected error deleting item")
	}
}

func TestFlush(t *testing.T) {
	// add an item to the cache
	table := cache2go.Cache("testFlush")
	table.Add(k, []byte(v), 10*time.Second)
	// flush the entire table
	table.Flush()

	// try to retrieve the item
	p, err := table.Value(k)
	if err == nil || p != nil {
		t.Error("Error flushing table")
	}
	// make sure there's really nothing else left in the cache
	if table.Count() != 0 {
		t.Error("Error verifying count of flushed table")
	}
}

func TestCount(t *testing.T) {
	// add a huge amount of items to the cache
	table := cache2go.Cache("testCount")
	count := 100000
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		table.Add(key, []byte(v), 10*time.Second)
	}
	// confirm every single item has been cached
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		p, err := table.Value(key)
		if err != nil {
			t.Error("Error retrieving data")
		}

		if string(p) != v {
			t.Errorf("Error retrieving data: %s", p)
		}
	}
	// make sure the item count matches (no dupes etc.)
	if table.Count() != count {
		t.Error("Data count mismatch")
	}
}

func TestDataLoader(t *testing.T) {
	// setup a cache with a configured data-loader
	table := cache2go.Cache("testDataLoader")
	table.SetDataLoader(func(key string, args ...interface{}) *cache2go.CacheItem {
		var item *cache2go.CacheItem
		if key != "nil" {
			val := k + key
			i := cache2go.NewCacheItem(k, []byte(val), 500*time.Millisecond)
			item = i
		}

		return item
	})

	// make sure data-loader works as expected and handles unloadable keys
	_, err := table.Value("nil")
	if err == nil || table.Exists("nil") {
		t.Error("Error validating data loader for nil values")
	}

	// retrieve a bunch of items via the data-loader
	for i := 0; i < 10; i++ {
		key := k + strconv.Itoa(i)
		vp := k + key
		p, err := table.Value(key)
		if err != nil {
			t.Error("Error validating data loader")
		}

		if string(p) != vp {
			t.Errorf("Error validating data loader: expected %s, got %s", vp, p)
		}
	}
}

func TestCallbacks(t *testing.T) {
	var m sync.Mutex
	addedKey := ""
	removedKey := ""
	calledAddedItem := false
	calledRemoveItem := false
	expired := false
	calledExpired := false

	// setup a cache with AddedItem & SetAboutToDelete handlers configured
	table := cache2go.Cache("testCallbacks")
	table.SetAddedItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		addedKey = item.Key()
		m.Unlock()
	})
	table.SetAddedItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		calledAddedItem = true
		m.Unlock()
	})
	table.SetAboutToDeleteItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		removedKey = item.Key()
		m.Unlock()
	})

	table.SetAboutToDeleteItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		calledRemoveItem = true
		m.Unlock()
	})
	// add an item to the cache and setup its AboutToExpire handler
	i := table.Add(k, []byte(v), 500*time.Millisecond)
	i.SetAboutToExpireCallback(func(key interface{}) {
		m.Lock()
		expired = true
		m.Unlock()
	})

	i.SetAboutToExpireCallback(func(key interface{}) {
		m.Lock()
		calledExpired = true
		m.Unlock()
	})

	// verify the AddedItem handler works
	time.Sleep(250 * time.Millisecond)
	m.Lock()
	if addedKey == k && !calledAddedItem {
		t.Error("AddedItem callback not working")
	}
	m.Unlock()
	// verify the AboutToDelete handler works
	time.Sleep(500 * time.Millisecond)
	m.Lock()
	if removedKey == k && !calledRemoveItem {
		t.Error("AboutToDeleteItem callback not working:" + k + "_" + removedKey)
	}
	// verify the AboutToExpire handler works
	if expired && !calledExpired {
		t.Error("AboutToExpire callback not working")
	}
	m.Unlock()

}

func TestCallbackQueue(t *testing.T) {
	var m sync.Mutex
	addedKey := ""
	addedkeyCallback2 := ""
	secondCallbackResult := "second"
	removedKey := ""
	removedKeyCallback := ""
	expired := false
	calledExpired := false
	// setup a cache with AddedItem & SetAboutToDelete handlers configured
	table := cache2go.Cache("testCallbacks")

	// test callback queue
	table.AddAddedItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		addedKey = item.Key()
		m.Unlock()
	})
	table.AddAddedItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		addedkeyCallback2 = secondCallbackResult
		m.Unlock()
	})

	table.AddAboutToDeleteItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		removedKey = item.Key()
		m.Unlock()
	})
	table.AddAboutToDeleteItemCallback(func(item *cache2go.CacheItem) {
		m.Lock()
		removedKeyCallback = secondCallbackResult
		m.Unlock()
	})

	i := table.Add(k, []byte(v), 500*time.Millisecond)
	i.AddAboutToExpireCallback(func(key interface{}) {
		m.Lock()
		expired = true
		m.Unlock()
	})
	i.AddAboutToExpireCallback(func(key interface{}) {
		m.Lock()
		calledExpired = true
		m.Unlock()
	})

	time.Sleep(250 * time.Millisecond)
	m.Lock()
	if addedKey != k && addedkeyCallback2 != secondCallbackResult {
		t.Error("AddedItem callback queue not working")
	}
	m.Unlock()

	time.Sleep(500 * time.Millisecond)
	m.Lock()
	if removedKey != k && removedKeyCallback != secondCallbackResult {
		t.Error("Item removed callback queue not working")
	}
	m.Unlock()

	// test removing of the callbacks
	table.RemoveAddedItemCallbacks()
	table.RemoveAboutToDeleteItemCallback()
	secondItemKey := "itemKey02"
	expired = false
	i = table.Add(secondItemKey, []byte(v), 500*time.Millisecond)
	i.SetAboutToExpireCallback(func(key interface{}) {
		m.Lock()
		expired = true
		m.Unlock()
	})
	i.RemoveAboutToExpireCallback()

	// verify if the callbacks were removed
	time.Sleep(250 * time.Millisecond)
	m.Lock()
	if addedKey == secondItemKey {
		t.Error("AddedItemCallbacks were not removed")
	}
	m.Unlock()

	// verify the AboutToDelete handler works
	time.Sleep(500 * time.Millisecond)
	m.Lock()
	if removedKey == secondItemKey {
		t.Error("AboutToDeleteItem not removed")
	}
	// verify the AboutToExpire handler works
	if !expired && !calledExpired {
		t.Error("AboutToExpire callback not working")
	}
	m.Unlock()
}

func TestLogger(t *testing.T) {
	// setup a logger
	out := new(bytes.Buffer)
	l := log.New(out, "cache2go ", log.Ldate|log.Ltime)

	// setup a cache with this logger
	table := cache2go.Cache("testLogger")
	table.SetLogger(l)
	table.Add(k, []byte(v), 0)

	time.Sleep(100 * time.Millisecond)

	// verify the logger has been used
	if out.Len() == 0 {
		t.Error("Logger is empty")
	}
}

func TestSanityChecks(t *testing.T) {
	item := cache2go.NewCacheItem("key", []byte("value"), 0)

	if item.LifeSpan() != 0 {
		t.Error("LifeSpan should be 0")
	}

	if item.Key() != "key" {
		t.Error("Key should be 'key'")
	}

	if string(item.Data()) != "value" {
		t.Error("Data should be 'value'")
	}

	if item.AccessedOn().Unix() != time.Now().Unix() {
		t.Error("AccessCount should be 0")
	}

	if item.CreatedOn().Unix() != time.Now().Unix() {
		t.Error("CreatedOn should be 0")
	}
}
