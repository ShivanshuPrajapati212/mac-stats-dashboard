package ws

import (
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

func HandleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer ws.Close()

	cmd := exec.Command("btop")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 220, Rows: 50})
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
