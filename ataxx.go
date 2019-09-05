/* The game of Ataxx implemented with minimax */
package main

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
