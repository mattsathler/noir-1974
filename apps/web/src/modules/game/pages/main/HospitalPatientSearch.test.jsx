import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import HospitalPatientSearch from "./HospitalPatientSearch.jsx";
import { searchHospitalPatients } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({ searchHospitalPatients: vi.fn() }));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("HospitalPatientSearch", () => {
  it("opens a patient file, compares the fingerprint, and sends the report and trace to print destinations", async () => {
    const onPrint = vi.fn();
    searchHospitalPatients.mockResolvedValue([{
      id: "npc:fr/lyon/victim",
      name: "Pascal Poulin",
      status: "dead",
      fingerprintCode: "AB12345",
      visual: { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", faceId: "face_round", accessoryIds: [] },
      postMortem: {
        reportId: "case:postmortem:case-1",
        estimatedDeathAt: "1974-10-23T23:55:00.000Z",
        cause: "Fatal blunt force injuries",
        bodyFoundAt: "Immeuble Saint-Paul, Central, Lyon, France",
        evidence: [{ id: "hospital:case-1:trace", source: "case", title: "Post-mortem trace", content: "A partial fingerprint code begins AB12." }],
      },
    }]);

    render(<HospitalPatientSearch accessAttemptId="71" timeZone="Europe/Paris" onPrint={onPrint} />);
    fireEvent.change(screen.getByLabelText("PATIENT NAME"), { target: { value: "Pascal" } });
    fireEvent.click(screen.getByRole("button", { name: "SEARCH" }));

    fireEvent.click(await screen.findByRole("button", { name: /Pascal Poulin.*FINGERPRINT AB12345/ }));
    expect(searchHospitalPatients).toHaveBeenCalledWith("Pascal", "71");
    expect(screen.getByText("AB12345")).toBeInTheDocument();
    expect(screen.getByText("Fatal blunt force injuries")).toBeInTheDocument();
    expect(screen.getByText("A partial fingerprint code begins AB12.")).toBeInTheDocument();
    expect(screen.queryByText(/killer|match confirmed|culprit/i)).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Print post-mortem report for Pascal Poulin" }));
    expect(onPrint).toHaveBeenLastCalledWith(expect.objectContaining({
      id: "case:postmortem:case-1",
      source: "case",
      content: expect.stringContaining("BODY FOUND · Immeuble Saint-Paul"),
    }));
    fireEvent.click(screen.getByRole("button", { name: "Print Post-mortem trace" }));
    expect(onPrint).toHaveBeenLastCalledWith(expect.objectContaining({ id: "hospital:case-1:trace", source: "case" }));
  });

  it("shows no homicide data for ordinary patient files", async () => {
    searchHospitalPatients.mockResolvedValue([{ id: "npc:lyon:one", name: "Ana Morel", status: "alive", fingerprintCode: "CD67890" }]);
    render(<HospitalPatientSearch accessAttemptId="71" timeZone="Europe/Paris" />);
    fireEvent.change(screen.getByLabelText("PATIENT NAME"), { target: { value: "Ana" } });
    fireEvent.click(screen.getByRole("button", { name: "SEARCH" }));
    fireEvent.click(await screen.findByRole("button", { name: /Ana Morel/ }));

    expect(screen.getByText("CD67890")).toBeInTheDocument();
    expect(screen.queryByText("POST-MORTEM REPORT")).not.toBeInTheDocument();
    expect(screen.queryByText("POST-MORTEM TRACES")).not.toBeInTheDocument();
  });
});
