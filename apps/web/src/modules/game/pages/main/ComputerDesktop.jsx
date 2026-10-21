import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { Building } from "pixelarticons/react/Building";
import { Files } from "pixelarticons/react/Files";
import { Map as MapIcon } from "pixelarticons/react/Map";
import { Mail } from "pixelarticons/react/Mail";
import { Notes } from "pixelarticons/react/Notes";
import { Phone } from "pixelarticons/react/Phone";
import { Globe } from "pixelarticons/react/Globe";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";
import { cachePinnedWhiteboardItem, cacheWhiteboardDraft, createWhiteboard, createWhiteboardDraft, getCachedWhiteboard, getWhiteboards, pinWhiteboardItem, setWhiteboardSaveState } from "../../../world/api.js";
import CaseFiles from "../../../cases/pages/files/CaseFiles.jsx";
import Whiteboard from "../../../cases/pages/files/Whiteboard.jsx";
import WorldDirectory from "../../../world/pages/directory/WorldDirectory.jsx";
import EmailInbox from "../../../mail/pages/inbox/EmailInbox.jsx";

const BOOT_LINES = [
  "NOIR PERSONAL TERMINAL · 1974",
  "CHECKING RELAY CONNECTION…",
  "LYON NODE FOUND",
  "PRIVATE SYSTEM READY",
];

const programs = [
  { id: "interpol", label: "PERSON SEARCH", Icon: Globe, state: "ONLINE", available: true },
  { id: "case-files", label: "CASE FILES", Icon: Files, state: "UNDER REVIEW", available: false },
  { id: "whiteboard", label: "WHITEBOARD", Icon: Notes, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: MapIcon, state: "ONLINE", available: true },
  { id: "bank-records", label: "MERIDIAN BANK", Icon: Building, state: "RESTRICTED", available: true },
  { id: "gate-records", label: "GATE NETWORK", Icon: Globe, state: "RESTRICTED", available: true },
  { id: "email", label: "MERIDIAN MAIL", Icon: Mail, state: "RESTRICTED", available: true },
  { id: "telephone", label: "TELEPHONE", Icon: Phone, state: "OFFLINE", available: false },
];

export default function ComputerDesktop({ location, display, serverTime, timeZone, onClose }) {
  const [windows, setWindows] = useState(() => {
    try {
      const saved = JSON.parse(sessionStorage.getItem("noir:computer:windows") || "null");
      if (!saved || Date.now() - saved.savedAt > 5 * 60 * 1000) return [];
      return saved.windows.flatMap((item) => {
        const program = programs.find((entry) => entry.id === item.programId);
        return program?.available ? [{ ...item, program }] : [];
      });
    } catch { return []; }
  });
  const [booting, setBooting] = useState(true);
  const [hacking, setHacking] = useState(false);
  const [pinDialog, setPinDialog] = useState(null);
  const [preferredBoardId, setPreferredBoardId] = useState(null);
  const modalRef = useRef(null);
  const closeButtonRef = useRef(null);
  const windowRefs = useRef(new Map());
  const dragRef = useRef(null);
  const bootTimerRef = useRef(null);
  const nextZIndexRef = useRef(Math.max(5, ...windows.map((item) => item.zIndex + 1)));
  const hackingRequestsRef = useRef(0);
  const windowsStateRef = useRef(windows);
  const pinDialogRef = useRef(pinDialog);
  windowsStateRef.current = windows;
  pinDialogRef.current = pinDialog;

  useEffect(() => {
    return () => window.clearTimeout(bootTimerRef.current);
  }, []);

  useEffect(() => {
    try {
      sessionStorage.setItem("noir:computer:windows", JSON.stringify({
        savedAt: Date.now(),
        windows: windows.map(({ id, program, boardId, zIndex, position }) => ({ id, programId: program.id, boardId, zIndex, position })),
      }));
    } catch { /* Storage may be disabled by the browser. */ }
  }, [windows]);

  useLayoutEffect(() => {
    function fitWindows() {
      let adjusted = false;
      const measurements = new Map();
      windows.forEach((item) => {
        const element = windowRefs.current.get(item.id);
        const parent = element?.parentElement;
        if (!element || !parent) return;
        element.style.maxHeight = `${Math.max(0, parent.clientHeight - 32)}px`;
        const parentBounds = parent.getBoundingClientRect();
        const bounds = element.getBoundingClientRect();
        const maxLeft = Math.max(0, parent.clientWidth - element.offsetWidth);
        const maxTop = Math.max(0, parent.clientHeight - element.offsetHeight);
        const left = Math.round(Math.max(0, Math.min(maxLeft, bounds.left - parentBounds.left)));
        const top = Math.round(Math.max(0, Math.min(maxTop, bounds.top - parentBounds.top)));
        if (!item.position || Math.abs(item.position.left - left) > 1 || Math.abs(item.position.top - top) > 1) {
          measurements.set(item.id, { left, top });
          adjusted = true;
        }
      });
      if (adjusted) {
        setWindows((previous) => previous.map((item) => measurements.has(item.id) ? { ...item, position: measurements.get(item.id) } : item));
      }
    }

    fitWindows();
    window.addEventListener("resize", fitWindows);
    return () => window.removeEventListener("resize", fitWindows);
  }, [windows]);

  function finishBoot() {
    bootTimerRef.current = window.setTimeout(() => setBooting(false), 650);
  }

  function persistWindows() {
    try {
      sessionStorage.setItem("noir:computer:windows", JSON.stringify({
        savedAt: Date.now(),
        windows: windowsStateRef.current.map(({ id, program, boardId, zIndex, position }) => ({ id, programId: program.id, boardId, zIndex, position })),
      }));
    } catch { /* Storage may be disabled by the browser. */ }
  }

  function exitComputer() {
    persistWindows();
    onClose?.();
  }

  const runHack = useCallback(async (loadRequest) => {
    hackingRequestsRef.current += 1;
    setHacking(true);
    try {
      const request = typeof loadRequest === "function" ? loadRequest() : Promise.resolve(loadRequest);
      const [result] = await Promise.all([
        request,
        new Promise((resolve) => window.setTimeout(resolve, 900)),
      ]);
      return result;
    } finally {
      hackingRequestsRef.current = Math.max(0, hackingRequestsRef.current - 1);
      setHacking(hackingRequestsRef.current > 0);
    }
  }, []);

  function startWindowDrag(event) {
    if (event.target.closest("button")) return;
    const parent = event.currentTarget.closest(".computer-desktop__content");
    const section = event.currentTarget.closest(".computer-window");
    const windowId = section?.dataset.windowId;
    const modal = windowRefs.current.get(windowId);
    if (!parent || !modal) return;
    const bounds = parent.getBoundingClientRect();
    const rect = modal.getBoundingClientRect();
    const left = rect.left - bounds.left;
    const top = rect.top - bounds.top;
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, position: { left, top } } : item));
    dragRef.current = { windowId, pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, left, top };
    focusWindow(windowId);
    event.currentTarget.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function moveWindow(event) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const modal = windowRefs.current.get(drag.windowId);
    const parent = modal?.parentElement;
    if (!parent || !modal) return;
    const maxLeft = Math.max(0, parent.clientWidth - modal.offsetWidth);
    const maxTop = Math.max(0, parent.clientHeight - modal.offsetHeight);
    const position = {
      left: Math.max(0, Math.min(maxLeft, drag.left + event.clientX - drag.startX)),
      top: Math.max(0, Math.min(maxTop, drag.top + event.clientY - drag.startY)),
    };
    setWindows((previous) => previous.map((item) => item.id === drag.windowId ? { ...item, position } : item));
  }

  function endWindowDrag(event) {
    if (dragRef.current?.pointerId === event.pointerId) dragRef.current = null;
  }

  useEffect(() => {
    const previousFocus = document.activeElement;
    closeButtonRef.current?.focus();

    function handleModalKeys(event) {
      if (event.key === "Escape") {
        if (pinDialogRef.current) { setPinDialog(null); return; }
        const topWindow = [...windowsStateRef.current].sort((a, b) => b.zIndex - a.zIndex)[0];
        if (topWindow) closeWindow(topWindow.id);
        else exitComputer();
        return;
      }
      if (event.key !== "Tab" || !modalRef.current) return;

      const focusable = modalRef.current.querySelectorAll("button:not(:disabled), a[href], input:not(:disabled), [tabindex]:not([tabindex='-1'])");
      if (!focusable.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
    window.addEventListener("keydown", handleModalKeys);
    return () => {
      window.removeEventListener("keydown", handleModalKeys);
      if (typeof previousFocus?.focus === "function") previousFocus.focus();
    };
  }, [onClose]);

  function focusWindow(windowId) {
    const zIndex = nextZIndexRef.current++;
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, zIndex } : item));
  }

  function launchProgram(program, boardId = null) {
    if (!program.available) return;
    if (boardId) setPreferredBoardId(boardId);
    const existing = windows.find((item) => item.program.id === program.id);
    if (existing) { focusWindow(existing.id); return; }
    const zIndex = nextZIndexRef.current++;
    const id = `${program.id}:${zIndex}`;
    const offset = windows.length;
    const compact = window.matchMedia("(max-width: 42rem)").matches;
    const position = program.id === "whiteboard" || offset === 0 ? null : {
      left: Math.min(compact ? 8 + offset * 10 : 40 + offset * 28, compact ? 32 : 140),
      top: Math.min(compact ? 8 + offset * 10 : 32 + offset * 24, compact ? 32 : 120),
    };
    setWindows((previous) => [...previous, { id, program, boardId, zIndex, position }]);
  }

  function closeWindow(windowId) {
    setWindows((previous) => previous.filter((item) => item.id !== windowId));
    windowRefs.current.delete(windowId);
  }

  function finishPin(boardId, item, board = null) {
    const destination = getCachedWhiteboard(boardId) ?? board ?? pinDialogRef.current?.boards.find((candidate) => candidate.id === boardId) ?? { id: boardId, pins: [] };
    const index = destination.pins?.length ?? 0;
    const x = 600 + (index % 3) * 3000;
    const y = 700 + (Math.floor(index / 3) % 3) * 2500;
    cachePinnedWhiteboardItem(boardId, item, x, y, destination);
    pinWhiteboardItem(boardId, item.id, x, y)
      .then(() => setWhiteboardSaveState(boardId, "saved"))
      .catch(() => setWhiteboardSaveState(boardId, "save"));
    setPinDialog(null);
    launchProgram(programs.find((item) => item.id === "whiteboard"), boardId);
  }

  async function requestPin(item) {
    let boards = [];
    try {
      boards = await getWhiteboards();
      if (!boards.length) {
        const draft = createWhiteboardDraft(0);
        cacheWhiteboardDraft(draft);
        setWhiteboardSaveState(draft.id, "saving");
        setPinDialog(null);
        launchProgram(programs.find((program) => program.id === "whiteboard"), draft.id);
        createWhiteboard(draft.title, draft.id)
          .then((created) => finishPin(draft.id, item, created))
          .catch(() => setWhiteboardSaveState(draft.id, "save"));
        return;
      }
      setPinDialog({ item, boards, busy: false, error: "" });
    } catch (error) {
      const availableBoards = await getWhiteboards().catch(() => boards);
      setPinDialog({ item, boards: availableBoards, busy: false, error: error.message || "COULD NOT OPEN PERSONAL BOARDS" });
    }
  }

  async function choosePinDestination(boardId) {
    if (!pinDialog) return;
    setPinDialog((previous) => ({ ...previous, busy: true, error: "" }));
    try {
      finishPin(boardId, pinDialog.item);
    } catch (error) {
      setPinDialog((previous) => previous ? { ...previous, busy: false, error: error.message || "COULD NOT PIN THIS RECORD" } : previous);
    }
  }

  async function createAndChooseBoard() {
    if (!pinDialog) return;
    const draft = createWhiteboardDraft(pinDialog.boards.length);
    const item = pinDialog.item;
    cacheWhiteboardDraft(draft);
    setWhiteboardSaveState(draft.id, "saving");
    setPinDialog(null);
    launchProgram(programs.find((program) => program.id === "whiteboard"), draft.id);
    createWhiteboard(draft.title, draft.id)
      .then((created) => finishPin(draft.id, item, created))
      .catch(() => setWhiteboardSaveState(draft.id, "save"));
  }

  return (
    <div
      className="modal-overlay computer-modal"
      ref={modalRef}
      role="dialog"
      aria-modal="true"
      aria-label="Detective computer desktop"
      onMouseDown={(event) => { if (event.target === event.currentTarget) exitComputer(); }}
    >
      <section className="computer-frame computer-frame--modal" aria-label="Desktop applications">
      <div className="computer-bezel">
        <div className="computer-screen">
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {booting && <div className="computer-sequence computer-sequence--boot" role="status">
            <Typewriter className="computer-sequence__typewriter" texts={BOOT_LINES} speed={26} lineDelay={220} onComplete={finishBoot} />
          </div>}
          <header className="computer-menubar">
            <p className="computer-brand"><span className="computer-brand__mark">N</span> NOIR TERMINAL <span className="computer-brand__divider">/</span> MERIDIAN GLOBAL RELAY</p>
            <div className="computer-menubar__status">
              <span className="computer-status-light" />
              <span>{location || "WORLD CONNECTION PENDING"}</span>
              {display && (
                <time dateTime={serverTime?.toISOString()} aria-label="Local date and time">
                  {display.date} · {display.time}
                </time>
              )}
            </div>
            <button ref={closeButtonRef} className="computer-modal__close" type="button" aria-label="Close computer" onClick={exitComputer}>×</button>
          </header>

          {!booting && <div className="computer-desktop__content">
            <div className="computer-desktop__heading">
              <p>PRIVATE INVESTIGATION SYSTEM</p>
              <span>PRIVATE LINE · ACCOUNT CHARGED</span>
            </div>
            <nav className="computer-programs" aria-label="Desktop applications">
              {programs.map((program) => (
                <button
                  key={program.id}
                  className={`computer-program${windows.some((item) => item.program.id === program.id) ? " computer-program--selected" : ""}`}
                  type="button"
                  aria-label={`${program.label}, ${program.state}`}
                  disabled={!program.available}
                  onClick={() => launchProgram(program)}
                >
                  <span className={`computer-program__icon computer-program__icon--${program.id}`} aria-hidden="true">
                    <program.Icon className="computer-program__glyph" aria-hidden="true" />
                  </span>
                  <span className="computer-program__label">{program.label}</span>
                  <span className="computer-program__state">{program.state}</span>
                </button>
              ))}
            </nav>

            {windows.map((desktopWindow) => {
              const { id, program, position, zIndex, boardId } = desktopWindow;
              return <section
                ref={(element) => { if (element) windowRefs.current.set(id, element); else windowRefs.current.delete(id); }}
                className={`computer-window${position ? " computer-window--positioned" : ""}`}
                data-window-id={id}
                key={id}
                style={{ ...(position || {}), zIndex }}
                aria-label={`${program.label} application window`}
                onMouseDown={() => focusWindow(id)}
              >
                <header className="computer-window__titlebar" onPointerDown={startWindowDrag} onPointerMove={moveWindow} onPointerUp={endWindowDrag} onPointerCancel={endWindowDrag}>
                  <span><span className="computer-status-light" /> {program.id === "whiteboard" ? "PERSONAL WHITEBOARD" : `${program.label} DATABASE`}</span>
                  <button type="button" className="computer-window__close" aria-label={`Close ${program.label}`} onClick={() => closeWindow(id)}>×</button>
                </header>
                <div className="computer-window__body">
                  {program.id === "case-files"
                    ? <CaseFiles key={id} onPin={requestPin} />
                    : program.id === "whiteboard"
                      ? <Whiteboard key={id} initialBoardId={boardId || preferredBoardId} focusBoard={(boardId) => setPreferredBoardId(boardId)} />
                      : program.id === "email"
                        ? <EmailInbox key={id} serverTime={serverTime} timeZone={timeZone} runHack={runHack} onPin={requestPin} />
                        : <WorldDirectory key={id} program={program.id} serverTime={serverTime} timeZone={timeZone} runHack={runHack} onPin={requestPin} />}
                </div>
              </section>;
            })}
            {hacking && <section className="computer-hack-modal" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} role="status" aria-label="Secure data access in progress">
              <header><span>MERIDIAN SECURE LINK</span><span className="computer-status-light" /></header>
              <div><p>ESTABLISHING SECURE TUNNEL</p><p>ROUTING THROUGH PRIVATE RELAY…</p><p>DECRYPTING REQUESTED FILE</p><div className="computer-sequence__progress"><i /></div></div>
            </section>}
            {pinDialog && <section className="whiteboard-destination" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} role="dialog" aria-modal="true" aria-labelledby="whiteboard-destination-title" onMouseDown={(event) => { if (event.target === event.currentTarget) setPinDialog(null); }}>
              <div className="whiteboard-destination__panel">
                <header><div><p className="computer-window__eyebrow">FILE DISCOVERED · {pinDialog.item.source.toUpperCase()}</p><h2 id="whiteboard-destination-title">PIN TO WHITEBOARD</h2></div><button type="button" aria-label="Close whiteboard selection" onClick={() => setPinDialog(null)}>×</button></header>
                <p className="whiteboard-destination__item">{pinDialog.item.title}</p>
                {pinDialog.error && <p className="whiteboard-destination__error" role="alert">{pinDialog.error}</p>}
                <div className="whiteboard-destination__list">
                  {pinDialog.boards.map((board) => <button key={board.id} type="button" disabled={pinDialog.busy} onClick={() => choosePinDestination(board.id)}><span>{board.title}</span><span>OPEN BOARD →</span></button>)}
                </div>
                <button className="whiteboard-destination__create" type="button" disabled={pinDialog.busy} onClick={createAndChooseBoard}>{pinDialog.busy ? "SAVING RECORD…" : "+ CREATE NEW BOARD"}</button>
              </div>
            </section>}
          </div>}

          <footer className="computer-taskbar">
            <div className="computer-taskbar__left">
              <span className="computer-taskbar__system"><span className="computer-brand__mark">N</span> SYSTEM</span>
              {windows.map((item) => <button key={item.id} className={`computer-taskbar__active${item.zIndex === Math.max(...windows.map((windowItem) => windowItem.zIndex)) ? " computer-taskbar__active--focused" : ""}`} type="button" onClick={() => focusWindow(item.id)}>{item.program.label}</button>)}
            </div>
            <button className="computer-logout" type="button" onClick={exitComputer}>Exit <span aria-hidden="true">↗</span></button>
          </footer>
        </div>
      </div>
      <div className="computer-indicator" aria-hidden="true"><span /> POWER</div>
      </section>
    </div>
  );
}
