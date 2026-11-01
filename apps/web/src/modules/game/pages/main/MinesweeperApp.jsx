import { useEffect, useMemo, useState } from "react";
import { Bomb } from "pixelarticons/react/Bomb";
import "./MinesweeperApp.css";

const SIZE = 9;
const MINES = 10;
const emptyBoard = () => Array.from({ length: SIZE * SIZE }, () => ({ revealed: false, flagged: false, mine: false, adjacent: 0 }));

function buildBoard(safeIndex) {
  const board = emptyBoard();
  const candidates = board.map((_, index) => index).filter((index) => index !== safeIndex);
  for (let index = candidates.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    [candidates[index], candidates[swapIndex]] = [candidates[swapIndex], candidates[index]];
  }
  candidates.slice(0, MINES).forEach((index) => { board[index].mine = true; });
  board.forEach((cell, index) => {
    if (cell.mine) return;
    cell.adjacent = neighbors(index).filter((neighbor) => board[neighbor].mine).length;
  });
  return board;
}

function neighbors(index) {
  const row = Math.floor(index / SIZE);
  const col = index % SIZE;
  const result = [];
  for (let y = Math.max(0, row - 1); y <= Math.min(SIZE - 1, row + 1); y += 1) {
    for (let x = Math.max(0, col - 1); x <= Math.min(SIZE - 1, col + 1); x += 1) {
      const neighbor = y * SIZE + x;
      if (neighbor !== index) result.push(neighbor);
    }
  }
  return result;
}

function revealArea(board, startIndex) {
  const next = board.map((cell) => ({ ...cell }));
  const pending = [startIndex];
  const visited = new Set();
  while (pending.length) {
    const index = pending.pop();
    if (visited.has(index)) continue;
    visited.add(index);
    const cell = next[index];
    if (cell.revealed || cell.flagged || cell.mine) continue;
    cell.revealed = true;
    if (!cell.adjacent) pending.push(...neighbors(index));
  }
  return next;
}

function formatTime(seconds) {
  return `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
}

export default function MinesweeperApp() {
  const [board, setBoard] = useState(emptyBoard);
  const [started, setStarted] = useState(false);
  const [status, setStatus] = useState("ready");
  const [seconds, setSeconds] = useState(0);
  const [focusIndex, setFocusIndex] = useState(0);
  const flags = useMemo(() => board.filter((cell) => cell.flagged).length, [board]);

  useEffect(() => {
    if (!started || status !== "playing") return undefined;
    const timer = window.setInterval(() => setSeconds((value) => value + 1), 1000);
    return () => window.clearInterval(timer);
  }, [started, status]);

  const reveal = (index) => {
    if (status === "won" || status === "lost" || board[index].flagged || board[index].revealed) return;
    const current = started ? board.map((cell) => ({ ...cell })) : buildBoard(index);
    const hitMine = current[index].mine;
    if (hitMine) {
      current.forEach((cell) => { if (cell.mine) cell.revealed = true; });
      setBoard(current);
      setStatus("lost");
      setStarted(true);
      return;
    }
    const next = revealArea(current, index);
    const won = next.every((cell) => cell.mine || cell.revealed);
    setBoard(next);
    setStarted(true);
    setStatus(won ? "won" : "playing");
  };

  const toggleFlag = (index) => {
    if (status === "won" || status === "lost" || board[index].revealed) return;
    setBoard((current) => current.map((cell, cellIndex) => cellIndex === index ? { ...cell, flagged: !cell.flagged } : cell));
  };

  const restart = () => {
    setBoard(emptyBoard());
    setStarted(false);
    setStatus("ready");
    setSeconds(0);
    setFocusIndex(0);
  };

  const moveFocus = (index, key) => {
    const row = Math.floor(index / SIZE);
    const col = index % SIZE;
    const nextRow = Math.max(0, Math.min(SIZE - 1, row + (key === "ArrowDown" ? 1 : key === "ArrowUp" ? -1 : 0)));
    const nextCol = Math.max(0, Math.min(SIZE - 1, col + (key === "ArrowRight" ? 1 : key === "ArrowLeft" ? -1 : 0)));
    const nextIndex = nextRow * SIZE + nextCol;
    setFocusIndex(nextIndex);
    document.getElementById(`minesweeper-cell-${nextIndex}`)?.focus();
  };

  const statusText = status === "won" ? "Board cleared." : status === "lost" ? "Mine hit. Start a new game to play again." : status === "playing" ? "Game in progress." : "Choose a square to begin. Your first square is always safe.";

  return <main className="minesweeper-app">
    <header className="minesweeper-app__topbar" aria-label="Game controls">
      <output className="minesweeper-app__counter" aria-label={`${MINES - flags} mines remaining`}>{String(MINES - flags).padStart(3, "0")}</output>
      <button type="button" className={`minesweeper-app__face minesweeper-app__face--${status}`} aria-label="Start a new game" title="New game" onClick={restart}>{status === "lost" ? "☹" : status === "won" ? "☻" : "☺"}</button>
      <output className="minesweeper-app__counter" aria-label={`Elapsed time ${formatTime(seconds)}`}>{String(Math.min(seconds, 999)).padStart(3, "0")}</output>
    </header>
    <p className="minesweeper-app__announcement" role="status" aria-live="polite">{statusText}</p>
    <div className="minesweeper-app__board" role="grid" aria-label="Minesweeper board, 9 rows by 9 columns">
      {board.map((cell, index) => {
        const row = Math.floor(index / SIZE) + 1;
        const col = index % SIZE + 1;
        const text = cell.flagged ? "Flagged" : cell.revealed ? cell.mine ? "Mine" : cell.adjacent ? `${cell.adjacent} adjacent mines` : "Empty" : "Covered";
        return <button
          key={index}
          id={`minesweeper-cell-${index}`}
          type="button"
          role="gridcell"
          aria-label={`Row ${row}, column ${col}: ${text}`}
          aria-pressed={cell.flagged}
          tabIndex={focusIndex === index ? 0 : -1}
          className={`minesweeper-app__cell${cell.revealed ? " is-revealed" : ""}${cell.mine && cell.revealed ? " is-mine" : ""}${cell.flagged ? " is-flagged" : ""}${cell.adjacent ? ` number-${cell.adjacent}` : ""}`}
          onFocus={() => setFocusIndex(index)}
          onClick={() => reveal(index)}
          onContextMenu={(event) => { event.preventDefault(); toggleFlag(index); }}
          onKeyDown={(event) => {
            if (event.key.startsWith("Arrow")) { event.preventDefault(); moveFocus(index, event.key); }
            else if (event.key.toLowerCase() === "f") { event.preventDefault(); toggleFlag(index); }
          }}
        >{cell.flagged ? "⚑" : cell.revealed ? cell.mine ? <Bomb className="minesweeper-app__bomb" aria-hidden="true" /> : cell.adjacent || "" : ""}</button>;
      })}
    </div>
    <p className="minesweeper-app__instructions">Use the arrow keys to move, Enter or Space to reveal, and F to place a flag. Right-click also places a flag.</p>
  </main>;
}
