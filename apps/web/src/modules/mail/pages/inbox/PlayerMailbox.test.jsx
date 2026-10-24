import { act, cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import PlayerMailbox from "./PlayerMailbox.jsx";
import { getPlayerEmails } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({ getPlayerEmails: vi.fn() }));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  vi.restoreAllMocks();
});

describe("PlayerMailbox", () => {
  it("refreshes while open so delayed reports arrive without reopening the mailbox", async () => {
    getPlayerEmails.mockResolvedValueOnce([]).mockResolvedValueOnce([{
      id: "warrant:1:mail", sentAt: "2026-10-23T23:03:00Z", subject: "Residential search report · Ana Costa",
      body: "The authorized search is complete.", attachments: [],
    }]);
    const setInterval = vi.spyOn(window, "setInterval");
    render(<PlayerMailbox />);
    expect(await screen.findByText("NO PERSONAL MAIL ON FILE")).toBeInTheDocument();

    const poll = setInterval.mock.calls.find(([, delay]) => delay === 15000)?.[0];
    expect(poll).toBeTypeOf("function");
    await act(async () => { poll(); });

    expect(screen.getByText("Residential search report · Ana Costa")).toBeInTheDocument();
    expect(getPlayerEmails).toHaveBeenCalledTimes(2);
  });
});
