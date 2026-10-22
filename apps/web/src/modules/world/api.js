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

export function saveCaseProgress(caseId, progress) {
  return requestCase(`/api/cases/${encodeURIComponent(caseId)}/progress`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(progress),
  });
}

async function whiteboardRequest(path, options) {
  const response = await fetch(path, { credentials: "same-origin", ...options });
  const payload = await responsePayload(response);
  if (!response.ok) throw new Error(payload.message?.trim() || "The whiteboard could not be reached.");
  return payload;
}

function readWhiteboardCache(key) {
  try {
    const cached = JSON.parse(sessionStorage.getItem(`noir:whiteboards:${key}`) || "null");
    return cached?.value ?? null;
  } catch { return null; }
}
function writeWhiteboardCache(key, value) {
  try { sessionStorage.setItem(`noir:whiteboards:${key}`, JSON.stringify({ savedAt: Date.now(), value })); } catch { /* Storage may be disabled. */ }
  return value;
}
const whiteboardSaveStates = new Map();
function announceWhiteboardUpdate(boardId, status) {
  if (typeof window !== "undefined") window.dispatchEvent(new CustomEvent("noir:whiteboard-updated", { detail: { boardId, status } }));
}
function updateWhiteboardCache(boardId, update) {
  const board = readWhiteboardCache(boardId);
  if (board) {
    writeWhiteboardCache(boardId, update(board));
    setWhiteboardSaveState(boardId, "saved");
  }
}

export function getCachedWhiteboard(boardId) {
  return readWhiteboardCache(boardId);
}

export function cachePinnedWhiteboardItem(boardId, item, x, y, fallbackBoard = null) {
  const board = readWhiteboardCache(boardId) ?? fallbackBoard ?? { id: boardId, title: "INVESTIGATION BOARD", pins: [], connections: [] };
  writeWhiteboardCache(boardId, ({
    ...board,
    pins: [...(board.pins ?? []).filter((pin) => pin.id !== item.id), { ...item, x, y }],
  }));
  setWhiteboardSaveState(boardId, "saving");
}

export function getWhiteboardSaveState(boardId) { return whiteboardSaveStates.get(boardId) ?? "saved"; }
export function setWhiteboardSaveState(boardId, status) {
  whiteboardSaveStates.set(boardId, status);
  announceWhiteboardUpdate(boardId, status);
}

export async function getWhiteboards() {
  const payload = await whiteboardRequest("/api/whiteboards");
  const remoteBoards = payload.results ?? [];
  const cachedBoards = readWhiteboardCache("list") ?? [];
  const pendingDrafts = cachedBoards.filter((board) =>
    getWhiteboardSaveState(board.id) === "saving" && !remoteBoards.some((remote) => remote.id === board.id),
  );
  return writeWhiteboardCache("list", [...remoteBoards, ...pendingDrafts]);
}

export function createWhiteboardDraft(index = 0) {
  const nonce = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}_${Math.random().toString(36).slice(2)}`;
  return { id: `board_${nonce}`, title: `INVESTIGATION ${index + 1}`, createdAt: new Date().toISOString(), pins: [], connections: [] };
}

export function cacheWhiteboardDraft(board) {
  const boards = readWhiteboardCache("list") ?? [];
  writeWhiteboardCache("list", [...boards.filter((item) => item.id !== board.id), board]);
  writeWhiteboardCache(board.id, board);
}
export function discardWhiteboardDraft(boardId) {
  const boards = readWhiteboardCache("list");
  if (boards) writeWhiteboardCache("list", boards.filter((board) => board.id !== boardId));
  try { sessionStorage.removeItem(`noir:whiteboards:${boardId}`); } catch { /* Storage may be disabled. */ }
  whiteboardSaveStates.delete(boardId);
}

export function createWhiteboard(title = "", boardId = null) {
  const id = boardId ?? createWhiteboardDraft().id;
  return whiteboardRequest("/api/whiteboards", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, id }),
  }).then((payload) => {
    const board = payload.whiteboard;
    const boards = readWhiteboardCache("list");
    if (boards) writeWhiteboardCache("list", [...boards.filter((item) => item.id !== board.id), board]);
    writeWhiteboardCache(board.id, board);
    setWhiteboardSaveState(board.id, "saved");
    return board;
  });
}

export function renameWhiteboard(boardId, title) {
  const previousBoard = readWhiteboardCache(boardId);
  const previousList = readWhiteboardCache("list");
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title }),
  }).then((result) => {
    if (previousBoard) writeWhiteboardCache(boardId, { ...previousBoard, title });
    if (previousList) writeWhiteboardCache("list", previousList.map((board) => board.id === boardId ? { ...board, title } : board));
    setWhiteboardSaveState(boardId, "saved");
    return result;
  });
}

export function deleteWhiteboard(boardId) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}`, { method: "DELETE" }).then((result) => {
    const boards = readWhiteboardCache("list");
    if (boards) writeWhiteboardCache("list", boards.filter((board) => board.id !== boardId));
    try { sessionStorage.removeItem(`noir:whiteboards:${boardId}`); } catch { /* Storage may be disabled. */ }
    whiteboardSaveStates.delete(boardId);
    announceWhiteboardUpdate(boardId, "deleted");
    return result;
  });
}

export async function getWhiteboard(boardId) {
  const cached = readWhiteboardCache(boardId);
  if (cached && getWhiteboardSaveState(boardId) === "saving") return cached;
  const payload = await whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}`);
  return writeWhiteboardCache(boardId, payload.whiteboard);
}

export function pinWhiteboardItem(boardId, itemId, x = 700, y = 700) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/pins`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ itemId, x, y }),
  });
}

export function createWhiteboardNote(boardId, x = 700, y = 700, optimisticId = null) {
  const nonce = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}_${Math.random().toString(36).slice(2)}`;
  const pendingId = optimisticId ?? `note:pending-${nonce}`;
  const pendingPin = { id: pendingId, source: "note", title: "UNTITLED NOTE", content: "", x, y, pending: true };
  const cachedBoard = readWhiteboardCache(boardId);
  if (cachedBoard) {
    writeWhiteboardCache(boardId, {
      ...cachedBoard,
      pins: [...(cachedBoard.pins ?? []).filter((pin) => pin.id !== pendingId), pendingPin],
    });
  }
  setWhiteboardSaveState(boardId, "saving");

  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/notes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ x, y }),
  }).then(({ pin }) => {
    const board = readWhiteboardCache(boardId);
    if (board) {
      writeWhiteboardCache(boardId, {
        ...board,
        pins: [...(board.pins ?? []).filter((item) => item.id !== pendingId && item.id !== pin.id), pin],
      });
    }
    setWhiteboardSaveState(boardId, "saved");
    return pin;
  }).catch((error) => {
    const board = readWhiteboardCache(boardId);
    if (board) {
      writeWhiteboardCache(boardId, { ...board, pins: (board.pins ?? []).filter((pin) => pin.id !== pendingId) });
    }
    setWhiteboardSaveState(boardId, "save");
    throw error;
  });
}

export function updateWhiteboardNote(boardId, itemId, title, content) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/notes/${encodeURIComponent(itemId)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, content }),
  }).then((result) => {
    updateWhiteboardCache(boardId, (board) => ({
      ...board,
      pins: board.pins.map((pin) => pin.id === itemId ? { ...pin, title, content } : pin),
    }));
    return result;
  });
}

export function moveWhiteboardPin(boardId, itemId, x, y) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/pins/${encodeURIComponent(itemId)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ x, y }),
  }).then((result) => { updateWhiteboardCache(boardId, (board) => ({ ...board, pins: board.pins.map((pin) => pin.id === itemId ? { ...pin, x, y } : pin) })); return result; });
}

export function removeWhiteboardPin(boardId, itemId) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/pins/${encodeURIComponent(itemId)}`, { method: "DELETE" }).then((result) => { updateWhiteboardCache(boardId, (board) => ({ ...board, pins: board.pins.filter((pin) => pin.id !== itemId), connections: board.connections.filter((edge) => edge.fromId !== itemId && edge.toId !== itemId) })); return result; });
}

export function connectWhiteboardPins(boardId, fromId, toId) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/connections`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fromId, toId }),
  }).then((result) => { updateWhiteboardCache(boardId, (board) => ({ ...board, connections: board.connections.some((edge) => edge.fromId === fromId && edge.toId === toId || edge.fromId === toId && edge.toId === fromId) ? board.connections : [...board.connections, { fromId, toId }] })); return result; });
}

export function disconnectWhiteboardPins(boardId, fromId, toId) {
  return whiteboardRequest(`/api/whiteboards/${encodeURIComponent(boardId)}/connections/${encodeURIComponent(fromId)}/${encodeURIComponent(toId)}`, { method: "DELETE" }).then((result) => { updateWhiteboardCache(boardId, (board) => ({ ...board, connections: board.connections.filter((edge) => !(edge.fromId === fromId && edge.toId === toId || edge.fromId === toId && edge.toId === fromId)) })); return result; });
}
