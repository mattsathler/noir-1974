import { afterEach, describe, expect, it, vi } from "vitest";
import { createCharacter, currentIdentity, loginAccount, logoutAccount, registerAccount } from "./api.js";

afterEach(() => vi.unstubAllGlobals());

describe("identity API", () => {
  it("sends registration details and returns the account response", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response('{"account":{"id":"1"}}', {
      status: 201,
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);

    await expect(registerAccount({ username: "detective", email: "d@example.com", password: "secret" }))
      .resolves.toEqual({ account: { id: "1" } });
    expect(fetch).toHaveBeenCalledWith("/api/auth/register", expect.objectContaining({
      method: "POST",
      credentials: "same-origin",
      body: JSON.stringify({ username: "detective", email: "d@example.com", password: "secret" }),
    }));
  });

  it("logs in and creates a character with JSON payloads", async () => {
    const fetch = vi.fn()
      .mockResolvedValueOnce(new Response('{"account":{"id":"1"}}', { headers: { "content-type": "application/json" } }))
      .mockResolvedValueOnce(new Response('{"character":{"portraitId":2}}', { headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", fetch);

    await loginAccount({ email: "d@example.com", password: "secret" });
    await createCharacter({ displayName: "Ada", portraitId: 2 });
    expect(fetch.mock.calls.map(([path]) => path)).toEqual(["/api/auth/login", "/api/characters"]);
    expect(fetch.mock.calls[1][1].body).toBe(JSON.stringify({ displayName: "Ada", portraitId: 2 }));
  });

  it("returns null after logout and reports server errors", async () => {
    const fetch = vi.fn()
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(new Response("duplicate email\n", { status: 409, headers: { "content-type": "text/plain" } }));
    vi.stubGlobal("fetch", fetch);

    await expect(logoutAccount()).resolves.toBeNull();
    await expect(loginAccount({ email: "d@example.com", password: "secret" })).rejects.toThrow("duplicate email");
  });

  it("requests the current authenticated identity", async () => {
    const identity = { character: null, world: { country: "France", city: "Lyon", timeZone: "Europe/Paris", locationId: "lyon", serverTime: "1974-11-12T20:15:00.000Z" } };
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify(identity), {
      headers: { "content-type": "application/json" },
    }));
    vi.stubGlobal("fetch", fetch);
    await expect(currentIdentity()).resolves.toEqual(identity);
    expect(fetch).toHaveBeenCalledWith("/api/me", expect.objectContaining({ credentials: "same-origin" }));
  });
});
