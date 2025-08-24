package main

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// AI CONSTANTS & CONFIGURATION
// ============================================================================

const (
	INFINITY          = 999999
	WIN_SCORE         = 100000
	DEFAULT_DEPTH     = 4
	MAX_THINKING_TIME = 30 * time.Second
)

// Piece values for material evaluation
var pieceValues = map[PieceType]int{
	Pawn:   100,
	Knight: 320,
	Bishop: 330,
	Rook:   500,
	Queen:  900,
	King:   20000,
}

// ============================================================================
// PIECE-SQUARE TABLES FOR POSITIONAL EVALUATION
// ============================================================================

// Pawn position values (Black perspective, White will be flipped)
var pawnTable = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

var knightTable = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}

var bishopTable = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var rookTable = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var queenTable = [8][8]int{
	{-20, -10, -10, -5, -5, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 5, 5, 5, 0, -10},
	{-5, 0, 5, 5, 5, 5, 0, -5},
	{0, 0, 5, 5, 5, 5, 0, -5},
	{-10, 5, 5, 5, 5, 5, 0, -10},
	{-10, 0, 5, 0, 0, 0, 0, -10},
	{-20, -10, -10, -5, -5, -10, -10, -20},
}

var kingTable = [8][8]int{
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-20, -30, -30, -40, -40, -30, -30, -20},
	{-10, -20, -20, -20, -20, -20, -20, -10},
	{20, 20, 0, 0, 0, 0, 20, 20},
	{20, 30, 10, 0, 0, 10, 30, 20},
}

// ============================================================================
// AI SERVICE
// ============================================================================

type AIService struct {
	depth           int
	nodesSearched   int64
	lastThinkingTime time.Duration
}

func NewAIService() *AIService {
	return &AIService{
		depth: DEFAULT_DEPTH,
	}
}

// ============================================================================
// MAIN AI INTERFACE METHODS
// ============================================================================

// GetBestMove finds the best move using minimax with alpha-beta pruning
func (ai *AIService) GetBestMove(ctx context.Context, game *ChessGame) (*Move, error) {
	if game.GameOver {
		return nil, fmt.Errorf("game is over")
	}

	moves := game.GetValidMoves(game.CurrentTurn)
	if len(moves) == 0 {
		return nil, fmt.Errorf("no valid moves available")
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(ctx, MAX_THINKING_TIME)
	defer cancel()

	// Channel to receive the result
	resultChan := make(chan struct {
		move *Move
		err  error
	}, 1)

	// Run AI calculation in goroutine
	go func() {
		start := time.Now()
		ai.nodesSearched = 0
		
		bestMove := ai.getBestMoveSync(game)
		ai.lastThinkingTime = time.Since(start)
		
		resultChan <- struct {
			move *Move
			err  error
		}{bestMove, nil}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result.move, result.err
	case <-ctx.Done():
		// Return first valid move if timeout
		return &moves[0], fmt.Errorf("AI thinking timeout, returning first valid move")
	}
}

func (ai *AIService) getBestMoveSync(game *ChessGame) *Move {
	moves := game.GetValidMoves(game.CurrentTurn)
	if len(moves) == 0 {
		return nil
	}

	bestMove := moves[0]
	bestValue := -INFINITY

	// Try each possible move
	for _, move := range moves {
		// Make a copy of the game to test the move
		gameCopy := game.CopyState()
		gameCopy.MakeMove(move)

		// Evaluate this position using minimax
		value := ai.minimax(gameCopy, ai.depth-1, -INFINITY, INFINITY, false)

		if value > bestValue {
			bestValue = value
			bestMove = move
		}
	}

	return &bestMove
}

// ============================================================================
// MINIMAX ALGORITHM WITH ALPHA-BETA PRUNING
// ============================================================================

func (ai *AIService) minimax(game *ChessGame, depth int, alpha, beta int, isMaximizing bool) int {
	ai.nodesSearched++

	// Terminal cases
	if depth == 0 {
		return ai.evaluatePosition(game)
	}

	if game.GameOver {
		if game.Winner == string(Black) {
			return WIN_SCORE + depth // Prefer faster wins
		} else if game.Winner == string(White) {
			return -WIN_SCORE - depth
		} else {
			return 0 // Draw
		}
	}

	if isMaximizing {
		// Black is maximizing (AI player)
		maxEval := -INFINITY
		moves := game.GetValidMoves(Black)

		for _, move := range moves {
			gameCopy := game.CopyState()
			gameCopy.MakeMove(move)

			eval := ai.minimax(gameCopy, depth-1, alpha, beta, false)
			maxEval = max(maxEval, eval)
			alpha = max(alpha, eval)

			// Alpha-beta pruning
			if beta <= alpha {
				break
			}
		}
		return maxEval

	} else {
		// White is minimizing (human player)
		minEval := INFINITY
		moves := game.GetValidMoves(White)

		for _, move := range moves {
			gameCopy := game.CopyState()
			gameCopy.MakeMove(move)

			eval := ai.minimax(gameCopy, depth-1, alpha, beta, true)
			minEval = min(minEval, eval)
			beta = min(beta, eval)

			// Alpha-beta pruning
			if beta <= alpha {
				break
			}
		}
		return minEval
	}
}

// ============================================================================
// POSITION EVALUATION
// ============================================================================

func (ai *AIService) evaluatePosition(game *ChessGame) int {
	if game.GameOver {
		if game.Winner == string(Black) {
			return WIN_SCORE
		} else if game.Winner == string(White) {
			return -WIN_SCORE
		} else {
			return 0 // Draw
		}
	}

	score := 0

	// Material and positional evaluation
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := game.Board[i][j]
			if piece == nil {
				continue
			}

			pieceScore := ai.evaluatePiece(piece, i, j)

			if piece.Color == Black {
				score += pieceScore
			} else {
				score -= pieceScore
			}
		}
	}

	// Additional positional factors
	score += ai.evaluatePositionalFactors(game)

	return score
}

func (ai *AIService) evaluatePiece(piece *Piece, row, col int) int {
	baseValue := pieceValues[piece.Type]

	// Get positional bonus using piece-square tables
	// Flip the row for white pieces (they start at bottom)
	evalRow := row
	if piece.Color == White {
		evalRow = 7 - row
	}

	var positionValue int
	switch piece.Type {
	case Pawn:
		positionValue = pawnTable[evalRow][col]
	case Knight:
		positionValue = knightTable[evalRow][col]
	case Bishop:
		positionValue = bishopTable[evalRow][col]
	case Rook:
		positionValue = rookTable[evalRow][col]
	case Queen:
		positionValue = queenTable[evalRow][col]
	case King:
		positionValue = kingTable[evalRow][col]
	}

	return baseValue + positionValue
}

func (ai *AIService) evaluatePositionalFactors(game *ChessGame) int {
	score := 0

	// Center control bonus
	centerSquares := []Position{{3, 3}, {3, 4}, {4, 3}, {4, 4}}
	extendedCenter := []Position{{2, 2}, {2, 3}, {2, 4}, {2, 5}, 
		{3, 2}, {3, 5}, {4, 2}, {4, 5}, {5, 2}, {5, 3}, {5, 4}, {5, 5}}

	for _, pos := range centerSquares {
		if ai.isSquareControlledBy(game, pos, Black) {
			score += 15
		}
		if ai.isSquareControlledBy(game, pos, White) {
			score -= 15
		}
	}

	for _, pos := range extendedCenter {
		if ai.isSquareControlledBy(game, pos, Black) {
			score += 5
		}
		if ai.isSquareControlledBy(game, pos, White) {
			score -= 5
		}
	}

	// King safety evaluation
	blackKing := game.findKing(Black)
	whiteKing := game.findKing(White)

	if blackKing != nil {
		score += ai.evaluateKingSafety(*blackKing, Black, game)
	}
	if whiteKing != nil {
		score -= ai.evaluateKingSafety(*whiteKing, White, game)
	}

	// Mobility (number of legal moves)
	blackMoves := len(game.GetValidMoves(Black))
	whiteMoves := len(game.GetValidMoves(White))
	score += (blackMoves - whiteMoves) * 2

	return score
}

func (ai *AIService) isSquareControlledBy(game *ChessGame, pos Position, color Color) bool {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := game.Board[i][j]
			if piece == nil || piece.Color != color {
				continue
			}

			if game.isValidPieceMove(Position{i, j}, pos, piece) {
				return true
			}
		}
	}
	return false
}

func (ai *AIService) evaluateKingSafety(kingPos Position, color Color, game *ChessGame) int {
	safety := 0

	// Penalty for king in center during middlegame
	if kingPos.Row >= 2 && kingPos.Row <= 5 && kingPos.Col >= 2 && kingPos.Col <= 5 {
		safety -= 30
	}

	// Bonus for king on back rank (not moved)
	expectedBackRank := 7
	if color == Black {
		expectedBackRank = 0
	}
	if kingPos.Row == expectedBackRank {
		safety += 20
	}

	// Count pawns in front of king for shelter
	pawnShield := 0
	direction := -1
	if color == Black {
		direction = 1
	}

	for colOffset := -1; colOffset <= 1; colOffset++ {
		checkCol := kingPos.Col + colOffset
		checkRow := kingPos.Row + direction

		if inBounds(Position{checkRow, checkCol}) {
			piece := game.Board[checkRow][checkCol]
			if piece != nil && piece.Type == Pawn && piece.Color == color {
				pawnShield++
			}
		}
	}

	safety += pawnShield * 10

	return safety
}

// ============================================================================
// AI SERVICE METHODS
// ============================================================================

func (ai *AIService) MakeAIMove(ctx context.Context, chessService *ChessService) (*GameResponse, error) {
	if chessService.game.GameOver {
		return nil, fmt.Errorf("game is over")
	}

	move, err := ai.GetBestMove(ctx, chessService.game)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI move: %w", err)
	}

	if move == nil {
		return nil, fmt.Errorf("no valid AI moves available")
	}

	// Make the move
	err = chessService.game.MakeMove(*move)
	if err != nil {
		return nil, fmt.Errorf("failed to execute AI move: %w", err)
	}

	// Return updated game state
	response := chessService.GetGameState()
	response.LastMove = move
	return response, nil
}

func (ai *AIService) GetStats() map[string]interface{} {
	difficulty := ai.getDifficultyString()
	
	return map[string]interface{}{
		"engine":           "Minimax with Alpha-Beta Pruning",
		"depth":            ai.depth,
		"difficulty":       difficulty,
		"timeout":          MAX_THINKING_TIME.String(),
		"nodes_searched":   ai.nodesSearched,
		"last_think_time":  ai.lastThinkingTime.String(),
	}
}

func (ai *AIService) getDifficultyString() string {
	switch ai.depth {
	case 1, 2:
		return "Easy"
	case 3, 4:
		return "Medium"
	case 5, 6:
		return "Hard"
	case 7, 8:
		return "Expert"
	default:
		return "Custom"
	}
}

func (ai *AIService) SetDifficulty(level string) error {
	switch level {
	case "easy", "Easy":
		ai.depth = 2
	case "medium", "Medium":
		ai.depth = 4
	case "hard", "Hard":
		ai.depth = 6
	case "expert", "Expert":
		ai.depth = 8
	default:
		return fmt.Errorf("invalid difficulty level: %s (use easy/medium/hard/expert)", level)
	}
	return nil
}

func (ai *AIService) SetDepth(depth int) error {
	if depth < 1 || depth > 10 {
		return fmt.Errorf("depth must be between 1 and 10, got %d", depth)
	}
	ai.depth = depth
	return nil
}

func (ai *AIService) GetDepth() int {
	return ai.depth
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}