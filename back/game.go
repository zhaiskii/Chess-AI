package main

import "fmt"

type Color string
type PieceType string

const (
	White Color = "white"
	Black Color = "black"
)

const (
	Pawn   PieceType = "pawn"
	Rook   PieceType = "rook"
	Knight PieceType = "knight"
	Bishop PieceType = "bishop"
	Queen  PieceType = "queen"
	King   PieceType = "king"
)

type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Piece struct {
	Type  PieceType `json:"type"`
	Color Color     `json:"color"`
	Icon  string    `json:"icon"`
}

type Square struct {
	Piece   *Piece `json:"piece"`
	IsWhite bool   `json:"isWhite"`
}

type Move struct {
	From          Position  `json:"from"`
	To            Position  `json:"to"`
	Piece         *Piece    `json:"piece,omitempty"`
	CapturedPiece *Piece    `json:"capturedPiece,omitempty"`
	IsEnPassant   bool      `json:"isEnPassant,omitempty"`
	IsCastle      bool      `json:"isCastle,omitempty"`
	IsPromotion   bool      `json:"isPromotion,omitempty"`
}

// API Types
type MoveRequest struct {
	From Position `json:"from"`
	To   Position `json:"to"`
}

type GameResponse struct {
	Board       [][]Square `json:"board"`
	IsGameOver  bool       `json:"isGameOver"`
	Winner      string     `json:"winner,omitempty"`
	IsCheck     bool       `json:"isCheck"`
	CurrentTurn string     `json:"currentTurn"`
	LastMove    *Move      `json:"lastMove,omitempty"`
	AIThinking  bool       `json:"aiThinking,omitempty"`
	MoveCount   int        `json:"moveCount"`
}

type ChessGame struct {
	Board       [8][8]*Piece
	CurrentTurn Color
	GameOver    bool
	Winner      string
	MoveHistory []Move
	EnPassant   *Position // For en passant captures
	KingMoved   map[Color]bool
	RookMoved   map[Color]map[int]bool // [color][column] -> has moved
}

type ChessService struct {
	game *ChessGame
}

func NewChessService() *ChessService {
	return &ChessService{
		game: NewChessGame(),
	}
}

func NewChessGame() *ChessGame {
	game := &ChessGame{
		CurrentTurn: White,
		GameOver:    false,
		KingMoved:   make(map[Color]bool),
		RookMoved:   make(map[Color]map[int]bool),
	}
	
	game.RookMoved[White] = make(map[int]bool)
	game.RookMoved[Black] = make(map[int]bool)
	game.RookMoved[White][0] = false  // Queen side
	game.RookMoved[White][7] = false  // King side
	game.RookMoved[Black][0] = false
	game.RookMoved[Black][7] = false
	
	game.initializeBoard()
	return game
}

func (g *ChessGame) initializeBoard() {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			g.Board[i][j] = nil
		}
	}
	
	pieceOrder := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	
	for j, pieceType := range pieceOrder {
		g.Board[0][j] = &Piece{Type: pieceType, Color: Black}
	}
	for j := 0; j < 8; j++ {
		g.Board[1][j] = &Piece{Type: Pawn, Color: Black}
	}
	
	for j := 0; j < 8; j++ {
		g.Board[6][j] = &Piece{Type: Pawn, Color: White}
	}
	for j, pieceType := range pieceOrder {
		g.Board[7][j] = &Piece{Type: pieceType, Color: White}
	}
}

func (s *ChessService) GetGameState() *GameResponse {
	return &GameResponse{
		Board:       s.game.GetBoardForFrontend(),
		IsGameOver:  s.game.GameOver,
		Winner:      s.game.Winner,
		IsCheck:     s.game.IsInCheck(s.game.CurrentTurn),
		CurrentTurn: string(s.game.CurrentTurn),
		LastMove:    s.game.GetLastMove(),
		MoveCount:   len(s.game.MoveHistory),
	}
}

func (s *ChessService) MakePlayerMove(moveReq MoveRequest) (*GameResponse, error) {
	move := Move{From: moveReq.From, To: moveReq.To}
	
	if !s.game.IsValidMove(move) {
		return nil, fmt.Errorf("invalid move from %v to %v", moveReq.From, moveReq.To)
	}
	
	err := s.game.MakeMove(move)
	if err != nil {
		return nil, err
	}
	
	response := s.GetGameState()
	response.LastMove = &move
	return response, nil
}

func (s *ChessService) NewGame() *GameResponse {
	s.game = NewChessGame()
	return s.GetGameState()
}

func (s *ChessService) GetGame() *ChessGame {
	return s.game
}

func (g *ChessGame) GetBoardForFrontend() [][]Square {
	result := make([][]Square, 8)
	for i := 0; i < 8; i++ {
		result[i] = make([]Square, 8)
		for j := 0; j < 8; j++ {
			var piece *Piece
			if g.Board[i][j] != nil {
				piece = &Piece{
					Type:  g.Board[i][j].Type,
					Color: g.Board[i][j].Color,
					Icon:  getPieceIcon(g.Board[i][j].Type),
				}
			}
			result[i][j] = Square{
				Piece:   piece,
				IsWhite: (i+j)%2 == 0,
			}
		}
	}
	return result
}

func getPieceIcon(pieceType PieceType) string {
	icons := map[PieceType]string{
		Pawn:   "faChessPawn",
		Rook:   "faChessRook",
		Knight: "faChessKnight",
		Bishop: "faChessBishop",
		Queen:  "faChessQueen",
		King:   "faChessKing",
	}
	return icons[pieceType]
}

func (g *ChessGame) IsValidMove(move Move) bool {
	from, to := move.From, move.To
	
	if !inBounds(from) || !inBounds(to) {
		return false
	}
	
	piece := g.Board[from.Row][from.Col]
	if piece == nil {
		return false
	}

	if piece.Color != g.CurrentTurn {
		return false
	}
	
	if from.Row == to.Row && from.Col == to.Col {
		return false
	}
	
	targetPiece := g.Board[to.Row][to.Col]
	if targetPiece != nil && targetPiece.Color == piece.Color {
		return false
	}
	
	if !g.isValidPieceMove(from, to, piece) {
		return false
	}
	
	return !g.wouldLeaveKingInCheck(move)
}

func (g *ChessGame) isValidPieceMove(from, to Position, piece *Piece) bool {
	dx := to.Col - from.Col
	dy := to.Row - from.Row
	
	switch piece.Type {
	case Pawn:
		return g.isValidPawnMove(from, to, dx, dy, piece.Color)
	case Rook:
		return g.isValidRookMove(from, to, dx, dy)
	case Knight:
		return g.isValidKnightMove(dx, dy)
	case Bishop:
		return g.isValidBishopMove(from, to, dx, dy)
	case Queen:
		return g.isValidQueenMove(from, to, dx, dy)
	case King:
		return g.isValidKingMove(from, to, dx, dy)
	}
	return false
}

func (g *ChessGame) isValidPawnMove(from, to Position, dx, dy int, color Color) bool {
	direction := 1  
	if color == White {
		direction = -1 
	}
	
	if dx == 0 {
		if dy == direction && g.Board[to.Row][to.Col] == nil {
			return true
		}
		startingRow := 1
		if color == White {
			startingRow = 6
		}
		if from.Row == startingRow && dy == 2*direction && g.Board[to.Row][to.Col] == nil {
			return true
		}
	}
	
	if abs(dx) == 1 && dy == direction {
		if g.Board[to.Row][to.Col] != nil {
			return true
		}
		if g.EnPassant != nil && to.Row == g.EnPassant.Row && to.Col == g.EnPassant.Col {
			return true
		}
	}
	
	return false
}

func (g *ChessGame) isValidRookMove(from, to Position, dx, dy int) bool {
	if dx != 0 && dy != 0 {
		return false
	}
	return g.isPathClear(from, to)
}

func (g *ChessGame) isValidKnightMove(dx, dy int) bool {
	return (abs(dx) == 2 && abs(dy) == 1) || (abs(dx) == 1 && abs(dy) == 2)
}

func (g *ChessGame) isValidBishopMove(from, to Position, dx, dy int) bool {
	if abs(dx) != abs(dy) {
		return false
	}
	return g.isPathClear(from, to)
}

func (g *ChessGame) isValidQueenMove(from, to Position, dx, dy int) bool {
	return g.isValidRookMove(from, to, dx, dy) || g.isValidBishopMove(from, to, dx, dy)
}

func (g *ChessGame) isValidKingMove(from, to Position, dx, dy int) bool {
	if abs(dx) <= 1 && abs(dy) <= 1 {
		return true
	}
		
	return false
}

func (g *ChessGame) isPathClear(from, to Position) bool {
	dx := sign(to.Col - from.Col)
	dy := sign(to.Row - from.Row)
	
	x, y := from.Col+dx, from.Row+dy
	
	for x != to.Col || y != to.Row {
		if g.Board[y][x] != nil {
			return false
		}
		x, y = x+dx, y+dy
	}
	
	return true
}

func (g *ChessGame) MakeMove(move Move) error {
	from, to := move.From, move.To
	
	piece := g.Board[from.Row][from.Col]
	capturedPiece := g.Board[to.Row][to.Col]
	
	move.Piece = piece
	move.CapturedPiece = capturedPiece
	
	if piece.Type == Pawn && g.EnPassant != nil && 
		to.Row == g.EnPassant.Row && to.Col == g.EnPassant.Col {
		captureRow := to.Row
		if piece.Color == White {
			captureRow = to.Row + 1
		} else {
			captureRow = to.Row - 1
		}
		g.Board[captureRow][to.Col] = nil
		move.IsEnPassant = true
	}
	
	g.Board[to.Row][to.Col] = piece
	g.Board[from.Row][from.Col] = nil
	
	if piece.Type == King {
		g.KingMoved[piece.Color] = true
	}
	if piece.Type == Rook {
		g.RookMoved[piece.Color][from.Col] = true
	}
	
	g.updateEnPassant(move)
	
	g.MoveHistory = append(g.MoveHistory, move)
	
	g.CurrentTurn = opponentColor(g.CurrentTurn)
	
	g.checkGameOver()
	
	return nil
}

func (g *ChessGame) updateEnPassant(move Move) {
	g.EnPassant = nil
	
	if move.Piece.Type == Pawn && abs(move.To.Row-move.From.Row) == 2 {
		g.EnPassant = &Position{
			Row: (move.From.Row + move.To.Row) / 2,
			Col: move.From.Col,
		}
	}
}

func (g *ChessGame) IsInCheck(color Color) bool {
	kingPos := g.findKing(color)
	if kingPos == nil {
		return false
	}
	
	opponentColor := opponentColor(color)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := g.Board[i][j]
			if piece == nil || piece.Color != opponentColor {
				continue
			}
			
			if g.isValidPieceMove(Position{i, j}, *kingPos, piece) {
				return true
			}
		}
	}
	
	return false
}

func (g *ChessGame) GetValidMoves(color Color) []Move {
	var validMoves []Move
	
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := g.Board[i][j]
			if piece == nil || piece.Color != color {
				continue
			}
			
			from := Position{i, j}
			
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					to := Position{x, y}
					move := Move{From: from, To: to}
					
					if g.IsValidMove(move) {
						validMoves = append(validMoves, move)
					}
				}
			}
		}
	}
	
	return validMoves
}

func (g *ChessGame) findKing(color Color) *Position {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := g.Board[i][j]
			if piece != nil && piece.Type == King && piece.Color == color {
				return &Position{Row: i, Col: j}
			}
		}
	}
	return nil
}

func (g *ChessGame) wouldLeaveKingInCheck(move Move) bool {
	from, to := move.From, move.To
	originalPiece := g.Board[to.Row][to.Col]
	movingPiece := g.Board[from.Row][from.Col]
	
	g.Board[to.Row][to.Col] = movingPiece
	g.Board[from.Row][from.Col] = nil
	
	inCheck := g.IsInCheck(g.CurrentTurn)
	
	g.Board[from.Row][from.Col] = movingPiece
	g.Board[to.Row][to.Col] = originalPiece
	
	return inCheck
}

func (g *ChessGame) checkGameOver() {
	validMoves := g.GetValidMoves(g.CurrentTurn)
	
	if len(validMoves) == 0 {
		g.GameOver = true
		if g.IsInCheck(g.CurrentTurn) {
			g.Winner = string(opponentColor(g.CurrentTurn))
		} else {
			g.Winner = "draw"
		}
	}
}

func (g *ChessGame) GetLastMove() *Move {
	if len(g.MoveHistory) == 0 {
		return nil
	}
	return &g.MoveHistory[len(g.MoveHistory)-1]
}

func (g *ChessGame) CopyState() *ChessGame {
	newGame := &ChessGame{
		CurrentTurn: g.CurrentTurn,
		GameOver:    g.GameOver,
		Winner:      g.Winner,
		MoveHistory: make([]Move, len(g.MoveHistory)),
		EnPassant:   g.EnPassant,
		KingMoved:   make(map[Color]bool),
		RookMoved:   make(map[Color]map[int]bool),
	}
	
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if g.Board[i][j] != nil {
				newGame.Board[i][j] = &Piece{
					Type:  g.Board[i][j].Type,
					Color: g.Board[i][j].Color,
				}
			}
		}
	}
	
	copy(newGame.MoveHistory, g.MoveHistory)
	
	newGame.KingMoved[White] = g.KingMoved[White]
	newGame.KingMoved[Black] = g.KingMoved[Black]
	
	newGame.RookMoved[White] = make(map[int]bool)
	newGame.RookMoved[Black] = make(map[int]bool)
	for col, moved := range g.RookMoved[White] {
		newGame.RookMoved[White][col] = moved
	}
	for col, moved := range g.RookMoved[Black] {
		newGame.RookMoved[Black][col] = moved
	}
	
	return newGame
}

func opponentColor(color Color) Color {
	if color == White {
		return Black
	}
	return White
}

func inBounds(pos Position) bool {
	return pos.Row >= 0 && pos.Row < 8 && pos.Col >= 0 && pos.Col < 8
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}