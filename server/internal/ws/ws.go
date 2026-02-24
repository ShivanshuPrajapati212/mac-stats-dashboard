package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

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

func HandleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer ws.Close()

	// Wait for the first message containing window size
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

	log.Printf("client window: %dx%d\n", size.Width, size.Height)

	cmd := exec.Command("btop")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: size.Width,
		Rows: size.Height,
	})
	if err != nil {
		log.Println("pty error:", err)
		return
	}
	defer f.Close()
	defer cmd.Wait()

	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		if err := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			break
		}
	}
}
