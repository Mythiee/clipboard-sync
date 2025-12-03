package main

import (
	"github.com/mythiee/clipboard-sync/clipboard-server/internal/ws"
	"os"

	"fmt"
	"net/http"
	"runtime/debug"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error:", err)
			debug.PrintStack()
		}
	}()

	http.HandleFunc("/", ws.Handler)

	fmt.Println("Listening on port 8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
