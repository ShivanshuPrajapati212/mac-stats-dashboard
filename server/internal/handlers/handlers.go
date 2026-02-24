package handlers

import (
	"fmt"
	"net/http"

	"github.com/ShivanshuPrajapati212/mac-stats-dashboard/internal/ws"
)

func SetupHandlers() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello Dashboard")
	})

	http.HandleFunc("/ws", ws.HandleWS)

	http.ListenAndServe(":42067", nil)
}
