async function search(path, query) {
  const response = await fetch(`${path}?q=${encodeURIComponent(query)}`, { credentials: "same-origin" });
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : { message: await response.text() };
  if (!response.ok) throw new Error(payload.message?.trim() || "The directory could not be reached.");
  return payload.results ?? [];
}

export const searchPeople = (query) => search("/api/world/npcs", query);
export const searchCompanies = (query) => search("/api/world/companies", query);
export const searchLocations = (query) => search("/api/world/locations", query);

export async function getGateDestinations() {
  const response = await fetch("/api/world/transit", { credentials: "same-origin" });
  const payload = await responsePayload(response);
  if (!response.ok) throw new Error(payload.message?.trim() || "Could not retrieve GATE destinations.");
  return payload.destinations ?? [];
}

export async function movePlayer(path, body) {
  const response = await fetch(path, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const payload = await responsePayload(response);
  if (!response.ok) throw new Error(payload.message?.trim() || "Could not change location.");
  return payload;
}

export async function getCurrentWorld() {
  const response = await fetch("/api/world", { credentials: "same-origin" });
  const payload = await responsePayload(response);
  if (!response.ok) throw new Error(payload.message?.trim() || "Could not retrieve current location.");
  return payload;
}

async function responsePayload(response) {
  const contentType = response.headers.get("content-type") ?? "";
  return contentType.includes("application/json")
    ? response.json()
    : { message: await response.text() };
}

export async function getNPCEmails(npcId, from, to) {
  const params = new URLSearchParams({ from, to });
  const response = await fetch(`/api/world/npcs/${encodeURIComponent(npcId)}/emails?${params}`, { credentials: "same-origin" });
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : { message: await response.text() };
  if (!response.ok) throw new Error(payload.message?.trim() || "The correspondence office could not be reached.");
  return payload.results ?? [];
}

export async function getNPCRecords(npcId, from, to, source) {
  const params = new URLSearchParams({ from, to });
  if (source) params.set("source", source);
  const response = await fetch(`/api/world/npcs/${encodeURIComponent(npcId)}/records?${params}`, { credentials: "same-origin" });
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : { message: await response.text() };
  if (!response.ok) throw new Error(payload.message?.trim() || "The records office could not be reached.");
  return payload.results ?? [];
}

export async function getNewspaperIssue() {
  const response = await fetch("/api/world/newspaper", { credentials: "same-origin" });
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : { message: await response.text() };
  if (!response.ok) throw new Error(payload.message?.trim() || "The newspaper could not be retrieved.");
  return payload;
}

async function requestCase(path, options) {
  const response = await fetch(path, { credentials: "same-origin", ...options });
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : { message: await response.text() };
  if (!response.ok) throw new Error(payload.message?.trim() || "The case service could not be reached.");
  return payload;
}

export async function getOpenCases() {
  const payload = await requestCase("/api/cases");
  return payload.results ?? [];
}

export function getCaseFile(caseId) {
  return requestCase(`/api/cases/${encodeURIComponent(caseId)}`);
}

export function saveCaseProgress(caseId, discoveredEvidenceIds) {
  return requestCase(`/api/cases/${encodeURIComponent(caseId)}/progress`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ discoveredEvidenceIds }),
  });
}
