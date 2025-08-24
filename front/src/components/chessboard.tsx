import React, { useState } from "react";
import { type SquareProps, Square } from "./square";
import { faChessBishop } from "@fortawesome/free-solid-svg-icons";
import { faChessRook } from "@fortawesome/free-solid-svg-icons";
import { faChessKnight } from "@fortawesome/free-solid-svg-icons";
import { faChessKing } from "@fortawesome/free-solid-svg-icons";
import { faChessQueen } from "@fortawesome/free-solid-svg-icons";
import { faChessPawn } from "@fortawesome/free-solid-svg-icons";
import type { IconProp } from "@fortawesome/fontawesome-svg-core";

export type PieceColor = 'white' | 'black';
export type PieceType = 'pawn' | 'rook' | 'knight' | 'bishop' | 'queen' | 'king';

export interface Piece {
  type: PieceType;
  color: PieceColor;
  icon: IconProp;
}

interface ChessBoardProps {
  squares: SquareProps[][];
}

function getDefaultBoard(): SquareProps[][] {
  let defaultBoard: SquareProps[][] = [];

  const backRankPieces = [
    faChessRook, faChessKnight, faChessBishop, faChessQueen, 
    faChessKing, faChessBishop, faChessKnight, faChessRook
  ];

  const backRankTypes: PieceType[] = [
    'rook', 'knight', 'bishop', 'queen',
    'king', 'bishop', 'knight', 'rook'
  ];

  for (let i = 0; i < 8; i++) {
    const row: SquareProps[] = [];
    for (let j = 0; j < 8; j++) {
      let piece: Piece | null = null;

      if (i === 0) {
        piece = {
          type: backRankTypes[j],
          color: 'black',
          icon: backRankPieces[j]
        };
      } else if (i === 1) {
        piece = {
          type: 'pawn',
          color: 'black',
          icon: faChessPawn
        };
      }

      else if (i === 6) {
        piece = {
          type: 'pawn',
          color: 'white',
          icon: faChessPawn
        };
      } else if (i === 7) {
        piece = {
          type: backRankTypes[j],
          color: 'white',
          icon: backRankPieces[j]
        };
      }

      let square: SquareProps = {
        piece: piece,
        isWhite: Boolean(!((i + j) % 2)),
        isSelected: false,
      };
      row.push(square);
    }
    defaultBoard.push(row);
  }

  return defaultBoard;
}

export function ChessBoard() {
  const [board, setBoard] = useState<ChessBoardProps>({
    squares: getDefaultBoard(),
  });
  const [selectedSquare, setSelectedSquare] = useState<{
    row: number;
    col: number;
  } | null>(null);

  const handleMove = async (fromRow: number, fromCol: number, toRow: number, toCol: number) => {
    try {
      const response = await fetch('/api/move', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          from: { row: fromRow, col: fromCol },
          to: { row: toRow, col: toCol }
        })
      });
      
      if (response.ok) {
        const newBoardState = await response.json();
        setBoard({ squares: newBoardState.board });
      }
    } catch (error) {
      console.error('error making move:', error);
    }
  };

  return (
    <div className="flex justify-center items-center w-screen h-screen bg-grey">
      <div className="grid grid-cols-8 w-120 h-120">
        {board.squares.map((row, rowI) =>
          row.map((square, colI) => {
            return (
              <Square
                key={`${rowI}-${colI}`}
                {...square}
                row={rowI}
                col={colI}
                isSelected={
                  selectedSquare !== null &&
                  selectedSquare.row === rowI &&
                  selectedSquare.col === colI
                }
                onSquareClick={(row, col) => {
                  if (selectedSquare) {
                    if (selectedSquare.row === row && selectedSquare.col === col) {
                      setSelectedSquare(null);
                    } else {
                      handleMove(selectedSquare.row, selectedSquare.col, row, col);
                      setSelectedSquare(null);
                    }
                  } else if (square.piece) {
                    setSelectedSquare({ row, col });
                  }
                }}
              />
            );
          })
        )}
      </div>
    </div>
  );
}