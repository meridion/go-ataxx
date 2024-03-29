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
		/* Setup pre-calculated bitboard tables */
		InitBitboards()

		/* Setup routes */
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

			/* Convert to bitboard for higher performance */
			bitboard := ply.Board.ToBitboard()

			/* Compute next computer move */
			newBoard, _ := AlphaBeta(&bitboard, ply.MaximizingPlayer, 4, -49, 49)

			/* Return resulting game state */
			var rply AtaxxPly
			rply.Board = newBoard.(*AtaxxBitboard).ToBoard()
			rply.MaximizingPlayer = !ply.MaximizingPlayer

			/* Marshal to JSON */
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)
			err = encoder.Encode(&rply)
			if err != nil {
				panic(err)
			}
		})

		/* Handle a player-made move */
		http.HandleFunc("/move", func(w http.ResponseWriter, r *http.Request) {
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
			var move AtaxxPlayerMove
			err := decoder.Decode(&move)
			if err != nil {
				panic(err)
			}

			/* Compute coordinates */
			srcX := move.Source % 7
			srcY := move.Source / 7
			tgtX := move.Target % 7
			tgtY := move.Target / 7

			/* Perform human move */
			newBoard, valid := HumanMove(&move.State.Board, move.State.MaximizingPlayer, srcX, srcY, tgtX, tgtY)

			/* Return resulting game state */
			var rply AtaxxPly
			rply.Board = newBoard
			/* No longer our turn */
			if valid {
				rply.MaximizingPlayer = !move.State.MaximizingPlayer
				/* Still our turn */
			} else {
				rply.MaximizingPlayer = move.State.MaximizingPlayer
			}

			/* Marshal to JSON */
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)
			err = encoder.Encode(&rply)
			if err != nil {
				panic(err)
			}
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
