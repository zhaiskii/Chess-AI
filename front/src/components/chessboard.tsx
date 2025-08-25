import React, { useState } from "react";
import { type SquareProps, Square } from "./square";
import { faChessBishop } from "@fortawesome/free-solid-svg-icons";
import { faChessRook } from "@fortawesome/free-solid-svg-icons";
import { faChessKnight } from "@fortawesome/free-solid-svg-icons";
import { faChessKing } from "@fortawesome/free-solid-svg-icons";
import { faChessQueen } from "@fortawesome/free-solid-svg-icons";
import { faChessPawn } from "@fortawesome/free-solid-svg-icons";
import type { IconProp } from "@fortawesome/fontawesome-svg-core";

export type PieceColor = "white" | "black";
export type PieceType =
  | "pawn"
  | "rook"
  | "knight"
  | "bishop"
  | "queen"
  | "king";

export interface Piece {
  type: PieceType;
  color: PieceColor;
  icon: IconProp;
}

interface ChessBoardProps {
  squares: SquareProps[][];
}

interface GameState {
  board: Array<
    Array<{
      piece: Piece | null;
      isWhite: boolean;
    }>
  >;
  isGameOver: boolean;
  winner?: string;
  isCheck: boolean;
  currentTurn: string;
  lastMove?: any;
  moveCount: number;
}

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

function convertBackendBoard(backendBoard: any[][]): SquareProps[][] {
  return backendBoard.map((row, rowIndex) =>
    row.map((square, colIndex) => ({
      piece: square.piece
        ? {
            type: square.piece.type as PieceType,
            color: square.piece.color as PieceColor,
            icon: getIconForPiece(square.piece.type),
          }
        : null,
      isWhite: square.isWhite,
      isSelected: false,
      row: rowIndex,
      col: colIndex,
    }))
  );
}

function getIconForPiece(pieceType: string): IconProp {
  const iconMap: Record<string, IconProp> = {
    pawn: faChessPawn,
    rook: faChessRook,
    knight: faChessKnight,
    bishop: faChessBishop,
    queen: faChessQueen,
    king: faChessKing,
  };
  return iconMap[pieceType] || faChessPawn;
}

export function ChessBoard() {
  const [menuOpen, setMenuOpen] = useState<boolean>(false);
  const [searchDepth, setSearchDepth] = useState<number>(-1);
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [selectedSquare, setSelectedSquare] = useState<{
    row: number;
    col: number;
  } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  console.log('edit2');

  React.useEffect(() => {
    if (searchDepth === -1) return;
    setLoading(true);
    (async () => {
      try {
        await fetch(`${API_URL}/api/change-depth`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ depth: searchDepth }),
        });
      } finally {
        setLoading(false);
      }
    })();
  }, [searchDepth]);

  // Initialize game
  React.useEffect(() => {
    setSearchDepth(3);
    fetchGameState();
  }, []);

  const fetchGameState = async () => {
    try {
      const response = await fetch(`${API_URL}/api/game`);
      if (!response.ok) throw new Error("Failed to fetch game state");
      const data = await response.json();
      setGameState(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  };

  const makeMove = async (
    from: { row: number; col: number },
    to: { row: number; col: number }
  ) => {
    setLoading(true);
    try {
      const response = await fetch(`${API_URL}/api/move`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ from, to }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Invalid move");
      }

      const data = await response.json();
      setGameState(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Move failed");
    } finally {
      setLoading(false);
    }
  };

  const newGame = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_URL}/api/new-game`, {
        method: "POST",
      });
      if (!response.ok) throw new Error("Failed to start new game");
      const data = await response.json();
      setGameState(data);
      setSelectedSquare(null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start new game");
    } finally {
      setLoading(false);
    }
  };

  const handleSquareClick = async (row: number, col: number) => {
    if (loading || !gameState) return;

    if (selectedSquare) {
      if (selectedSquare.row === row && selectedSquare.col === col) {
        // Deselect if clicking same square
        setSelectedSquare(null);
      } else {
        // Attempt move
        await makeMove(selectedSquare, { row, col });
        setSelectedSquare(null);
      }
    } else {
      // Select square if it has a piece and it's player's turn
      const square = gameState.board[row][col];
      if (
        square.piece &&
        gameState.currentTurn === "white" &&
        square.piece.color === "white"
      ) {
        setSelectedSquare({ row, col });
      }
    }
  };

  if (!gameState) {
    return (
      <div className="flex justify-center items-center w-screen h-screen bg-gray-100">
        <div className="text-center">
          <div className="text-xl mb-4">Loading Chess AI...</div>
          {error && <div className="text-red-600 mb-4">{error}</div>}
          <button
            onClick={fetchGameState}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const squares = convertBackendBoard(gameState.board);

  return (
    <div className="flex flex-col justify-center items-center w-screen h-screen bg-gray-100 p-4">
      {/* Game Status */}
      <div className="mb-4 text-center">
        <div className="text-2xl font-bold mb-2">Chess AI</div>
        <div className="text-lg mb-2">
          Turn:{" "}
          <span className="font-semibold">
            {gameState.currentTurn === "white" ? "Your Turn" : "AI Thinking..."}
          </span>
        </div>
        {gameState.isCheck && (
          <div className="text-red-600 font-bold">Check!</div>
        )}
        {gameState.isGameOver && (
          <div className="text-xl font-bold">
            Game Over!{" "}
            {gameState.winner === "white"
              ? "You Win!"
              : gameState.winner === "black"
              ? "AI Wins!"
              : "Draw!"}
          </div>
        )}
        {error && <div className="text-red-600 mt-2">{error}</div>}
      </div>

      {/* Chess Board */}
      <div className="grid grid-cols-8 w-96 h-96 border-4 border-gray-800 mb-4">
        {squares.map((row, rowI) =>
          row.map((square, colI) => {
            const isSelected =
              selectedSquare !== null &&
              selectedSquare.row === rowI &&
              selectedSquare.col === colI;

            return (
              <Square
                key={`${rowI}-${colI}`}
                {...square}
                isSelected={isSelected}
                onSquareClick={() => handleSquareClick(rowI, colI)}
                disabled={loading}
              />
            );
          })
        )}
      </div>

      {/* Controls */}
      <div className="flex space-x-4">
        <button
          onClick={newGame}
          disabled={loading}
          className="px-6 py-3 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {loading ? "Loading..." : "New Game"}
        </button>
        <button
          onClick={fetchGameState}
          disabled={loading}
          className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          Refresh
        </button>
        <div className="relative inline-block px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed">
          <button
            onClick={() => { setMenuOpen(!menuOpen); }}
            disabled={loading}
            className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Minimax algo's search depth {searchDepth}
          </button>

          {menuOpen && (
            <div className="absolute top-full right-0 mt-2 w-40 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5">
              <button onClick={() => setSearchDepth(1)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">1</button>
              <button onClick={() => setSearchDepth(2)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">2</button>
              <button onClick={() => setSearchDepth(3)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">3</button>
              <button onClick={() => setSearchDepth(4)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">4</button>
              <button onClick={() => setSearchDepth(5)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">5</button>
              <button onClick={() => setSearchDepth(6)} className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">6</button>
            </div>
          )}
        </div>
      </div>

      {/* Game Info */}
      <div className="mt-4 text-sm text-gray-600 text-center">
        <div>Move Count: {gameState.moveCount}</div>
        <div>You are White (bottom), AI is Black (top)</div>
      </div>
    </div>
  );
}