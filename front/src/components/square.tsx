import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import type { Piece } from "./ChessBoard";

export interface SquareProps {
  row?: number;
  col?: number;
  piece: Piece | null;
  isWhite: boolean;
  isSelected: boolean;
  onSquareClick?: () => void;
  disabled?: boolean;
}

export function Square({
  piece,
  isWhite,
  isSelected,
  onSquareClick,
  disabled = false,
}: SquareProps) {
  const handleClick = () => {
    if (!disabled && onSquareClick) {
      onSquareClick();
    }
  };

  return (
    <div
      className={`
        w-full aspect-square flex items-center justify-center cursor-pointer
        transition-all duration-200
        ${isWhite ? "bg-[#f0d9b5]" : "bg-[#b58863]"}
        ${!disabled && "hover:opacity-80"}
        ${disabled && "cursor-not-allowed opacity-50"}
        ${isSelected ? "ring-4 ring-yellow-400 ring-inset" : ""}
      `}
      onClick={handleClick}
      style={{
        fontSize: "2.5rem",
      }}
    >
      {piece && (
        <FontAwesomeIcon 
          icon={piece.icon} 
          className={`
            ${piece.color === 'white' ? "text-white drop-shadow-[2px_2px_0_#000]" : "text-black drop-shadow-[2px_2px_0_#fff]"}
            transition-transform duration-200
            ${!disabled && "hover:scale-110"}
          `}
        />
      )}
    </div>
  );
}