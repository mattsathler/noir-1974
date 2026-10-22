import { useEffect, useRef, useState } from "react";
import { Pencil } from "pixelarticons/react/Pencil";
import { Trash } from "pixelarticons/react/Trash";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import {
  connectWhiteboardPins,
  cacheWhiteboardDraft,
  createWhiteboard,
  createWhiteboardDraft,
  createWhiteboardNote,
  deleteWhiteboard,
  discardWhiteboardDraft,
  disconnectWhiteboardPins,
  getWhiteboardSaveState,
  getWhiteboard,
  getWhiteboards,
  moveWhiteboardPin,
  renameWhiteboard,
  removeWhiteboardPin,
  updateWhiteboardNote,
} from "../../../world/api.js";

function clamp(value, min, max) { return Math.max(min, Math.min(max, value)); }

export default function Whiteboard({ initialBoardId, focusBoard }) {
  const [boards, setBoards] = useState([]);
  const [board, setBoard] = useState(null);
  const [selection, setSelection] = useState([]);
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
  const dragRef = useRef(null);
  const panRef = useRef(null);
  const renameInputRef = useRef(null);
  const noteDraftsRef = useRef({});
  const noteTimersRef = useRef(new Map());
  const noteVersionsRef = useRef(new Map());
  const activeNoteSavesRef = useRef(0);

  useEffect(() => () => noteTimersRef.current.forEach((timer) => clearTimeout(timer)), []);

  async function loadBoard(id) {
    const result = await getWhiteboard(id);
    setBoard(result);
    setSelection([]);
    setSaveState(getWhiteboardSaveState(id));
    setMessage(`${result.pins.length} RECORD${result.pins.length === 1 ? "" : "S"} ON THIS BOARD`);
  }

  useEffect(() => {
    let active = true;
    getWhiteboards().then(async (results) => {
      if (!active) return;
      setBoards(results);
      const preferred = results.find((item) => item.id === initialBoardId)?.id ?? results[0]?.id;
      if (preferred) {
        const result = await getWhiteboard(preferred);
        if (active) {
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
    if (!initialBoardId || !boards.some((item) => item.id === initialBoardId) || board?.id === initialBoardId) return;
    loadBoard(initialBoardId).catch((error) => setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED"));
  }, [initialBoardId, boards, board?.id]);

  useEffect(() => {
    function handleExternalBoardUpdate(event) {
      if (!board || event.detail?.boardId !== board.id) return;
      if (event.detail.status) setSaveState(event.detail.status);
      getWhiteboard(board.id).then(setBoard).catch(() => setSaveState("save"));
    }
    window.addEventListener("noir:whiteboard-updated", handleExternalBoardUpdate);
    return () => window.removeEventListener("noir:whiteboard-updated", handleExternalBoardUpdate);
  }, [board?.id]);

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
    setSelection([]);
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
    setRenamingBoardId(item.id);
    setRenameValue(item.title);
    if (item.id !== board?.id) loadBoard(item.id).catch((error) => setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED"));
  }

  function finishRename() {
    const boardID = renamingBoardId;
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
    if (!board || busy || saveState === "saving") return;
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
    if (!boardID) return;
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
    if (!board) return;
    const pending = board.pins.filter((pin) => pin.source === "note" && noteDraftsRef.current[pin.id]);
    if (!pending.length) return;
    setSaveState("saving");
    Promise.all(pending.map((pin) => saveNote(pin, noteVersionsRef.current.get(pin.id) ?? 0, board.id))).then(() => {
      if (pending.every((pin) => !noteDraftsRef.current[pin.id])) setSaveState("saved");
    });
  }

  async function confirmDeleteBoard() {
    if (!board || busy) return;
    const deletedId = board.id;
    setBusy(true);
    try {
      await deleteWhiteboard(deletedId);
      const remaining = boards.filter((item) => item.id !== deletedId);
      setBoards(remaining);
      setConfirmDelete(false);
      setSelection([]);
      if (remaining.length) {
        await loadBoard(remaining[0].id);
      } else {
        setBoard(null);
        setMessage("BOARD DELETED · CREATE A NOTE OR PIN A RECORD TO START AGAIN");
        setSaveState("saved");
      }
    } catch (error) {
      setMessage(error.message || "THE BOARD COULD NOT BE DELETED");
    } finally {
      setBusy(false);
    }
  }

  async function removePin(pinID) {
    if (!board || busy) return;
    setBusy(true);
    try {
      await removeWhiteboardPin(board.id, pinID);
      await loadBoard(board.id);
    } catch (error) { setSaveState("save"); setMessage(error.message || "THE PIN COULD NOT BE REMOVED"); }
    finally { setBusy(false); }
  }

  async function connectSelected() {
    if (!board || busy || selection.length !== 2) return;
    const [fromId, toId] = selection;
    const previousConnections = board.connections;
    const exists = previousConnections.some((edge) => edge.fromId === fromId && edge.toId === toId || edge.fromId === toId && edge.toId === fromId);
    if (exists) { setSelection([]); return; }
    const optimisticConnections = [...previousConnections, { fromId, toId }];
    setBoard((previous) => previous?.id === board.id ? { ...previous, connections: optimisticConnections } : previous);
    setBusy(true);
    setSaveState("saving");
    try {
      await connectWhiteboardPins(board.id, fromId, toId);
      setSelection([]);
      setSaveState("saved");
    } catch (error) {
      setBoard((previous) => previous?.id === board.id ? { ...previous, connections: previousConnections } : previous);
      setSaveState("save"); setMessage(error.message || "THE RECORDS COULD NOT BE CONNECTED");
    }
    finally { setBusy(false); }
  }

  async function removeConnection(edge) {
    if (!board || busy) return;
    setBusy(true);
    try {
      await disconnectWhiteboardPins(board.id, edge.fromId, edge.toId);
      await loadBoard(board.id);
    } catch (error) { setSaveState("save"); setMessage(error.message || "THE STRING COULD NOT BE REMOVED"); }
    finally { setBusy(false); }
  }

  function beginDrag(event, pin) {
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
    if (!board || busy) return;
    const x = clamp(pin.x + dx, 0, 7200);
    const y = clamp(pin.y + dy, 0, 8200);
    setBoard((previous) => previous ? { ...previous, pins: previous.pins.map((item) => item.id === pin.id ? { ...item, x, y } : item) } : previous);
    setSaveState("saving");
    try { await moveWhiteboardPin(board.id, pin.id, x, y); setSaveState("saved"); }
    catch (error) { setSaveState("save"); setMessage(error.message || "PIN POSITION COULD NOT BE SAVED"); }
  }

  function beginPan(event) {
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
    <section className="personal-whiteboard" aria-label="Personal investigation whiteboard">
      <header className="personal-whiteboard__header">
        <div><p className="computer-window__eyebrow">PRIVATE INVESTIGATION · PERSONAL RECORD</p>
          {renamingBoardId === board?.id
            ? <input ref={renameInputRef} className="personal-whiteboard__title-input" aria-label="Rename investigation board" maxLength={64} value={renameValue} onChange={(event) => setRenameValue(event.target.value)} onBlur={finishRename} onKeyDown={(event) => {
              if (event.key === "Enter") { event.preventDefault(); event.currentTarget.blur(); }
              if (event.key === "Escape") { setRenameValue(board.title); setRenamingBoardId(null); }
            }} />
            : <h1 className="personal-whiteboard__title">{board?.title || "WHITEBOARD"}</h1>}
          {board && renamingBoardId !== board.id && <button className="personal-whiteboard__rename" type="button" aria-label={`Rename ${board.title}`} title="Rename board" onClick={() => startRename(board)}><Pencil aria-hidden="true" /></button>}
        </div>
        <div className="personal-whiteboard__header-actions">
          <button type="button" className="personal-whiteboard__new" disabled={busy} onClick={addBoard}>+ NEW BOARD</button>
          <button type="button" className="personal-whiteboard__delete-board" disabled={!board || busy} aria-label={`Delete ${board?.title || "board"}`} title="Delete board" onClick={() => setConfirmDelete(true)}><Trash aria-hidden="true" /></button>
        </div>
      </header>
      <nav className="personal-whiteboard__pages" aria-label="Investigation boards">
        {boards.map((item) => <button type="button" key={item.id} aria-current={item.id === board?.id ? "page" : undefined} onClick={() => loadBoard(item.id).catch((error) => setMessage(error.message || "THE SELECTED BOARD COULD NOT BE OPENED"))}>{item.title}</button>)}
      </nav>
      <p className="personal-whiteboard__status" role="status" aria-live="polite">{message}</p>
      {!board ? <div className="personal-whiteboard__empty"><span aria-hidden="true">＋</span><p>Create a board, then add a blank note or pin a personnel file, email, or record.</p></div> : <>
        <div className="personal-whiteboard__tools">
          <span>DRAG PINS TO ARRANGE · SELECT TWO PINS TO CONNECT</span>
          {saveState === "save" ? <button type="button" className="personal-whiteboard__save personal-whiteboard__save--save" onClick={retryNoteSaves}>SAVE</button> : <span className={`personal-whiteboard__save personal-whiteboard__save--${saveState}`} role="status" aria-live="polite">{saveState === "saving" ? "SAVING" : "SAVED"}</span>}
          <button type="button" disabled={busy || saveState === "saving"} onClick={createBlankNote}>+ BLANK NOTE</button>
          {selection.length === 2 && <button type="button" disabled={busy} onClick={connectSelected}>CONNECT SELECTED</button>}
        </div>
        <div className="personal-whiteboard__viewport" ref={viewportRef} onPointerDown={beginPan} onPointerMove={movePan} onPointerUp={endPan} onPointerCancel={endPan}>
        <div className="personal-whiteboard__stage" ref={stageRef}>
          <svg className="personal-whiteboard__strings" aria-hidden="true"><defs><filter id="string-shadow"><feGaussianBlur stdDeviation="1.2" /></filter></defs>
            {connectorLines.map((line) => <g key={line.id}><polyline className="personal-whiteboard__string-shadow" points={line.points} /><polyline points={line.points} /></g>)}
          </svg>
          {board.pins.map((pin) => {
            const selected = selection.includes(pin.id);
            return <article
              ref={(node) => { if (node) pinRefs.current.set(pin.id, node); else pinRefs.current.delete(pin.id); }}
              className={`personal-whiteboard__pin${selected ? " personal-whiteboard__pin--selected" : ""}${pin.source === "note" ? " personal-whiteboard__pin--note" : ""}`}
              key={pin.id}
              style={{ left: `${pin.x / 100}%`, top: `${pin.y / 100}%` }}
              tabIndex={0}
              onPointerDown={(event) => beginDrag(event, pin)}
              onPointerMove={moveDrag}
              onPointerUp={endDrag}
              onPointerCancel={endDrag}
              onKeyDown={(event) => {
                const directions = { ArrowLeft: [-200, 0], ArrowRight: [200, 0], ArrowUp: [0, -300], ArrowDown: [0, 300] };
                if (directions[event.key]) { event.preventDefault(); nudgePin(pin, ...directions[event.key]); }
              }}
            >
              <span className="personal-whiteboard__tack" aria-hidden="true" />
              {pin.source === "note" ? <>
                <header><span>NOTE - PINNED</span><button className="personal-whiteboard__remove-pin" type="button" aria-label="Remove note" disabled={busy} onClick={() => removePin(pin.id)}><Trash aria-hidden="true" /></button></header>
                <textarea
                  className="personal-whiteboard__note-content"
                  aria-label="Note text"
                  maxLength={5000}
                  value={(noteDrafts[pin.id] ?? pin).content}
                  disabled={pin.pending}
                  onChange={(event) => updateNoteDraft(pin, "content", event.target.value)}
                />
              </> : <>
                <header><span>{pin.source.toUpperCase()} · PINNED</span><button className="personal-whiteboard__remove-pin" type="button" aria-label={`Remove ${pin.title} from board`} disabled={busy} onClick={() => removePin(pin.id)}><Trash aria-hidden="true" /></button></header>
                {pin.source === "person" && pin.metadata?.npc?.visual && <div className="personal-whiteboard__person-portrait"><ModularFace visual={pin.metadata.npc.visual} name={pin.metadata.npc.name || pin.title} /></div>}
                <h2>{pin.title}</h2><p>{pin.content}</p>
              </>}
              <button className="personal-whiteboard__connect" type="button" aria-pressed={selected} onClick={() => setSelection((previous) => selected ? previous.filter((id) => id !== pin.id) : previous.length < 2 ? [...previous, pin.id] : [previous[1], pin.id])}>{selected ? "SELECTED" : "SELECT TO CONNECT"}</button>
            </article>;
          })}
        </div>
        </div>
        {board.connections.length > 0 && <section className="personal-whiteboard__links" aria-label="Board strings">
          <h2>CONNECTED RECORDS</h2>
          {board.connections.map((edge) => {
            const from = board.pins.find((pin) => pin.id === edge.fromId);
            const to = board.pins.find((pin) => pin.id === edge.toId);
            return <div key={`${edge.fromId}-${edge.toId}`}><span>{from?.title} <b aria-hidden="true">↔</b> {to?.title}</span><button type="button" disabled={busy} onClick={() => removeConnection(edge)}>REMOVE STRING</button></div>;
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
