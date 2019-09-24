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
type AtaxxBoard [7][7]int8

/* This single bitboard type allows us
 * to define some bithacking methods
 * on the standard uint64 type
 */
type SingleBitboard uint64

/* The 7 by 7 bitboard
 *
 * Two arrays formed by bits.
 * One 49-bit array for all maximizingPlayer pieces.
 * One 49-bit array for all minimizingPlayer pieces.
 *
 * The arrays follow the same ordering as the original int array board.
 * 7 contiguous bits are a single line of X-coords.
 * 0 1 2 3 4 5 6
 * 7 8 9 A B C D etc.
 */
type AtaxxBitboard struct {
	maximizingPlayer SingleBitboard
	minimizingPlayer SingleBitboard
}

/* Bitboard lookup tables */
var moveMask, subdivideMask, jumpMask [49]SingleBitboard

/* A bitboard for storing board data by player on the move
 * instead of player strategy.
 * This type simplifies a bit of code, and allows us to
 * implement performing of moves disregarding wether
 * maximizingPlayer or minimizingPlayer makes the move.
 */
type MoveBitboard struct {
	movingPlayer  SingleBitboard
	waitingPlayer SingleBitboard
}

/* A transposition table for storing Ataxx boards */
type AtaxxTranspositionTable struct {
	transpositionMap map[AtaxxTransposition]AtaxxTranspositionResult
	maxSize          int
}

/* A single Ataxx ply (board + player on turn), used by HTTP server */
type AtaxxPly struct {
	board            AtaxxBoard
	maximizingPlayer bool
}

/* A single Ataxx Transposition */
type AtaxxTransposition struct {
	AtaxxPly
	depth, alpha, beta int
}

/* A stored transposition result */
type AtaxxTranspositionResult struct {
	resultBoard AtaxxBoard
	resultScore int
}

/* A transposition table for storing AtaxxBit boards */
type AtaxxBitTranspositionTable struct {
	transpositionMap map[AtaxxBitTransposition]AtaxxBitTranspositionResult
	maxSize          int
}

/* A single AtaxxBit Transposition */
type AtaxxBitTransposition struct {
	board              AtaxxBitboard
	maximizingPlayer   bool
	depth, alpha, beta int
}

/* A stored bit transposition result */
type AtaxxBitTranspositionResult struct {
	resultBoard AtaxxBitboard
	resultScore int
}

/* Compute Ataxx score
 *
 * This is a simple sumation of all player pieces, since the player with the
 * most pieces wins.
 */
func (board *AtaxxBoard) Score() (score int) {
	score = 0

	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			score += int(board[y][x])
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
func (board *AtaxxBoard) NextBoards(maximizingPlayer bool) []MinimaxableGameboard {
	results := make([]MinimaxableGameboard, 0)

	var color int8 = 1
	if !maximizingPlayer {
		color = -1
	}

	/* This is for checking wether the entire board is full
	 * and the game is therefore finished.
	 */
	hasEmptyCell := false

	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			/* Found an empty cell */
			if board[y][x] == 0 {
				hasEmptyCell = true
				hasSubdivision := false

				/* This is another optimization
				 * Every move to a certain position
				 * has a few effects that are always the same:
				 * - The enemy pieces are infected, that is,
				 *   no directly neighbouring pieces are of enemy color.
				 * - The center piece is taken by the player also.
				 * These are the same, wether we jump or subdivide.
				 * This being the case, we can cache this state and use it as a
				 * template for setting up the final board positions for moves
				 * to the center of this neighbourhood.
				 */
				var newBoardTemplate *AtaxxBoard = nil

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
				for iy := -2; iy <= 2; iy++ {
					/* Clamp bounds of Y neighbourhood */
					if iy+y < 0 || iy+y >= 7 {
						continue
					}
					for ix := -2; ix <= 2; ix++ {
						/* Clamp bounds of X neighbourhood */
						if ix+x < 0 || ix+x >= 7 {
							continue
						}

						/* Found a piece that can move to the center */
						if board[iy+y][ix+x] == color {
							/* Setup move cache if it was not initialized
							 * see explanation near declaration for details.
							 */
							if newBoardTemplate == nil {
								newBoardTemplate = &AtaxxBoard{}
								*newBoardTemplate = *board

								/* Infect enemy pieces
								 *
								 * As with the movable piece search
								 * neighbourhood also clamp this neighbourhood
								 * to remain within the board.
								 */
								for iiy := -1; iiy <= 1; iiy++ {
									if y+iiy < 0 || y+iiy >= 7 {
										continue
									}
									for iix := -1; iix <= 1; iix++ {
										if x+iix < 0 || x+iix >= 7 {
											continue
										}
										if newBoardTemplate[y+iiy][x+iix] == -color {
											newBoardTemplate[y+iiy][x+iix] = color
										}
									}
								}

								/* Add piece to neighbourhood center completing
								 * the template.
								 */
								newBoardTemplate[y][x] = color
							} /* Setup template */

							/* Establish wether we are jumping or subdividing */
							isSubdivision := true
							if ix < -1 || ix > 1 {
								isSubdivision = false
							}
							if iy < -1 || iy > 1 {
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
								 *
								 * To make move order returned equal to the move
								 * order of the bitboards, return subdivision last.
								 * (also making minimax favor jumps in the process)
								 */
								hasSubdivision = true

							} else { /* Handle jumping */
								/* Every jump is unique, as the piece
								 * performing the jump leaves its original
								 * position.
								 * Therefore we add every jump as a new board
								 * configuration.
								 *
								 * However since we have our board template set
								 * up we only need to copy the template and
								 * remove the piece that jumped from it.
								 */
								/* Copy template data */
								newBoard := &AtaxxBoard{}
								*newBoard = *newBoardTemplate

								/* Remove piece that jumped */
								newBoard[iy+y][ix+x] = 0

								/* Add to total moves available */
								results = append(results, newBoard)
							}
						}
					} /* inner-X loop */
				} /* inner-Y loop */

				if hasSubdivision {
					/* Since subdivision is equal to the board template
					 * state we have cached simply add the template
					 * to the set.
					 */
					results = append(results, newBoardTemplate)
				}
			} /* if empty */
		} /* for x */
	} /* for y */

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
func (board *AtaxxBoard) Finished() bool {
	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			if board[y][x] == 0 {
				return false
			}
		}
	}

	return true
}

/* Return a freshly initialized game board in starting positions */
func NewGame() *AtaxxBoard {
	newBoard := AtaxxBoard{}

	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			newBoard[y][x] = 0
		}
	}

	/* Initialize corner positions */
	newBoard[0][0] = 1
	newBoard[6][6] = 1

	newBoard[0][6] = -1
	newBoard[6][0] = -1

	return &newBoard
}

/* Print board */
func (board *AtaxxBoard) Print() {
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

/* Load a previously computed board from our cache */
func (table *AtaxxTranspositionTable) Load(game MinimaxableGameboard, maximizingPlayer bool, depth int, alpha int, beta int) (MinimaxableGameboard, int, bool) {
	key := AtaxxTransposition{AtaxxPly{*(game.(*AtaxxBoard)), maximizingPlayer}, depth, alpha, beta}

	/* Maps return "zero" values, so in our case an empty board and a 0 score */
	res, found := table.transpositionMap[key]
	return &res.resultBoard, res.resultScore, found
}

/* Store a board to the cache
 *
 * This function is where the "magic" happens.
 * For now use an incredibly simple replacement strategy.
 * Whenever our hash table hits the maximum size, we clear the hash table.
 */
func (table *AtaxxTranspositionTable) Store(game MinimaxableGameboard, maximizingPlayer bool, depth int, alpha int, beta int, resultBoard MinimaxableGameboard, resultScore int) {
	key := AtaxxTransposition{AtaxxPly{*(game.(*AtaxxBoard)), maximizingPlayer}, depth, alpha, beta}

	/* Clear hash table if we are about to grow past maximum size */
	if len(table.transpositionMap) == table.maxSize {
		table.transpositionMap = make(map[AtaxxTransposition]AtaxxTranspositionResult)
	}

	table.transpositionMap[key] = AtaxxTranspositionResult{*(resultBoard.(*AtaxxBoard)), resultScore}
}

/* Build a new table with the predefined size */
func NewTranspositionTable(size int) *AtaxxTranspositionTable {
	table := AtaxxTranspositionTable{}
	table.transpositionMap = make(map[AtaxxTransposition]AtaxxTranspositionResult)
	table.maxSize = size

	return &table
}

/* Score player status.
 *
 * Returns the heuristic board score.
 *
 * Since the scoring is done by subtracting the minimizing player total pieces
 * from the maximizing player total pieces, we need to count the bits in both
 * bit arrays.
 */
func (board *AtaxxBitboard) Score() int {
	return board.maximizingPlayer.PiecesPlaced() - board.minimizingPlayer.PiecesPlaced()
}

/* Count the number of bits set in a bitboard array.
 *
 * In other words, the number of pieces placed within the array.
 * For counting we use Kernighan's method:
 * https://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetNaive
 */
func (board SingleBitboard) PiecesPlaced() int {
	var pieces int

	/* This loop clears the least significant bit set, until no bits remain */
	for pieces = 0; board != 0; pieces++ {
		board &= board - 1
	}

	return pieces
}

/* Return valid board states that can be reached by the given player for the
 * given board.
 *
 * This function operates on bitboards by using a precomputed lookup table
 * containing bitboard neighbourhoods for all 49 possibly empty cells.
 *
 * These are then used for determining wether a move is possible, and then
 * for efficiently computing the new board state.
 *
 * Arguments:
 *  maximizingPlayer: true if the maximizingPlayer is making the move
 *  false otherwise.
 */
func (board *AtaxxBitboard) NextBoards(maximizingPlayer bool) []MinimaxableGameboard {
	results := make([]MinimaxableGameboard, 0)

	/* Handle case where we are finished already */
	if board.Finished() {
		results = append(results, board)
		return results
	}

	var move MoveBitboard

	if maximizingPlayer {
		move.movingPlayer = board.maximizingPlayer
		move.waitingPlayer = board.minimizingPlayer
	} else {
		move.movingPlayer = board.minimizingPlayer
		move.waitingPlayer = board.maximizingPlayer
	}

	//fmt.Println("movingPlayer")
	//move.movingPlayer.Print()
	//fmt.Println("waitingPlayer")
	//move.waitingPlayer.Print()
	emptyCells := (^(move.movingPlayer | move.waitingPlayer)) & ((1 << 49) - 1)
	//fmt.Println("emtpyCells")
	//emptyCells.Print()

	/* Loop over all 49 possible cells (in the 7x7 bitboard) */
	for bit := uint(0); bit < 49; bit++ {
		//fmt.Println("bit:", bit)
		//moveMask[bit].Print()
		/* To know if we can make a move, we need to know 2 things:
		 * 1. Is the cell empty? (no player pieces set to 1)
		 * 2. Are any moving player pieces in range? (check neighbourhood mask)
		 */
		//(emptyCells & (1 << bit)).Print()
		if (emptyCells&(1<<bit)) != 0 && move.movingPlayer&moveMask[bit] != 0 {
			//fmt.Println("empty movable cell")
			/* Compute move template.
			 *
			 * Since every move infects target pieces and places a piece of the
			 * moving player in the empty cell, we can cache this part.
			 */

			/* Copy data */
			newMoveTemplate := move

			/* Set empty cell to piece */
			newMoveTemplate.movingPlayer |= (1 << bit)

			/* Infect surrounding cells
			 *
			 * First we compute all enemy pieces in range by
			 * using the subdivide mask to gather those pieces.
			 *
			 * Second we add those pieces to the moving player mask
			 * gaining control of them.
			 *
			 * Finally we delete those pieces from the waiting player
			 * losing control.
			 */
			infectionMask := newMoveTemplate.waitingPlayer & subdivideMask[bit]
			newMoveTemplate.movingPlayer |= infectionMask
			newMoveTemplate.waitingPlayer &= ^infectionMask

			/* Handle jumps, if any
			 *
			 * Jumps are handled by using a bit hack to fetch the LSB (least
			 * significant bit) from the jumpingMask and clearing this
			 * simultaneously from the computed move template (appending the
			 * result) and from the jumpMask, creating a new LSB.
			 *
			 * This proces continuous until no more jump capable pieces remain
			 * in the jump mask, and all possible board jumps have been
			 * performed.
			 */
			jumpingMask := move.movingPlayer & jumpMask[bit]
			for jumpingMask != 0 {
				/* Fetch LSB from jumpingMask */
				nextJump := (^jumpingMask + 1) & jumpingMask

				/* Append next move board, clearing jumped piece */
				newMove := newMoveTemplate
				newMove.movingPlayer ^= nextJump
				results = append(results, newMove.ToMinimaxBoard(maximizingPlayer))

				/* Finally clear jumped piece from jumping mask */
				jumpingMask ^= nextJump
			}

			/* Now check if this player can subdivide
			 *
			 * We also intentionally return subdivision last,
			 * to make move order identical to non-bitboard computation.
			 * This also does make minimax favor jumps, which we might want
			 * to change later on.
			 */
			if move.movingPlayer&subdivideMask[bit] != 0 {
				/* Add subdivided board to results */
				results = append(results, newMoveTemplate.ToMinimaxBoard(maximizingPlayer))
			}

		}
	}

	return results
}

/* The game is finished if no more empty cells remain.
 *
 * That is, if both arrays have the first 49 bits set together, the game is
 * over.
 */
func (board *AtaxxBitboard) Finished() bool {
	return (board.maximizingPlayer | board.minimizingPlayer) == ((1 << 49) - 1)
}

/* Initialize bitboard lookup tables */
func InitBitboards() {
	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			/* Compute cell index in mask array */
			maskIndex := y*7 + x

			/* Clear initial mask contents */
			moveMask[maskIndex] = 0
			subdivideMask[maskIndex] = 0
			jumpMask[maskIndex] = 0

			/* Compute mask neighbourhood
			 *
			 * The board has this shape:
			 * . . . . . . .
			 * . . . . . . .
			 * . . . . . . .
			 * . . . . . . .
			 * . . . . . . .
			 * . . . . . . .
			 * . . . . . . .
			 *
			 * The masks are used to determine:
			 * - What stones can move to the neighbourhood center
			 * - What stones can subdivide to the neighbourhood center
			 * - What stones can jump to the neighbourhood center
			 *
			 * e.g.
			 * The different masks for the upper left corner then become
			 *
			 *   move mask     subdivide mask    jump mask
			 * . M M . . . .   . S . . . . .   . . J . . . .
			 * M M M . . . .   S S . . . . .   . . J . . . .
			 * M M M . . . .   . . . . . . .   J J J . . . .
			 * . . . . . . .   . . . . . . .   . . . . . . .
			 * . . . . . . .   . . . . . . .   . . . . . . .
			 * . . . . . . .   . . . . . . .   . . . . . . .
			 * . . . . . . .   . . . . . . .   . . . . . . .
			 *
			 * The subdivide mask also doubles as a mask for determining
			 * infected stones.
			 */
			for iy := -2; iy <= 2; iy++ {
				/* Clamp bounds of Y neighbourhood */
				if iy+y < 0 || iy+y >= 7 {
					continue
				}
				for ix := -2; ix <= 2; ix++ {
					/* Clamp bounds of X neighbourhood */
					if ix+x < 0 || ix+x >= 7 {
						continue
					}
					/* Skip neighbourhood center */
					if ix == 0 && iy == 0 {
						continue
					}

					/* Determine subdivision status */
					isSubdivision := true
					if ix < -1 || ix > 1 {
						isSubdivision = false
					}
					if iy < -1 || iy > 1 {
						isSubdivision = false
					}

					/* Compute current mask bit */
					maskBit := SingleBitboard(1 << uint((y+iy)*7+(x+ix)))

					/* Set masks */
					moveMask[maskIndex] |= maskBit
					if isSubdivision {
						subdivideMask[maskIndex] |= maskBit
					} else {
						jumpMask[maskIndex] |= maskBit
					}
				}
			}
			moveMask[maskIndex].Print()
		}
	}
}

/* Conversion function used for simplifying the bitboard next move computation
 * code
 */
func (move MoveBitboard) ToMinimaxBoard(maximizingPlayer bool) MinimaxableGameboard {
	minimax := AtaxxBitboard{}

	if maximizingPlayer {
		minimax.maximizingPlayer = move.movingPlayer
		minimax.minimizingPlayer = move.waitingPlayer
	} else {
		minimax.maximizingPlayer = move.waitingPlayer
		minimax.minimizingPlayer = move.movingPlayer
	}

	return &minimax
}

/* Initialize a new game, bitboard style */
func NewBitGame() *AtaxxBitboard {
	board := AtaxxBitboard{}
	board.maximizingPlayer = SingleBitboard((1 << 48) | 1)
	board.minimizingPlayer = SingleBitboard((1 << 42) | (1 << 6))

	return &board
}

/* Print bitboard */
func (board *AtaxxBitboard) Print() {
	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			maskBit := SingleBitboard(1 << uint((y)*7+x))
			if board.maximizingPlayer&maskBit != 0 {
				fmt.Print(" X")
			} else if board.minimizingPlayer&maskBit != 0 {
				fmt.Print(" O")
			} else {
				fmt.Print(" .")
			}
		}
		fmt.Printf("\n")
	}

	return
}

/* Print single bitboard */
func (board SingleBitboard) Print() {
	/* Iterate board */
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			maskBit := SingleBitboard(1 << uint((y)*7+x))
			if board&maskBit != 0 {
				fmt.Print(" #")
			} else {
				fmt.Print(" .")
			}
		}
		fmt.Printf("\n")
	}

	return
}

/* Load a previously computed bitboard from our cache */
func (table *AtaxxBitTranspositionTable) Load(game MinimaxableGameboard, maximizingPlayer bool, depth int, alpha int, beta int) (MinimaxableGameboard, int, bool) {
	key := AtaxxBitTransposition{*(game.(*AtaxxBitboard)), maximizingPlayer, depth, alpha, beta}

	/* Maps return "zero" values, so in our case an empty board and a 0 score */
	res, found := table.transpositionMap[key]
	return &res.resultBoard, res.resultScore, found
}

/* Store a board to the cache
 *
 * This function is where the "magic" happens.
 * For now use an incredibly simple replacement strategy.
 * Whenever our hash table hits the maximum size, we clear the hash table.
 */
func (table *AtaxxBitTranspositionTable) Store(game MinimaxableGameboard, maximizingPlayer bool, depth int, alpha int, beta int, resultBoard MinimaxableGameboard, resultScore int) {
	key := AtaxxBitTransposition{*(game.(*AtaxxBitboard)), maximizingPlayer, depth, alpha, beta}

	/* Clear hash table if we are about to grow past maximum size */
	if len(table.transpositionMap) == table.maxSize {
		table.transpositionMap = make(map[AtaxxBitTransposition]AtaxxBitTranspositionResult)
	}

	table.transpositionMap[key] = AtaxxBitTranspositionResult{*(resultBoard.(*AtaxxBitboard)), resultScore}
}

/* Build a new table with the predefined size */
func NewBitTranspositionTable(size int) *AtaxxBitTranspositionTable {
	table := AtaxxBitTranspositionTable{}
	table.transpositionMap = make(map[AtaxxBitTransposition]AtaxxBitTranspositionResult)
	table.maxSize = size

	return &table
}
