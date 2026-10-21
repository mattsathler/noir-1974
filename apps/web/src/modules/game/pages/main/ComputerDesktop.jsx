import { useCallback, useEffect, useRef, useState } from "react";
import { Building } from "pixelarticons/react/Building";
import { Files } from "pixelarticons/react/Files";
import { Map } from "pixelarticons/react/Map";
import { Mail } from "pixelarticons/react/Mail";
import { Phone } from "pixelarticons/react/Phone";
import { Globe } from "pixelarticons/react/Globe";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";
import CaseFiles from "../../../cases/pages/files/CaseFiles.jsx";
import WorldDirectory from "../../../world/pages/directory/WorldDirectory.jsx";
import EmailInbox from "../../../mail/pages/inbox/EmailInbox.jsx";

const programs = [
  { id: "interpol", label: "PERSON SEARCH", Icon: Globe, state: "ONLINE", available: true },
  { id: "case-files", label: "CASE FILES", Icon: Files, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: Map, state: "ONLINE", available: true },
  { id: "bank-records", label: "MERIDIAN BANK", Icon: Building, state: "RESTRICTED", available: true },
  { id: "gate-records", label: "GATE NETWORK", Icon: Globe, state: "RESTRICTED", available: true },
  { id: "email", label: "MERIDIAN MAIL", Icon: Mail, state: "RESTRICTED", available: true },
  { id: "telephone", label: "TELEPHONE", Icon: Phone, state: "OFFLINE", available: false },
];

export default function ComputerDesktop({ location, display, serverTime, timeZone, onClose }) {
  const [openProgram, setOpenProgram] = useState(null);
  const [selectedProgram, setSelectedProgram] = useState(null);
  const [booting, setBooting] = useState(true);
  const [hacking, setHacking] = useState(false);
  const [windowPosition, setWindowPosition] = useState(null);
  const modalRef = useRef(null);
  const closeButtonRef = useRef(null);
  const windowRef = useRef(null);
  const dragRef = useRef(null);

  useEffect(() => {
    const timer = window.setTimeout(() => setBooting(false), 2900);
    return () => window.clearTimeout(timer);
  }, []);

  const runHack = useCallback(async (loadRequest) => {
    setHacking(true);
    try {
      const request = typeof loadRequest === "function" ? loadRequest() : Promise.resolve(loadRequest);
      const [result] = await Promise.all([
        request,
        new Promise((resolve) => window.setTimeout(resolve, 2800)),
      ]);
      return result;
    } finally {
      setHacking(false);
    }
  }, []);

  function startWindowDrag(event) {
    if (event.target.closest("button")) return;
    const parent = event.currentTarget.closest(".computer-desktop__content");
    const modal = windowRef.current;
    if (!parent || !modal) return;
    const bounds = parent.getBoundingClientRect();
    const rect = modal.getBoundingClientRect();
    const left = rect.left - bounds.left;
    const top = rect.top - bounds.top;
    setWindowPosition({ left, top });
    dragRef.current = { pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, left, top };
    event.currentTarget.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function moveWindow(event) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const parent = windowRef.current?.parentElement;
    if (!parent || !windowRef.current) return;
    const maxLeft = Math.max(0, parent.clientWidth - windowRef.current.offsetWidth);
    const maxTop = Math.max(0, parent.clientHeight - windowRef.current.offsetHeight);
    setWindowPosition({
      left: Math.max(0, Math.min(maxLeft, drag.left + event.clientX - drag.startX)),
      top: Math.max(0, Math.min(maxTop, drag.top + event.clientY - drag.startY)),
    });
  }

  function endWindowDrag(event) {
    if (dragRef.current?.pointerId === event.pointerId) dragRef.current = null;
  }

  useEffect(() => {
    const previousFocus = document.activeElement;
    closeButtonRef.current?.focus();

    function handleModalKeys(event) {
      if (event.key === "Escape") {
        onClose?.();
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

  function launchProgram(program) {
    if (!program.available) return;
    setSelectedProgram(program.id);
    setOpenProgram(program);
    setWindowPosition(null);
  }

  return (
    <div
      className="modal-overlay computer-modal"
      ref={modalRef}
      role="dialog"
      aria-modal="true"
      aria-label="Detective computer desktop"
      onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}
    >
      <section className="computer-frame computer-frame--modal" aria-label="Desktop applications">
      <div className="computer-bezel">
        <div className="computer-screen">
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {booting && <div className="computer-sequence computer-sequence--boot" role="status">
            <span><Typewriter text="NOIR PERSONAL TERMINAL · 1974" speed={32} /></span>
            <span><Typewriter text="CHECKING RELAY CONNECTION…" speed={32} startDelay={220} /></span>
            <span><Typewriter text="LYON NODE FOUND" speed={32} startDelay={460} /></span>
            <span><Typewriter text="PRIVATE SYSTEM READY" speed={32} startDelay={700} /> <i>_</i></span>
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
            <button ref={closeButtonRef} className="computer-modal__close" type="button" aria-label="Close computer" onClick={onClose}>×</button>
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
                  className={`computer-program${selectedProgram === program.id ? " computer-program--selected" : ""}`}
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

            {openProgram && (
              <section ref={windowRef} className={`computer-window${windowPosition ? " computer-window--positioned" : ""}`} style={windowPosition || undefined} aria-label={`${openProgram.label} application window`}>
                <header className="computer-window__titlebar" onPointerDown={startWindowDrag} onPointerMove={moveWindow} onPointerUp={endWindowDrag} onPointerCancel={endWindowDrag}>
                  <span><span className="computer-status-light" /> {openProgram.label} DATABASE</span>
                  <button type="button" className="computer-window__close" aria-label="Close application" onClick={() => setOpenProgram(null)}>×</button>
                </header>
                <div className="computer-window__body">
                  {openProgram.id === "case-files"
                    ? <CaseFiles key={openProgram.id} />
                    : openProgram.id === "email"
                      ? <EmailInbox key={openProgram.id} serverTime={serverTime} timeZone={timeZone} runHack={runHack} />
                      : <WorldDirectory key={openProgram.id} program={openProgram.id} serverTime={serverTime} timeZone={timeZone} runHack={runHack} />}
                </div>
              </section>
            )}
            {hacking && <section className="computer-hack-modal" role="status" aria-label="Secure data access in progress">
              <header><span>MERIDIAN SECURE LINK</span><span className="computer-status-light" /></header>
              <div><p>ESTABLISHING SECURE TUNNEL</p><p>ROUTING THROUGH PRIVATE RELAY…</p><p>DECRYPTING REQUESTED FILE</p><div className="computer-sequence__progress"><i /></div></div>
            </section>}
          </div>}

          <footer className="computer-taskbar">
            <div className="computer-taskbar__left">
              <span className="computer-taskbar__system"><span className="computer-brand__mark">N</span> SYSTEM</span>
              {openProgram && <span className="computer-taskbar__active">{openProgram.label}</span>}
            </div>
            <button className="computer-logout" type="button" onClick={onClose}>Exit <span aria-hidden="true">↗</span></button>
          </footer>
        </div>
      </div>
      <div className="computer-indicator" aria-hidden="true"><span /> POWER</div>
      </section>
    </div>
  );
}
