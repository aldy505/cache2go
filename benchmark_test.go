/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2022, Reinaldy Rafli <aldy505@tutanota.com>
 *
 *   For license see LICENSE.txt
 */

package cache2go_test

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aldy505/cache2go"
)

func BenchmarkNotFoundAdd(b *testing.B) {
	table := cache2go.Cache("testNotFoundAdd")

	var finish sync.WaitGroup
	var added int32
	var idle int32

	fn := func(id int) {
		for i := 0; i < b.N; i++ {
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

}
