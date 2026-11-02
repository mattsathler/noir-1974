import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import MinesweeperApp, { buildBoard, MINESWEEPER_MINES, MINESWEEPER_SIZE, neighbors, revealArea } from "./MinesweeperApp.jsx";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe("Minesweeper rules", () => {
  it("builds a 9 by 9 board with ten mines and correct neighboring counts", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.42);
    const safeIndex = 40;
    const board = buildBoard(safeIndex);

    expect(board).toHaveLength(MINESWEEPER_SIZE * MINESWEEPER_SIZE);
    expect(board.filter((cell) => cell.mine)).toHaveLength(MINESWEEPER_MINES);
    expect(board[safeIndex].mine).toBe(false);
    board.forEach((cell, index) => {
      if (!cell.mine) expect(cell.adjacent).toBe(neighbors(index).filter((neighbor) => board[neighbor].mine).length);
    });
  });

  it("keeps neighbors within the board and expands empty areas without uncovering flagged cells", () => {
    expect(neighbors(0)).toEqual([1, 9, 10]);
    expect(neighbors(80)).toEqual([70, 71, 79]);
    const board = Array.from({ length: 81 }, () => ({ revealed: false, flagged: false, mine: false, adjacent: 0 }));
    board[80] = { revealed: false, flagged: true, mine: false, adjacent: 0 };

    const revealed = revealArea(board, 40);

    expect(revealed.every((cell) => cell.revealed || cell.flagged)).toBe(true);
    expect(revealed[80]).toMatchObject({ revealed: false, flagged: true });
    expect(board.every((cell) => !cell.revealed)).toBe(true);
  });
});

describe("Minesweeper app", () => {
  it("lets keyboard and pointer users move, flag, and reveal accessible cells", () => {
    render(<MinesweeperApp />);
    const firstCell = screen.getByRole("gridcell", { name: "Row 1, column 1: Covered" });
    firstCell.focus();
    fireEvent.keyDown(firstCell, { key: "ArrowRight" });
    const secondCell = screen.getByRole("gridcell", { name: "Row 1, column 2: Covered" });
    expect(document.activeElement).toBe(secondCell);

    fireEvent.keyDown(secondCell, { key: "f" });
    expect(screen.getByRole("gridcell", { name: "Row 1, column 2: Flagged" })).toHaveAttribute("aria-pressed", "true");
    fireEvent.contextMenu(firstCell);
    expect(screen.getByRole("gridcell", { name: "Row 1, column 1: Flagged" })).toHaveAttribute("aria-pressed", "true");
  });

  it("reveals covered squares with Enter and Space", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.42);
    const { container } = render(<MinesweeperApp />);
    const firstCell = container.querySelector("#minesweeper-cell-0");
    firstCell.focus();
    fireEvent.keyDown(firstCell, { key: "Enter" });
    expect(firstCell).toHaveClass("is-revealed");

    const coveredCell = container.querySelector('.minesweeper-app__cell[aria-label$="Covered"]');
    expect(coveredCell).toBeTruthy();
    fireEvent.keyDown(coveredCell, { key: " " });
    expect(coveredCell).toHaveClass("is-revealed");
  });

  it("always protects the first reveal and reaches a win after all safe squares are opened", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const safeIndex = 40;
    const expectedBoard = buildBoard(safeIndex);
    const { container } = render(<MinesweeperApp />);
    fireEvent.click(container.querySelector("#minesweeper-cell-40"));
    expect(container.querySelector("#minesweeper-cell-40")).toHaveClass("is-revealed");

    expectedBoard.forEach((cell, index) => {
      if (!cell.mine) fireEvent.click(container.querySelector("#minesweeper-cell-" + index));
    });
    expect(container.querySelector(".minesweeper-app__announcement")).toHaveTextContent("Board cleared.");
    expect(container.querySelectorAll(".minesweeper-app__cell.is-revealed")).toHaveLength(81 - MINESWEEPER_MINES);
  });

  it("ends on a mine, blocks further moves, and restarts a fresh board", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.73);
    const expectedBoard = buildBoard(0);
    const mineIndex = expectedBoard.findIndex((cell) => cell.mine);
    const { container } = render(<MinesweeperApp />);
    fireEvent.click(container.querySelector("#minesweeper-cell-0"));
    fireEvent.click(container.querySelector("#minesweeper-cell-" + mineIndex));
    expect(container.querySelector(".minesweeper-app__announcement")).toHaveTextContent("Mine hit.");

    const revealedAfterLoss = container.querySelectorAll(".minesweeper-app__cell.is-revealed").length;
    fireEvent.click(container.querySelector("#minesweeper-cell-1"));
    expect(container.querySelectorAll(".minesweeper-app__cell.is-revealed")).toHaveLength(revealedAfterLoss);

    fireEvent.click(screen.getByRole("button", { name: "Start a new game" }));
    expect(container.querySelector(".minesweeper-app__announcement")).toHaveTextContent("Choose a square to begin.");
    expect(screen.getByLabelText("Elapsed time 00:00")).toBeInTheDocument();
    expect(container.querySelectorAll(".minesweeper-app__cell.is-revealed")).toHaveLength(0);
    expect(container.querySelectorAll(".minesweeper-app__cell.is-flagged")).toHaveLength(0);
  });

  it("counts only the active round and resets its elapsed time", () => {
    vi.useFakeTimers();
    const { container } = render(<MinesweeperApp />);
    fireEvent.click(container.querySelector("#minesweeper-cell-0"));
    act(() => { vi.advanceTimersByTime(2000); });
    expect(screen.getByLabelText("Elapsed time 00:02")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Start a new game" }));
    expect(screen.getByLabelText("Elapsed time 00:00")).toBeInTheDocument();
  });
});
