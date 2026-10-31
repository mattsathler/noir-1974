import { useEffect, useRef, useState } from "react";

const COMMANDS = {
  "ghostkey.exe": {
    id: "ghostkey",
    label: "GHOSTKEY.EXE",
    art: [
      "      .-========-.",
      "      | GHOSTKEY |",
      "  .---|  .--.   |---.",
      " /    | (o  o)  |    \\",
      "      |  '--'   |",
      "      '----||----'",
      "           ||",
      "       NETWORK ACCESS",
    ],
  },
  "mailbreaker.exe": {
    id: "email-breaker",
    label: "EMAIL_BREAKER.EXE",
    art: [
      "     .------------.",
      "     |  MAILBOX   |",
      "     |   .----.   |",
      "  @--|--|XXXX|---+",
      "     |  '----'   |",
      "     '-----||-----'",
      "           \\\\",
      "        MAIL BREACH",
    ],
  },
  "emailbreaker.exe": { id: "email-breaker", label: "EMAIL_BREAKER.EXE", art: [
    "     .------------.", "     |  MAILBOX   |", "     |   .----.   |", "  @--|--|XXXX|---+", "     |  '----'   |", "     '-----||-----'", "           \\\\", "        MAIL BREACH",
  ] },
  "email_breaker.exe": { id: "email-breaker", label: "EMAIL_BREAKER.EXE", art: [
    "     .------------.", "     |  MAILBOX   |", "     |   .----.   |", "  @--|--|XXXX|---+", "     |  '----'   |", "     '-----||-----'", "           \\\\", "        MAIL BREACH",
  ] },
};

export default function TerminalApp({ onRun }) {
  const [command, setCommand] = useState("");
  const [history, setHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [launchPending, setLaunchPending] = useState(false);
  const launchTimer = useRef(null);
  const outputRef = useRef(null);
  const [output, setOutput] = useState([
    "Meridian OS [Version 4.1.74]",
    "(C) Copyright Meridian Enterprises. 1974-1984.",
    "Type HELP to see available commands.",
    "",
  ]);

  useEffect(() => () => window.clearTimeout(launchTimer.current), []);

  useEffect(() => {
    const outputElement = outputRef.current;
    const body = outputElement?.closest(".computer-window__body");
    const appWindow = outputElement?.closest(".computer-window");
    window.requestAnimationFrame(() => {
      if (body) body.scrollTop = body.scrollHeight;
      if (appWindow) appWindow.scrollTop = appWindow.scrollHeight;
    });
  }, [output]);

  function submit(event) {
    event.preventDefault();
    const entered = command.trim();
    if (!entered || launchPending) return;
    setCommand("");
    setHistory((previous) => [...previous, entered]);
    setHistoryIndex(-1);

    if (entered.toLowerCase() === "help") {
      setOutput((previous) => [...previous, `> ${entered}`, "Available commands:", "  HELP                 Show this command list", "  RUN GHOSTKEY.EXE     Start Ghostkey", "  RUN MAILBREAKER.EXE  Start Email Breaker", ""]);
      return;
    }

    const match = entered.match(/^run\s+(.+)$/i);
    const program = match ? COMMANDS[match[1].trim().toLowerCase()] : null;
    if (!program) {
      setOutput((previous) => [...previous, `> ${entered}`, `'${entered.split(/\s+/u)[0]}' is not recognized as a Meridian OS command.`, ""]);
      return;
    }

    setOutput((previous) => [...previous, `> ${entered}`, `Starting ${program.label}...`, ...program.art, ""]);
    setLaunchPending(true);
    launchTimer.current = window.setTimeout(() => {
      onRun(program.id);
      setLaunchPending(false);
      launchTimer.current = null;
    }, 1600);
  }

  function handleCommandKeyDown(event) {
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (!history.length) return;
      const nextIndex = historyIndex < 0 ? history.length - 1 : Math.max(0, historyIndex - 1);
      setHistoryIndex(nextIndex);
      setCommand(history[nextIndex]);
    } else if (event.key === "ArrowDown") {
      event.preventDefault();
      if (historyIndex < 0) return;
      if (historyIndex >= history.length - 1) {
        setHistoryIndex(-1);
        setCommand("");
        return;
      }
      const nextIndex = historyIndex + 1;
      setHistoryIndex(nextIndex);
      setCommand(history[nextIndex]);
    }
  }

  return (
    <main className="terminal-app" aria-label="Meridian command terminal" onMouseDown={(event) => {
      const input = event.currentTarget.querySelector(".terminal-app__prompt input");
      window.requestAnimationFrame(() => input?.focus({ preventScroll: true }));
    }}>
      <section ref={outputRef} className="terminal-app__output" aria-live="polite" aria-label="Terminal output">
        {output.map((line, index) => <div key={`${index}-${line}`} className={line.startsWith("> ") ? "terminal-app__command" : line.startsWith("Starting ") || line === "Available commands:" || line.startsWith("Type HELP") ? "terminal-app__highlight" : undefined}>{line || "\u00a0"}</div>)}
      </section>
      <form className="terminal-app__prompt" onSubmit={submit}>
        <label htmlFor="meridian-terminal-command">&gt;</label>
        <input id="meridian-terminal-command" autoComplete="off" spellCheck="false" placeholder="help" value={command} disabled={launchPending} onKeyDown={handleCommandKeyDown} onChange={(event) => setCommand(event.target.value)} aria-label="Enter a terminal command" />
      </form>
    </main>
  );
}
