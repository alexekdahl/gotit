package server

import (
	"io"
	"sync"
)

type Tunnel struct {
	w      io.Writer
	donech chan struct{}
}

type TunnelStorer struct {
	mu      sync.RWMutex
	tunnels map[string]chan Tunnel
}

func (ts *TunnelStorer) Get(id string) (chan Tunnel, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	tunnel, ok := ts.tunnels[id]
	return tunnel, ok
}

func (ts *TunnelStorer) Put(id string, tunnel chan Tunnel) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tunnels[id] = tunnel
}

func (ts *TunnelStorer) Delete(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tunnels, id)
}

func NewTunnel() *TunnelStorer {
	tunnelStorer := &TunnelStorer{
		tunnels: make(map[string]chan Tunnel),
	}
	return tunnelStorer
}
