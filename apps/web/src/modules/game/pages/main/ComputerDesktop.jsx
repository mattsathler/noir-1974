import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { Files } from "pixelarticons/react/Files";
import { Cross } from "pixelarticons/react/Cross";
import { Dollar } from "pixelarticons/react/Dollar";
import { Mail } from "pixelarticons/react/Mail";
import { Map as MapIcon } from "pixelarticons/react/Map";
import { Message } from "pixelarticons/react/Message";
import { Note } from "pixelarticons/react/Note";
import { Phone } from "pixelarticons/react/Phone";
import { Printer } from "pixelarticons/react/Printer";
import { User } from "pixelarticons/react/User";
import { Users } from "pixelarticons/react/Users";
import { Wifi } from "pixelarticons/react/Wifi";
import { Brush } from "pixelarticons/react/Brush";
import { Bug } from "pixelarticons/react/Bug";
import { Target } from "pixelarticons/react/Target";
import { Mic } from "pixelarticons/react/Mic";
import { BatteryFull } from "pixelarticons/react/BatteryFull";
import { Terminal } from "pixelarticons/react/Terminal";
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
import CompanyWebsite from "./CompanyWebsite.jsx";
import DesktopSettings from "./DesktopSettings.jsx";
import ProxyConnection from "../../../world/components/ProxyConnection.jsx";
import GhostkeyExplorer from "./GhostkeyExplorer.jsx";
import EmailBreakerApp from "./EmailBreakerApp.jsx";
import TerminalApp from "./TerminalApp.jsx";
import EmployeeDirectoryApp from "./EmployeeDirectoryApp.jsx";
import CharacterAppearanceEditor from "../../../identity/components/CharacterAppearanceEditor.jsx";
import { DEFAULT_DESKTOP_PREFERENCES, DESKTOP_WALLPAPERS } from "./desktopPreferences.js";

const BOOT_LINES = [
  "PhoenixBIOS Meridian Systems 4.1 Release",
  "Copyright 1974-1984 Meridian Enterprises",
  "All Rights Reserved",
  "BIOS version 4.1.74",
  "System ID = M_OS-00250850",
  "Build Time: 10/18/74 09:09:52",
  "CPU: Meridian 16-bit Processor 1066 MHz",
  "640K System RAM Passed",
  "254M Extended RAM Passed",
  "512K Cache SRAM Passed",
  "System BIOS shadowed",
  "Video BIOS shadowed",
  "Detecting primary IDE controller… OK",
  "Checking local file system… OK",
  "Initializing Meridian network stack…",
  "Private relay connection established",
  "Loading Meridian OS workstation services…",
  "System startup complete.",
];
const LOGIN_BOOT_STORAGE_KEY = "noir:computer:booted-this-login";

function hasBootedThisLogin() {
  try { return sessionStorage.getItem(LOGIN_BOOT_STORAGE_KEY) === "true"; }
  catch { return false; }
}

function markBootedThisLogin() {
  try { sessionStorage.setItem(LOGIN_BOOT_STORAGE_KEY, "true"); }
  catch { /* Boot remains available when browser storage is disabled. */ }
}
const PERSONAL_PROGRAMS = [
  { id: "web-browser", label: "WEB BROWSER", Icon: Target, state: "ONLINE", available: true },
  { id: "terminal", label: "TERMINAL", Icon: Terminal, state: "READY", available: true },
  { id: "police-website", label: "MERIDIAN POLICE", Icon: Files, state: "ONLINE", available: true },
  { id: "newspaper", label: "THE EVENING COURIER", Icon: Note, state: "ONLINE", available: true },
  { id: "personal-mail", label: "PERSONAL MAIL", Icon: Mail, state: "ONLINE", available: true },
  { id: "company-portal", label: "COMPANY PORTAL", Icon: MapIcon, state: "ONLINE", available: true },
  { id: "public-services", label: "CITY SERVICES DIRECTORY", Icon: Cross, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: MapIcon, state: "ONLINE", available: true },
];
const WEB_BROWSER_PROGRAM = PERSONAL_PROGRAMS[0];
const TERMINAL_PROGRAM = PERSONAL_PROGRAMS.find((program) => program.id === "terminal");
const SETTINGS_PROGRAM = { id: "settings", label: "SETTINGS", Icon: Brush, state: "READY", available: true };
const PROXY_PROGRAM = { id: "proxy", label: "ROGUE DNS", Icon: Wifi, state: "UNLISTED", available: true };
const GHOSTKEY_PROGRAM = { id: "ghostkey", label: "GHOSTKEY.EXE", Icon: Bug, state: "READY", available: true };
const EMAIL_BREAKER_PROGRAM = { id: "email-breaker", label: "EMAIL_BREAKER.EXE", Icon: Mail, state: "UNSIGNED", available: true };
const PERSON_SEARCH_PROGRAM = { id: "person-search", label: "PERSON_SEARCH_APP.EXE", Icon: User, state: "PRIVILEGED", available: true };
const PATIENT_SEARCH_PROGRAM = { id: "patient-search", label: "PATIENT_SEARCH_APP.EXE", Icon: Cross, state: "PRIVILEGED", available: true };
const GATE_RECORDS_PROGRAM = { id: "gate-records", label: "GATE_ACCESS_LOG_APP.EXE", Icon: MapIcon, state: "PRIVILEGED", available: true };
const EMPLOYEE_DIRECTORY_PROGRAM = { id: "employee-directory", label: "EMPLOYEE_DIRECTORY_APP.EXE", Icon: Users, state: "PRIVILEGED", available: true };
const RESIDENTIAL_SEARCH_PROGRAM = { id: "warrant-request", label: "SEARCH_WARRANT.EXE", Icon: Files, state: "PRIVILEGED", available: true };
const ARREST_WARRANT_PROGRAM = { id: "murder-accusation", label: "ARREST_WARRANT_APP.EXE", Icon: Files, state: "PRIVILEGED", available: true };
const DESKTOP_APP_IDS = ["newspaper", "settings", "character-appearance", "proxy", "terminal", "ghostkey", "email-breaker", "person-search", "patient-search", "gate-records", "employee-directory", "warrant-request", "murder-accusation"];
const PERSONAL_START_MENU_PROGRAMS = [
  WEB_BROWSER_PROGRAM,
  PERSONAL_PROGRAMS.find((program) => program.id === "newspaper"),
  SETTINGS_PROGRAM,
  ...PERSONAL_PROGRAMS.filter((program) => !["web-browser", "newspaper"].includes(program.id)),
];
const CHARACTER_APPEARANCE_PROGRAM = { id: "character-appearance", label: "USER APPEARANCE", Icon: User, state: "READY", available: true };
const MERIDIAN_ENTERPRISES_PROGRAM = { id: "meridian-enterprises", label: "MERIDIAN ENTERPRISES", Icon: MapIcon, state: "ONLINE", available: true };
const START_MENU_DESCRIPTIONS = {
  "web-browser": "Browse connected websites",
  terminal: "Run installed programs with commands",
  "police-website": "Public records and notices",
  newspaper: "Local news and city reports",
  "personal-mail": "Private correspondence",
  "company-portal": "National business directory",
  "public-services": "Hospitals and civic offices",
  "city-records": "People and places index",
  settings: "Desktop and account settings",
  "interpol": "Personnel search",
  emails: "Institutional mail archive",
  "patient-search": "Hospital patient records",
  "bank-records": "Financial records",
  "gate-records": "Transit access logs",
};

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
    GHOSTKEY_PROGRAM,
    ...offlinePrograms,
  ];
  if (facilityContext.kind === "bank") activePrograms = [{ id: "bank-records", label: "FINANCIAL RECORDS", Icon: Dollar, state: "ONLINE", available: true }];
  else if (facilityContext.kind === "gate") activePrograms = [GHOSTKEY_PROGRAM];
  else if (facilityContext.kind === "company") activePrograms = [GHOSTKEY_PROGRAM];
  if (activePrograms) return [...activePrograms, ...offlinePrograms];
  return PERSONAL_PROGRAMS;
}

function institutionalVariant(programId) {
  if (programId === "person-search") return "police";
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

function siteProviderIndex(programId) {
  if (["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports", "interpol", "person-search", "warrant-request", "murder-accusation"].includes(programId)) return 0;
  if (["newspaper", "personal-mail", "emails", "hospital-website", "patient-search", "hospital-patient-record"].includes(programId)) return 1;
  if (["meridian-enterprises", "gate-records", "bank-records"].includes(programId)) return 2;
  return [...programId].reduce((total, character) => total + character.charCodeAt(0), 0) % 3;
}

function networkSiteAddress(program, countryId, path = "") {
  if (program.siteAddress) return program.siteAddress;
  if (program.id === "company-website") {
    const address = program.website?.siteAddress || program.website?.SiteAddress;
    if (address) return `${address}${program.pageId && program.pageId !== "home" ? program.pageId : ""}`;
  }
  const providerIndex = siteProviderIndex(program.id);
  const providerHosts = ["atlasglobal", "northstar", "continental"];
  const countryHash = [...(countryId || "fr")].reduce((total, character) => total + character.charCodeAt(0), 0);
  const ip = `123.223.${40 + (countryHash % 180)}.${providerIndex * 2 + 1}`;
  const normalizedPath = path.startsWith("/") ? path.slice(1) : path;
  return `merid//${providerHosts[providerIndex]}.${ip}:4000/${normalizedPath}`;
}

function browserAddress(program, facilityContext, cityId) {
  const countryId = cityId?.split("/")[0] || "fr";
  if (program.id === "web-browser") return networkSiteAddress(program, countryId, `portal/${countryId}`);
  if (program.id === "newspaper") return networkSiteAddress(program, countryId);
  if (program.id === "meridian-enterprises") return networkSiteAddress(program, countryId);
  if (program.id === "gate-records") return networkSiteAddress(program, countryId, "gate/access-logs");
  if (program.id === "deep-link") return networkSiteAddress(program, countryId, program.linkId || "");
  if (["hospital-website", "patient-search", "hospital-patient-record"].includes(program.id)) {
    const institution = program.service?.address?.name || facilityContext?.label || "Hôpital Sainte-Claire";
    const city = institution.normalize("NFD").replace(/\p{Diacritic}/gu, "").replace(/\b(?:hopital|hospital)\b/iu, "").trim();
    const slug = cityId?.split("/").at(-1) || city.toLowerCase().replace(/[^a-z0-9]+/gu, "-").replace(/^-|-$/gu, "") || "sainte-claire";
    const recordPath = program.id === "hospital-patient-record"
      ? program.patient?.postMortem
        ? `/patient-services/post-mortem/${encodeURIComponent(program.patient.name)}`
        : `/patient-services/records/${encodeURIComponent(program.patient?.id || "record")}`
      : "";
    return networkSiteAddress(program, countryId, `hospital/${slug}${recordPath}`);
  }
  if (program.id === "company-website") return networkSiteAddress(program, countryId, `${program.website?.domain || "company"}${program.pageId && program.pageId !== "home" ? `/${program.pageId}` : ""}`);
  if (["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports"].includes(program.id)) {
    const paths = { "police-notices": "/notices", "police-about": "/about", "police-bureaus": "/district-bureaus", "police-policies": "/records-policy", "police-contact": "/contact", "crime-reports": "/public-records/crime-reports" };
    return networkSiteAddress(program, countryId, paths[program.id] || "");
  }
  if (program.id === "personal-mail") return networkSiteAddress(program, countryId);
  return networkSiteAddress(program, countryId, program.id);
}

export default function ComputerDesktop({ location, display, serverTime, timeZone, cityId, countryName, issueDay, facilityContext, onClose, onOpenCompany, onOpenCompanyStaff, onOpenService, onProxyConnected, onPrepareProxy, persistent = false, initialAccessAttemptId = "", initialFacilityAccessGranted = false, onOpenTelephone, onOpenWhiteboard, onLogout, accountUsername = "", character, onCharacterSaved, desktopPreferences = DEFAULT_DESKTOP_PREFERENCES, onDesktopPreferencesChange }) {
  const programs = programsFor(facilityContext);
  const isCompanyComputer = Boolean(facilityContext && facilityContext.kind !== "personal");
  const visibleDesktopPrograms = persistent ? [WEB_BROWSER_PROGRAM, PROXY_PROGRAM, TERMINAL_PROGRAM, PERSONAL_PROGRAMS.find((program) => program.id === "newspaper"), ...(!isCompanyComputer ? [SETTINGS_PROGRAM] : [])] : programs;
  const startMenuPrograms = persistent ? PERSONAL_START_MENU_PROGRAMS : programs;
  const facilityAccessGranted = !isCompanyComputer || initialFacilityAccessGranted;
  const [startMenuOpen, setStartMenuOpen] = useState(false);
  const [ghostkeyWarningOpen, setGhostkeyWarningOpen] = useState(false);
  const [securityWarningProgram, setSecurityWarningProgram] = useState(null);
  const [proxyBoot, setProxyBoot] = useState({ active: false, target: "" });
  const [characterPreviewVisual, setCharacterPreviewVisual] = useState(character?.visual);
  const [accessAttemptId, setAccessAttemptId] = useState(initialAccessAttemptId);
  const [windows, setWindows] = useState(() => {
    try {
      const saved = JSON.parse(sessionStorage.getItem("noir:computer:windows") || "null");
      if (!saved || Date.now() - saved.savedAt > 5 * 60 * 1000) return [];
      const restored = saved.windows.flatMap((item) => {
        if (persistent && [GHOSTKEY_PROGRAM.id, "email-breaker", PATIENT_SEARCH_PROGRAM.id].includes(item.programId)) return [];
        const program = item.programId === SETTINGS_PROGRAM.id && persistent
          ? SETTINGS_PROGRAM
          : item.programId === PROXY_PROGRAM.id && persistent
            ? PROXY_PROGRAM
            : item.programId === GHOSTKEY_PROGRAM.id && persistent
              ? GHOSTKEY_PROGRAM
              : item.programId === PERSON_SEARCH_PROGRAM.id && persistent
                ? PERSON_SEARCH_PROGRAM
            : programs.find((entry) => entry.id === item.programId);
        return program?.available ? [{ ...item, program }] : [];
      });
      return persistent
        ? restored.map((item) => ({ ...item, id: item.program.id === "newspaper" ? `newspaper:${item.zIndex}` : item.id.startsWith("noir-browser:") ? item.id : `noir-browser:${item.zIndex}` }))
        : restored;
    } catch { return []; }
  });
  const [booting, setBooting] = useState(() => !hasBootedThisLogin());
  const [bootLinesVisible, setBootLinesVisible] = useState(0);
  const [showMeridianLogo, setShowMeridianLogo] = useState(false);
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
    if (!booting) return undefined;
    markBootedThisLogin();
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reducedMotion) {
      setBootLinesVisible(BOOT_LINES.length);
      setShowMeridianLogo(true);
      bootTimerRef.current = window.setTimeout(finishBoot, 4300);
      return () => window.clearTimeout(bootTimerRef.current);
    }
    let line = 0;
    const reveal = () => {
      line += 1;
      setBootLinesVisible(line);
      if (line < BOOT_LINES.length) {
        bootTimerRef.current = window.setTimeout(reveal, 190);
      } else {
        bootTimerRef.current = window.setTimeout(() => {
          setShowMeridianLogo(true);
          bootTimerRef.current = window.setTimeout(finishBoot, 4300);
        }, 450);
      }
    };
    bootTimerRef.current = window.setTimeout(reveal, 150);
    return () => window.clearTimeout(bootTimerRef.current);
  }, [booting]);

  useEffect(() => {
    const timers = windows.filter((item) => item.loading).map((item) => window.setTimeout(() => {
      setWindows((previous) => previous.map((current) => current.id === item.id ? { ...current, loading: false } : current));
    }, 700));
    return () => timers.forEach((timer) => window.clearTimeout(timer));
  }, [windows]);

  useLayoutEffect(() => {
    const topWindow = [...windows].sort((left, right) => right.zIndex - left.zIndex)[0];
    if (topWindow) document.documentElement.dataset.noirActiveComputerWindow = topWindow.id;
    else delete document.documentElement.dataset.noirActiveComputerWindow;
    return () => { delete document.documentElement.dataset.noirActiveComputerWindow; };
  }, [windows]);

  useEffect(() => {
    setCharacterPreviewVisual(character?.visual);
  }, [character?.visual]);

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
    setBooting(false);
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
    document.documentElement.dataset.noirActiveComputerWindow = windowId;
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, zIndex } : item));
  }

  function setWebsiteNavigationPending(windowId, navigationPending) {
    if (navigationPending) {
      const body = windowRefs.current.get(windowId)?.querySelector(".computer-window__body");
      if (body) body.scrollTop = 0;
    }
    setWindows((previous) => previous.map((item) => item.id === windowId ? { ...item, navigationPending } : item));
  }

  function openHackedMailbox(account) {
    const zIndex = nextZIndexRef.current++;
    const offset = windows.length;
    const compact = window.matchMedia("(max-width: 42rem)").matches;
    const position = offset === 0 ? null : {
      left: Math.min(compact ? 8 + offset * 10 : 40 + offset * 28, compact ? 32 : 140),
      top: Math.min(compact ? 8 + offset * 10 : 32 + offset * 24, compact ? 32 : 120),
    };
    const id = `${persistent ? "noir-browser" : "personal-mail"}:${zIndex}`;
    const program = {
      ...PERSONAL_PROGRAMS.find((item) => item.id === "personal-mail"),
      label: `PERSONAL MAIL · ${account.name.toUpperCase()}`,
      state: "BREACHED",
      accessedViaGhost: true,
      hackedAccount: { id: account.id, name: account.name, email: account.email },
    };
    document.documentElement.dataset.noirActiveComputerWindow = id;
    setWindows((previous) => [...previous, { id, program, zIndex, position, history: [], loading: false }]);
  }

  function launchProgram(program, sourceWindowId = null, sourceState = null, skipGhostkeyWarning = false) {
    if (!program?.available) return;
    if (["ghostkey", "proxy", "email-breaker"].includes(program.id) && !skipGhostkeyWarning && !windows.some((item) => item.program.id === program.id)) {
      setSecurityWarningProgram(program);
      setGhostkeyWarningOpen(true);
      return;
    }
    if (persistent) {
      if (DESKTOP_APP_IDS.includes(program.id)) {
        const existingApp = windows.find((item) => item.program.id === program.id && (program.id !== "employee-directory" || item.program.companyRecord?.company?.id === program.companyRecord?.company?.id));
        if (existingApp) { focusWindow(existingApp.id); return; }
        const zIndex = nextZIndexRef.current++;
        const id = `${program.id}:${zIndex}`;
        document.documentElement.dataset.noirActiveComputerWindow = id;
        setWindows((previous) => [...previous, { id, program, zIndex, position: null }]);
        return;
      }
      const browserWindow = sourceWindowId
        ? windows.find((item) => item.id === sourceWindowId)
        : program.id === "web-browser" ? null : windows.find((item) => item.id.startsWith("noir-browser:"));
      if (browserWindow) {
        const zIndex = nextZIndexRef.current++;
        document.documentElement.dataset.noirActiveComputerWindow = browserWindow.id;
        setWindows((previous) => previous.map((item) => item.id === browserWindow.id
          ? { ...item, program, zIndex, loading: program.id !== "web-browser", navigationPending: false, history: program.id === "web-browser" ? [] : [...(item.history || []), sourceState ? { ...item.program, ...sourceState } : item.program] }
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
      const id = `noir-browser:${zIndex}`;
      document.documentElement.dataset.noirActiveComputerWindow = id;
      setWindows((previous) => [...previous, { id, program, zIndex, position, history: [], loading: program.id !== "web-browser" }]);
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
    document.documentElement.dataset.noirActiveComputerWindow = id;
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
    if (windowsStateRef.current.find((item) => item.id === windowId)?.program.id === "character-appearance") {
      setCharacterPreviewVisual(character?.visual);
    }
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
      <section className={`computer-frame ${persistent ? "computer-frame--persistent" : "computer-frame--modal"}`} data-desktop-theme={persistent ? desktopPreferences.theme : undefined} aria-label="Desktop">
      <div className="computer-bezel">
        <div className={`computer-screen${persistent ? " computer-screen--desktop" : ""}`} style={persistent ? { "--noir-desktop-wallpaper": `url("${DESKTOP_WALLPAPERS.find((wallpaper) => wallpaper.id === desktopPreferences.wallpaper)?.image || DESKTOP_WALLPAPERS[0].image}") center / cover` } : undefined} onMouseDown={(event) => {
          if (startMenuOpen && !event.target.closest(".computer-start-menu") && !event.target.closest(".computer-start")) setStartMenuOpen(false);
        }}>
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {booting && <div className={`computer-sequence computer-sequence--boot${showMeridianLogo ? " computer-sequence--logo" : ""}`} role="status" aria-live="polite">
            {showMeridianLogo ? <strong className="computer-sequence__brand" aria-label="Meridian OS"><span>M</span><span>OS</span></strong> : <div className="computer-sequence__bios">{BOOT_LINES.slice(0, bootLinesVisible).map((line, index) => <span key={`${line}-${index}`}>{line}</span>)}</div>}
          </div>}
          <header className={`computer-menubar${persistent ? " computer-menubar--desktop-hidden" : ""}`}>
            <p className="computer-brand"><span className="computer-brand__mark">M</span> MERID_OS <span className="computer-brand__divider">/</span> MERIDIAN ENTERPRISES</p>
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
                  <button type="button" onClick={() => launchProgram(PROXY_PROGRAM)}>ROGUE DNS</button>
              <button type="button" onClick={onOpenTelephone}>TELEPHONE</button>
              <button type="button" onClick={onOpenWhiteboard}>CASE BOARD</button>
            </nav>}
            {!persistent && <button ref={closeButtonRef} className="computer-modal__close" type="button" aria-label="Close computer" onClick={exitComputer}>×</button>}
          </header>

          {!booting && facilityAccessGranted && <div className="computer-desktop__content" inert={proxyBoot.active}>
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
                  aria-label={`${program.label}${DESKTOP_APP_IDS.includes(program.id) ? " application" : " website"}, ${program.state}`}
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
              const companySiteName = program.id === "company-website"
                ? program.website?.company?.name || program.website?.Company?.name || program.label
                : program.label;
              return <section
                ref={(element) => { if (element) windowRefs.current.set(id, element); else windowRefs.current.delete(id); }}
                className={`computer-window${position ? " computer-window--positioned" : ""}${persistent && !DESKTOP_APP_IDS.includes(program.id) ? " computer-window--website" : ""}${persistent && DESKTOP_APP_IDS.includes(program.id) ? " computer-window--app" : ""}${program.id === "ghostkey" ? " computer-window--ghostkey" : ""}${program.id === "terminal" ? " computer-window--terminal" : ""}${program.accessedViaGhost ? " computer-window--ghost-compromised" : ""}${program.id === "settings" ? " computer-window--settings" : ""}${program.id === "character-appearance" ? " computer-window--character-appearance" : ""}${program.id === "proxy" ? " computer-window--proxy" : ""}`}
                data-site-variant={institutionalVariant(program.id)}
                data-window-id={id}
                key={id}
                style={{ ...(position || {}), zIndex, ...(program.id === "character-appearance" ? { height: "min(48vh, 30rem)" } : {}) }}
                aria-label={`${companySiteName} ${DESKTOP_APP_IDS.includes(program.id) ? "application" : "website"} window`}
                onMouseDown={() => {
                  focusWindow(id);
                  if (program.id === "terminal") window.requestAnimationFrame(() => windowRefs.current.get(id)?.querySelector(".terminal-app__prompt input")?.focus({ preventScroll: true }));
                }}
                onClickCapture={() => {
                  if (program.id === "terminal") window.requestAnimationFrame(() => windowRefs.current.get(id)?.querySelector(".terminal-app__prompt input")?.focus({ preventScroll: true }));
                }}
              >
                <header className="computer-window__titlebar" onPointerDown={startWindowDrag} onPointerMove={moveWindow} onPointerUp={endWindowDrag} onPointerCancel={endWindowDrag}>
                  <span>{persistent && !["web-browser", ...DESKTOP_APP_IDS].includes(program.id) && <button type="button" className="computer-browser-back" aria-label="Go back" title="Go back" onPointerDown={(event) => event.stopPropagation()} onClick={() => goBackInBrowser(id)}>←</button>}<span className="computer-status-light" /> {program.id === "web-browser" ? "WEB BROWSER" : program.id === "settings" ? "MERID_OS SETTINGS" : program.id === "proxy" ? "ROGUE DNS" : companySiteName}</span>
                  <button type="button" className="computer-window__close" aria-label={`Close ${companySiteName}`} onClick={() => closeWindow(id)}>×</button>
                </header>
                <div className={`computer-browser-address${DESKTOP_APP_IDS.includes(program.id) ? " computer-browser-address--app" : ""}`} aria-label="Website address"><span aria-hidden="true">▣</span><span>{program.id === "settings" ? "control://desktop" : program.id === "proxy" ? "control://rogue-dns" : program.id === "ghostkey" ? "control://ghostkey" : program.id === "terminal" ? "control://terminal" : browserAddress(program, facilityContext, cityId)}</span><span className="computer-browser-address__secure">{program.id === "proxy" ? "UNLISTED NETWORK" : "PRIVATE RELAY"}</span></div>
                <div className={`computer-window__body${persistent && !DESKTOP_APP_IDS.includes(program.id) ? " computer-window__body--website" : ""}${desktopWindow.loading || desktopWindow.navigationPending ? " computer-window__body--loading" : ""}`} aria-busy={desktopWindow.loading || desktopWindow.navigationPending}>
                  {desktopWindow.loading ? <div className="computer-browser-loading" role="status" aria-label="Loading webpage" /> : <>
                  {program.id === "settings"
                    ? <DesktopSettings preferences={desktopPreferences} accountUsername={accountUsername} character={character} characterPreviewVisual={characterPreviewVisual} location={location} onEditCharacter={() => launchProgram(CHARACTER_APPEARANCE_PROGRAM)} onChange={(patch) => onDesktopPreferencesChange?.(patch)} onLogout={onLogout} />
                    : program.id === "character-appearance"
                    ? <CharacterAppearanceEditor windowed character={character} onPreview={setCharacterPreviewVisual} onClose={() => closeWindow(id)} onSaved={(savedCharacter) => { onCharacterSaved?.(savedCharacter); closeWindow(id); setCharacterPreviewVisual(savedCharacter.visual); }} />
                    : program.id === "terminal"
                      ? <TerminalApp onRun={(programId) => launchProgram(programId === "ghostkey" ? GHOSTKEY_PROGRAM : EMAIL_BREAKER_PROGRAM)} />
                    : program.id === "ghostkey"
                      ? <GhostkeyExplorer countryId={cityId?.split("/")[0] || ""} countryName={countryName} onOpenPersonSearch={(siteAddress) => launchProgram({ ...PERSON_SEARCH_PROGRAM, siteAddress, accessedViaGhost: true }, id)} onOpenPatientSearch={(siteAddress) => launchProgram({ ...PATIENT_SEARCH_PROGRAM, siteAddress, accessedViaGhost: true }, id)} onOpenGateRecords={(siteAddress) => launchProgram({ ...GATE_RECORDS_PROGRAM, siteAddress, accessedViaGhost: true }, id)} onOpenEmployeeDirectory={(companyRecord) => launchProgram({ ...EMPLOYEE_DIRECTORY_PROGRAM, companyRecord, siteAddress: companyRecord.siteAddress, accessedViaGhost: true }, id)} onOpenWarrantRequest={(siteAddress) => launchProgram({ ...RESIDENTIAL_SEARCH_PROGRAM, siteAddress, accessedViaGhost: true }, id)} onOpenArrestWarrant={(siteAddress) => launchProgram({ ...ARREST_WARRANT_PROGRAM, siteAddress, accessedViaGhost: true }, id)} />
                      : program.id === "email-breaker"
                        ? <EmailBreakerApp serverTime={serverTime} onOpenHackedMailbox={openHackedMailbox} />
                      : program.id === "person-search"
                        ? <PoliceWebsite page="person-search" navigate={(destination) => launchProgram(destination, id)} searchId={`police-site-search-${id}`}><WorldDirectory key={`${id}:person-search`} program="interpol" countryName={countryName} cityId={cityId} serverTime={serverTime} timeZone={timeZone} onPrint={requestPrint} /></PoliceWebsite>
                      : program.id === "employee-directory"
                        ? <EmployeeDirectoryApp companyRecord={program.companyRecord} onOpenCompanyStaff={onOpenCompanyStaff} onPrint={requestPrint} />
                    : program.id === "proxy"
                    ? <ProxyConnection
                        world={{ city: location?.split(",")[0], cityId, country: countryName }}
                        hasOpenBrowser={windows.some((item) => !["proxy", "settings"].includes(item.program.id))}
                        onConnected={onProxyConnected}
                        onRouteChanged={() => setWindows((previous) => previous.filter((item) => ["proxy", "settings"].includes(item.program.id)))}
                        onBusyChange={(active, node) => setProxyBoot({ active, target: node ? `${node.name}, ${node.country}` : "" })}
                      />
                    : program.id === "web-browser"
                    ? <LegacyCountryPortal
                        countryName={countryName}
                        location={location}
                        cityId={cityId}
                        programs={programs}
                        navigate={(destination) => launchProgram(destination, id)}
                        onOpenCompany={onOpenCompany}
                        initialSearch={program.searchTerm}
                        initialCompanies={program.searchResults}
                        onNavigateCompanyWebsite={(site, searchState) => launchProgram({ id: "company-website", label: site.company.name, website: site, pageId: "home", state: "ONLINE", available: true }, id, { searchTerm: searchState.query, searchResults: searchState.companies })}
                      />
                    : ["police-website", "police-notices", "police-about", "police-bureaus", "police-policies", "police-contact", "crime-reports"].includes(program.id)
                      ? <PoliceWebsite page={program.id} navigate={(destination) => launchProgram(destination, id)}>{program.id === "crime-reports" && <CaseFiles key={`${id}:crime-reports:${program.searchTerm || ""}`} initialSearch={program.searchTerm} cityId={cityId} issueDay={issueDay} timeZone={timeZone} />}</PoliceWebsite>
                      : program.id === "deep-link"
                        ? <InstitutionalWebsiteFrame variant="civic" organizationName={program.label}><DeepLinkWebsite linkId={program.linkId} /></InstitutionalWebsiteFrame>
                      : program.id === "newspaper"
                        ? <NewspaperWindow cityId={cityId} issueDay={issueDay} embedded />
                      : program.id === "meridian-enterprises"
                        ? <MeridianEnterprisesWebsite location={location} countryName={countryName} />
                      : program.id === "hospital-website"
                        ? <HospitalWebsite service={program.service} onConnect={onOpenService} />
                      : program.id === "company-website"
                        ? <CompanyWebsite
                            site={program.website}
                            pageId={program.pageId}
                            onNavigate={(pageId) => launchProgram({ ...program, pageId, label: `${program.website.company.name} · ${pageId.toUpperCase()}` }, id)}
                          />
                      : ["warrant-request", "murder-accusation"].includes(program.id)
                        ? <PoliceWebsite page={program.id} navigate={(destination) => launchProgram(destination, id)} searchId={`police-site-search-${id}`}>
                            {program.id === "warrant-request"
                              ? <ResidentialSearchWindow key={id} cityId={cityId} stationLocationId={facilityContext?.locationId} standalone onClose={() => closeWindow(id)} />
                              : <AccusationWindow key={id} cityId={cityId} standalone onClose={() => closeWindow(id)} />}
                          </PoliceWebsite>
                      : <InstitutionalWebsiteFrame
                        variant={institutionalVariant(program.id)}
                        organizationName={["patient-search", "hospital-patient-record"].includes(program.id) ? facilityContext?.label : ""}
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
                                ? <PlayerMailbox key={id} serverTime={serverTime} timeZone={timeZone} onPrint={requestPrint} hackedAccount={program.hackedAccount} />
                                : program.id === "public-services"
                                  ? <PublicServices key={id} cityId={cityId} countryName={countryName} onOpenService={onOpenService} onOpenHospitalWebsite={(service) => launchProgram({ id: "hospital-website", label: service.address.name, Icon: Cross, state: "ONLINE", available: true, service }, id)} />
                                  : program.id === "warrant-request"
                                    ? <ResidentialSearchWindow key={id} cityId={cityId} stationLocationId={facilityContext?.locationId} onClose={() => closeWindow(id)} />
                                    : program.id === "murder-accusation"
                                      ? <AccusationWindow key={id} cityId={cityId} onClose={() => closeWindow(id)} />
                                          : <WorldDirectory key={`${id}:${program.id}`} program={program.id} facilityLocationId={facilityContext?.locationId} accessAttemptId={accessAttemptId} cityId={cityId} countryName={countryName} serverTime={serverTime} timeZone={timeZone} onPrint={requestPrint} onOpenCompany={onOpenCompany} onBeginCompanyNavigation={program.id === "company-portal" ? () => setWebsiteNavigationPending(id, true) : undefined} onCancelCompanyNavigation={program.id === "company-portal" ? () => setWebsiteNavigationPending(id, false) : undefined} initialSearch={program.searchTerm} initialCompanies={program.searchResults} onNavigateCompanyWebsite={(site, searchState) => launchProgram({ id: "company-website", label: site.company.name, website: site, pageId: "home", state: "ONLINE", available: true }, id, { searchTerm: searchState.query, searchResults: searchState.companies })} />}
                    {desktopWindow.navigationPending && <div className="computer-browser-loading" role="status" aria-label="Loading webpage" />}
                      </InstitutionalWebsiteFrame>}
                  </>}
                </div>
              </section>;
            })}
            {printDialog && <section className="page-print-dialog" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} role="dialog" aria-modal="true" aria-labelledby="page-print-title" onMouseDown={(event) => { if (event.target === event.currentTarget) setPrintDialog(null); }}>
              <div className="page-print-dialog__panel">
                <header className="page-print-dialog__titlebar"><span><Printer aria-hidden="true" /> WEB BROWSER · PRINT</span><button type="button" aria-label="Close print dialog" onClick={() => setPrintDialog(null)}>×</button></header>
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
            {startMenuOpen && persistent && <nav className="computer-start-menu" style={{ zIndex: Math.max(20, ...windows.map((item) => item.zIndex + 1)) }} aria-label="Start menu">
              <header className="computer-start-menu__header">
                <span className="computer-start-menu__brand" aria-hidden="true">M</span>
                <span><strong>MERID_OS</strong><small>PERSONAL COMPUTER</small></span>
              </header>
              <section className="computer-start-menu__programs" aria-labelledby="start-menu-programs-title">
                <h2 id="start-menu-programs-title">PROGRAMS</h2>
                {startMenuPrograms.map((program) => <button key={program.id} type="button" onClick={() => { launchProgram(program); setStartMenuOpen(false); }}>
                  <span className="computer-start-menu__icon" aria-hidden="true"><program.Icon /></span>
                  <span className="computer-start-menu__copy"><strong>{program.label}</strong><small>{START_MENU_DESCRIPTIONS[program.id] || "Open this system program"}</small></span>
                  <span className="computer-start-menu__arrow" aria-hidden="true">›</span>
                </button>)}
              </section>
              <section className="computer-start-menu__tools" aria-labelledby="start-menu-tools-title">
                <h2 id="start-menu-tools-title">SYSTEM TOOLS</h2>
                <div>
                  <button type="button" onClick={() => { launchProgram(PROXY_PROGRAM); setStartMenuOpen(false); }}><Wifi aria-hidden="true" /><span>ROGUE DNS</span></button>
                  <button type="button" onClick={() => { onOpenTelephone?.(); setStartMenuOpen(false); }}><Phone aria-hidden="true" /><span>TELEPHONE</span></button>
                  <button type="button" onClick={() => { onOpenWhiteboard?.(); setStartMenuOpen(false); }}><Files aria-hidden="true" /><span>CASE BOARD</span></button>
                </div>
              </section>
            </nav>}
          </div>}

          <footer className={`computer-taskbar${persistent ? " computer-taskbar--desktop" : ""}`} inert={proxyBoot.active}>
            <div className="computer-taskbar__left">
              {persistent && <button type="button" className="computer-start" aria-expanded={startMenuOpen} onClick={() => setStartMenuOpen((open) => !open)}><span className="computer-brand__mark">M</span> START</button>}
              {!persistent && <span className="computer-taskbar__system"><span className="computer-brand__mark">M</span> {isCompanyComputer ? "WORKSTATION SESSION" : "PERSONAL SESSION"}</span>}
              {windows.map((item) => <button key={item.id} className={`computer-taskbar__active${item.zIndex === Math.max(...windows.map((windowItem) => windowItem.zIndex)) ? " computer-taskbar__active--focused" : ""}`} type="button" onClick={() => focusWindow(item.id)}>{item.program.label}</button>)}
            </div>
            {persistent ? <div className="computer-taskbar__tray" role="group" aria-label="System status">
              <span className="computer-taskbar__status-icon" role="img" aria-label="Wi-Fi connected" title="Wi-Fi connected"><Wifi aria-hidden="true" /></span>
              <span className="computer-taskbar__status-icon" role="img" aria-label="Microphone available" title="Microphone available"><Mic aria-hidden="true" /></span>
              <span className="computer-taskbar__status-icon" role="img" aria-label="Battery 92 percent" title="Battery 92%"><BatteryFull aria-hidden="true" /></span>
              <span className="computer-taskbar__clock">{display?.time}</span>
            </div>
              : <button className="computer-logout" type="button" onClick={exitComputer}>Exit <span aria-hidden="true">↗</span></button>}
          </footer>
          {proxyBoot.active && <div className="computer-sequence computer-sequence--proxy" role="status" aria-label="Rogue DNS route switch in progress">
            <p className="computer-sequence--proxy__heading">ROGUE DNS · ROUTE HANDSHAKE</p>
            <Typewriter className="computer-sequence__typewriter" texts={[
              "rogue-dns://bootstrap --unregistered",
              "[cache] flushing resolver records ...... OK",
              "[key] rotating ephemeral route token .. OK",
              "[udp:53] opening shadow tunnel ........ ACK",
              `[exit] ${proxyBoot.target || "UNKNOWN NODE"} · VERIFYING`,
              "[dns] recursive chain established ...... OK",
            ]} speed={10} lineDelay={80} />
          </div>}
          {ghostkeyWarningOpen && <div className="ghostkey__error-backdrop computer-security-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setGhostkeyWarningOpen(false); }}>
            <section className="ghostkey__error-dialog" role="alertdialog" aria-modal="true" aria-labelledby="ghostkey-warning-title" aria-describedby="ghostkey-warning-message">
              <header className="ghostkey__error-titlebar"><span><i aria-hidden="true">!</i> MERID_OS — APPLICATION SECURITY</span><button type="button" aria-label="Close security warning" onClick={() => setGhostkeyWarningOpen(false)}>×</button></header>
              <div className="ghostkey__error-content"><strong id="ghostkey-warning-title">Application not authorized</strong><p id="ghostkey-warning-message">MeridOS cannot verify the publisher of {securityWarningProgram?.label || "this application"}. This application may bypass network access controls.</p></div>
              <footer className="ghostkey__warning-actions"><button type="button" onClick={() => { setGhostkeyWarningOpen(false); setSecurityWarningProgram(null); }}>CANCEL</button><button type="button" autoFocus onClick={() => { const program = securityWarningProgram; setGhostkeyWarningOpen(false); setSecurityWarningProgram(null); if (program?.id === "proxy") onPrepareProxy?.(); launchProgram(program, null, null, true); }}>RUN ANYWAY</button></footer>
            </section>
          </div>}
        </div>
      </div>
      {!persistent && <div className="computer-indicator" aria-hidden="true"><span /> POWER</div>}
      </section>
    </div>
  );
}
