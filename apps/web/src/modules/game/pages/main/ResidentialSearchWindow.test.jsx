import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import ResidentialSearchWindow from "./ResidentialSearchWindow.jsx";
import { requestResidentialSearch, searchPeople } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({
  requestResidentialSearch: vi.fn(),
  searchPeople: vi.fn(),
}));

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("ResidentialSearchWindow", () => {
  it("queues a search and does not reveal the inventory before the report arrives by mail", async () => {
    searchPeople.mockResolvedValue([{
      npc: { id: "npc:br/rio-de-janeiro/people/ana", name: "Ana Costa", status: "alive" },
    }]);
    requestResidentialSearch.mockResolvedValue({ status: "pending", readyAt: "2026-10-23T23:03:00Z" });

    render(<ResidentialSearchWindow cityId="br/rio-de-janeiro" stationLocationId="service:br/rio-de-janeiro/police-station" onClose={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("PERSON NAME"), { target: { value: "Ana" } });
    fireEvent.click(screen.getByRole("button", { name: "SEARCH" }));
    fireEvent.click(await screen.findByRole("button", { name: /Ana Costa/ }));
    fireEvent.click(screen.getByRole("button", { name: "REQUEST SEARCH" }));

    expect(requestResidentialSearch).toHaveBeenCalledWith("npc:br/rio-de-janeiro/people/ana", "service:br/rio-de-janeiro/police-station");
    expect(await screen.findByText(/AUTHORIZATION REVIEW IS IN PROGRESS.*ABOUT 3 MINUTES/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "REQUEST PENDING" })).toBeDisabled();
    expect(screen.queryByText(/RECOVERED AND SECURED/)).not.toBeInTheDocument();
  });
});
