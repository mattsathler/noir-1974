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
      id: "warrant:1:mail", sentAt: "1974-10-23T23:03:00Z", subject: "Residential search report · Ana Costa",
      body: "The authorized search is complete.", attachments: [],
    }]);
    const setInterval = vi.spyOn(window, "setInterval");
    render(<PlayerMailbox serverTime={new Date("1974-10-23T12:00:00Z")} />);
    expect(await screen.findByText("NO MESSAGES IN THIS DATE RANGE")).toBeInTheDocument();

    const poll = setInterval.mock.calls.find(([, delay]) => delay === 15000)?.[0];
    expect(poll).toBeTypeOf("function");
    await act(async () => { poll(); });

    expect(screen.getByText("Residential search report · Ana Costa")).toBeInTheDocument();
    expect(getPlayerEmails).toHaveBeenCalledTimes(2);
  });

  it("notifies when a new message arrives after the mailbox has loaded", async () => {
    const message = { id: "new-mail", sentAt: "1974-10-23T23:03:00Z", subject: "A new dispatch", sender: "Bill Bouer", body: "Read this report.", attachments: [] };
    const onMessageReceived = vi.fn();
    getPlayerEmails.mockResolvedValueOnce([]).mockResolvedValueOnce([message]);
    const setInterval = vi.spyOn(window, "setInterval");
    render(<PlayerMailbox serverTime={new Date("1974-10-23T12:00:00Z")} onMessageReceived={onMessageReceived} />);
    expect(await screen.findByText("NO MESSAGES IN THIS DATE RANGE")).toBeInTheDocument();
    expect(onMessageReceived).not.toHaveBeenCalled();

    const poll = setInterval.mock.calls.find(([, delay]) => delay === 15000)?.[0];
    await act(async () => { poll(); });
    expect(onMessageReceived).toHaveBeenCalledWith(message);
  });
});
