import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import BankruptApp from "./BankruptApp.jsx";
import { breachBankAccount, findNPCByBankAccount } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  breachBankAccount: vi.fn(),
  findNPCByBankAccount: vi.fn(),
}));

describe("Bankrupt account breach", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    findNPCByBankAccount.mockResolvedValue([{ id: "npc:ada", name: "Ada Morel", role: "accountant", status: "alive", bankAccountNumber: "12345678" }]);
    breachBankAccount.mockResolvedValue({ npcId: "npc:ada", name: "Ada Morel", accountNumber: "12345678" });
  });

  it("looks up an eight-digit account and opens its ledger after the breach", async () => {
    const onOpenBankLedger = vi.fn();
    render(<BankruptApp onOpenBankLedger={onOpenBankLedger} />);
    fireEvent.change(screen.getByLabelText(/TARGET ACCOUNT/i), { target: { value: "12345678" } });
    fireEvent.click(screen.getByRole("button", { name: /query account/i }));
    expect(await screen.findByText("Ada Morel")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /force account access/i }));
    expect(breachBankAccount).toHaveBeenCalledWith("npc:ada", "12345678");
    await waitFor(() => expect(onOpenBankLedger).toHaveBeenCalledWith(expect.objectContaining({ id: "npc:ada" })));
  });
});
