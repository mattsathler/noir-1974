import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import TerminalApp from "./TerminalApp.jsx";

describe("Meridian terminal command launcher", () => {
  afterEach(() => vi.useRealTimers());

  it("lists Bankrupt in HELP and launches it with the run command", async () => {
    vi.useFakeTimers();
    const onRun = vi.fn();
    render(<TerminalApp onRun={onRun} />);
    const input = screen.getByLabelText("Enter a terminal command");
    fireEvent.change(input, { target: { value: "help" } });
    fireEvent.submit(input.closest("form"));
    expect(screen.getByText(/RUN BANKRUPT\.EXE\s+Start Bankrupt/u)).toBeInTheDocument();
    fireEvent.change(input, { target: { value: "run bankrupt.exe" } });
    fireEvent.submit(input.closest("form"));
    expect(screen.getByText("Starting BANKRUPT.EXE...")).toBeInTheDocument();
    expect(screen.getByText(/BANK BREACH/u)).toBeInTheDocument();
    await vi.advanceTimersByTimeAsync(1600);
    expect(onRun).toHaveBeenCalledWith("bankrupt");
  });
});
