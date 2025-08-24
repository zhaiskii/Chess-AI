package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"fmt"
)

// ============================================================================
// HANDLERS STRUCT & CONSTRUCTOR
// ============================================================================

type Handlers struct {
	chessService *ChessService
	aiService    *AIService
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func NewHandlers(chessService *ChessService, aiService *AIService) *Handlers {
	return &Handlers{
		chessService: chessService,
		aiService:    aiService,
	}
}

// ============================================================================
// HEALTH & STATUS ENDPOINTS
// ============================================================================

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":      "healthy",
		"service":     "chess-ai-backend",
		"version":     "3.0.0",
		"timestamp":   time.Now(),
		"uptime":      "running", // Could calculate actual uptime
		"endpoints": map[string]string{
			"game":       "GET /api/game",
			"move":       "POST /api/move", 
			"new_game":   "POST /api/new-game",
			"ai_move":    "POST /api/ai/move",
			"ai_stats":   "GET /api/ai/stats",
		},
	}
	
	h.writeJSON(w, response)
}

// ============================================================================
// GAME STATE ENDPOINTS
// ============================================================================

func (h *Handlers) GetGameState(w http.ResponseWriter, r *http.Request) {
	response := h.chessService.GetGameState()
	h.writeJSON(w, response)
}

func (h *Handlers) NewGame(w http.ResponseWriter, r *http.Request) {
	log.Println("🎮 Starting new game")
	response := h.chessService.NewGame()
	h.writeJSON(w, response)
}

func (h *Handlers) GetValidMoves(w http.ResponseWriter, r *http.Request) {
	game := h.chessService.GetGame()
	moves := game.GetValidMoves(game.CurrentTurn)
	
	response := map[string]interface{}{
		"valid_moves":  moves,
		"count":        len(moves),
		"current_turn": string(game.CurrentTurn),
		"is_check":     game.IsInCheck(game.CurrentTurn),
	}
	
	h.writeJSON(w, response)
}

func (h *Handlers) GetGameHistory(w http.ResponseWriter, r *http.Request) {
	game := h.chessService.GetGame()
	
	response := map[string]interface{}{
		"moves":         game.MoveHistory,
		"move_count":    len(game.MoveHistory),
		"current_turn":  string(game.CurrentTurn),
		"game_over":     game.GameOver,
		"winner":        game.Winner,
		"last_move":     game.GetLastMove(),
	}
	
	h.writeJSON(w, response)
}

// ============================================================================
// MOVE ENDPOINTS
// ============================================================================

func (h *Handlers) MakeMove(w http.ResponseWriter, r *http.Request) {
	var moveReq MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&moveReq); err != nil {
		h.writeError(w, "Invalid JSON format", http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("🎯 Player move: %+v", moveReq)

	// Validate the move request
	if !inBounds(moveReq.From) || !inBounds(moveReq.To) {
		h.writeError(w, "Move coordinates out of bounds", http.StatusBadRequest, "")
		return
	}

	// Make player move
	response, err := h.chessService.MakePlayerMove(moveReq)
	if err != nil {
		h.writeError(w, "Invalid move", http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("✅ Player move successful")

	// If game is over, return immediately
	if response.IsGameOver {
		log.Printf("🏁 Game over: %s", response.Winner)
		h.writeJSON(w, response)
		return
	}

	// Make AI move if it's AI's turn (Black)
	if h.chessService.game.CurrentTurn == Black {
		log.Println("🤖 AI thinking...")
		
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		
		aiResponse, err := h.aiService.MakeAIMove(ctx, h.chessService)
		if err != nil {
			log.Printf("⚠️ AI move failed: %v", err)
			// Return current state even if AI fails
			response.AIThinking = false
			h.writeJSON(w, response)
			return
		}
		
		log.Printf("🤖 AI move completed")
		response = aiResponse
		response.AIThinking = false
	}

	h.writeJSON(w, response)
}

func (h *Handlers) ForceAIMove(w http.ResponseWriter, r *http.Request) {
	if h.chessService.game.GameOver {
		h.writeError(w, "Cannot make AI move: game is over", http.StatusBadRequest, "")
		return
	}

	if h.chessService.game.CurrentTurn != Black {
		h.writeError(w, "Not AI's turn", http.StatusBadRequest, "Current turn: "+string(h.chessService.game.CurrentTurn))
		return
	}

	log.Println("🤖 Forced AI move requested")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	response, err := h.aiService.MakeAIMove(ctx, h.chessService)
	if err != nil {
		h.writeError(w, "AI move failed", http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("🤖 Forced AI move completed")
	h.writeJSON(w, response)
}

// ============================================================================
// AI CONFIGURATION ENDPOINTS
// ============================================================================

func (h *Handlers) GetAIStats(w http.ResponseWriter, r *http.Request) {
	stats := h.aiService.GetStats()
	
	// Add game-specific stats
	game := h.chessService.GetGame()
	stats["game_stats"] = map[string]interface{}{
		"moves_played":   len(game.MoveHistory),
		"current_turn":   string(game.CurrentTurn),
		"game_over":      game.GameOver,
		"is_check":       game.IsInCheck(game.CurrentTurn),
		"valid_moves":    len(game.GetValidMoves(game.CurrentTurn)),
	}
	
	h.writeJSON(w, stats)
}

func (h *Handlers) SetDifficulty(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Difficulty string `json:"difficulty"`
		Depth      *int   `json:"depth,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON format", http.StatusBadRequest, err.Error())
		return
	}

	// Set difficulty by name or custom depth
	if req.Depth != nil {
		if err := h.aiService.SetDepth(*req.Depth); err != nil {
			h.writeError(w, "Invalid depth", http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("🎯 AI depth set to %d", *req.Depth)
	} else if req.Difficulty != "" {
		if err := h.aiService.SetDifficulty(req.Difficulty); err != nil {
			h.writeError(w, "Invalid difficulty", http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("🎯 AI difficulty set to %s", req.Difficulty)
	} else {
		h.writeError(w, "Must provide either 'difficulty' or 'depth'", http.StatusBadRequest, "")
		return
	}

	response := map[string]interface{}{
		"message":     "AI configuration updated successfully",
		"difficulty":  h.aiService.getDifficultyString(),
		"depth":       h.aiService.GetDepth(),
		"stats":       h.aiService.GetStats(),
	}
	
	h.writeJSON(w, response)
}

// ============================================================================
// ANALYSIS ENDPOINTS
// ============================================================================

func (h *Handlers) EvaluatePosition(w http.ResponseWriter, r *http.Request) {
	game := h.chessService.GetGame()
	evaluation := h.aiService.evaluatePosition(game)
	
	response := map[string]interface{}{
		"evaluation":    evaluation,
		"current_turn":  string(game.CurrentTurn),
		"description":   getEvaluationDescription(evaluation),
		"material_only": h.getMaterialBalance(game),
		"game_phase":    h.getGamePhase(game),
	}
	
	h.writeJSON(w, response)
}

func (h *Handlers) GetBestMoves(w http.ResponseWriter, r *http.Request) {
	// Get depth from query parameter (default to AI's current depth)
	depthStr := r.URL.Query().Get("depth")
	depth := h.aiService.GetDepth()
	
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d >= 1 && d <= 6 {
			depth = d
		}
	}
	
	game := h.chessService.GetGame()
	
	// Get top 3 moves (simplified analysis)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	originalDepth := h.aiService.GetDepth()
	h.aiService.SetDepth(depth)
	
	bestMove, err := h.aiService.GetBestMove(ctx, game)
	
	h.aiService.SetDepth(originalDepth) // Restore original depth
	
	if err != nil {
		h.writeError(w, "Failed to analyze position", http.StatusInternalServerError, err.Error())
		return
	}
	
	response := map[string]interface{}{
		"best_move":     bestMove,
		"analysis_depth": depth,
		"evaluation":    h.aiService.evaluatePosition(game),
		"current_turn":  string(game.CurrentTurn),
	}
	
	h.writeJSON(w, response)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

func (h *Handlers) getMaterialBalance(game *ChessGame) map[string]int {
	whiteMaterial := 0
	blackMaterial := 0
	
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := game.Board[i][j]
			if piece == nil {
				continue
			}
			
			value := pieceValues[piece.Type]
			if piece.Color == White {
				whiteMaterial += value
			} else {
				blackMaterial += value
			}
		}
	}
	
	return map[string]int{
		"white": whiteMaterial,
		"black": blackMaterial,
		"difference": blackMaterial - whiteMaterial,
	}
}

func (h *Handlers) getGamePhase(game *ChessGame) string {
	moveCount := len(game.MoveHistory)
	
	if moveCount < 20 {
		return "opening"
	} else if moveCount < 60 {
		return "middlegame"
	} else {
		return "endgame"
	}
}

func getEvaluationDescription(eval int) string {
	absEval := eval
	if absEval < 0 {
		absEval = -absEval
	}
	
	var advantage string
	var magnitude string
	
	if eval > 0 {
		advantage = "Black"
	} else if eval < 0 {
		advantage = "White"
	} else {
		return "Position is equal"
	}
	
	switch {
	case absEval > 1000:
		magnitude = "winning"
	case absEval > 500:
		magnitude = "has a decisive advantage"
	case absEval > 200:
		magnitude = "has a significant advantage"
	case absEval > 100:
		magnitude = "has a moderate advantage"
	case absEval > 50:
		magnitude = "has a slight advantage"
	default:
		magnitude = "is slightly better"
	}
	
	return fmt.Sprintf("%s %s", advantage, magnitude)
}

// ============================================================================
// UTILITY METHODS
// ============================================================================

func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("❌ Error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handlers) writeError(w http.ResponseWriter, message string, statusCode int, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Details: details,
	}
	
	log.Printf("⚠️ HTTP Error %d: %s", statusCode, message)
	if details != "" {
		log.Printf("   Details: %s", details)
	}
	
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("❌ Error encoding error response: %v", err)
	}
}

// ============================================================================
// DEBUG & DEVELOPMENT ENDPOINTS
// ============================================================================

func (h *Handlers) DebugBoard(w http.ResponseWriter, r *http.Request) {
	game := h.chessService.GetGame()
	
	// Create ASCII representation of the board
	boardStr := ""
	for i := 0; i < 8; i++ {
		boardStr += fmt.Sprintf("%d ", 8-i)
		for j := 0; j < 8; j++ {
			piece := game.Board[i][j]
			if piece == nil {
				boardStr += ". "
			} else {
				symbol := h.getPieceSymbol(piece)
				boardStr += symbol + " "
			}
		}
		boardStr += "\n"
	}
	boardStr += "  a b c d e f g h\n"
	
	response := map[string]interface{}{
		"board_ascii":   boardStr,
		"current_turn":  string(game.CurrentTurn),
		"move_count":    len(game.MoveHistory),
		"game_over":     game.GameOver,
		"in_check":      game.IsInCheck(game.CurrentTurn),
		"valid_moves":   len(game.GetValidMoves(game.CurrentTurn)),
	}
	
	h.writeJSON(w, response)
}

func (h *Handlers) getPieceSymbol(piece *Piece) string {
	symbols := map[PieceType]map[Color]string{
		King:   {White: "K", Black: "k"},
		Queen:  {White: "Q", Black: "q"},
		Rook:   {White: "R", Black: "r"},
		Bishop: {White: "B", Black: "b"},
		Knight: {White: "N", Black: "n"},
		Pawn:   {White: "P", Black: "p"},
	}
	
	if typeSymbols, exists := symbols[piece.Type]; exists {
		if symbol, exists := typeSymbols[piece.Color]; exists {
			return symbol
		}
	}
	
	return "?"
}