import { useEffect, useRef, useState } from "react";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";

const ACCESS_LINES = [
  "LOADING BYPASSDECRYPT.EXE",
  "SCANNING AUTHENTICATION LAYERS…",
  "RECONSTRUCTING LOCAL CREDENTIAL TOKEN…",
  "DECRYPTION COMPLETE · LOCAL SESSION UNLOCKED",
  "ACCESS GRANTED!",
];

export default function LocalTerminal({ facility, onAccessGranted, onClose, variant = "monitor" }) {
  const [lines, setLines] = useState([
    `${facility.label} · LOCAL ACCESS NODE`,
    "PASSWORD REQUIRED!",
    "HELP · type help to see available commands.",
    "TRY: run bypassdecrypt.exe",
  ]);
  const [command, setCommand] = useState("");
  const [busy, setBusy] = useState(false);
  const [showAccess, setShowAccess] = useState(false);
  const inputRef = useRef(null);
  const outputRef = useRef(null);
  const timerRef = useRef(null);

  useEffect(() => { inputRef.current?.focus(); }, []);
  useEffect(() => { if (outputRef.current) outputRef.current.scrollTop = outputRef.current.scrollHeight; }, [lines, busy]);
  useEffect(() => () => window.clearTimeout(timerRef.current), []);

  function submitCommand(event) {
    event.preventDefault();
    const entered = command.trim();
    if (!entered || busy) return;
    setCommand("");
    setLines((current) => [...current, `> ${entered}`]);
    const normalized = entered.toLowerCase();
    if (normalized === "exit") { onClose?.(); return; }
    if (normalized === "clear") {
      setLines([`${facility.label} · LOCAL ACCESS NODE`, "PASSWORD REQUIRED!", "HELP · type help to see available commands.", "TRY: run bypassdecrypt.exe"]);
      return;
    }
    if (normalized === "help") {
      setLines((current) => [...current, "COMMANDS", "run bypassdecrypt.exe · unlock this computer", "clear · clear terminal output", "exit · close terminal"]);
      return;
    }
    if (normalized !== "run bypassdecrypt.exe") {
      setLines((current) => [...current, "ACCESS DENIED · UNKNOWN COMMAND", "TRY: run bypassdecrypt.exe"]);
      return;
    }
    setBusy(true);
    setShowAccess(true);
  }

  function finishAccess() {
    setBusy(false);
    timerRef.current = window.setTimeout(() => onAccessGranted?.(), 450);
  }

  const terminalPanel = (
    <div className={`local-terminal${variant === "app" ? " local-terminal--app" : ""}`} aria-busy={busy}>
      <header className="local-terminal__header">
        <div><p>MERIDIAN GLOBAL · SECURE NODE</p><h2>{facility.label}</h2><span>{facility.city?.toUpperCase()} · ON-SITE SESSION</span></div>
        {variant !== "app" && <button type="button" aria-label="Close terminal" disabled={busy} onClick={onClose}>×</button>}
      </header>
      <div ref={outputRef} className="local-terminal__output" role="log" aria-live="polite">
        {lines.map((line, index) => <p key={`${index}:${line}`} className={line.startsWith("> ") ? "local-terminal__command" : /ACCESS GRANTED|UNLOCKED/.test(line) ? "local-terminal__success" : /HELP|COMMANDS|TRY:|clear ·|exit ·|run bypassdecrypt/.test(line) ? "local-terminal__help" : undefined}>{line}</p>)}
        {showAccess && <div className="local-terminal__activity"><Typewriter texts={ACCESS_LINES} speed={23} lineDelay={180} onComplete={finishAccess} /></div>}
      </div>
      <form className="local-terminal__prompt" onSubmit={submitCommand}>
        <label htmlFor="local-terminal-command"><span aria-hidden="true">&gt;</span><input ref={inputRef} id="local-terminal-command" autoComplete="off" autoCapitalize="off" spellCheck="false" maxLength={160} value={command} onChange={(event) => setCommand(event.target.value)} placeholder="run bypassdecrypt.exe" disabled={busy} /></label>
        <button type="submit" disabled={busy || !command.trim()}>{busy ? "RUNNING…" : "EXECUTE"}</button>
      </form>
    </div>
  );

  if (variant === "app") return (
    <div className="computer-app-modal" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) onClose?.(); }}>
      <section className="computer-window computer-window--terminal" role="dialog" aria-modal="true" aria-label={`${facility.label} local terminal`}>
        <header className="computer-window__titlebar"><span><span className="computer-status-light" /> LOCAL TERMINAL · BYPASSDECRYPT</span><button type="button" className="computer-window__close" aria-label="Close terminal" disabled={busy} onClick={onClose}>×</button></header>
        {terminalPanel}
      </section>
    </div>
  );

  return (
    <div className="modal-overlay local-terminal-overlay" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) onClose?.(); }}>
      <section className="computer-frame computer-frame--modal local-terminal-frame" role="dialog" aria-modal="true" aria-label={`${facility.label} local terminal`}>
        <div className="computer-bezel"><div className="computer-screen local-terminal-screen">
          <div className="computer-screen__glass texture-scanlines" aria-hidden="true" />
          {terminalPanel}
        </div></div>
      </section>
    </div>
  );
}
