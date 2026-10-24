const inFlightGETs = new Map();

export class APIRequestError extends Error {
  constructor(payload, status) {
    const message = typeof payload?.message === "string" ? payload.message.trim() : "";
    super(message || "The request could not be completed.");
    this.name = "APIRequestError";
    this.code = payload?.code;
    this.status = status;
    this.requestId = payload?.requestId;
    this.details = payload?.details;
  }
}

async function readPayload(response) {
  if (response.status === 204) return null;
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    try {
      return await response.json();
    } catch {
      return {};
    }
  }
  return { message: await response.text() };
}

function dedupeKey(path, options) {
  const headers = new Headers(options.headers);
  const stableHeaders = [...headers.entries()].sort(([left], [right]) => left.localeCompare(right));
  return JSON.stringify([
    path,
    options.credentials ?? "same-origin",
    options.cache ?? "default",
    options.mode ?? "cors",
    options.redirect ?? "follow",
    stableHeaders,
    options.dedupeKey ?? "",
  ]);
}

async function performRequest(path, options) {
  const response = await fetch(path, options);
  const payload = await readPayload(response);
  if (!response.ok) throw new APIRequestError(payload, response.status);
  return payload;
}

// Identical GETs share only their in-flight promise. Mutations are never cached,
// and requests with an AbortSignal stay independent so one caller cannot cancel another.
export function apiRequest(path, options = {}) {
  const { dedupeKey: contextKey, ...fetchOptions } = options;
  const method = (fetchOptions.method ?? "GET").toUpperCase();
  const requestOptions = {
    credentials: "same-origin",
    ...fetchOptions,
  };
  if (fetchOptions.method) requestOptions.method = method;

  if (method !== "GET" || requestOptions.signal) return performRequest(path, requestOptions);

  const key = dedupeKey(path, { ...requestOptions, dedupeKey: contextKey });
  const current = inFlightGETs.get(key);
  if (current) return current;

  const request = performRequest(path, requestOptions);
  inFlightGETs.set(key, request);
  request.then(
    () => { if (inFlightGETs.get(key) === request) inFlightGETs.delete(key); },
    () => { if (inFlightGETs.get(key) === request) inFlightGETs.delete(key); },
  );
  return request;
}
