import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import type { Piece } from "./chessboard";

export interface SquareProps {
  row?: number;
  col?: number;
  piece: Piece | null;
  isWhite: boolean;
  isSelected: boolean;
  onSquareClick?: (row: number, col: number) => void;
}

export function Square({
  row,
  col,
  piece,
  isWhite,
  isSelected,
  onSquareClick,
}: SquareProps) {
  const handleClick = () => {
    if (onSquareClick && row !== undefined && col !== undefined) {
      onSquareClick(row, col);
    }
  };

  return (
    <div
      className={`w-full aspect-square ${
        isWhite ? "bg-[#f0d9b5]" : "bg-[#b58863]"
      } flex items-center justify-center cursor-pointer hover:opacity-80 transition-opacity`}
      onClick={handleClick}
      style={{
        color: piece?.color === 'white' ? "#ffffff" : "#000000",
        border: isSelected ? "3px solid #FFD700" : "none",
        fontSize: "2rem",
      }}
    >
      {piece && <FontAwesomeIcon icon={piece.icon} />}
    </div>
  );
}