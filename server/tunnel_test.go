package server

import (
	"fmt"
	"sync"
	"testing"
)

func TestTunnelStore(t *testing.T) {
	tunnelStore := NewTunnel()

	// Test Put method
	tunnel := make(chan Tunnel)
	tunnelStore.Put("testID", tunnel)

	// Test Get method
	retrievedTunnel, ok := tunnelStore.Get("testID")
	if !ok || retrievedTunnel != tunnel {
		t.Errorf("Get method failed, expected %v and got %v", tunnel, retrievedTunnel)
	}

	// Test Delete method
	tunnelStore.Delete("testID")
	_, ok = tunnelStore.Get("testID")
	if ok {
		t.Errorf("Delete method failed, expected tunnel to be deleted")
	}
}

func TestTunnelStoreConcurrent(t *testing.T) {
	tunnelStore := NewTunnel()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Test Put method
			tunnel := make(chan Tunnel)
			tunnelStore.Put(fmt.Sprint(i), tunnel)

			// Test Get method
			retrievedTunnel, ok := tunnelStore.Get(fmt.Sprint(i))
			if !ok || retrievedTunnel != tunnel {
				t.Errorf("Get method failed for ID %d, expected %v and got %v", i, tunnel, retrievedTunnel)
			}

			// Test Delete method
			tunnelStore.Delete(fmt.Sprint(i))
			_, ok = tunnelStore.Get(fmt.Sprint(i))
			if ok {
				t.Errorf("Delete method failed for ID %d, expected tunnel to be deleted", i)
			}
		}(i)
	}

	wg.Wait()
}
