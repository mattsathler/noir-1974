import { afterEach, describe, expect, it, vi } from "vitest";
import { getGateDestinations, getNeighborhoodDestinations, getNPCRecords, movePlayer } from "./api.js";

afterEach(() => vi.unstubAllGlobals());

describe("NPC activity records API", () => {
  it("requests one NPC for the selected date interval", async () => {
    const results = [{ id: "record:1", source: "bank", description: "A payment was processed." }];
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ results }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await expect(getNPCRecords("npc:ada/01", "2026-11-01", "2026-11-12")).resolves.toEqual(results);
    expect(fetch).toHaveBeenCalledWith(
      "/api/world/npcs/npc%3Aada%2F01/records?from=2026-11-01&to=2026-11-12",
      { credentials: "same-origin" },
    );
  });

  it("scopes the request to one records system", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ results: [] }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await getNPCRecords("npc:ada", "2026-11-01", "2026-11-12", "gate");
    expect(fetch).toHaveBeenCalledWith(
      "/api/world/npcs/npc%3Aada/records?from=2026-11-01&to=2026-11-12&source=gate",
      { credentials: "same-origin" },
    );
  });

  it("surfaces API errors", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response("date range must be valid\n", { status: 400 }));
    vi.stubGlobal("fetch", fetch);
    await expect(getNPCRecords("npc:ada", "2026-12-01", "2026-11-01")).rejects.toThrow("date range must be valid");
  });
});

describe("world navigation APIs", () => {
  it("preserves structured API diagnostics for travel failures", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      code: "GATE_CITY_MATERIALIZATION_FAILED",
      message: "The destination could not be prepared.",
      requestId: "request-123",
      details: "foreign key constraint",
    }), { status: 500, headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", fetch);

    await expect(movePlayer("/api/world/gate/travel", { cityId: "us/los-angeles" })).rejects.toMatchObject({
      message: "The destination could not be prepared.",
      code: "GATE_CITY_MATERIALIZATION_FAILED",
      requestId: "request-123",
      details: "foreign key constraint",
    });
  });

  it("loads city-only destinations from the GATE registry endpoint", async () => {
    const destinations = [{ id: "jp/tokyo", countryId: "jp", country: "Japan", name: "Tokyo" }];
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ destinations }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await expect(getGateDestinations()).resolves.toEqual(destinations);
    expect(fetch).toHaveBeenCalledWith("/api/world/gate/destinations", { credentials: "same-origin" });
  });

  it("loads local neighborhoods from the separate navigation endpoint", async () => {
    const destinations = [{ id: "fr/lyon/district/central", name: "Central" }];
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ destinations }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await expect(getNeighborhoodDestinations("fr/lyon")).resolves.toEqual(destinations);
    expect(fetch).toHaveBeenCalledWith("/api/world/transit?cityId=fr%2Flyon", { credentials: "same-origin" });
  });
});
