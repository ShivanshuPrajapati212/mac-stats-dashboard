package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (safe for local dev)
	},
}

type Message struct {
	CPU      float64 `json:"cpu"`
	TotalRAM float64 `json:"total_ram"`
	UsedRAM  float64 `json:"used_ram"`
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		cpuPercent, _ := cpu.Percent(0, false)
		memStats, _ := mem.VirtualMemory()

		stats := Message{CPU: cpuPercent[0], TotalRAM: float64(memStats.Total), UsedRAM: float64(memStats.Used)}

		err = conn.WriteJSON(stats)
		if err != nil {
			log.Println("Write error:", err)
			break
		}

		time.Sleep(1 * time.Second)
	}
}

func getProcessByName(name string) (*process.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		n, _ := p.Name()
		if n == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("process not found")
}
