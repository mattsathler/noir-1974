import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import CaseFiles from "./CaseFiles.jsx";
import { destroyCaseEvidence, getCaseFile, getCrimeReports } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  destroyCaseEvidence: vi.fn(),
  getCaseFile: vi.fn(),
  getCrimeReports: vi.fn(),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("CaseFiles source destruction", () => {
  it("requires confirmation and removes the source after a successful destruction", async () => {
    getCrimeReports.mockResolvedValue({ cases: [{ id: "case-1", cityId: "fr/lyon", caseType: "embezzlement", status: "open", organizationName: "Aurore Industries", crimeAt: "1974-10-20T12:00:00Z" }], missing: [] });
    getCaseFile.mockResolvedValue({ evidence: [{ id: "case-1/evidence/register", presentation: "Altered company register", content: "A revised ledger page." }] });
    destroyCaseEvidence.mockResolvedValue(null);

    render(<CaseFiles cityId="fr/lyon" issueDay="1974-10-27" timeZone="UTC" />);
    fireEvent.click(await screen.findByRole("button", { name: "Aurore Industries" }));
    expect(await screen.findByText("Altered company register")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "DESTROY SOURCE…" }));
    expect(screen.getByRole("alertdialog")).toHaveTextContent("permanently removes");
    fireEvent.click(screen.getByRole("button", { name: "DESTROY SOURCE" }));

    expect(destroyCaseEvidence).toHaveBeenCalledWith("case-1/evidence/register");
    expect(await screen.findByText("No further public records are available.")).toBeInTheDocument();
  });
});
