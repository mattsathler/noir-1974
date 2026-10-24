import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import Whiteboard from "./Whiteboard.jsx";
import { getWhiteboard, getWhiteboards } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  getWhiteboard: vi.fn(),
  getWhiteboards: vi.fn(),
  getWhiteboardSaveState: vi.fn(() => "saved"),
  getCachedWhiteboard: vi.fn(() => null),
  subscribeWhiteboardUpdates: vi.fn(() => () => {}),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("Whiteboard pin designations", () => {
  it("does not offer proof designation for blank notes", async () => {
    const board = {
      id: "board-1",
      title: "Case notes",
      pins: [
        { id: "note-1", source: "note", title: "UNTITLED NOTE", content: "", designation: "proof", x: 2000, y: 2000 },
        { id: "person-1", source: "person", title: "Ada Morel", content: "PERSON FILE", x: 4000, y: 2000 },
      ],
      connections: [],
    };
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title }]);
    getWhiteboard.mockResolvedValue(board);

    render(<Whiteboard />);

    expect(await screen.findByRole("heading", { name: "Case notes" })).toBeInTheDocument();
    expect(screen.queryByRole("combobox", { name: "Mark note" })).not.toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Mark Ada Morel" })).toBeInTheDocument();
  });
});
