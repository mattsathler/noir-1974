import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import ArchivePage from "./ArchivePage.jsx";
import { currentIdentity, logoutAccount } from "../../api.js";

vi.mock("../../api.js", () => ({ currentIdentity: vi.fn(), logoutAccount: vi.fn() }));
afterEach(cleanup);

describe("ArchivePage", () => {
  it("loads the saved character dossier before entering the world", async () => {
    const world = { city: "Lyon", country: "France", timeZone: "Europe/Paris", locationId: "lyon", serverTime: "1974-11-12T20:15:00.000Z" };
    currentIdentity.mockResolvedValue({ account: { username: "detective" }, character: { displayName: "Ada", portraitId: 4 }, world });
    const onLoaded = vi.fn();
    render(<ArchivePage onLoaded={onLoaded} />);

    await screen.findByText("AGENT · Ada");
    expect(screen.getByText("PRIVATE FILE · LYON")).toBeInTheDocument();
    await waitFor(() => expect(onLoaded).toHaveBeenCalledWith(world), { timeout: 2500 });
  });
});
