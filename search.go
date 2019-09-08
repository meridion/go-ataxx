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

///* The most naive playing algorithm. Use the score heuristic to immediately
// * select the "best" move.
// *
// * This can be considered greedy play, and will most likely lead to a weak opponent.
// * Only a search heuristic of really high quality can save the player here.
// */
//func BestMoveByRawScore(game MinimaxableGameboard, maximizingPlayer bool) (bestBoard MinimaxableGameboard, bestScore int) {
//}

/* Naive minimax
 *
 * This algorithm will search the whole tree every move up to a certain depth.
 *
 * Return the best possible move a player can make based on given search depth
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

/* Minimax with alpha-beta pruning
 *
 * This algorithm will cut off further search if it has been determined
 * no better moves can be found in a certain path using current heuristics.
 *
 * Return the best possible move a player can make based on given search depth
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
 * alpha-beta:
 * Alpha-beta pruning works by keeping tabs on the best moves available to both players.
 * alpha: The best score player A can achieve to current knowlegde.
 *        When starting search this move should be A's worst possible score.
 *        In A's case that mostly likely means a low or even negative value.
 * beta: The best score player B can achieve to current knowledge.
 *        When starting search this move should be B's worst possible score.
 *        In B's case that mostly likely means a positive, possibly high value.
 *
 * player A is the score maximizing player, and wants alpha to become as high as possible.
 * player B is the score minimizing player and wants beta to become as low as possible.
 *
 * Alpha-beta pruning now allows us to drop search whenever we discover a
 * branch that yields moves we know will never be accepted by the other player
 * (as both players are assumed to play perfectly, and thus would never make a
 * move leading to this situation).
 *
 * That is, whenever in a search we find that the alpha value becomes more than
 * the beta value, we know we are in a branch that will never be reached
 * through normal play, and therefore searching here any further is pointless.
 *
 * alpha being more than beta means in this branch Player A has a series of moves
 * available that will lead to a better score than any of the moves in this branch,
 * since we now know that in this branch Player B will be able to make a move that
 * guarantees any options in this branch to be worse than what Player A can optimally do.
 *
 * As both players know this, we will never reach this branch in standard play,
 * and no more searching is needed here.
 *
 * Of course, all this still rests upon the heuristics used for searching. In
 * theory, if we were to search deeper, maybe an even better move would be
 * uncovered for Player A, however the knowledge gained on alpha and beta are
 * based on already having hit maximum searched depth. Therefore it's fine to
 * assume this is the best we will do with our current heuristics and search
 * can be terminated prematurely. Doing so computations and time can be spared.
 */
func AlphaBeta(game MinimaxableGameboard, maximizingPlayer bool, depth int, alpha int, beta int) (bestBoard MinimaxableGameboard, bestScore int) {
	color := 1
	if !maximizingPlayer {
		color = -1
	}

	boards := game.NextBoards(color)

	var maxBoard, minBoard MinimaxableGameboard
	var maxScore, minScore int

	/* In case the game has finish, return current game state */
	if len(boards) == 0 {
		return game, game.Score()
	}

	/* If we have reached maximum search depth, heuristically evaluate game
	 * boards.
	 */
	if depth == 0 {
		/* Handle maximizing player */
		if maximizingPlayer {
			for i, board := range boards {
				/* Compute position heurstic */
				newScore := board.Score()

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
		} else { /* Handle minimizing player */
			for i, board := range boards {
				/* Compute position heurstic */
				newScore := board.Score()

				/* Store best board seen (in our case, lowest score possible) */
				if i == 0 {
					minScore = newScore
					minBoard = board
				} else {
					if newScore < minScore {
						minScore = newScore
						minBoard = board
					}
				}
			}
			return minBoard, minScore
		}
	}

	/* If we are not at maximum search depth, iterate the various boards and
	 * score them by recursively evaluating the underlying game trees.
	 */

	/* Handle maximizing player */
	if maximizingPlayer {
		for i, board := range boards {
			/* Compute enemy score by recursing */
			_, newScore := AlphaBeta(board, !maximizingPlayer, depth-1, alpha, beta)

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
			/* Update alpha if necessary */
			if maxScore > alpha {
				alpha = maxScore
			}
			/* Terminate if known suboptimal branch found */
			if alpha >= beta {
				return maxBoard, maxScore
			}
		}
		return maxBoard, maxScore
	} else { /* Handle minimizing player */
		for i, board := range boards {
			/* Compute enemy score by recursing */
			_, newScore := AlphaBeta(board, !maximizingPlayer, depth-1, alpha, beta)

			/* Store best board seen (in our case, lowest score possible) */
			if i == 0 {
				minScore = newScore
				minBoard = board
			} else {
				if newScore < minScore {
					minScore = newScore
					minBoard = board
				}
			}
			/* Update beta if necessary */
			if minScore < beta {
				beta = minScore
			}
			/* Terminate if known suboptimal branch found */
			if alpha >= beta {
				return minBoard, minScore
			}
		}
		return minBoard, minScore
	}
}
