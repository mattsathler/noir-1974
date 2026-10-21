import { afterEach, describe, expect, it, vi } from "vitest";
import { getNPCRecords } from "./api.js";

afterEach(() => vi.unstubAllGlobals());

describe("NPC activity records API", () => {
  it("requests one NPC for the selected date interval", async () => {
    const results = [{ id: "record:1", source: "bank", description: "A payment was processed." }];
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ results }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await expect(getNPCRecords("npc:ada/01", "1974-11-01", "1974-11-12")).resolves.toEqual(results);
    expect(fetch).toHaveBeenCalledWith(
      "/api/world/npcs/npc%3Aada%2F01/records?from=1974-11-01&to=1974-11-12",
      { credentials: "same-origin" },
    );
  });

  it("scopes the request to one records system", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({ results: [] }), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await getNPCRecords("npc:ada", "1974-11-01", "1974-11-12", "gate");
    expect(fetch).toHaveBeenCalledWith(
      "/api/world/npcs/npc%3Aada/records?from=1974-11-01&to=1974-11-12&source=gate",
      { credentials: "same-origin" },
    );
  });

  it("surfaces API errors", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response("date range must be valid\n", { status: 400 }));
    vi.stubGlobal("fetch", fetch);
    await expect(getNPCRecords("npc:ada", "1974-12-01", "1974-11-01")).rejects.toThrow("date range must be valid");
  });
});
