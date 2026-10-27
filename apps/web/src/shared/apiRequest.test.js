import { afterEach, describe, expect, it, vi } from "vitest";
import { apiRequest, APIRequestError } from "./apiRequest.js";

afterEach(() => vi.unstubAllGlobals());

describe("apiRequest", () => {
  it("shares an identical in-flight GET and releases it when complete", async () => {
    let resolveResponse;
    const fetch = vi.fn()
      .mockImplementationOnce(() => new Promise((resolve) => { resolveResponse = resolve; }))
      .mockResolvedValue(new Response('{"city":"Lyon"}', { headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", fetch);

    const first = apiRequest("/api/world?cityId=fr%2Flyon");
    const second = apiRequest("/api/world?cityId=fr%2Flyon");
    expect(fetch).toHaveBeenCalledTimes(1);

    resolveResponse(new Response('{"city":"Lyon"}', { headers: { "content-type": "application/json" } }));
    await expect(Promise.all([first, second])).resolves.toEqual([{ city: "Lyon" }, { city: "Lyon" }]);

    await apiRequest("/api/world?cityId=fr%2Flyon");
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("does not share mutations, cancellable requests, or different request contexts", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response("{}", { headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", fetch);

    await Promise.all([
      apiRequest("/api/action", { method: "POST", body: "{}" }),
      apiRequest("/api/action", { method: "POST", body: "{}" }),
    ]);
    await Promise.all([
      apiRequest("/api/profile", { headers: { "X-Context": "one" } }),
      apiRequest("/api/profile", { headers: { "X-Context": "two" } }),
    ]);
    await Promise.all([
      apiRequest("/api/world/newspaper", { dedupeKey: "fr/lyon:1974-10-24" }),
      apiRequest("/api/world/newspaper", { dedupeKey: "us/austin:1974-10-24" }),
    ]);
    await Promise.all([
      apiRequest("/api/cancellable", { signal: new AbortController().signal }),
      apiRequest("/api/cancellable", { signal: new AbortController().signal }),
    ]);

    expect(fetch).toHaveBeenCalledTimes(8);
    expect(fetch.mock.calls.every(([, options]) => !("dedupeKey" in options))).toBe(true);
  });

  it("preserves structured diagnostics and permits retry after a failed GET", async () => {
    const fetch = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({
        code: "LOCAL_ACCESS_REQUIRED",
        message: "Authenticate at the local terminal.",
        requestId: "request-456",
        details: "private server detail",
      }), { status: 403, headers: { "content-type": "application/json" } }))
      .mockResolvedValueOnce(new Response('{"ok":true}', { headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", fetch);

    await expect(apiRequest("/api/private")).rejects.toMatchObject({
      name: "APIRequestError",
      message: "Authenticate at the local terminal.",
      code: "LOCAL_ACCESS_REQUIRED",
      status: 403,
      requestId: "request-456",
      details: "private server detail",
    });
    await expect(apiRequest("/api/private")).resolves.toEqual({ ok: true });
    expect(fetch).toHaveBeenCalledTimes(2);
    expect(APIRequestError).toBeDefined();
  });
});
