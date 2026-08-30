package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestGrantStoreConcurrentAccess(t *testing.T) {
	store := &GrantStore{Grants: make(map[GrantID]PasteID)}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		grant := GrantID(fmt.Sprintf("grant-%d", i))
		wg.Add(1)
		go func(grant GrantID) {
			defer wg.Done()
			store.mu.Lock()
			store.Grants[grant] = "paste"
			store.mu.Unlock()
			if paste, ok := store.Get(grant); !ok || paste != "paste" {
				t.Errorf("Get(%q) = %q, %v", grant, paste, ok)
			}
			store.mu.Lock()
			delete(store.Grants, grant)
			store.mu.Unlock()
		}(grant)
	}
	wg.Wait()
}
