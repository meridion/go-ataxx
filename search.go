/* An abstract implementation of the minimax algorithm */
package main

type MinimaxableGameboard interface {
	/* Function that returns a heuristic estimate on the board positions
	 *
	 * The value should be more positive if player A is likely to win,
	 * and more negative if player B is likely to win,
	 * with close to zero meaning the players are tied.
	 *
	 * Other than that any non-Nan value is fine.
	 */
	Score() int

	/* Return an array of all valid boards a given player can move to given the
	 * current game state.
	 *
	 * Color indicates the player making the move:
	 *  1 -> Player A
	 * -1 -> Player B
	 *
	 * The function can return an empty slice if the game has reached
	 * terminal state.
	 */
	NextBoards(color int) []MinimaxableGameboard

	/* Return true if the game has reached a terminal state */
	Finished() bool
}

/* Return the best possible move a player can make based on given search depth
 * and its according score.
 *
 * color:
 *  1 -> Player A
 * -1 -> Player B
 *
 * depth:
 * Maximum search depth.
 * 0 meaning, do not recurse and simply evaluate all current moves
 * heuristically choosing the best among them.
 *
 * Technically speaking this function implements negamax.
 * Which mostly means we exploit the color value of the player
 * to spare branches. But since we're using floats, it doesn't really matter.
 */
func Minimax(game MinimaxableGameboard, color int, depth int) (maxBoard MinimaxableGameboard, maxScore int) {
	boards := game.NextBoards(color)

	/* In case the game has finish, return current game state */
	if len(boards) == 0 {
		return game, color * game.Score()
	}

	/* If we have reached maximum search depth, heuristically evaluate game
	 * boards.
	 */
	if depth == 0 {
		for i, board := range boards {
			/* Compute position heurstic */
			newScore := color * board.Score()

			/* Store best board seen */
			if i == 0 {
				maxScore = newScore
				maxBoard = board
			} else {
				if newScore > maxScore {
					maxScore = newScore
					maxBoard = board
				}
			}
		}

		/* Return best board available */
		return maxBoard, maxScore
	}

	/* If we are not at maximum search depth, iterate the various boards and
	 * score them by recursively evaluating the underlying game trees.
	 */
	for i, board := range boards {
		/* Compute enemy score by recursing */
		_, newScore := Minimax(board, -color, depth-1)

		/* Negate best enemy score, to get our worst score
		 * (this is the negamax part)
		 */
		newScore = -newScore

		/* Store best board seen */
		if i == 0 {
			maxScore = newScore
			maxBoard = board
		} else {
			if newScore > maxScore {
				maxScore = newScore
				maxBoard = board
			}
		}
	}

    return
}
