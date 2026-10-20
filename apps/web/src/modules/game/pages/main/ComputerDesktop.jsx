import { useEffect, useRef, useState } from "react";
import { Database } from "pixelarticons/react/Database";
import { Files } from "pixelarticons/react/Files";
import { Map } from "pixelarticons/react/Map";
import { Phone } from "pixelarticons/react/Phone";
import CaseFiles from "../../../cases/pages/files/CaseFiles.jsx";
import WorldDirectory from "../../../world/pages/directory/WorldDirectory.jsx";

const programs = [
  { id: "interpol", label: "INTERPOL", Icon: Database, state: "ONLINE", available: true },
  { id: "case-files", label: "CASE FILES", Icon: Files, state: "ONLINE", available: true },
  { id: "city-records", label: "CITY RECORDS", Icon: Map, state: "ONLINE", available: true },
  { id: "telephone", label: "TELEPHONE", Icon: Phone, state: "OFFLINE", available: false },
];

export default function ComputerDesktop({ location, display, serverTime, onClose }) {
  const [openProgram, setOpenProgram] = useState(null);
  const [selectedProgram, setSelectedProgram] = useState(null);
  const modalRef = useRef(null);
  const closeButtonRef = useRef(null);

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
  }

  return (
    <div
      className="computer-modal"
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

          <div className="computer-desktop__content">
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
              <section className="computer-window" aria-label={`${openProgram.label} application window`}>
                <header className="computer-window__titlebar">
                  <span><span className="computer-status-light" /> {openProgram.label} DATABASE</span>
                  <button type="button" className="computer-window__close" aria-label="Close application" onClick={() => setOpenProgram(null)}>×</button>
                </header>
                <div className="computer-window__body">
                  {openProgram.id === "case-files"
                    ? <CaseFiles key={openProgram.id} />
                    : <WorldDirectory key={openProgram.id} program={openProgram.id} />}
                </div>
              </section>
            )}
          </div>

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
