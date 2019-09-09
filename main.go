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
	fmt.Println("Start of game")
	board.Print()

	turn := 1
	color := 1

	//for i, newBoard := range board.NextBoards(color) {
	//	fmt.Println("Next position", i)
	//	(newBoard.(AtaxxBoard)).Print()
	//}

	/* Self play until finished. */
	transposition := NewTranspositionTable(160000)
	for !board.Finished() {
		var currentPlayer string
		if color == 1 {
			currentPlayer = "X"
		} else {
			currentPlayer = "O"
		}

		fmt.Println("Turn", turn, currentPlayer, "moves")
		//newBoard, _ := Minimax(board, color, 3)
		//newBoard, _ := AlphaBeta(board, color == 1, 4, -49, 49)
		//newBoard, _ := AlphaBetaTransposition(board, color == 1, 4, -49, 49, NewTranspositionTable(60000))
		newBoard, _ := AlphaBetaTransposition(board, color == 1, 3, -49, 49, transposition)

		board = newBoard.(*AtaxxBoard)
		board.Print()
		color = -color
		turn += 1
	}
	fmt.Println("Hash tables items:", len(transposition.transpositionMap))

	return
}
