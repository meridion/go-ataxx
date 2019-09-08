/* The game of Ataxx implemented with minimax */
package main

import (
	"fmt"
)

/* Ataxx is a board game, played on a 7 by 7 grid and included as a puzzle in
 * the 7th guest.
 *
 * 2 players battle for battle for dominance on the board, trying to, in a
 * reversi-esque fashion, gain the largest number of pieces.
 *
 * The game starts with players occupying opposing corners on the grid:
 *
 *  X * * * * * O
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  O * * * * * X
 *
 * Like some sort of bacteria (this game is also called infection) a single
 * piece can either subdivide to a neighbouring cell on the grid. (including
 * diagonally)
 *
 * e.g. X subdivides
 *  X * * * * * O
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * X *
 *  O * * * * * X
 *
 * Or a piece can jump over a neighbouring cell to a cell 2 moves away.
 * In this case the piece doesn't subdivide.
 *
 * e.g. X jumps
 *  X * * * * * O
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * * * *
 *  * * * * X * *
 *  * * * * * * *
 *  O * * * * * *
 *
 * On reaching a new cell the piece in question infects all neighbouring cell's
 * opposing pieces.
 *
 * e.g.
 * Starting from
 *  * * * * * * *
 *  * * O * * * *
 *  * O O * X * *
 *  * * O * * * *
 *  * * * * * * *
 *
 * X subdivides left
 *  * * * * * * *
 *  * * O * * * *
 *  * O O X X * *
 *  * * O * * * *
 *  * * * * * * *
 *
 * Infection takes place
 *  * * * * * * *
 *  * * X * * * *
 *  * O X X X * *
 *  * * X * * * *
 *  * * * * * * *
 *
 * Ending X's turn.
 *
 * The game continues until neither player is able to move (no empty cells
 * remain). Upon which the player with the most pieces wins. As the grid
 * contains an odd number of cells, there will always be a victor.
 */

/* The 7 by 7 board.
 *
 * We use the following integer values.
 *  0 -> empty cell
 *  1 -> player X
 * -1 -> player O
 */
type AtaxxBoard [7][7]int

/* Compute Ataxx score
 *
 * This is a simple sumation of all player pieces, since the player with the
 * most pieces wins.
 */
func (board AtaxxBoard) Score() (score int) {
	score = 0

	/* Iterate board */
	for x := 0; x < 7; x++ {
		for y := 0; y < 7; y++ {
			score += board[y][x]
		}
	}

	return
}

/* Return valid board states that can be reached by the given player for the
 * given board.
 *
 * This function will iterate all empty cells and check if pieces of the
 * given color are in range to either subdivide or jump to the position.
 * Appending all those boards to the resulting boards returned.
 *
 * In case empty cells remain, but the player cannot reach any of them return
 * the same board as only result. (forced passed turn)
 * e.g.
 * X X X X X X X
 * X O O O O O X
 * X O O O O O X
 * X O O . O O X
 * X O O O O O X
 * X O O O O O X
 * X X X X X X X
 * needless to say these instances are rare, but they do exist.
 *
 * In case no empty cells remain, return empty slice, signalling end of game.
 */
func (board AtaxxBoard) NextBoards(color int) []MinimaxableGameboard {
	results := make([]MinimaxableGameboard, 0)

	hasEmptyCell := false

	/* Iterate board */
	for x := 0; x < 7; x++ {
		for y := 0; y < 7; y++ {
			fmt.Println(x, y)
			/* Found an empty cell */
			if board[y][x] == 0 {
				hasEmptyCell = true
				hasSubdivided := false

				/* Now iterate the neighbourhood around this cell
				 * with X and/or Y having a disposition of 1 meaning
				 * subdivide distance.
				 * and X and/or Y having a disposition of 2 meaning
				 * jump distance.
				 *
				 * i.e.
				 * J J J J J
				 * J S S S J
				 * J S . S J
				 * J S S S J
				 * J J J J J
				 *
				 * with J a position to jump from, and S a position to
				 * subdivide from.
				 *
				 * Of course we cannot actually leave the board, so near the
				 * board edges this neighbourhood is clamped.
				 * We iterate the neighbourhood using inner x and inner y
				 * alternatively these variables can be called x-offset and y-offset.
				 */
				for ix := -2; ix <= 2; ix++ {
					/* Clamp bounds of X neighbourhood */
					if ix+x < 0 || ix+x >= 7 {
						continue
					}
					for iy := -2; iy <= 2; iy++ {
						/* Clamp bounds of Y neighbourhood */
						if iy+y < 0 || iy+y >= 7 {
							continue
						}

						/* Found a piece that can move to the center */
						if board[iy+y][ix+x] == color {
							isSubdivision := true

							/* Establish wether we are jumping or subdividing */
							if ix < 1 || ix > 1 {
								isSubdivision = false
							}
							if iy < 1 || iy > 1 {
								isSubdivision = false
							}

							if isSubdivision {
								/* All subdivisions add a piece to the center
								 * of the neighbourhood, not deleting any old
								 * pieces, and then mutating opposing pieces.
								 * Therefore, it doesn't matter which piece
								 * subdivided to the center. They yield the
								 * same end board state.  Since we want to
								 * prevent duplicate boards, only return the
								 * first possible subdivision detected.
								 */
								if hasSubdivided {
									continue
								}

								/* Add piece to neighbourhood center and add to
								 * resulting set.
								 */
								newBoard := board.Copy()

								/* Infect enemy pieces
								 *
								 * As with the movable piece search
								 * neighbourhood also clamp this neighbourhood
								 * to remain within the board.
								 */
								for iix := -1; iix <= 1; iix++ {
									if x+iix < 0 || x+iix >= 7 {
										continue
									}
									for iiy := -1; iiy <= 1; iiy++ {
										if y+iiy < 0 || y+iiy >= 7 {
											continue
										}
										if board[y+iiy][x+iix] == -color {
											board[y+iiy][x+iix] = color
										}
									}
								}

								newBoard[y][x] = color
								results = append(results, newBoard)

								/* Handle jumping */
							} else {
								/* Every jump is unique, as the piece
								 * performing the jump leaves its original
								 * position.
								 * Therefore we add every jump as a new board
								 * configuration.
								 */
								newBoard := board.Copy()
								/* Add center piece */
								newBoard[y][x] = color
								/* Removing piece that jumped */
								newBoard[iy+y][ix+x] = 0
								/* Infect enemy pieces
								 *
								 * As with the movable piece search
								 * neighbourhood also clamp this neighbourhood
								 * to remain within the board.
								 */
								for iix := -1; iix <= 1; iix++ {
									if x+iix < 0 || x+iix >= 7 {
										continue
									}
									for iiy := -1; iiy <= 1; iiy++ {
										if y+iiy < 0 || y+iiy >= 7 {
											continue
										}
										if board[y+iiy][x+iix] == -color {
											board[y+iiy][x+iix] = color
										}
									}
								}
								results = append(results, newBoard)
							}
						}
					}
				}
			}
		}
	}

	/* Return same board state as singular result
	 * in case moves remain, but current player cannot make them.
	 *
	 * NOTE: this hasEmptyCell variable is an optimisation.
	 * We could also simply call the board.Finished() function,
	 * however, we have already performed the necessary computations.
	 * So this saves time.
	 * It does clutter the function's code, which is the downside here.
	 */
	if hasEmptyCell && len(results) == 0 {
		results = append(results, board)
	}

	return results
}

/* Return true if the game is completed
 *
 * in practice this means checking if no empty cells (zeroes) remain.
 */
func (board AtaxxBoard) Finished() bool {
	/* Iterate board */
	for x := 0; x < 7; x++ {
		for y := 0; y < 7; y++ {
			if board[y][x] == 0 {
				return false
			}
		}
	}

	return true
}

/* This function returns a deep copy of the given board. */
func (board AtaxxBoard) Copy() (newBoard AtaxxBoard) {
	// Do we even need this?
	//newBoard = make([7][7]int

	/* Iterate board */
	for x := 0; x < 7; x++ {
		for y := 0; y < 7; y++ {
			newBoard[y][x] = board[y][x]
		}
	}

	return
}

/* Return a freshly initialized game board in starting positions */
func NewGame() AtaxxBoard {
	newBoard := AtaxxBoard{}

	/* Iterate board */
	for x := 0; x < 7; x++ {
		for y := 0; y < 7; y++ {
			newBoard[y][x] = 0
		}
	}

	/* Initialize corner positions */
	newBoard[0][0] = 1
	newBoard[6][6] = 1

	newBoard[0][6] = -1
	newBoard[6][0] = -1

	return newBoard
}

/* Print board */
func (board AtaxxBoard) Print() {
	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			if board[y][x] > 0 {
				fmt.Print(" X")
			} else if board[y][x] < 0 {
				fmt.Print(" O")
			} else {
				fmt.Print(" .")
			}
		}
		fmt.Printf("\n")
	}

	return
}
