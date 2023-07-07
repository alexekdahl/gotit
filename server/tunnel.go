package server

import (
	"io"
	"sync"
)

type TunnelStorer interface {
	Get(id string) (chan Tunnel, bool)
	Put(id string, tunnel chan Tunnel)
	Delete(id string)
}

type Tunnel struct {
	w      io.Writer
	donech chan struct{}
}

type TunnelStore struct {
	mu      sync.RWMutex
	tunnels map[string]chan Tunnel
}

func (ts *TunnelStore) Get(id string) (chan Tunnel, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	tunnel, ok := ts.tunnels[id]
	return tunnel, ok
}

func (ts *TunnelStore) Put(id string, tunnel chan Tunnel) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tunnels[id] = tunnel
}

func (ts *TunnelStore) Delete(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tunnels, id)
}

func NewTunnel() TunnelStorer {
	tunnelStorer := &TunnelStore{
		tunnels: make(map[string]chan Tunnel),
	}
	return tunnelStorer
}
