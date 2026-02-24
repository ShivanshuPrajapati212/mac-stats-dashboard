package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (safe for local dev)
	},
}

type WindowSize struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func (h *Hub) add(ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ws] = struct{}{}
}

func (h *Hub) remove(ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ws)
}

func (h *Hub) broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ws := range h.clients {
		if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
			ws.Close()
			delete(h.clients, ws)
		}
	}
}

func (h *Hub) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

var hub = &Hub{clients: make(map[*websocket.Conn]struct{})}

func startBtop(cols, rows uint16) {
	cmd := exec.Command("btop")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		log.Println("pty error:", err)
		return
	}
	defer f.Close()
	defer cmd.Wait()

	var (
		buf    = make([]byte, 4096)
		chunk  []byte
		ticker = time.NewTicker(5 * time.Second)
	)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if len(chunk) == 0 {
				continue
			}
			hub.broadcast(chunk)
			chunk = nil
		}
	}()

	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		chunk = append(chunk, buf[:n]...)
	}
}

var (
	btopOnce    sync.Once
	btopRunning bool
	btopMu      sync.Mutex
)

func ensureBtop(cols, rows uint16) {
	btopMu.Lock()
	defer btopMu.Unlock()
	if btopRunning {
		return
	}
	btopRunning = true
	go func() {
		startBtop(cols, rows)
		btopMu.Lock()
		btopRunning = false
		btopMu.Unlock()
	}()
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer ws.Close()

	// Read window size from client
	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("read size error:", err)
		return
	}

	var size WindowSize
	if err := json.Unmarshal(msg, &size); err != nil {
		log.Println("invalid size message:", err)
		return
	}

	log.Printf("client connected: %dx%d\n", size.Width, size.Height)

	hub.add(ws)
	defer func() {
		hub.remove(ws)
		log.Printf("client disconnected, remaining: %d\n", hub.count())
	}()

	// Start btop only if not already running
	ensureBtop(size.Width, size.Height)

	// Keep connection alive until client disconnects
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}
