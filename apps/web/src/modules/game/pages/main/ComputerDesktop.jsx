import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { Files } from "pixelarticons/react/Files";
import { Cross } from "pixelarticons/react/Cross";
import { Dollar } from "pixelarticons/react/Dollar";
import { Mail } from "pixelarticons/react/Mail";
import { Map as MapIcon } from "pixelarticons/react/Map";
import { Message } from "pixelarticons/react/Message";
import { Note } from "pixelarticons/react/Note";
import { Printer } from "pixelarticons/react/Printer";
import { User } from "pixelarticons/react/User";
import { Users } from "pixelarticons/react/Users";
import { Wifi } from "pixelarticons/react/Wifi";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";
import { cachePinnedWhiteboardItem, cacheWhiteboardDraft, createWhiteboard, createWhiteboardDraft, getCachedWhiteboard, getWhiteboards, pinWhiteboardItem, setWhiteboardSaveState } from "../../../world/api.js";
import CaseFiles from "../../../cases/pages/files/CaseFiles.jsx";
import WorldDirectory from "../../../world/pages/directory/WorldDirectory.jsx";
import EmailInbox from "../../../mail/pages/inbox/EmailInbox.jsx";
import PlayerMailbox from "../../../mail/pages/inbox/PlayerMailbox.jsx";
import LocalTerminal from "./LocalTerminal.jsx";
import HospitalPatientSearch from "./HospitalPatientSearch.jsx";

const BOOT_LINES = [
  "NOIR PERSONAL TERMINAL · PRIVATE SESSION",
  "CHECKING RELAY CONNECTION…",
  "LYON NODE FOUND",
  "PRIVATE SYSTEM READY",
];
const PERSONAL_BOOT_STORAGE_KEY = "noir:computer:personal-booted";

function hasPersonalComputerBooted() {
  try { return sessionStorage.getItem(PERSONAL_BOOT_STORAGE_KEY) === "true"; }
  catch { return false; }
}

function markPersonalComputerBooted() {
  try { sessionStorage.setItem(PERSONAL_BOOT_STORAGE_KEY, "true"); }
  catch { /* Keep the boot sequence available when browser storage is disabled. */ }
}

const PERSONAL_PROGRAMS = [
  { id: "case-files", label: "CRIME REPORTS", Icon: Files, state: "ONLINE", available: true },
  { id: "personal-mail", label: "PERSONAL MAIL", Icon: Mail, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: MapIcon, state: "ONLINE", available: true },
];

const BYPASSED_PROGRAM_LABELS = {
  "case-files": "CRI\uE000E REPO\uE001TS",
  "city-records": "CI\uE003Y RECO\uE004DS",
  interpol: "PERSON\uE001EL SEARCH",
  emails: "E\uE002AIL ARCHIVE",
  "bank-records": "FINANCI\uE003L RECORDS",
  "gate-records": "GATE ACCESS L\uE004GS",
  "staff-directory": "STAFF DIREC\uE001ORY",
  "internal-memos": "INTERNAL ME\uE002OS",
  "print-queue": "PRINT Q\uE003EUE",
  "network-status": "NETW\uE004RK STATUS",
  "voice-mail": "VOICE MESSA\uE000ES",
};

function visibleProgramLabel(program, bypassed) {
  const label = bypassed ? BYPASSED_PROGRAM_LABELS[program.id] || program.label : program.label;
  if (!bypassed) return label;
  return Array.from(label).map((character, index) => /[\uE000-\uF8FF]/u.test(character)
    ? <span className="computer-program__tofu" key={`${program.id}:tofu:${index}`}>{character}</span>
    : character);
}

function programsFor(facilityContext) {
  if (!facilityContext || facilityContext.kind === "personal") return PERSONAL_PROGRAMS;
  const offlinePrograms = [
    { id: "internal-memos", label: "INTERNAL MEMOS", Icon: Note, state: "OFFLINE", available: false },
    { id: "print-queue", label: "PRINT QUEUE", Icon: Printer, state: "OFFLINE", available: false },
    { id: "network-status", label: "NETWORK STATUS", Icon: Wifi, state: "OFFLINE", available: false },
    { id: "voice-mail", label: "VOICE MESSAGES", Icon: Message, state: "OFFLINE", available: false },
  ];
  let activePrograms;
  if (facilityContext.kind === "police") return [
    { id: "interpol", label: "PERSONNEL SEARCH", Icon: User, state: "ONLINE", available: true },
    { id: "emails", label: "EMAIL ARCHIVE", Icon: Mail, state: "ONLINE", available: true },
    ...offlinePrograms,
  ];
  if (facilityContext.kind === "hospital") return [
    { id: "patient-search", label: "PATIENT SEARCH", Icon: Cross, state: "ONLINE", available: true },
    ...offlinePrograms,
  ];
  if (facilityContext.kind === "bank") activePrograms = [{ id: "bank-records", label: "FINANCIAL RECORDS", Icon: Dollar, state: "ONLINE", available: true }];
  else if (facilityContext.kind === "gate") activePrograms = [{ id: "gate-records", label: "GATE ACCESS LOGS", Icon: MapIcon, state: "ONLINE", available: true }];
  else if (facilityContext.kind === "company") activePrograms = [
    { id: "staff-directory", label: "STAFF DIRECTORY", Icon: Users, state: "ONLINE", available: true },
  ];
  if (activePrograms) return [...activePrograms, ...offlinePrograms];
  return PERSONAL_PROGRAMS;
}

export default function ComputerDesktop({ location, display, serverTime, timeZone, cityId, countryName, issueDay, facilityContext, onClose }) {
  const programs = programsFor(facilityContext);
  const isCompanyComputer = Boolean(facilityContext && facilityContext.kind !== "personal");
  const bootLines = isCompanyComputer ? [`${facilityContext.label} WORKSTATION`, "MERIDIAN AUTHENTICATION SERVICE…", "LOCAL SESSION READY"] : BOOT_LINES;
  const [facilityUnlocked, setFacilityUnlocked] = useState(!isCompanyComputer);
  const isBypassed = isCompanyComputer && facilityUnlocked;
  const [terminalOpen, setTerminalOpen] = useState(false);
  const [accessAttemptId, setAccessAttemptId] = useState("");
  const [loginError, setLoginError] = useState("");
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
  const [booting, setBooting] = useState(() => isCompanyComputer || !hasPersonalComputerBooted());
  const [pinDialog, setPinDialog] = useState(null);
  const modalRef = useRef(null);
  const closeButtonRef = useRef(null);
  const windowRefs = useRef(new Map());
  const dragRef = useRef(null);
  const bootTimerRef = useRef(null);
  const nextZIndexRef = useRef(Math.max(5, ...windows.map((item) => item.zIndex + 1)));
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
        windows: windows.map(({ id, program, zIndex, position }) => ({ id, programId: program.id, zIndex, position })),
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
    bootTimerRef.current = window.setTimeout(() => {
      setBooting(false);
      if (!isCompanyComputer) markPersonalComputerBooted();
    }, 650);
  }

  function persistWindows() {
    try {
      sessionStorage.setItem("noir:computer:windows", JSON.stringify({
        savedAt: Date.now(),
        windows: windowsStateRef.current.map(({ id, program, zIndex, position }) => ({ id, programId: program.id, zIndex, position })),
      }));
    } catch { /* Storage may be disabled by the browser. */ }
  }

  function exitComputer() {
    persistWindows();
    onClose?.();
  }

  const runHack = useCallback(async (loadRequest) => {
    const request = typeof loadRequest === "function" ? loadRequest() : Promise.resolve(loadRequest);
    const [result] = await Promise.all([
      request,
      new Promise((resolve) => window.setTimeout(resolve, 900)),
    ]);
    return result;
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

  function launchProgram(program) {
    if (!program.available) return;
    const existing = windows.find((item) => item.program.id === program.id);
    if (existing) { focusWindow(existing.id); return; }
    const zIndex = nextZIndexRef.current++;
    const id = `${program.id}:${zIndex}`;
    const offset = windows.length;
    const compact = window.matchMedia("(max-width: 42rem)").matches;
    const position = offset === 0 ? null : {
      left: Math.min(compact ? 8 + offset * 10 : 40 + offset * 28, compact ? 32 : 140),
      top: Math.min(compact ? 8 + offset * 10 : 32 + offset * 24, compact ? 32 : 120),
    };
    setWindows((previous) => [...previous, { id, program, zIndex, position }]);
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
  }

  async function requestPin(item) {
    let boards = [];
    try {
      boards = await getWhiteboards();
      const editableBoards = boards.filter((board) => !board.lockedAt);
      if (!editableBoards.length) {
        const draft = createWhiteboardDraft(0);
        cacheWhiteboardDraft(draft);
        setWhiteboardSaveState(draft.id, "saving");
        setPinDialog(null);
        createWhiteboard(draft.title, draft.id)
          .then((created) => finishPin(draft.id, item, created))
          .catch(() => setWhiteboardSaveState(draft.id, "save"));
        return;
      }
      setPinDialog({ item, boards: editableBoards, busy: false, error: "" });
    } catch (error) {
      const availableBoards = await getWhiteboards().catch(() => boards);
      setPinDialog({ item, boards: availableBoards.filter((board) => !board.lockedAt), busy: false, error: error.message || "COULD NOT OPEN PERSONAL BOARDS" });
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
        <div className={`computer-screen${isBypassed ? " computer-screen--bypassed" : ""}`}>
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {isBypassed && <div className="computer-bypass-overlay" aria-hidden="true">
            <span>BYPASSDECRYPT.EXE // SESSION HIJACKED</span>
            <span>AUTH LAYER: OVERRIDDEN // BYPASS</span>
            <span>TRACE DETECTED… FALSE ROUTE ENGAGED</span>
            <span className="computer-bypass-garble">▓▒░ 0x7F // �#%&amp;? ⌧ ▒▓█</span>
            <span className="computer-bypass-garble">�� 1010:ERR // �&lt;!~ #�</span>
            <span className="computer-bypass-garble">KERNEL? NO // BYPASS::BYPASS</span>
            <span className="computer-bypass-garble">▒#@! 0x0000:DECRYPT_FAIL?=FALSE</span>
            <span className="computer-bypass-tofu computer-bypass-tofu--one">{"\uE000\uE001\uE002"}</span>
            <span className="computer-bypass-tofu computer-bypass-tofu--two">{"\uF8FF\uE017\uE0FF"}</span>
            <span className="computer-bypass-tofu computer-bypass-tofu--three">{"\uE100\uE101\uE102\uE103"}</span>
            <span className="computer-bypass-tofu computer-bypass-tofu--four">{"\uE777\uE778\uE779"}</span>
          </div>}
          {booting && <div className="computer-sequence computer-sequence--boot" role="status">
            <Typewriter className="computer-sequence__typewriter" texts={bootLines} speed={26} lineDelay={220} onComplete={finishBoot} />
          </div>}
          <header className="computer-menubar">
            <p className="computer-brand"><span className="computer-brand__mark">N</span> NOIR TERMINAL <span className="computer-brand__divider">/</span> MERIDIAN GLOBAL RELAY {isBypassed && <span className="computer-bypass-tag">BYPASS ACTIVE</span>}</p>
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

          {!booting && isCompanyComputer && !facilityUnlocked && <section className="company-login" aria-label={`${facilityContext.label} employee login`}>
            <div className="company-login__panel">
              <p className="company-login__eyebrow">{facilityContext.label} · EMPLOYEE ACCESS</p>
              <h1>Sign in to your workstation</h1>
              <p className="company-login__copy">Use your company credentials to access internal applications.</p>
              <form onSubmit={(event) => { event.preventDefault(); setLoginError("INVALID EMPLOYEE CREDENTIALS"); }}>
                <label htmlFor="company-login-email">EMPLOYEE EMAIL</label>
                <input id="company-login-email" type="email" autoComplete="username" placeholder="name@meridian.global" />
                <label htmlFor="company-login-password">PASSWORD</label>
                <input id="company-login-password" type="password" autoComplete="current-password" placeholder="Enter your password" />
                {loginError && <p className="company-login__error" role="alert">{loginError}</p>}
                <button type="submit">SIGN IN</button>
              </form>
            </div>
            <button className="company-login__terminal" type="button" aria-label="Open local terminal" title="Open local terminal" onClick={() => setTerminalOpen(true)}><span aria-hidden="true">&gt;_</span></button>
          </section>}

          {!booting && (!isCompanyComputer || facilityUnlocked) && <div className="computer-desktop__content">
            <div className="computer-desktop__heading">
              <p>{facilityContext?.label || "PRIVATE INVESTIGATION SYSTEM"} {isBypassed && <span className="computer-bypass-inline">// SYSTEM OVERRIDDEN</span>}</p>
              <span>{isBypassed ? "UNAUTHORIZED SESSION · BYPASSDECRYPT RUNNING" : "PRIVATE LINE · ACCOUNT CHARGED"}</span>
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
                  <span className="computer-program__label">{visibleProgramLabel(program, isBypassed)}</span>
                  <span className="computer-program__state">{program.state}</span>
                </button>
              ))}
            </nav>

            {windows.map((desktopWindow) => {
              const { id, program, position, zIndex } = desktopWindow;
              return <section
                ref={(element) => { if (element) windowRefs.current.set(id, element); else windowRefs.current.delete(id); }}
                className={`computer-window${position ? " computer-window--positioned" : ""}`}
                data-window-id={id}
                key={id}
                style={{ ...(position || {}), zIndex }}
                aria-label={`${program.label} application window`}
                onMouseDown={() => focusWindow(id)}
              >
                {isBypassed && <div className="computer-window__corruption" aria-hidden="true">
                  <span>ERR://BYPASS_TOKEN_REJECTED</span>
                  <span>0x7F ▓▒░ �⿻� AUTH::FALSE</span>
                  <span>DATA STREAM DESYNC // RETRY</span>
                  <span>{"\uE000\uE001 // \uE017\uE018\uE019"}</span>
                </div>}
                <header className="computer-window__titlebar" onPointerDown={startWindowDrag} onPointerMove={moveWindow} onPointerUp={endWindowDrag} onPointerCancel={endWindowDrag}>
                  <span><span className="computer-status-light" /> {program.id === "personal-mail" ? "PERSONAL MAILBOX" : `${program.label} DATABASE`} {isBypassed && <i className="computer-bypass-window-tag">BYPASS LINK</i>}</span>
                  <button type="button" className="computer-window__close" aria-label={`Close ${program.label}`} onClick={() => closeWindow(id)}>×</button>
                </header>
                <div className="computer-window__body">
                  {program.id === "case-files"
                    ? <CaseFiles key={id} cityId={cityId} issueDay={issueDay} timeZone={timeZone} />
                    : program.id === "patient-search"
                      ? <HospitalPatientSearch key={id} accessAttemptId={accessAttemptId} timeZone={timeZone} onPin={requestPin} />
                      : program.id === "emails"
                        ? <EmailInbox key={id} serverTime={serverTime} timeZone={timeZone} cityId={cityId} facilityLocationId={facilityContext?.locationId} accessAttemptId={accessAttemptId} runHack={runHack} onPin={requestPin} />
                        : program.id === "personal-mail"
                          ? <PlayerMailbox key={id} timeZone={timeZone} onPin={requestPin} />
                        : <WorldDirectory key={id} program={program.id} companyLocationId={program.id === "staff-directory" ? facilityContext?.locationId : undefined} companyId={program.id === "staff-directory" ? facilityContext?.companyId : undefined} facilityLocationId={facilityContext?.locationId} accessAttemptId={accessAttemptId} cityId={cityId} countryName={countryName} serverTime={serverTime} timeZone={timeZone} runHack={runHack} onPin={requestPin} />}
                </div>
              </section>;
            })}
            {pinDialog && <section className="whiteboard-destination" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} role="dialog" aria-modal="true" aria-labelledby="whiteboard-destination-title" onMouseDown={(event) => { if (event.target === event.currentTarget) setPinDialog(null); }}>
              <div className="whiteboard-destination__panel">
                <header><div><p className="computer-window__eyebrow">FILE DISCOVERED · {pinDialog.item.source.toUpperCase()}</p><h2 id="whiteboard-destination-title">PIN TO WHITEBOARD</h2></div><button type="button" aria-label="Close whiteboard selection" onClick={() => setPinDialog(null)}>×</button></header>
                <p className="whiteboard-destination__item">{pinDialog.item.title}</p>
                {pinDialog.error && <p className="whiteboard-destination__error" role="alert">{pinDialog.error}</p>}
                <div className="whiteboard-destination__list">
                  {pinDialog.boards.map((board) => <button key={board.id} type="button" disabled={pinDialog.busy} onClick={() => choosePinDestination(board.id)}><span>{board.title}</span><span>PIN TO BOARD →</span></button>)}
                </div>
                <button className="whiteboard-destination__create" type="button" disabled={pinDialog.busy} onClick={createAndChooseBoard}>{pinDialog.busy ? "SAVING RECORD…" : "+ CREATE NEW BOARD"}</button>
              </div>
            </section>}
          </div>}

          {terminalOpen && <LocalTerminal facility={facilityContext} variant="app" onAccessGranted={(attempt) => { setAccessAttemptId(String(attempt.attemptId)); setFacilityUnlocked(true); setTerminalOpen(false); }} onClose={() => setTerminalOpen(false)} />}

          <footer className="computer-taskbar">
            <div className="computer-taskbar__left">
              <span className="computer-taskbar__system"><span className="computer-brand__mark">N</span> {isBypassed ? "BYPASS SESSION" : isCompanyComputer ? "WORKSTATION LOGIN" : "PERSONAL SESSION"}</span>
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
