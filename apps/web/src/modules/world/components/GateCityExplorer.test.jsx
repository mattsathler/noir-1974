import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import GateCityExplorer from "./GateCityExplorer.jsx";
import { getGateDestinations } from "../api.js";

vi.mock("../api.js", () => ({ getGateDestinations: vi.fn() }));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("GateCityExplorer", () => {
  it("lists city destinations by country and starts travel with a city ID", async () => {
    const tokyo = { id: "jp/tokyo", countryId: "jp", country: "Japan", name: "Tokyo", timeZone: "Asia/Tokyo" };
    const lyon = { id: "fr/lyon", countryId: "fr", country: "France", name: "Lyon", timeZone: "Europe/Paris" };
    getGateDestinations.mockResolvedValue([lyon, tokyo]);
    const onTravel = vi.fn().mockResolvedValue(undefined);
    render(<GateCityExplorer world={{ cityId: lyon.id, city: lyon.name, country: lyon.country }} onTravel={onTravel} onClose={vi.fn()} />);

    expect(await screen.findByRole("heading", { name: "City transit" })).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Cities grouped by country" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Tokyo/ })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Tokyo/ }));
    fireEvent.click(screen.getByRole("button", { name: /TRAVEL BY GATE/ }));
    await waitFor(() => expect(onTravel).toHaveBeenCalledWith(tokyo));
    expect(screen.queryByText("Neighborhood access")).not.toBeInTheDocument();
  });
});
