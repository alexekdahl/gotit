package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gliderlabs/ssh"
)

type Tunnel struct {
	w      io.Writer
	donech chan struct{}
}

var tunnels = map[int]chan Tunnel{}

func main() {
	go func() {
		http.HandleFunc("/", handleReq)
		http.ListenAndServe(":3000", nil)
	}()
	ssh.Handle(func(s ssh.Session) {
		id := rand.Intn(math.MaxInt)
		tunnels[id] = make(chan Tunnel)

		fmt.Println("tunnel ID -> %i", id)
		tunnel := <-tunnels[id]
		fmt.Println("tunnel ready")

		_, err := io.Copy(tunnel.w, s)
		if err != nil {
			log.Fatal(err)
		}

		close(tunnel.donech)

		_, err = s.Write([]byte("done"))
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Fatal(ssh.ListenAndServe(":2222", nil))
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	idstr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idstr)
	tunnel, ok := tunnels[id]
	if !ok {
		w.Write([]byte("not found"))
		return
	}
	donech := make(chan struct{})
	tunnel <- Tunnel{
		w:      w,
		donech: donech,
	}
	<-donech
}
