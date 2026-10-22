import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import GateCityExplorer from "./GateCityExplorer.jsx";
import { getGateDestinations, getNeighborhoodDestinations } from "../api.js";

vi.mock("../api.js", () => ({ getGateDestinations: vi.fn(), getNeighborhoodDestinations: vi.fn() }));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("GateCityExplorer", () => {
  it("keeps country, city, and neighborhood columns visible and starts at the current city", async () => {
    const tokyo = { id: "jp/tokyo", countryId: "jp", country: "Japan", name: "Tokyo", timeZone: "Asia/Tokyo" };
    const lyon = { id: "fr/lyon", countryId: "fr", country: "France", name: "Lyon", timeZone: "Europe/Paris" };
    const lyonCentral = { id: "fr/lyon/district/central", name: "Central", locations: [] };
    const tokyoCentral = { id: "jp/tokyo/district/central", name: "Central", locations: [] };
    const tokyoOldQuarter = { id: "jp/tokyo/district/old-quarter", name: "Old Quarter", locations: [] };
    getGateDestinations.mockResolvedValue([lyon, tokyo]);
    getNeighborhoodDestinations.mockImplementation(async (cityId) => cityId === tokyo.id ? [tokyoCentral, tokyoOldQuarter] : [lyonCentral]);
    const onTravel = vi.fn().mockResolvedValue(undefined);
    render(<GateCityExplorer world={{ cityId: lyon.id, city: lyon.name, country: lyon.country }} onTravel={onTravel} onClose={vi.fn()} />);

    expect(await screen.findByRole("heading", { name: "Travel registry" })).toBeInTheDocument();
    expect(screen.getByText("01 / COUNTRY")).toBeInTheDocument();
    expect(screen.getByText("02 / CITY")).toBeInTheDocument();
    expect(screen.getByText("03 / NEIGHBORHOOD")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /France/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: /Lyon/ })).toHaveAttribute("aria-pressed", "true");
    expect(getNeighborhoodDestinations).toHaveBeenCalledWith(lyon.id);
    expect(screen.queryByRole("button", { name: /BACK/ })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Japan/ }));
    expect(screen.getByRole("button", { name: /Tokyo/ })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Tokyo/ }));
    const oldQuarter = await screen.findByRole("button", { name: /Old Quarter/ });
    fireEvent.click(oldQuarter);
    fireEvent.click(screen.getByRole("button", { name: /TRAVEL BY GATE/ }));
    await waitFor(() => expect(onTravel).toHaveBeenCalledWith({ city: tokyo, district: tokyoOldQuarter }));
  });
});
