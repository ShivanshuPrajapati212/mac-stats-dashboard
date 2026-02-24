package handlers

import (
	"fmt"
	"net/http"
)

func SetupHandlers() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello Dashboard")
	})

	http.ListenAndServe(":42067", nil)
}
