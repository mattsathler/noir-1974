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
  it("queries and presents only the bank app's unmarked ledger", async () => {
    searchPeople.mockResolvedValue([{
      npc: { id: "npc:ada", name: "Ada Morel", age: 41, appearance: "dark hair", status: "alive" },
      employment: [],
      personalRelationships: [],
    }]);
    getNPCRecords.mockResolvedValue([
      { id: "record:other", source: "bank", occurredAt: "1974-11-10T16:42:00.000Z", operator: "Meridian Global Settlement Office", description: "A private remittance cleared.", amount: 68, direction: "outgoing" },
    ]);
    render(<div className="computer-window"><WorldDirectory program="bank-records" serverTime={new Date("1974-11-12T20:15:00Z")} /></div>);
    expect(screen.getByText("MERIDIAN INTERNAL CREDENTIALS DETECTED")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "USE OPERATIVE CREDENTIALS" }));
    await screen.findByLabelText("PERSON NAME");
    fireEvent.change(screen.getByLabelText("PERSON NAME"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "SEARCH" }));
    await screen.findByText("Ada Morel");
    fireEvent.click(screen.getByRole("button", { name: "OPEN FINANCIAL LEDGER" }));
    fireEvent.click(screen.getByRole("button", { name: "Select date FROM" }));
    fireEvent.click(screen.getByRole("button", { name: "1974-11-08" }));
    fireEvent.click(screen.getByRole("button", { name: "READ RECORDS" }));

    expect(await screen.findByText("A private remittance cleared.")).toBeInTheDocument();
    expect(getNPCRecords).toHaveBeenCalledWith("npc:ada", "1974-11-08", "1974-11-12", "bank");
    expect(screen.queryByText("MAXIMUM INTERVAL · 7 DAYS")).not.toBeInTheDocument();
    expect(screen.getByText("DEBIT · SENT")).toBeInTheDocument();
    expect(screen.getByText("−68 CREDITS")).toBeInTheDocument();
    expect(screen.getByText("Sun, Nov 10, 5:42 PM")).toBeInTheDocument();
    expect(screen.getByText("FINANCIAL LEDGER")).toBeInTheDocument();
    expect(screen.queryByText(/case|extortion|evidence|clue/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/1974/)).not.toBeInTheDocument();
  });

  it("requests only GATE records in the GATE app", async () => {
    searchPeople.mockResolvedValue([{
      npc: { id: "npc:ada", name: "Ada Morel", age: 41, appearance: "dark hair", status: "alive" },
      employment: [],
      personalRelationships: [],
    }]);
    getNPCRecords.mockResolvedValue([
      { id: "record:gate", source: "gate", occurredAt: "1974-11-10T13:00:00.000Z", operator: "Meridian Global · GATE Network", description: "A public terminal session ended.", origin: "Bellecour", destination: "Part-Dieu" },
    ]);
    render(<div className="computer-window"><WorldDirectory program="gate-records" serverTime={new Date("1974-11-12T20:15:00Z")} /></div>);
    fireEvent.click(screen.getByRole("button", { name: "USE OPERATIVE CREDENTIALS" }));
    await screen.findByLabelText("PERSON NAME");
    fireEvent.change(screen.getByLabelText("PERSON NAME"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "SEARCH" }));
    await screen.findByText("Ada Morel");
    fireEvent.click(screen.getByRole("button", { name: "OPEN GATE TRAVEL FILE" }));
    fireEvent.click(screen.getByRole("button", { name: "READ RECORDS" }));

    await screen.findByText("A public terminal session ended.");
    expect(getNPCRecords).toHaveBeenCalledWith("npc:ada", "1974-11-06", "1974-11-12", "gate");
    expect(screen.getByText("RECORDED PASSAGE")).toBeInTheDocument();
    expect(screen.getByText("Bellecour")).toBeInTheDocument();
    expect(screen.getByText("Part-Dieu")).toBeInTheDocument();
    expect(screen.queryByText("FINANCIAL LEDGER")).not.toBeInTheDocument();
  });
});
