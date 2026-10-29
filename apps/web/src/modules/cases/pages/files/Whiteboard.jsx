import { useEffect, useRef, useState } from "react";
import { Pencil } from "pixelarticons/react/Pencil";
import { Trash } from "pixelarticons/react/Trash";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import {
  connectWhiteboardPins,
  cacheWhiteboardDraft,
  createWhiteboard,
  createWhiteboardDraft,
  designateWhiteboardPin,
  createWhiteboardNote,
  deleteWhiteboard,
  discardWhiteboardDraft,
  disconnectWhiteboardPins,
  getCachedWhiteboard,
  getWhiteboardSaveState,
  getWhiteboard,
  getWhiteboards,
  moveWhiteboardPin,
  renameWhiteboard,
  removeWhiteboardPin,
  subscribeWhiteboardUpdates,
  updateWhiteboardNote,
} from "../../../world/api.js";

function clamp(value, min, max) { return Math.max(min, Math.min(max, value)); }

function displayPinContent(pin) {
  const content = pin.content ?? "";
  if (pin.source !== "gate" || content.includes("GATE USED ·")) return content;
  const occurredAt = pin.metadata?.occurredAt;
  const date = new Date(occurredAt);
  if (!occurredAt || Number.isNaN(date.getTime())) return content;
  const timestamp = new Intl.DateTimeFormat("en-US", {
    timeZone: "UTC",
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
    timeZoneName: "short",
  }).format(date);
  return `GATE USED · ${timestamp} · ${content}`;
}

export default function Whiteboard({ initialBoardId, focusBoard, standalone = false }) {
  const [boards, setBoards] = useState([]);
  const [board, setBoard] = useState(null);
  const boardLocked = Boolean(board?.lockedAt);
  const [connectionStartID, setConnectionStartID] = useState(null);
  const [connectionPreview, setConnectionPreview] = useState(null);
  const [message, setMessage] = useState("CONNECTING TO PERSONAL BOARDS…");
  const [busy, setBusy] = useState(false);
  const [saveState, setSaveState] = useState("saved");
  const [renamingBoardId, setRenamingBoardId] = useState(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [renameValue, setRenameValue] = useState("");
  const [noteDrafts, setNoteDrafts] = useState({});
  const [connectorLines, setConnectorLines] = useState([]);
  const viewportRef = useRef(null);
  const stageRef = useRef(null);
  const pinRefs = useRef(new Map());
  const connectionDragRef = useRef(null);
  const suppressConnectionClickRef = useRef(false);
  const dragRef = useRef(null);
  const panRef = useRef(null);
  const renameInputRef = useRef(null);
  const noteDraftsRef = useRef({});
  const noteTimersRef = useRef(new Map());
  const noteVersionsRef = useRef(new Map());
  const activeNoteSavesRef = useRef(0);
  const lastSyncSnapshotRef = useRef("");
  const boardSelectionRequestRef = useRef(0);

  useEffect(() => () => noteTimersRef.current.forEach((timer) => clearTimeout(timer)), []);


  async function loadBoard(id, selectionRequest = null) {
    const result = await getWhiteboard(id);
    if (selectionRequest !== null && selectionRequest !== boardSelectionRequestRef.current) return;
    lastSyncSnapshotRef.current = JSON.stringify(result);
    setBoard(result);
    setConnectionStartID(null);
    setConnectionPreview(null);
    setSaveState(getWhiteboardSaveState(id));
    setMessage(`${result.pins.length} RECORD${result.pins.length === 1 ? "" : "S"} ON THIS BOARD`);
  }

  function selectBoard(id) {
    const selectionRequest = ++boardSelectionRequestRef.current;
    focusBoard?.(id);
    const cached = getCachedWhiteboard(id);
    if (cached) {
      lastSyncSnapshotRef.current = JSON.stringify(cached);
      setBoard(cached);
      setConnectionStartID(null);
      setConnectionPreview(null);
      setSaveState(getWhiteboardSaveState(id));
      setMessage(`${cached.pins.length} RECORD${cached.pins.length === 1 ? "" : "S"} ON THIS BOARD`);
    } else {
      setMessage("OPENING SELECTED BOARD…");
    }
    loadBoard(id, selectionRequest).catch((error) => {
      if (selectionRequest === boardSelectionRequestRef.current) setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED");
    });
  }

  useEffect(() => {
    let active = true;
    getWhiteboards().then(async (results) => {
      if (!active) return;
      setBoards(results);
      const preferred = results.find((item) => item.id === initialBoardId)?.id ?? results[0]?.id;
      if (preferred) {
        const selectionAtStart = boardSelectionRequestRef.current;
        const result = await getWhiteboard(preferred);
        if (active && selectionAtStart === boardSelectionRequestRef.current) {
          lastSyncSnapshotRef.current = JSON.stringify(result);
          setBoard(result);
          setSaveState(getWhiteboardSaveState(preferred));
          setMessage(`${result.pins.length} RECORD${result.pins.length === 1 ? "" : "S"} ON THIS BOARD`);
        }
      } else {
        setMessage("PIN A PERSONNEL FILE OR RECORD TO CREATE YOUR FIRST BOARD");
      }
    }).catch((error) => { if (active) setMessage(error.message || "PERSONAL BOARDS ARE UNAVAILABLE"); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    function handleExternalBoardUpdate(event) {
      if (!board || event.detail?.boardId !== board.id) return;
      if (event.detail.status === "deleted") return;
      if (event.detail.status) setSaveState(event.detail.status);
      const cached = getCachedWhiteboard(board.id);
      if (cached) setBoard((current) => {
        if (!current || current.id !== cached.id) return current;
        const knownPins = new Set(current.pins.map((pin) => pin.id));
        const connectionKey = (edge) => [edge.fromId, edge.toId].sort().join(":");
        const knownConnections = new Set(current.connections.map(connectionKey));
        return {
          ...current,
          title: cached.title ?? current.title,
          lockedAt: cached.lockedAt ?? current.lockedAt,
          pins: [...current.pins, ...cached.pins.filter((pin) => !knownPins.has(pin.id))],
          connections: [...current.connections, ...cached.connections.filter((edge) => !knownConnections.has(connectionKey(edge)))],
        };
      });
    }
    window.addEventListener("noir:whiteboard-updated", handleExternalBoardUpdate);
    return () => window.removeEventListener("noir:whiteboard-updated", handleExternalBoardUpdate);
  }, [board?.id]);

  useEffect(() => {
    if (board) lastSyncSnapshotRef.current = JSON.stringify(board);
  }, [board]);

  useEffect(() => {
    if (!board?.id) return undefined;
    const boardId = board.id;
    let active = true;
    const refreshFromServer = async () => {
      if (document.visibilityState === "hidden" || busy || saveState === "saving" || activeNoteSavesRef.current > 0 || noteTimersRef.current.size > 0 || Object.keys(noteDraftsRef.current).length > 0 || dragRef.current) return;
      try {
        const latest = await getWhiteboard(boardId);
        if (!active) return;
        const snapshot = JSON.stringify(latest);
        if (snapshot === lastSyncSnapshotRef.current) return;
        lastSyncSnapshotRef.current = snapshot;
        setBoard((current) => current?.id === boardId ? latest : current);
        setBoards((previous) => previous.map((item) => item.id === boardId ? { ...item, title: latest.title, lockedAt: latest.lockedAt } : item));
        setMessage(`${latest.pins.length} RECORD${latest.pins.length === 1 ? "" : "S"} · UPDATED FROM ANOTHER WINDOW`);
      } catch { /* Keep the current board visible during temporary network failures. */ }
    };
    const unsubscribe = subscribeWhiteboardUpdates((update) => {
      if (update.boardId === boardId && ["saved", "locked"].includes(update.status)) refreshFromServer();
    });
    const onVisibilityChange = () => { if (document.visibilityState === "visible") refreshFromServer(); };
    if (standalone) {
      window.addEventListener("focus", refreshFromServer);
      document.addEventListener("visibilitychange", onVisibilityChange);
    }
    const timer = standalone ? window.setInterval(refreshFromServer, 5000) : null;
    return () => {
      active = false;
      unsubscribe();
      if (standalone) {
        window.removeEventListener("focus", refreshFromServer);
        document.removeEventListener("visibilitychange", onVisibilityChange);
      }
      if (timer) window.clearInterval(timer);
    };
  }, [board?.id, busy, saveState, standalone]);

  useEffect(() => {
    if (renamingBoardId) renameInputRef.current?.focus();
  }, [renamingBoardId]);

  useEffect(() => {
    const stage = stageRef.current;
    if (!stage || !board) return undefined;
    const measure = () => {
      const bounds = stage.getBoundingClientRect();
      setConnectorLines(board.connections.flatMap((edge) => {
        const from = pinRefs.current.get(edge.fromId);
        const to = pinRefs.current.get(edge.toId);
        if (!from || !to) return [];
        const a = from.getBoundingClientRect();
        const b = to.getBoundingClientRect();
        const x1 = a.left + a.width / 2 - bounds.left;
        const y1 = a.top - bounds.top - 2;
        const x2 = b.left + b.width / 2 - bounds.left;
        const y2 = b.top - bounds.top - 2;
        const deltaX = x2 - x1;
        const deltaY = y2 - y1;
        const length = Math.max(1, Math.hypot(deltaX, deltaY));
        let normalX = -deltaY / length;
        let normalY = deltaX / length;
        if (normalY < 0) { normalX *= -1; normalY *= -1; }
        const slack = Math.min(24, Math.max(8, length * 0.035));
        const midX = Math.round((x1 + x2) / 2 + normalX * slack);
        const midY = Math.round((y1 + y2) / 2 + normalY * slack);
        return [{ id: `${edge.fromId}:${edge.toId}`, points: `${x1},${y1} ${midX},${midY} ${x2},${y2}` }];
      }));
    };
    measure();
    const observer = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(measure);
    observer?.observe(stage);
    pinRefs.current.forEach((pin) => observer?.observe(pin));
    window.addEventListener("resize", measure);
    return () => { observer?.disconnect(); window.removeEventListener("resize", measure); };
  }, [board]);

  function addBoard() {
    const draft = createWhiteboardDraft(boards.length);
    const previousBoard = board;
    cacheWhiteboardDraft(draft);
    setBoards((previous) => [...previous.filter((item) => item.id !== draft.id), draft]);
    setBoard(draft);
    setConnectionStartID(null);
    setConnectionPreview(null);
    setSaveState("saving");
    setMessage("0 RECORDS ON THIS BOARD");
    focusBoard?.(draft.id);
    createWhiteboard(draft.title, draft.id).then((created) => {
      setBoards((previous) => previous.map((item) => item.id === draft.id ? created : item));
      setBoard((previous) => previous?.id === draft.id ? created : previous);
    }).catch((error) => {
      discardWhiteboardDraft(draft.id);
      setBoards((previous) => previous.filter((item) => item.id !== draft.id));
      setBoard((previous) => previous?.id === draft.id ? previousBoard : previous);
      setSaveState("save");
      setMessage(error.message || "A NEW WHITEBOARD COULD NOT BE CREATED");
    });
  }

  function startRename(item) {
    if (item.lockedAt) return;
    setRenamingBoardId(item.id);
    setRenameValue(item.title);
    if (item.id !== board?.id) loadBoard(item.id).catch((error) => setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED"));
  }

  function finishRename() {
    const boardID = renamingBoardId;
    if (boards.find((item) => item.id === boardID)?.lockedAt) { setRenamingBoardId(null); return; }
    const previousTitle = boards.find((item) => item.id === boardID)?.title;
    const title = renameValue.trim().slice(0, 64);
    setRenamingBoardId(null);
    if (!boardID || !title || title === previousTitle) { setRenameValue(previousTitle ?? ""); return; }
    setBoards((previous) => previous.map((item) => item.id === boardID ? { ...item, title } : item));
    setBoard((previous) => previous?.id === boardID ? { ...previous, title } : previous);
    setSaveState("saving");
    renameWhiteboard(boardID, title).then(() => setSaveState("saved")).catch((error) => {
      setBoards((previous) => previous.map((item) => item.id === boardID ? { ...item, title: previousTitle } : item));
      setBoard((previous) => previous?.id === boardID ? { ...previous, title: previousTitle } : previous);
      setSaveState("save");
      setMessage(error.message || "THE BOARD TITLE COULD NOT BE SAVED");
    });
  }

  async function createBlankNote() {
    if (!board || boardLocked || busy || saveState === "saving") return;
    const boardID = board.id;
    const index = board.pins.length;
    const x = 700 + (index % 3) * 3000;
    const y = 700 + Math.floor(index / 3) * 2500;
    const nonce = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}_${Math.random().toString(36).slice(2)}`;
    const pendingID = `note:pending-${nonce}`;
    const pendingPin = { id: pendingID, source: "note", title: "UNTITLED NOTE", content: "", x, y, pending: true };
    const initialDraft = { title: pendingPin.title, content: pendingPin.content };
    setBoard((previous) => previous?.id === boardID ? { ...previous, pins: [...previous.pins, pendingPin] } : previous);
    noteDraftsRef.current = { ...noteDraftsRef.current, [pendingID]: initialDraft };
    setNoteDrafts((previous) => ({ ...previous, [pendingID]: initialDraft }));
    setBusy(true);
    setSaveState("saving");
    try {
      const pin = await createWhiteboardNote(boardID, x, y, pendingID);
      setBoard((previous) => previous?.id === boardID
        ? { ...previous, pins: previous.pins.map((item) => item.id === pendingID ? pin : item) }
        : previous);
      noteDraftsRef.current = { ...noteDraftsRef.current, [pin.id]: initialDraft };
      delete noteDraftsRef.current[pendingID];
      setNoteDrafts((previous) => {
        const next = { ...previous, [pin.id]: initialDraft };
        delete next[pendingID];
        return next;
      });
      setMessage("BLANK NOTE ADDED · WRITE YOUR OWN RECORD");
    } catch (error) {
      setBoard((previous) => previous?.id === boardID
        ? { ...previous, pins: previous.pins.filter((item) => item.id !== pendingID) }
        : previous);
      delete noteDraftsRef.current[pendingID];
      setNoteDrafts((previous) => { const next = { ...previous }; delete next[pendingID]; return next; });
      setSaveState("save");
      setMessage(error.message || "A BLANK NOTE COULD NOT BE ADDED");
    } finally {
      setBusy(false);
    }
  }

  function updateNoteDraft(pin, field, value) {
    if (boardLocked) return;
    const draft = { title: pin.title, content: pin.content, ...noteDraftsRef.current[pin.id], [field]: value };
    noteDraftsRef.current = { ...noteDraftsRef.current, [pin.id]: draft };
    setNoteDrafts((previous) => ({ ...previous, [pin.id]: draft }));
    const version = (noteVersionsRef.current.get(pin.id) ?? 0) + 1;
    noteVersionsRef.current.set(pin.id, version);
    clearTimeout(noteTimersRef.current.get(pin.id));
    const boardID = board?.id;
    noteTimersRef.current.set(pin.id, setTimeout(() => {
      noteTimersRef.current.delete(pin.id);
      saveNote(pin, version, boardID);
    }, 800));
  }

  async function saveNote(pin, version = noteVersionsRef.current.get(pin.id) ?? 0, boardID = board?.id) {
    if (!boardID || boardLocked) return;
    const draft = noteDraftsRef.current[pin.id] ?? noteDrafts[pin.id] ?? { title: pin.title, content: pin.content };
    const title = "UNTITLED NOTE";
    setSaveState("saving");
    activeNoteSavesRef.current += 1;
    try {
      await updateWhiteboardNote(boardID, pin.id, title, draft.content);
      if (noteVersionsRef.current.get(pin.id) === version) {
        setBoard((previous) => previous?.id === boardID ? { ...previous, pins: previous.pins.map((item) => item.id === pin.id ? { ...item, title, content: draft.content } : item) } : previous);
        delete noteDraftsRef.current[pin.id];
        setNoteDrafts((previous) => { const next = { ...previous }; delete next[pin.id]; return next; });
      }
    } catch (error) {
      setSaveState("save");
      setMessage(error.message || "NOTE COULD NOT BE SAVED · SELECT SAVE TO RETRY");
    } finally {
      activeNoteSavesRef.current -= 1;
      if (activeNoteSavesRef.current === 0 && noteTimersRef.current.size === 0 && Object.keys(noteDraftsRef.current).length === 0) {
        setSaveState("saved");
      }
    }
  }

  function retryNoteSaves() {
    if (!board || boardLocked) return;
    const pending = board.pins.filter((pin) => pin.source === "note" && noteDraftsRef.current[pin.id]);
    if (!pending.length) return;
    setSaveState("saving");
    Promise.all(pending.map((pin) => saveNote(pin, noteVersionsRef.current.get(pin.id) ?? 0, board.id))).then(() => {
      const hasUnfinishedNotes = Object.keys(noteDraftsRef.current).length > 0
        || noteTimersRef.current.size > 0
        || activeNoteSavesRef.current > 0;
      setSaveState(hasUnfinishedNotes ? "save" : "saved");
    });
  }

  async function confirmDeleteBoard() {
    if (!board || boardLocked || busy) return;
    const deletedId = board.id;
    const previousBoard = board;
    const previousBoards = boards;
    const remaining = boards.filter((item) => item.id !== deletedId);
    const nextBoard = remaining.length ? getCachedWhiteboard(remaining[0].id) : null;
    setBoards(remaining);
    setBoard(nextBoard);
    setConnectionStartID(null);
    setConnectionPreview(null);
    setConfirmDelete(false);
    setMessage(remaining.length
      ? nextBoard ? `${nextBoard.pins.length} RECORD${nextBoard.pins.length === 1 ? "" : "S"} ON THIS BOARD` : "BOARD DELETED · OPENING NEXT BOARD…"
      : "BOARD DELETED · CREATE A NOTE OR PIN A RECORD TO START AGAIN");
    setSaveState("saved");
    setBusy(true);
    try {
      await deleteWhiteboard(deletedId);
    } catch (error) {
      setBoards(previousBoards);
      setBoard(previousBoard);
      setMessage(error.message || "THE BOARD COULD NOT BE DELETED");
      setSaveState("save");
      return;
    } finally {
      setBusy(false);
    }
    if (remaining.length && !nextBoard) {
      loadBoard(remaining[0].id).catch((error) => setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED"));
    }
  }

  async function removePin(pinID) {
    if (!board || boardLocked || busy) return;
    setBusy(true);
    try {
      await removeWhiteboardPin(board.id, pinID);
      await loadBoard(board.id);
    } catch (error) { setSaveState("save"); setMessage(error.message || "THE PIN COULD NOT BE REMOVED"); }
    finally { setBusy(false); }
  }

  async function changeDesignation(pin, designation) {
    if (!board || boardLocked || busy) return;
    const boardID=board.id;
    setBoard((previous) => previous?.id===boardID ? {...previous,pins:previous.pins.map((item)=>item.id===pin.id?{...item,designation}:((designation==="accused"||designation==="victim")&&item.designation===designation)?{...item,designation:""}:item)} : previous);
    setBusy(true);
    try { await designateWhiteboardPin(boardID,pin.id,designation); await loadBoard(boardID); }
    catch(error) { await loadBoard(boardID).catch(()=>{}); setMessage(error.message||"THE PIN DESIGNATION COULD NOT BE SAVED"); }
    finally { setBusy(false); }
  }

  async function connectPins(fromId, toId) {
    if (!board || boardLocked || busy || fromId === toId) return;
    const previousConnections = board.connections;
    const exists = previousConnections.some((edge) => edge.fromId === fromId && edge.toId === toId || edge.fromId === toId && edge.toId === fromId);
    setConnectionStartID(null);
    if (exists) { setMessage("THESE PINS ARE ALREADY CONNECTED"); return; }
    const optimisticConnections = [...previousConnections, { fromId, toId }];
    setBoard((previous) => previous?.id === board.id ? { ...previous, connections: optimisticConnections } : previous);
    setBusy(true);
    setSaveState("saving");
    try {
      await connectWhiteboardPins(board.id, fromId, toId);
      setSaveState("saved");
    } catch (error) {
      setBoard((previous) => previous?.id === board.id ? { ...previous, connections: previousConnections } : previous);
      setSaveState("save"); setMessage(error.message || "THE RECORDS COULD NOT BE CONNECTED");
    }
    finally { setBusy(false); }
  }

  function activateConnectionHandle(pinID) {
    if (boardLocked || busy) return;
    if (!connectionStartID) {
      setConnectionStartID(pinID);
      setMessage("SELECT ANOTHER PIN OR DRAG THIS TACK TO CONNECT IT");
      return;
    }
    const fromId = connectionStartID;
    setConnectionStartID(null);
    if (fromId !== pinID) connectPins(fromId, pinID);
  }

  function beginConnectionDrag(event, pin) {
    if (event.button !== 0 || boardLocked || busy) return;
    event.stopPropagation();
    const stage = stageRef.current;
    if (!stage) return;
    const stageBounds = stage.getBoundingClientRect();
    const handleBounds = event.currentTarget.getBoundingClientRect();
    const startX = handleBounds.left + handleBounds.width / 2 - stageBounds.left;
    const startY = handleBounds.top + handleBounds.height / 2 - stageBounds.top;
    connectionDragRef.current = {
      pointerId: event.pointerId,
      pinID: pin.id,
      startClientX: event.clientX,
      startClientY: event.clientY,
      startX,
      startY,
      moved: false,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
    setConnectionStartID(null);
    setConnectionPreview({ x1: startX, y1: startY, x2: startX, y2: startY });
    event.preventDefault();
  }

  function moveConnectionDrag(event) {
    const drag = connectionDragRef.current;
    const stage = stageRef.current;
    if (!drag || !stage || drag.pointerId !== event.pointerId) return;
    event.stopPropagation();
    const distance = Math.hypot(event.clientX - drag.startClientX, event.clientY - drag.startClientY);
    if (distance > 5) drag.moved = true;
    const bounds = stage.getBoundingClientRect();
    setConnectionPreview({ x1: drag.startX, y1: drag.startY, x2: event.clientX - bounds.left, y2: event.clientY - bounds.top });
  }

  function endConnectionDrag(event) {
    const drag = connectionDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    event.stopPropagation();
    connectionDragRef.current = null;
    setConnectionPreview(null);
    const moved = drag.moved || Math.hypot(event.clientX - drag.startClientX, event.clientY - drag.startClientY) > 5;
    if (!moved) return;

    suppressConnectionClickRef.current = true;
    window.setTimeout(() => { suppressConnectionClickRef.current = false; }, 0);
    event.preventDefault();
    const target = document.elementFromPoint(event.clientX, event.clientY)?.closest(".personal-whiteboard__pin");
    const targetID = target?.dataset.pinId;
    if (targetID && targetID !== drag.pinID) {
      connectPins(drag.pinID, targetID);
    } else {
      setMessage("DROP THE STRING ON ANOTHER PIN TO CONNECT");
    }
  }

  async function removeConnection(edge) {
    if (!board || boardLocked || busy) return;
    setBusy(true);
    try {
      await disconnectWhiteboardPins(board.id, edge.fromId, edge.toId);
      await loadBoard(board.id);
    } catch (error) { setSaveState("save"); setMessage(error.message || "THE STRING COULD NOT BE REMOVED"); }
    finally { setBusy(false); }
  }

  function beginDrag(event, pin) {
    if (boardLocked) return;
    if (event.target.closest("button, input, textarea, select")) return;
    event.stopPropagation();
    const stage = stageRef.current;
    if (!stage) return;
    dragRef.current = { pointerId: event.pointerId, itemID: pin.id, startX: event.clientX, startY: event.clientY, x: pin.x, y: pin.y };
    event.currentTarget.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function moveDrag(event) {
    const drag = dragRef.current;
    const stage = stageRef.current;
    if (!drag || !stage || drag.pointerId !== event.pointerId) return;
    const bounds = stage.getBoundingClientRect();
    const x = clamp(Math.round(drag.x + (event.clientX - drag.startX) / bounds.width * 10000), 0, 7200);
    const y = clamp(Math.round(drag.y + (event.clientY - drag.startY) / bounds.height * 10000), 0, 8200);
    drag.x = x;
    drag.y = y;
    drag.startX = event.clientX;
    drag.startY = event.clientY;
    setBoard((previous) => previous ? { ...previous, pins: previous.pins.map((pin) => pin.id === drag.itemID ? { ...pin, x, y } : pin) } : previous);
  }

  async function endDrag(event) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId || !board) return;
    dragRef.current = null;
    setSaveState("saving");
    try {
      await moveWhiteboardPin(board.id, drag.itemID, drag.x, drag.y);
      setSaveState("saved");
    } catch (error) { setSaveState("save"); setMessage(error.message || "PIN POSITION COULD NOT BE SAVED"); }
  }

  async function nudgePin(pin, dx, dy) {
    if (!board || boardLocked || busy) return;
    const x = clamp(pin.x + dx, 0, 7200);
    const y = clamp(pin.y + dy, 0, 8200);
    setBoard((previous) => previous ? { ...previous, pins: previous.pins.map((item) => item.id === pin.id ? { ...item, x, y } : item) } : previous);
    setSaveState("saving");
    try { await moveWhiteboardPin(board.id, pin.id, x, y); setSaveState("saved"); }
    catch (error) { setSaveState("save"); setMessage(error.message || "PIN POSITION COULD NOT BE SAVED"); }
  }

  function beginPan(event) {
    if (boardLocked) return;
    const viewport = viewportRef.current;
    if (!viewport || event.button !== 0 || event.target === viewport || event.target.closest(".personal-whiteboard__pin")) return;
    panRef.current = { pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, left: viewport.scrollLeft, top: viewport.scrollTop };
    viewport.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function movePan(event) {
    const pan = panRef.current;
    const viewport = viewportRef.current;
    if (!pan || !viewport || pan.pointerId !== event.pointerId) return;
    viewport.scrollLeft = pan.left - (event.clientX - pan.startX);
    viewport.scrollTop = pan.top - (event.clientY - pan.startY);
  }

  function endPan(event) {
    if (panRef.current?.pointerId === event.pointerId) panRef.current = null;
  }

  return (
    <section className={`personal-whiteboard${standalone ? " personal-whiteboard--standalone" : ""}`} aria-label="Personal investigation whiteboard">
      <header className="personal-whiteboard__header">
        <div><p className="computer-window__eyebrow">NOIR CASEWORK · INVESTIGATION BOARD</p>
          {renamingBoardId === board?.id && !boardLocked
            ? <input ref={renameInputRef} className="personal-whiteboard__title-input" aria-label="Rename investigation board" maxLength={64} value={renameValue} onChange={(event) => setRenameValue(event.target.value)} onBlur={finishRename} onKeyDown={(event) => {
              if (event.key === "Enter") { event.preventDefault(); event.currentTarget.blur(); }
              if (event.key === "Escape") { setRenameValue(board.title); setRenamingBoardId(null); }
            }} />
        : <h1 className="personal-whiteboard__title">{board?.title || "WHITEBOARD"}</h1>}
          {board && !boardLocked && renamingBoardId !== board.id && <button className="personal-whiteboard__rename" type="button" aria-label={`Rename ${board.title}`} title="Rename board" onClick={() => startRename(board)}><Pencil aria-hidden="true" /></button>}
        </div>
        <div className="personal-whiteboard__header-actions">
          <button type="button" className="personal-whiteboard__new" disabled={busy} onClick={addBoard}>+ NEW BOARD</button>
          <button type="button" className="personal-whiteboard__delete-board" disabled={!board || boardLocked || busy} aria-label={`Delete ${board?.title || "board"}`} title="Delete board" onClick={() => setConfirmDelete(true)}><Trash aria-hidden="true" /></button>
        </div>
      </header>
      <nav className="personal-whiteboard__pages" aria-label="Investigation boards">
        {boards.map((item) => <button type="button" key={item.id} aria-current={item.id === board?.id ? "page" : undefined} onClick={() => selectBoard(item.id)}>{item.title}</button>)}
      </nav>
      <p className="personal-whiteboard__status" role="status" aria-live="polite">{boardLocked ? "SUBMITTED FOR REVIEW · THIS BOARD IS LOCKED. CREATE A NEW BOARD TO PREPARE ANOTHER ACCUSATION." : message}</p>
      {!board ? <div className="personal-whiteboard__empty"><span aria-hidden="true">＋</span><p>Create a board, then add a blank note or pin a personnel file, email, or record.</p></div> : <>
        <div className="personal-whiteboard__tools">
          <span>DRAG HEADERS TO ARRANGE · DRAG RED TACKS TO CONNECT</span>
          {saveState === "save" ? <button type="button" className="personal-whiteboard__save personal-whiteboard__save--save" disabled={boardLocked} onClick={retryNoteSaves}>SAVE</button> : <span className={`personal-whiteboard__save personal-whiteboard__save--${saveState}`} role="status" aria-live="polite">{saveState === "saving" ? "SAVING" : "SAVED"}</span>}
          <button type="button" disabled={boardLocked || busy || saveState === "saving"} onClick={createBlankNote}>+ BLANK NOTE</button>
        </div>
        <div className="personal-whiteboard__viewport" ref={viewportRef} onPointerDown={beginPan} onPointerMove={movePan} onPointerUp={endPan} onPointerCancel={endPan}>
        <div className="personal-whiteboard__stage" ref={stageRef}>
          <svg className="personal-whiteboard__strings" aria-hidden="true"><defs><filter id="string-shadow"><feGaussianBlur stdDeviation="1.2" /></filter></defs>
            {connectorLines.map((line) => <g key={line.id}><polyline className="personal-whiteboard__string-shadow" points={line.points} /><polyline points={line.points} /></g>)}
            {connectionPreview && <polyline className="personal-whiteboard__string-preview" points={`${connectionPreview.x1},${connectionPreview.y1} ${connectionPreview.x2},${connectionPreview.y2}`} />}
          </svg>
          {board.pins.map((pin) => {
            const selected = connectionStartID === pin.id;
            return <article
              ref={(node) => { if (node) pinRefs.current.set(pin.id, node); else pinRefs.current.delete(pin.id); }}
              className={`personal-whiteboard__pin${selected ? " personal-whiteboard__pin--selected" : ""}${pin.source === "note" ? " personal-whiteboard__pin--note" : ""}`}
              data-pin-id={pin.id}
              key={pin.id}
              style={{ left: `${pin.x / 100}%`, top: `${pin.y / 100}%` }}
              tabIndex={0}
              onKeyDown={(event) => {
                const directions = { ArrowLeft: [-200, 0], ArrowRight: [200, 0], ArrowUp: [0, -300], ArrowDown: [0, 300] };
                if (directions[event.key]) { event.preventDefault(); nudgePin(pin, ...directions[event.key]); }
              }}
            >
              <button
                className="personal-whiteboard__tack"
                type="button"
                aria-label={`Connect ${pin.title} to another pin`}
                aria-pressed={selected}
                title="Drag to another pin to connect; activate twice with the keyboard"
                disabled={boardLocked || busy}
                onClick={() => {
                  if (suppressConnectionClickRef.current) { suppressConnectionClickRef.current = false; return; }
                  activateConnectionHandle(pin.id);
                }}
                onPointerDown={(event) => beginConnectionDrag(event, pin)}
                onPointerMove={moveConnectionDrag}
                onPointerUp={endConnectionDrag}
                onPointerCancel={endConnectionDrag}
              />
              {pin.source === "note" ? <>
                <header className="personal-whiteboard__pin-drag-handle" onPointerDown={(event) => beginDrag(event, pin)} onPointerMove={moveDrag} onPointerUp={endDrag} onPointerCancel={endDrag}><span>NOTE - PINNED</span><button className="personal-whiteboard__remove-pin" type="button" aria-label="Remove note" disabled={boardLocked || busy} onClick={() => removePin(pin.id)}><Trash aria-hidden="true" /></button></header>
                <textarea
                  className="personal-whiteboard__note-content"
                  aria-label="Note text"
                  maxLength={5000}
                  spellCheck={false}
                  value={(noteDrafts[pin.id] ?? pin).content}
                  disabled={boardLocked || pin.pending}
                  onChange={(event) => updateNoteDraft(pin, "content", event.target.value)}
                />
              </> : <>
                <header className="personal-whiteboard__pin-drag-handle" onPointerDown={(event) => beginDrag(event, pin)} onPointerMove={moveDrag} onPointerUp={endDrag} onPointerCancel={endDrag}><span>{pin.source.toUpperCase()} · PINNED</span><button className="personal-whiteboard__remove-pin" type="button" aria-label={`Remove ${pin.title} from board`} disabled={boardLocked || busy} onClick={() => removePin(pin.id)}><Trash aria-hidden="true" /></button></header>
                {pin.metadata?.npc?.visual && <div className="personal-whiteboard__person-portrait"><ModularFace visual={pin.metadata.npc.visual} name={pin.metadata.npc.name || pin.title} /></div>}
                <h2>{pin.title}</h2><p>{displayPinContent(pin)}</p>
                <label className="personal-whiteboard__designation">MARK PIN
                  <select aria-label={`Mark ${pin.title}`} value={pin.designation||""} disabled={boardLocked || busy} onChange={(event)=>changeDesignation(pin,event.target.value)}>
                    <option value="">UNMARKED</option><option value="proof">PROOF</option>
                    {pin.source==="person"&&<><option value="accused">ACCUSED</option><option value="victim">VICTIM</option></>}
                  </select>
                </label>
              </>}
            </article>;
          })}
        </div>
        </div>
        {board.connections.length > 0 && <section className="personal-whiteboard__links" aria-label="Board strings">
          <h2>CONNECTED RECORDS</h2>
          {board.connections.map((edge) => {
            const from = board.pins.find((pin) => pin.id === edge.fromId);
            const to = board.pins.find((pin) => pin.id === edge.toId);
            return <div key={`${edge.fromId}-${edge.toId}`}><span>{from?.title} <b aria-hidden="true">↔</b> {to?.title}</span><button type="button" disabled={boardLocked || busy} onClick={() => removeConnection(edge)}>REMOVE STRING</button></div>;
          })}
        </section>}
      </>}
      {confirmDelete && board && <div className="personal-whiteboard__confirm-backdrop">
        <section className="personal-whiteboard__confirm" role="alertdialog" aria-modal="true" aria-labelledby="whiteboard-delete-title" aria-describedby="whiteboard-delete-description">
          <p className="computer-window__eyebrow">PERMANENT ACTION</p>
          <h2 id="whiteboard-delete-title">DELETE {board.title}?</h2>
          <p id="whiteboard-delete-description">This will permanently delete this board, all pinned records, and their connections.</p>
          <div>
            <button type="button" disabled={busy} onClick={() => setConfirmDelete(false)}>CANCEL</button>
            <button type="button" className="personal-whiteboard__confirm-delete" disabled={busy} onClick={confirmDeleteBoard}>{busy ? "DELETING…" : "DELETE BOARD"}</button>
          </div>
        </section>
      </div>}
    </section>
  );
}
