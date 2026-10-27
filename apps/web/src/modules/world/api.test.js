import { afterEach, describe, expect, it, vi } from "vitest";
import { cacheWhiteboardDraft, discardWhiteboardDraft, getCachedWhiteboard, getGateDestinations, getNeighborhoodDestinations, getNPCRecords, getWhiteboard, movePlayer, setWhiteboardSaveState, subscribeWhiteboardUpdates } from "./api.js";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  localStorage.clear();
  sessionStorage.clear();
});

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

describe("whiteboard local cache", () => {
  it("notifies other same-origin tabs only after a board save succeeds", () => {
    class TestBroadcastChannel {
      static channels = [];
      constructor(name) { this.name = name; this.listeners = new Set(); TestBroadcastChannel.channels.push(this); }
      addEventListener(type, listener) { if (type === "message") this.listeners.add(listener); }
      removeEventListener(type, listener) { if (type === "message") this.listeners.delete(listener); }
      postMessage(data) {
        for (const channel of TestBroadcastChannel.channels) {
          if (channel === this || channel.name !== this.name) continue;
          for (const listener of channel.listeners) listener({ data });
        }
      }
    }
    vi.stubGlobal("BroadcastChannel", TestBroadcastChannel);
    const listener = vi.fn();
    const unsubscribe = subscribeWhiteboardUpdates(listener);
    const thisTabChannel = TestBroadcastChannel.channels[0];
    const postMessage = vi.spyOn(thisTabChannel, "postMessage");

    setWhiteboardSaveState("board:sync", "saving");
    expect(postMessage).not.toHaveBeenCalled();
    setWhiteboardSaveState("board:sync", "saved");
    expect(postMessage).toHaveBeenCalledWith({ boardId: "board:sync", status: "saved" });
    const otherTab = new TestBroadcastChannel("noir:whiteboards");
    otherTab.postMessage({ boardId: "board:sync", status: "saved" });
    expect(listener).toHaveBeenCalledWith({ boardId: "board:sync", status: "saved" });
    unsubscribe();
  });

  it("does not replace a newer local mutation with a stale in-flight response", async () => {
    let finishRequest;
    const fetch = vi.fn(() => new Promise((resolve) => { finishRequest = resolve; }));
    vi.stubGlobal("fetch", fetch);
    const oldBoard = { id: "board:race", title: "Old title", pins: [], connections: [] };
    const updatedBoard = { ...oldBoard, title: "Updated title" };

    const pending = getWhiteboard(oldBoard.id);
    cacheWhiteboardDraft(updatedBoard);
    finishRequest(new Response(JSON.stringify({ whiteboard: oldBoard }), {
      headers: { "content-type": "application/json" },
    }));

    await expect(pending).resolves.toEqual(updatedBoard);
    expect(getCachedWhiteboard(oldBoard.id)).toEqual(updatedBoard);
  });

  it("expires cached boards and removes drafts when discarded", () => {
    let now = 1_000_000;
    vi.spyOn(Date, "now").mockImplementation(() => now);
    const board = { id: "board:qa", title: "QA board", pins: [], connections: [] };

    cacheWhiteboardDraft(board);
    expect(getCachedWhiteboard(board.id)).toEqual(board);
    discardWhiteboardDraft(board.id);
    expect(getCachedWhiteboard(board.id)).toBeNull();

    cacheWhiteboardDraft(board);
    now += 24 * 60 * 60 * 1000 + 1;
    expect(getCachedWhiteboard(board.id)).toBeNull();
  });
});
