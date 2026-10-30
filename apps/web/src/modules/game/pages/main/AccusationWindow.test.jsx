import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import AccusationWindow from "./AccusationWindow.jsx";
import { getWhiteboard, getWhiteboards, submitCaseAccusation } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  getWhiteboard: vi.fn(),
  getWhiteboards: vi.fn(),
  markWhiteboardSubmitted: vi.fn(),
  submitCaseAccusation: vi.fn(),
}));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("AccusationWindow", () => {
  it("marks missing board requirements in red and blocks submission", async () => {
    getWhiteboards.mockResolvedValue([{ id: "board-1", title: "Case notes" }]);
    getWhiteboard.mockResolvedValue({
      id: "board-1", title: "Case notes", pins: [
        { id: "person-1", source: "person", designation: "accused" },
      ],
    });

    render(<AccusationWindow cityId="fr/lyon" onClose={vi.fn()} />);

    expect(await screen.findByText(/AT LEAST ONE PIN MARKED PROOF/)).toHaveClass("is-missing");
    expect(screen.getByText(/ONE PERSON MARKED VICTIM/)).toHaveClass("is-missing");
    expect(screen.getByRole("button", { name: "SUBMIT ACCUSATION" })).toBeDisabled();
    expect(submitCaseAccusation).not.toHaveBeenCalled();
  });

  it("does not count a blank note marked as proof toward accusation readiness", async () => {
    getWhiteboards.mockResolvedValue([{ id: "board-note", title: "Case notes" }]);
    getWhiteboard.mockResolvedValue({
      id: "board-note", title: "Case notes", pins: [
        { id: "note-1", source: "note", designation: "proof" },
        { id: "person-1", source: "person", designation: "accused" },
        { id: "person-2", source: "person", designation: "victim" },
      ],
    });

    render(<AccusationWindow cityId="fr/lyon" onClose={vi.fn()} />);

    expect(await screen.findByText(/AT LEAST ONE PIN MARKED PROOF/)).toHaveClass("is-missing");
    expect(screen.getByRole("button", { name: "SUBMIT ACCUSATION" })).toBeDisabled();
    expect(submitCaseAccusation).not.toHaveBeenCalled();
  });

  it("submits the selected complete board from the police station", async () => {
    getWhiteboards.mockResolvedValue([{ id: "board-2", title: "Case file" }]);
    getWhiteboard.mockResolvedValue({
      id: "board-2", title: "Case file", pins: [
        { id: "proof-1", source: "record", designation: "proof" },
        { id: "person-1", source: "person", designation: "accused" },
        { id: "person-2", source: "person", designation: "victim" },
      ],
    });
    submitCaseAccusation.mockResolvedValue({ status: "pending" });

    render(<AccusationWindow cityId="fr/lyon" onClose={vi.fn()} />);
    fireEvent.click(await screen.findByRole("button", { name: "SUBMIT ACCUSATION" }));

    expect(submitCaseAccusation).toHaveBeenCalledWith("board-2", "service:lyon/police-station");
    expect(await screen.findByText(/REVIEW IS IN PROGRESS.*ABOUT 3 MINUTES/)).toBeInTheDocument();
  });

  it("keeps a submitted board locked and offers a fresh board for another review", async () => {
    getWhiteboards.mockResolvedValue([
      { id: "board-old", title: "First submission", lockedAt: "1974-10-27T10:00:00Z" },
      { id: "board-new", title: "Revised evidence" },
    ]);
    getWhiteboard.mockImplementation(async (id) => ({
      id,
      title: id === "board-old" ? "First submission" : "Revised evidence",
      lockedAt: id === "board-old" ? "1974-10-27T10:00:00Z" : undefined,
      pins: [
        { id: `${id}-proof`, source: "record", designation: "proof" },
        { id: `${id}-accused`, source: "person", designation: "accused" },
        { id: `${id}-victim`, source: "person", designation: "victim" },
      ],
    }));
    submitCaseAccusation.mockResolvedValue({ status: "pending" });

    render(<AccusationWindow cityId="fr/lyon" onClose={vi.fn()} />);
    expect(await screen.findByRole("combobox", { name: "INVESTIGATION BOARD" })).toHaveValue("board-new");
    expect(screen.getByRole("option", { name: /First submission.*SUBMITTED · LOCKED/ })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "SUBMIT ACCUSATION" }));

    expect(submitCaseAccusation).toHaveBeenCalledWith("board-new", "service:lyon/police-station");
    expect(await screen.findByText(/REVIEW IS IN PROGRESS/)).toBeInTheDocument();
  });
});
