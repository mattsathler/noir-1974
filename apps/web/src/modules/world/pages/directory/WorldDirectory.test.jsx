import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import WorldDirectory from "./WorldDirectory.jsx";
import { getNPCRecords, searchPeople } from "../../api.js";

vi.mock("../../api.js", () => ({
  getNPCRecords: vi.fn(),
  searchCompanies: vi.fn(),
  searchLocations: vi.fn(),
  searchPeople: vi.fn(),
}));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("WorldDirectory activity records", () => {
  it("queries the bank ledger while showing only the account holder's name", async () => {
    searchPeople.mockResolvedValue([{
      npc: { id: "npc:ada", name: "Ada Morel", age: 41, appearance: "dark hair" },
      workplaces: [{ company: "Aurore Industries", address: "Aurore Industries, Bellecour, Lyon, France" }],
      maritalStatus: "married",
    }]);
    getNPCRecords.mockResolvedValue([
      { id: "record:other", source: "bank", occurredAt: "1974-11-10T16:42:00.000Z", operator: "Union Mercantile Settlement Office", description: "A private remittance cleared.", amount: 68, direction: "outgoing" },
    ]);
    render(<div className="computer-window"><WorldDirectory program="bank-records" serverTime={new Date("1974-11-12T20:15:00Z")} runHack={(action) => action()} /></div>);
    await screen.findByLabelText("PERSON NAME");
    fireEvent.change(screen.getByLabelText("PERSON NAME"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Press Enter to search" }));
    await screen.findByText("Ada Morel");
    expect(screen.queryByText(/WORKPLACE|MARITAL STATUS|41 YEARS|dark hair/i)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Ada Morel.*ACCOUNT HOLDER/ }));
    fireEvent.click(screen.getByRole("button", { name: "Select date FROM" }));
    fireEvent.click(screen.getByRole("button", { name: "Nov 8" }));
    fireEvent.click(screen.getByRole("button", { name: "Press Enter to search records" }));

    expect(await screen.findByText("A private remittance cleared.")).toBeInTheDocument();
    expect(getNPCRecords).toHaveBeenCalledWith("npc:ada", "1974-11-08", "1974-11-12", "bank", "", "");
    expect(screen.queryByText("MAXIMUM INTERVAL · 7 DAYS")).not.toBeInTheDocument();
    expect(screen.getByText("DEBIT · SENT")).toBeInTheDocument();
    expect(screen.getByText("−68 CREDITS")).toBeInTheDocument();
    expect(screen.getByText("Sun, Nov 10, 5:42 PM")).toBeInTheDocument();
    expect(screen.getByText("FINANCIAL LEDGER")).toBeInTheDocument();
    expect(screen.queryByText(/case|extortion|evidence|clue/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/\b\d{4}\b/)).not.toBeInTheDocument();
  });

  it("requests only GATE records in the GATE app and sends them to print destinations", async () => {
    const onPrint = vi.fn();
    searchPeople.mockResolvedValue([{
      npc: { id: "npc:ada", name: "Ada Morel", age: 41, appearance: "dark hair" },
      workplaces: [],
      maritalStatus: "single",
    }]);
    getNPCRecords.mockResolvedValue([
      { id: "record:gate", source: "gate", occurredAt: "1974-11-10T13:00:00.000Z", operator: "Meridian Global · GATE Network", description: "A public terminal session ended.", origin: "Bellecour", destination: "Part-Dieu" },
    ]);
    render(<div className="computer-window"><WorldDirectory program="gate-records" serverTime={new Date("1974-11-12T20:15:00Z")} runHack={(action) => action()} onPrint={onPrint} /></div>);
    await screen.findByLabelText("PERSON NAME");
    fireEvent.change(screen.getByLabelText("PERSON NAME"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Press Enter to search" }));
    await screen.findByText("Ada Morel");
    fireEvent.click(screen.getByRole("button", { name: /Ada Morel.*TRAVEL SUBJECT/ }));
    fireEvent.click(screen.getByRole("button", { name: "Press Enter to search records" }));

    await screen.findByText("A public terminal session ended.");
    expect(getNPCRecords).toHaveBeenCalledWith("npc:ada", "1974-11-06", "1974-11-12", "gate", "", "");
    expect(screen.getByText("RECORDED PASSAGE")).toBeInTheDocument();
    expect(screen.getByText("Bellecour")).toBeInTheDocument();
    expect(screen.getByText("Part-Dieu")).toBeInTheDocument();
    expect(screen.queryByText("FINANCIAL LEDGER")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Print GATE log" }));
    expect(onPrint).toHaveBeenCalledWith(expect.objectContaining({
      id: "record:gate",
      content: expect.stringContaining("GATE USED · Sun, Nov 10, 2:00 PM"),
    }));
  });
});
