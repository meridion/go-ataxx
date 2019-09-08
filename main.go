package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func main() {
	if false {
		http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	board := NewGame()
	board.Print()

	return
}
