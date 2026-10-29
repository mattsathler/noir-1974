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
import PublicServices from "../../../world/pages/directory/PublicServices.jsx";
import EmailInbox from "../../../mail/pages/inbox/EmailInbox.jsx";
import PlayerMailbox from "../../../mail/pages/inbox/PlayerMailbox.jsx";
import HospitalPatientSearch from "./HospitalPatientSearch.jsx";
import HospitalPatientRecord from "./HospitalPatientRecord.jsx";
import ResidentialSearchWindow from "./ResidentialSearchWindow.jsx";
import AccusationWindow from "./AccusationWindow.jsx";
import LegacyCountryPortal from "./LegacyCountryPortal.jsx";
import PoliceWebsite from "./PoliceWebsite.jsx";
import InstitutionalWebsiteFrame from "./InstitutionalWebsiteFrame.jsx";
import HospitalWebsite from "./HospitalWebsite.jsx";
import MeridianEnterprisesWebsite from "./MeridianEnterprisesWebsite.jsx";
import NewspaperWindow from "../../../world/pages/newspaper/NewspaperWindow.jsx";
import DeepLinkWebsite from "./DeepLinkWebsite.jsx";

const BOOT_LINES = [
  "NOIR PERSONAL TERMINAL · PRIVATE SESSION",
  "CHECKING RELAY CONNECTION…",
  "ACTIVE PROXY NODE FOUND",
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
  { id: "web-browser", label: "NOIR WEB", Icon: Wifi, state: "ONLINE", available: true },
  { id: "police-website", label: "MERIDIAN POLICE", Icon: Files, state: "ONLINE", available: true },
  { id: "newspaper", label: "THE EVENING COURIER", Icon: Note, state: "ONLINE", available: true },
  { id: "personal-mail", label: "PERSONAL MAIL", Icon: Mail, state: "ONLINE", available: true },
  { id: "company-portal", label: "COMPANY PORTAL", Icon: MapIcon, state: "ONLINE", available: true },
  { id: "public-services", label: "CITY SERVICES DIRECTORY", Icon: Cross, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: MapIcon, state: "ONLINE", available: true },
];
const WEB_BROWSER_PROGRAM = PERSONAL_PROGRAMS[0];
const MERIDIAN_ENTERPRISES_PROGRAM = { id: "meridian-enterprises", label: "MERIDIAN ENTERPRISES", Icon: MapIcon, state: "ONLINE", available: true };

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
    { id: "warrant-request", label: "WARRANT REQUEST", Icon: Files, state: "ONLINE", available: true },
    { id: "murder-accusation", label: "MURDER ACCUSATION", Icon: Note, state: "ONLINE", available: true },
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

function institutionalVariant(programId) {
  if (programId === "meridian-enterprises") return "gate";
  if (["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports", "interpol", "emails", "warrant-request", "murder-accusation"].includes(programId)) return "police";
  if (programId === "patient-search") return "hospital";
  if (programId === "hospital-patient-record") return "hospital";
  if (programId === "hospital-website") return "hospital";
  if (programId === "bank-records") return "bank";
  if (programId === "gate-records") return "gate";
  if (["public-services", "city-records"].includes(programId)) return "civic";
  if (programId === "personal-mail") return "mail";
  return "company";
}

function browserAddress(program, countryName, facilityContext) {
  if (program.id === "web-browser") return `noir://portal/${countryName || "country"}`;
  if (program.id === "newspaper") return "www.eveningcourier.noir.local";
  if (program.id === "meridian-enterprises") return "www.meridian-enterprises.noir.local";
  if (program.id === "gate-records") return "www.meridian-enterprises.noir.local/gate/access-logs";
  if (program.id === "deep-link") return `www.${program.linkId}.noir.local`;
  if (["hospital-website", "patient-search", "hospital-patient-record"].includes(program.id)) {
    const institution = program.service?.address?.name || facilityContext?.label || "Hôpital Sainte-Claire";
    const city = institution.normalize("NFD").replace(/\p{Diacritic}/gu, "").replace(/\b(?:hopital|hospital)\b/iu, "").trim();
    const slug = city.toLowerCase().replace(/[^a-z0-9]+/gu, "-").replace(/^-|-$/gu, "") || "sainte-claire";
    const recordPath = program.id === "hospital-patient-record"
      ? program.patient?.postMortem
        ? `/patient-services/post-mortem/${encodeURIComponent(program.patient.name)}`
        : `/patient-services/records/${encodeURIComponent(program.patient?.id || "record")}`
      : "";
    return `www.${slug}-hospital.noir.local${recordPath}`;
  }
  if (program.id === "staff-directory") {
    const company = facilityContext?.label?.toLowerCase().replace(/[^a-z0-9]+/gu, "-").replace(/^-|-$/gu, "") || "company";
    return `www.${company}.noir.local/staff`;
  }
  if (["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports"].includes(program.id)) {
    const paths = { "police-notices": "/notices", "police-about": "/about", "police-bureaus": "/district-bureaus", "police-policies": "/records-policy", "police-contact": "/contact", "crime-reports": "/public-records/crime-reports" };
    return `www.police.noir.local${paths[program.id] || ""}`;
  }
  if (program.id === "personal-mail") return "www.meridian-mail.noir.local";
  return `www.${program.id}.noir.local`;
}

export default function ComputerDesktop({ location, display, serverTime, timeZone, cityId, countryName, issueDay, facilityContext, onClose, onOpenCompany, onOpenService, persistent = false, initialAccessAttemptId = "", initialFacilityAccessGranted = false, onOpenProxy, onOpenTelephone, onOpenWhiteboard, onAppearance, onLogout }) {
  const programs = programsFor(facilityContext);
  const visibleDesktopPrograms = persistent ? [WEB_BROWSER_PROGRAM, PERSONAL_PROGRAMS.find((program) => program.id === "newspaper")] : programs;
  const isCompanyComputer = Boolean(facilityContext && facilityContext.kind !== "personal");
  const bootLines = isCompanyComputer ? [`${facilityContext.label} WORKSTATION`, "MERIDIAN AUTHENTICATION SERVICE…", "LOCAL SESSION READY"] : BOOT_LINES;
  const facilityAccessGranted = !isCompanyComputer || initialFacilityAccessGranted;
  const [startMenuOpen, setStartMenuOpen] = useState(false);
  const [accessAttemptId, setAccessAttemptId] = useState(initialAccessAttemptId);
  const [windows, setWindows] = useState(() => {
    try {
      const saved = JSON.parse(sessionStorage.getItem("noir:computer:windows") || "null");
      if (!saved || Date.now() - saved.savedAt > 5 * 60 * 1000) return [];
      const restored = saved.windows.flatMap((item) => {
        const program = programs.find((entry) => entry.id === item.programId);
        return program?.available ? [{ ...item, program }] : [];
      });
      return persistent
        ? restored.map((item) => ({ ...item, id: item.program.id === "newspaper" ? `newspaper:${item.zIndex}` : item.id.startsWith("noir-browser:") ? item.id : `noir-browser:${item.zIndex}` }))
        : restored;
    } catch { return []; }
  });
  const [booting, setBooting] = useState(() => isCompanyComputer || !hasPersonalComputerBooted());
  const [printDialog, setPrintDialog] = useState(null);
  const printDialogRef = useRef(printDialog);
  const modalRef = useRef(null);
  const closeButtonRef = useRef(null);
  const windowRefs = useRef(new Map());
  const dragRef = useRef(null);
  const bootTimerRef = useRef(null);
  const nextZIndexRef = useRef(Math.max(5, ...windows.map((item) => item.zIndex + 1)));
  const routedFacilityAccessRef = useRef("");
  const windowsStateRef = useRef(windows);
  printDialogRef.current = printDialog;
  windowsStateRef.current = windows;

  useEffect(() => {
    return () => window.clearTimeout(bootTimerRef.current);
  }, []);

  useEffect(() => {
    if (!persistent || !isCompanyComputer || !initialFacilityAccessGranted) return;
    const accessKey = `${facilityContext?.kind}:${facilityContext?.locationId}:${initialAccessAttemptId}`;
    if (routedFacilityAccessRef.current === accessKey) return;
    const destination = programs.find((program) => program.available && !program.id.startsWith("internal-") && !["print-queue", "network-status", "voice-mail"].includes(program.id));
    const browserWindow = windowsStateRef.current.find((item) => item.id.startsWith("noir-browser:"));
    if (!destination || !browserWindow || browserWindow.program.id === destination.id) return;
    routedFacilityAccessRef.current = accessKey;
    const zIndex = nextZIndexRef.current++;
    setWindows((previous) => previous.map((item) => item.id === browserWindow.id
      ? { ...item, program: destination, zIndex, history: [...(item.history || []), item.program] }
      : item));
  }, [facilityContext?.companyId, facilityContext?.kind, facilityContext?.locationId, initialAccessAttemptId, initialFacilityAccessGranted, isCompanyComputer, persistent, programs]);

  useEffect(() => {
    setAccessAttemptId(initialAccessAttemptId);
  }, [initialAccessAttemptId]);

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
    if (persistent) return undefined;
    const previousFocus = document.activeElement;
    closeButtonRef.current?.focus();

    function handleModalKeys(event) {
      if (event.key === "Escape") {
        if (printDialogRef.current) { setPrintDialog(null); return; }
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
  }, [onClose, persistent]);

  function focusWindow(windowId) {
    const zIndex = nextZIndexRef.current++;
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, zIndex } : item));
  }

  function launchProgram(program, sourceWindowId = null) {
    if (!program?.available) return;
    if (persistent) {
      if (program.id === "newspaper") {
        const existingNewspaper = windows.find((item) => item.program.id === "newspaper");
        if (existingNewspaper) { focusWindow(existingNewspaper.id); return; }
        const zIndex = nextZIndexRef.current++;
        setWindows((previous) => [...previous, { id: `newspaper:${zIndex}`, program, zIndex, position: null }]);
        return;
      }
      const browserWindow = sourceWindowId
        ? windows.find((item) => item.id === sourceWindowId)
        : program.id === "web-browser" ? null : windows.find((item) => item.id.startsWith("noir-browser:"));
      if (browserWindow) {
        const zIndex = nextZIndexRef.current++;
        setWindows((previous) => previous.map((item) => item.id === browserWindow.id
          ? { ...item, program, zIndex, history: program.id === "web-browser" ? [] : [...(item.history || []), item.program] }
          : item));
        return;
      }
      const zIndex = nextZIndexRef.current++;
      const offset = windows.length;
      const compact = window.matchMedia("(max-width: 42rem)").matches;
      const position = offset === 0 ? null : {
        left: Math.min(compact ? 8 + offset * 10 : 40 + offset * 28, compact ? 32 : 140),
        top: Math.min(compact ? 8 + offset * 10 : 32 + offset * 24, compact ? 32 : 120),
      };
      setWindows((previous) => [...previous, { id: `noir-browser:${zIndex}`, program, zIndex, position, history: [] }]);
      return;
    }
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

  function goBackInBrowser(windowId) {
    const current = windows.find((item) => item.id === windowId);
    const previousProgram = current?.history?.at(-1) || WEB_BROWSER_PROGRAM;
    const history = current?.history?.slice(0, -1) || [];
    const zIndex = nextZIndexRef.current++;
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, program: previousProgram, history, zIndex } : item));
  }

  function closeWindow(windowId) {
    setWindows((previous) => previous.filter((item) => item.id !== windowId));
    windowRefs.current.delete(windowId);
  }

  function finishPrint(boardId, item, board = null) {
    const destination = getCachedWhiteboard(boardId) ?? board ?? printDialogRef.current?.boards.find((candidate) => candidate.id === boardId) ?? { id: boardId, pins: [] };
    const index = destination.pins?.length ?? 0;
    const x = 600 + (index % 3) * 3000;
    const y = 700 + (Math.floor(index / 3) % 3) * 2500;
    cachePinnedWhiteboardItem(boardId, item, x, y, destination);
    pinWhiteboardItem(boardId, item.id, x, y)
      .then(() => setWhiteboardSaveState(boardId, "saved"))
      .catch(() => setWhiteboardSaveState(boardId, "save"));
    setPrintDialog(null);
  }

  async function requestPrint(item) {
    const printableItem = { ...item, source: item.source || "page" };
    let boards = [];
    try {
      boards = await getWhiteboards();
      const editableBoards = boards.filter((board) => !board.lockedAt);
      if (!editableBoards.length) {
        const draft = createWhiteboardDraft(0);
        cacheWhiteboardDraft(draft);
        setWhiteboardSaveState(draft.id, "saving");
        setPrintDialog(null);
        createWhiteboard(draft.title, draft.id)
          .then((created) => finishPrint(draft.id, printableItem, created))
          .catch(() => setWhiteboardSaveState(draft.id, "save"));
        return;
      }
      setPrintDialog({ item: printableItem, boards: editableBoards, busy: false, error: "" });
    } catch (error) {
      const availableBoards = await getWhiteboards().catch(() => boards);
      setPrintDialog({ item: printableItem, boards: availableBoards.filter((board) => !board.lockedAt), busy: false, error: error.message || "COULD NOT OPEN CASE BOARDS" });
    }
  }

  async function choosePrintDestination(boardId) {
    if (!printDialog) return;
    setPrintDialog((previous) => ({ ...previous, busy: true, error: "" }));
    try {
      finishPrint(boardId, printDialog.item);
    } catch (error) {
      setPrintDialog((previous) => previous ? { ...previous, busy: false, error: error.message || "COULD NOT SEND THIS PAGE TO THE BOARD" } : previous);
    }
  }

  async function createBoardAndPrint() {
    if (!printDialog) return;
    const draft = createWhiteboardDraft(printDialog.boards.length);
    const item = printDialog.item;
    cacheWhiteboardDraft(draft);
    setWhiteboardSaveState(draft.id, "saving");
    setPrintDialog(null);
    createWhiteboard(draft.title, draft.id)
      .then((created) => finishPrint(draft.id, item, created))
      .catch(() => setWhiteboardSaveState(draft.id, "save"));
  }

  return (
    <div
      className={persistent ? "computer-shell" : "modal-overlay computer-modal"}
      ref={modalRef}
      role={persistent ? "region" : "dialog"}
      aria-modal={persistent ? undefined : "true"}
      aria-label="Detective computer desktop"
      onMouseDown={(event) => { if (!persistent && event.target === event.currentTarget) exitComputer(); }}
    >
      <section className={`computer-frame ${persistent ? "computer-frame--persistent" : "computer-frame--modal"}`} aria-label="Desktop">
      <div className="computer-bezel">
        <div className={`computer-screen${persistent ? " computer-screen--desktop" : ""}`}>
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {booting && <div className="computer-sequence computer-sequence--boot" role="status">
            <Typewriter className="computer-sequence__typewriter" texts={bootLines} speed={26} lineDelay={220} onComplete={finishBoot} />
          </div>}
          <header className={`computer-menubar${persistent ? " computer-menubar--desktop-hidden" : ""}`}>
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
            {persistent && <nav className="computer-menubar__actions" aria-label="System controls">
              <button type="button" onClick={onOpenProxy}>PROXY</button>
              <button type="button" onClick={onOpenTelephone}>TELEPHONE</button>
              <button type="button" onClick={onOpenWhiteboard}>CASE BOARD</button>
            </nav>}
            {!persistent && <button ref={closeButtonRef} className="computer-modal__close" type="button" aria-label="Close computer" onClick={exitComputer}>×</button>}
          </header>

          {!booting && facilityAccessGranted && <div className="computer-desktop__content">
            <div className="computer-desktop__heading">
              <p>{facilityContext?.label || "PRIVATE INVESTIGATION SYSTEM"}</p>
              <span>PRIVATE LINE · ACCOUNT CHARGED</span>
            </div>
            <nav className="computer-programs" aria-label={persistent ? "Desktop browser" : "Computer programs"}>
              {visibleDesktopPrograms.map((program) => (
                <button
                  key={program.id}
                  className={`computer-program${windows.some((item) => item.program.id === program.id) ? " computer-program--selected" : ""}`}
                  type="button"
                  aria-label={`${program.label}${persistent ? " website" : " application"}, ${program.state}`}
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
              const { id, program, position, zIndex } = desktopWindow;
              return <section
                ref={(element) => { if (element) windowRefs.current.set(id, element); else windowRefs.current.delete(id); }}
                className={`computer-window${position ? " computer-window--positioned" : ""}${persistent && program.id !== "newspaper" ? " computer-window--website" : ""}${persistent && program.id === "newspaper" ? " computer-window--app" : ""}`}
                data-site-variant={institutionalVariant(program.id)}
                data-window-id={id}
                key={id}
                style={{ ...(position || {}), zIndex }}
                aria-label={`${program.label} ${program.id === "newspaper" ? "application" : "website"} window`}
                onMouseDown={() => focusWindow(id)}
              >
                <header className="computer-window__titlebar" onPointerDown={startWindowDrag} onPointerMove={moveWindow} onPointerUp={endWindowDrag} onPointerCancel={endWindowDrag}>
                  <span>{persistent && !["web-browser", "newspaper"].includes(program.id) && <button type="button" className="computer-browser-back" aria-label="Go back" title="Go back" onPointerDown={(event) => event.stopPropagation()} onClick={() => goBackInBrowser(id)}>←</button>}<span className="computer-status-light" /> {program.id === "web-browser" ? "NOIR BROWSER" : program.label}</span>
                  <button type="button" className="computer-window__close" aria-label={`Close ${program.label}`} onClick={() => closeWindow(id)}>×</button>
                </header>
                <div className="computer-browser-address" aria-label="Website address"><span aria-hidden="true">▣</span><span>{browserAddress(program, countryName, facilityContext)}</span><span className="computer-browser-address__secure">PRIVATE RELAY</span></div>
                <div className={`computer-window__body${persistent && program.id !== "newspaper" ? " computer-window__body--website" : ""}`}>
                  {program.id === "web-browser"
                    ? <LegacyCountryPortal countryName={countryName} location={location} cityId={cityId} programs={programs} navigate={(destination) => launchProgram(destination, id)} />
                    : ["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports"].includes(program.id)
                      ? <PoliceWebsite page={program.id} navigate={(destination) => launchProgram(destination, id)}>{program.id === "crime-reports" && <CaseFiles key={`${id}:crime-reports:${program.searchTerm || ""}`} initialSearch={program.searchTerm} cityId={cityId} issueDay={issueDay} timeZone={timeZone} />}</PoliceWebsite>
                      : program.id === "deep-link"
                        ? <InstitutionalWebsiteFrame variant="civic" organizationName={program.label}><DeepLinkWebsite linkId={program.linkId} /></InstitutionalWebsiteFrame>
                      : program.id === "newspaper"
                        ? <NewspaperWindow cityId={cityId} issueDay={issueDay} embedded />
                      : program.id === "meridian-enterprises"
                        ? <MeridianEnterprisesWebsite location={location} cityId={cityId} countryName={countryName} onOpenService={onOpenService} />
                      : program.id === "hospital-website"
                        ? <HospitalWebsite service={program.service} onConnect={onOpenService} />
                      : <InstitutionalWebsiteFrame
                        variant={institutionalVariant(program.id)}
                        organizationName={["staff-directory", "patient-search", "hospital-patient-record"].includes(program.id) ? facilityContext?.label : ""}
                        onHome={program.id === "gate-records" ? () => launchProgram(MERIDIAN_ENTERPRISES_PROGRAM, id) : ["patient-search", "hospital-patient-record"].includes(program.id) ? () => launchProgram({
                          id: "hospital-website",
                          label: "HOSPITAL WEBSITE",
                          Icon: Cross,
                          state: "ONLINE",
                          available: true,
                          service: { kind: "hospital", address: {
                            name: facilityContext?.label || "Hôpital Sainte-Claire",
                            city: facilityContext?.city || "",
                            country: countryName || "",
                            locationId: facilityContext?.locationId || "",
                            full: [facilityContext?.label, facilityContext?.city, countryName].filter(Boolean).join(", "),
                          } },
                        }, id) : undefined}
                      >
                        {program.id === "case-files"
                          ? <CaseFiles key={id} cityId={cityId} issueDay={issueDay} timeZone={timeZone} />
                          : program.id === "patient-search"
                            ? <HospitalPatientSearch key={id} accessAttemptId={accessAttemptId} timeZone={timeZone} onPrint={requestPrint} onOpenPatient={(patient) => launchProgram({ id: "hospital-patient-record", label: patient.postMortem ? `POST-MORTEM FILE · ${patient.name}` : `PATIENT FILE · ${patient.name}`, state: "ONLINE", available: true, patient }, id)} />
                            : program.id === "hospital-patient-record"
                              ? <HospitalPatientRecord key={`${id}:${program.patient?.id}`} patient={program.patient} timeZone={timeZone} onPrint={requestPrint} />
                            : program.id === "emails"
                              ? <EmailInbox key={id} serverTime={serverTime} timeZone={timeZone} cityId={cityId} facilityLocationId={facilityContext?.locationId} accessAttemptId={accessAttemptId} onPrint={requestPrint} />
                              : program.id === "personal-mail"
                                ? <PlayerMailbox key={id} timeZone={timeZone} onPrint={requestPrint} />
                                : program.id === "public-services"
                                  ? <PublicServices key={id} cityId={cityId} countryName={countryName} onOpenService={onOpenService} onOpenHospitalWebsite={(service) => launchProgram({ id: "hospital-website", label: service.address.name, Icon: Cross, state: "ONLINE", available: true, service }, id)} />
                                  : program.id === "warrant-request"
                                    ? <ResidentialSearchWindow key={id} cityId={cityId} stationLocationId={facilityContext?.locationId} onClose={() => closeWindow(id)} />
                                    : program.id === "murder-accusation"
                                      ? <AccusationWindow key={id} stationLocationId={facilityContext?.locationId} onClose={() => closeWindow(id)} />
                                      : <WorldDirectory key={`${id}:${program.id}`} program={program.id} companyLocationId={program.id === "staff-directory" ? facilityContext?.locationId : undefined} companyId={program.id === "staff-directory" ? facilityContext?.companyId : undefined} facilityLocationId={facilityContext?.locationId} accessAttemptId={accessAttemptId} cityId={cityId} countryName={countryName} serverTime={serverTime} timeZone={timeZone} onPrint={requestPrint} onOpenCompany={onOpenCompany} />}
                      </InstitutionalWebsiteFrame>}
                </div>
              </section>;
            })}
            {printDialog && <section className="page-print-dialog" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} role="dialog" aria-modal="true" aria-labelledby="page-print-title" onMouseDown={(event) => { if (event.target === event.currentTarget) setPrintDialog(null); }}>
              <div className="page-print-dialog__panel">
                <header className="page-print-dialog__titlebar"><span><Printer aria-hidden="true" /> MERIDIAN WEB BROWSER · PRINT</span><button type="button" aria-label="Close print dialog" onClick={() => setPrintDialog(null)}>×</button></header>
                <div className="page-print-dialog__content">
                  <p className="page-print-dialog__eyebrow">PAGE READY · {printDialog.item.source.toUpperCase()}</p>
                  <h2 id="page-print-title">PRINT PAGE</h2>
                  <p className="page-print-dialog__item">{printDialog.item.title || "UNTITLED PAGE"}</p>
                  {printDialog.error && <p className="page-print-dialog__error" role="alert">{printDialog.error}</p>}
                  <div className="page-print-dialog__destinations">
                    {printDialog.boards.map((board) => <button key={board.id} type="button" disabled={printDialog.busy} onClick={() => choosePrintDestination(board.id)}><span>{board.title}</span><span>PRINT TO BOARD →</span></button>)}
                  </div>
                  <button className="page-print-dialog__create" type="button" disabled={printDialog.busy} onClick={createBoardAndPrint}>{printDialog.busy ? "SENDING PAGE…" : "+ CREATE BOARD AND PRINT"}</button>
                  <footer className="page-print-dialog__actions"><button type="button" onClick={() => setPrintDialog(null)}>CANCEL</button></footer>
                </div>
              </div>
            </section>}
          </div>}

          <footer className={`computer-taskbar${persistent ? " computer-taskbar--desktop" : ""}`}>
            <div className="computer-taskbar__left">
              {persistent && <button type="button" className="computer-start" aria-expanded={startMenuOpen} onClick={() => setStartMenuOpen((open) => !open)}><span className="computer-brand__mark">N</span> START</button>}
              {!persistent && <span className="computer-taskbar__system"><span className="computer-brand__mark">N</span> {isCompanyComputer ? "WORKSTATION SESSION" : "PERSONAL SESSION"}</span>}
              {windows.map((item) => <button key={item.id} className={`computer-taskbar__active${item.zIndex === Math.max(...windows.map((windowItem) => windowItem.zIndex)) ? " computer-taskbar__active--focused" : ""}`} type="button" onClick={() => focusWindow(item.id)}>{item.program.label}</button>)}
            </div>
            {startMenuOpen && persistent && <nav className="computer-start-menu" aria-label="System menu">
              <strong>NOIR PERSONAL COMPUTER</strong>
              <button type="button" onClick={() => { launchProgram(PERSONAL_PROGRAMS[0]); setStartMenuOpen(false); }}>OPEN NOIR WEB</button>
              <button type="button" onClick={() => { launchProgram(PERSONAL_PROGRAMS.find((program) => program.id === "newspaper")); setStartMenuOpen(false); }}>THE EVENING COURIER</button>
              <button type="button" onClick={() => { onOpenProxy?.(); setStartMenuOpen(false); }}>PROXY CONNECTION</button>
              <button type="button" onClick={() => { onOpenTelephone?.(); setStartMenuOpen(false); }}>TELEPHONE</button>
              <button type="button" onClick={() => { onOpenWhiteboard?.(); setStartMenuOpen(false); }}>CASE BOARD</button>
              <button type="button" onClick={() => { onAppearance?.(); setStartMenuOpen(false); }}>APPEARANCE</button>
              {!isCompanyComputer && <button type="button" onClick={() => { onLogout?.(); setStartMenuOpen(false); }}>LOG OUT ↗</button>}
            </nav>}
            {persistent ? isCompanyComputer
              ? <span className="computer-taskbar__clock">{display?.time}</span>
              : <span className="computer-taskbar__clock">{display?.time}</span>
              : <button className="computer-logout" type="button" onClick={exitComputer}>Exit <span aria-hidden="true">↗</span></button>}
          </footer>
        </div>
      </div>
      {!persistent && <div className="computer-indicator" aria-hidden="true"><span /> POWER</div>}
      </section>
    </div>
  );
}
