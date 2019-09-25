package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
)

func main() {
	if true {
		http.Handle("/", http.FileServer(http.Dir(".")))

		http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		})

		http.HandleFunc("/ply", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				fmt.Println("Received method", r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var lr io.LimitedReader
			/* Allow a post-body of maximum size 1kB
			 * This should be more than enough for the JSON POST requests we handle.
			 */
			lr.R = r.Body
			lr.N = 1024
			decoder := json.NewDecoder(&lr)

			/* Decode board state + player on turn */
			var ply AtaxxPly
			err := decoder.Decode(&ply)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
			fmt.Println("Received board", ply)
		})

		/* Return a new Game board in JSON AtaxxPly format over GET request */
		http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				fmt.Println("Received method", r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)
			var newGame AtaxxPly
			newGame = AtaxxPly{*NewGame(), true}
			//board := NewGame()
			//fmt.Println(newGame)
			jsonPly, err := json.Marshal(newGame)
			if err != nil {
				panic(err)
			}
			fmt.Println(jsonPly)
			err = encoder.Encode(&newGame)
			if err != nil {
				panic(err)
			}
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	/* Setup pre-calculated bitboard tables */
	InitBitboards()

	/* Initialize a new game board */
	//board := NewGame()
	board := NewBitGame()
	fmt.Println("Start of game")
	board.Print()

	turn := 1
	color := 1

	//for i, newBoard := range board.NextBoards(true) {
	//	fmt.Println("Next position", i)
	//	(newBoard.(*AtaxxBitboard)).Print()
	//}
	//return

	/* Self play until finished. */
	//transposition := NewTranspositionTable(160000)
	transposition := NewBitTranspositionTable(160000)
	for !board.Finished() {
		var currentPlayer string
		if color == 1 {
			currentPlayer = "X"
		} else {
			currentPlayer = "O"
		}

		fmt.Println("Turn", turn, currentPlayer, "moves")
		//newBoard, _ := Minimax(board, color, 3)
		//newBoard, _ := AlphaBeta(board, color == 1, 5, -49, 49)
		//newBoard, _ := AlphaBetaTransposition(board, color == 1, 4, -49, 49, NewTranspositionTable(60000))
		//newBoard, _ := AlphaBetaTransposition(board, color == 1, 5, -49, 49, transposition)
		newBoard, _ := AlphaBetaTransposition(board, color == 1, 3, -49, 49, transposition)

		//board = newBoard.(*AtaxxBoard)
		board = newBoard.(*AtaxxBitboard)
		board.Print()
		color = -color
		turn += 1
	}
	fmt.Println("Hash tables items:", len(transposition.transpositionMap))

	return
}
