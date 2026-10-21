import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import NeighborhoodExplorer from "./NeighborhoodExplorer.jsx";
import { getGateDestinations } from "../api.js";

vi.mock("../api.js", () => ({ getGateDestinations: vi.fn() }));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

const central = {
  id: "fr/lyon/district/central",
  name: "Central",
  locations: [
    { address: { locationId: "service:lyon/police-station", name: "Lyon Police Station", kind: "building" } },
    { address: { locationId: "service:lyon/meridian-bank", name: "Meridian Bank", kind: "building" } },
    { address: { locationId: "service:lyon/gate-station", name: "Central GATE Station", kind: "building" } },
  ],
};
const bellecour = {
  id: "fr/lyon/district/bellecour",
  name: "Bellecour",
  locations: [{ address: { locationId: "location:bellecour-office", name: "Bellecour Office", kind: "building" } }],
};

describe("NeighborhoodExplorer", () => {
  it("shows buildings only in the current district and allows visiting its service points", async () => {
    const onVisit = vi.fn().mockResolvedValue(undefined);
    getGateDestinations.mockResolvedValue([central, bellecour]);
    render(<NeighborhoodExplorer
      world={{ districtId: central.id, district: "Central", locationId: central.id }}
      onVisit={onVisit}
      onTravel={vi.fn()}
      onClose={vi.fn()}
    />);

    fireEvent.click(await screen.findByRole("button", { name: /Bellecour/ }));
    expect(screen.getByText("Building records are available after arriving in this district.")).toBeInTheDocument();
    expect(screen.queryByText("Bellecour Office")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Central/ }));
    const bankRow = screen.getByText("Meridian Bank").closest(".gate-interface__place");
    fireEvent.click(within(bankRow).getByRole("button", { name: "ENTER" }));
    await waitFor(() => expect(onVisit).toHaveBeenCalledWith(central.locations[1].address));
  });

  it("returns a player from a service point to the district without changing districts", async () => {
    const onVisit = vi.fn().mockResolvedValue(undefined);
    getGateDestinations.mockResolvedValue([central, bellecour]);
    render(<NeighborhoodExplorer
      world={{ districtId: central.id, district: "Central", locationId: "service:lyon/meridian-bank", placeName: "Meridian Bank" }}
      onVisit={onVisit}
      onTravel={vi.fn()}
      onClose={vi.fn()}
    />);

    fireEvent.click(await screen.findByRole("button", { name: /Central/ }));
    expect(screen.getByText("CURRENT LOCATION")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /RETURN TO DISTRICT/ }));
    await waitFor(() => expect(onVisit).toHaveBeenCalledWith({ locationId: central.id }));
  });
});
