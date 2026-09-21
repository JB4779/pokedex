package pokecache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	for _, tc := range []struct{ key, val string }{
		{"https://example.com", "testdata"},
		{"https://example.com/path", "moretestdata"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			c := NewCache(5 * time.Second)
			defer c.Close()
			if _, ok := c.Get(tc.key); ok {
				t.Fatal("unexpected cache hit")
			}
			input := []byte(tc.val)
			c.Add(tc.key, input)
			input[0] = '!'
			got, ok := c.Get(tc.key)
			if !ok || string(got) != tc.val {
				t.Fatalf("Get = %q, %v", got, ok)
			}
			got[0] = '!'
			again, _ := c.Get(tc.key)
			if string(again) != tc.val {
				t.Fatal("caller mutated cache")
			}
			c.Add(tc.key, []byte("replacement"))
			got, ok = c.Get(tc.key)
			if !ok || string(got) != "replacement" {
				t.Fatal("overwrite failed")
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	c := NewCache(5 * time.Millisecond)
	defer c.Close()
	c.Add("old", []byte("data"))
	// Age the entry under the lock to avoid depending on a narrow sleep window.
	c.mu.Lock()
	entry := c.entries["old"]
	entry.createdAt = time.Now().Add(-time.Second)
	c.entries["old"] = entry
	c.mu.Unlock()
	deadline := time.After(time.Second)
	poll := time.NewTicker(time.Millisecond)
	defer poll.Stop()
	for {
		select {
		case <-poll.C:
			c.mu.Lock()
			_, exists := c.entries["old"]
			c.mu.Unlock()
			if !exists {
				return
			}
		case <-deadline:
			t.Fatal("reaper did not remove expired entry")
		}
	}
}

func TestGetExpired(t *testing.T) {
	c := NewCache(time.Hour)
	defer c.Close()
	c.Add("old", []byte("data"))
	c.mu.Lock()
	entry := c.entries["old"]
	entry.createdAt = time.Now().Add(-2 * time.Hour)
	c.entries["old"] = entry
	c.mu.Unlock()
	if _, ok := c.Get("old"); ok {
		t.Fatal("returned expired entry")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := NewCache(time.Millisecond)
	defer c.Close()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprint(j % 10)
				c.Add(key, []byte("data"))
				c.Get(key)
			}
		}()
	}
	wg.Wait()
	c.Close()
	c.Close()
}
