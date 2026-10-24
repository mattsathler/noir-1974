import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import APIErrorNotice from "./APIErrorNotice.jsx";

describe("APIErrorNotice", () => {
  it("shows a safe message and request reference, never internal details, and retries on demand", () => {
    const retry = vi.fn();
    render(<APIErrorNotice error={{ message: "The request exceeded its time limit.", code: "REQUEST_TIMEOUT", requestId: "req-123", details: "SELECT * FROM private_table" }} onRetry={retry} />);

    expect(screen.getByRole("alert")).toHaveTextContent("The request exceeded its time limit.");
    expect(screen.getByText("REFERENCE req-123 · REQUEST_TIMEOUT")).toBeInTheDocument();
    expect(screen.queryByText(/private_table|SELECT/i)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "TRY AGAIN" }));
    expect(retry).toHaveBeenCalledOnce();
  });
});
