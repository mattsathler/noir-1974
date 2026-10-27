import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import Whiteboard from "./Whiteboard.jsx";
import { connectWhiteboardPins, getWhiteboard, getWhiteboards, updateWhiteboardNote } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  getWhiteboard: vi.fn(),
  getWhiteboards: vi.fn(),
  connectWhiteboardPins: vi.fn(),
  getWhiteboardSaveState: vi.fn(() => "saved"),
  getCachedWhiteboard: vi.fn(() => null),
  subscribeWhiteboardUpdates: vi.fn(() => () => {}),
  updateWhiteboardNote: vi.fn(),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("Whiteboard pin designations", () => {
  it("shows submitted boards as locked and disables editing controls", async () => {
    const board = {
      id: "board-locked",
      title: "Submitted case",
      lockedAt: "1974-10-27T10:00:00Z",
      pins: [
        { id: "note-1", source: "note", title: "UNTITLED NOTE", content: "Evidence note", x: 2000, y: 2000 },
        { id: "person-1", source: "person", title: "Ada Morel", content: "PERSON FILE", x: 4000, y: 2000 },
      ],
      connections: [],
    };
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title, lockedAt: board.lockedAt }]);
    getWhiteboard.mockResolvedValue(board);

    render(<Whiteboard />);

    expect(await screen.findByText(/SUBMITTED FOR REVIEW · THIS BOARD IS LOCKED/)).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Note text" })).toBeDisabled();
    expect(screen.getByRole("combobox", { name: "Mark Ada Morel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "+ BLANK NOTE" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Delete Submitted case" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "+ NEW BOARD" })).toBeEnabled();
  });

  it("connects pins by dragging one red tack onto another card", async () => {
    const board = {
      id: "board-strings",
      title: "Case notes",
      pins: [
        { id: "person-1", source: "person", title: "Ada Morel", content: "PERSON FILE", x: 2000, y: 2000 },
        { id: "email-1", source: "email", title: "A message", content: "EMAIL RECORD", x: 5000, y: 2000 },
      ],
      connections: [],
    };
    const previousSetPointerCapture = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "setPointerCapture");
    const previousElementFromPoint = Object.getOwnPropertyDescriptor(document, "elementFromPoint");
    const previousPointerEvent = Object.getOwnPropertyDescriptor(window, "PointerEvent");
    if (!window.PointerEvent) Object.defineProperty(window, "PointerEvent", { configurable: true, value: class PointerEvent extends MouseEvent {
      constructor(type, init = {}) { super(type, init); this.pointerId = init.pointerId ?? 1; }
    } });
    Object.defineProperty(HTMLElement.prototype, "setPointerCapture", { configurable: true, value: vi.fn() });
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title }]);
    getWhiteboard.mockResolvedValue(board);
    connectWhiteboardPins.mockResolvedValue({});

    try {
      render(<Whiteboard />);
      expect(await screen.findByRole("heading", { name: "Case notes" })).toBeInTheDocument();
      const firstTack = screen.getByRole("button", { name: "Connect Ada Morel to another pin" });
      const targetCard = screen.getByText("A message").closest("article");
      Object.defineProperty(document, "elementFromPoint", { configurable: true, value: () => targetCard });

      fireEvent.pointerDown(firstTack, { pointerId: 7, button: 0, clientX: 20, clientY: 20 });
      fireEvent.pointerMove(firstTack, { pointerId: 7, clientX: 80, clientY: 60 });
      fireEvent.pointerUp(firstTack, { pointerId: 7, clientX: 80, clientY: 60 });

      expect(connectWhiteboardPins).toHaveBeenCalledWith(board.id, "person-1", "email-1");
      expect(await screen.findByRole("heading", { name: "CONNECTED RECORDS" })).toBeInTheDocument();
      expect(screen.queryByRole("button", { name: /connect selected/i })).not.toBeInTheDocument();
    } finally {
      if (previousSetPointerCapture) Object.defineProperty(HTMLElement.prototype, "setPointerCapture", previousSetPointerCapture);
      else delete HTMLElement.prototype.setPointerCapture;
      if (previousElementFromPoint) Object.defineProperty(document, "elementFromPoint", previousElementFromPoint);
      else delete document.elementFromPoint;
      if (previousPointerEvent) Object.defineProperty(window, "PointerEvent", previousPointerEvent);
      else delete window.PointerEvent;
    }
  });

  it("keeps keyboard connection available through the tack buttons", async () => {
    const board = {
      id: "board-keyboard-strings",
      title: "Case notes",
      pins: [
        { id: "person-1", source: "person", title: "Ada Morel", content: "PERSON FILE", x: 2000, y: 2000 },
        { id: "email-1", source: "email", title: "A message", content: "EMAIL RECORD", x: 5000, y: 2000 },
      ],
      connections: [],
    };
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title }]);
    getWhiteboard.mockResolvedValue(board);
    connectWhiteboardPins.mockResolvedValue({});

    render(<Whiteboard />);
    expect(await screen.findByRole("heading", { name: "Case notes" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Connect Ada Morel to another pin" }));
    fireEvent.click(screen.getByRole("button", { name: "Connect A message to another pin" }));

    expect(connectWhiteboardPins).toHaveBeenCalledWith(board.id, "person-1", "email-1");
  });

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

  it("shows timestamps on legacy GATE pins and starts dragging only from the header", async () => {
    const board = {
      id: "board-2",
      title: "Case notes",
      pins: [{ id: "gate-1", source: "gate", title: "GATE ACCESS LOG", content: "GATE passage at 2:00 PM", metadata: { occurredAt: "1974-11-10T13:00:00.000Z" }, x: 2000, y: 2000 }],
      connections: [],
    };
    const setPointerCapture = vi.fn();
    const previousDescriptor = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "setPointerCapture");
    Object.defineProperty(HTMLElement.prototype, "setPointerCapture", { configurable: true, value: setPointerCapture });
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title }]);
    getWhiteboard.mockResolvedValue(board);

    try {
      render(<Whiteboard />);
      expect(await screen.findByText(/GATE USED · Tue, Nov 10, 1:00 PM UTC/)).toBeInTheDocument();
      const description = screen.getByText(/GATE passage at 2:00 PM/);

      fireEvent.pointerDown(description, { pointerId: 1, button: 0, clientX: 20, clientY: 20 });
      expect(setPointerCapture).not.toHaveBeenCalled();

      fireEvent.pointerDown(screen.getByText("GATE · PINNED"), { pointerId: 2, button: 0, clientX: 20, clientY: 20 });
      expect(setPointerCapture).toHaveBeenCalledOnce();
    } finally {
      if (previousDescriptor) Object.defineProperty(HTMLElement.prototype, "setPointerCapture", previousDescriptor);
      else delete HTMLElement.prototype.setPointerCapture;
    }
  });

  it("changes SAVE to SAVING to SAVED after a successful retry", async () => {
    const board = {
      id: "board-save",
      title: "Case notes",
      pins: [{ id: "note-save", source: "note", title: "UNTITLED NOTE", content: "old text", x: 2000, y: 2000 }],
      connections: [],
    };
    let resolveRetry;
    getWhiteboards.mockResolvedValue([{ id: board.id, title: board.title }]);
    getWhiteboard.mockResolvedValue(board);
    updateWhiteboardNote
      .mockRejectedValueOnce(new Error("temporary failure"))
      .mockImplementationOnce(() => new Promise((resolve) => { resolveRetry = resolve; }));

    render(<Whiteboard />);
    expect(await screen.findByRole("heading", { name: "Case notes" })).toBeInTheDocument();
    fireEvent.change(screen.getByRole("textbox", { name: "Note text" }), { target: { value: "new observation" } });
    fireEvent.click(await screen.findByRole("button", { name: "SAVE", timeout: 2000 }));

    expect(screen.getByText("SAVING")).toBeInTheDocument();
    expect(updateWhiteboardNote).toHaveBeenCalledTimes(2);
    expect(updateWhiteboardNote).toHaveBeenLastCalledWith(board.id, "note-save", "UNTITLED NOTE", "new observation");

    await act(async () => { resolveRetry({}); });
    expect(screen.getByText("SAVED")).toBeInTheDocument();
  });
});
