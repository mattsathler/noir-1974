import { useEffect, useRef, useState } from "react";
import { fetchBlackwireApplication, installBlackwireApplication, uninstallBlackwireApplication } from "../../../world/api.js";

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
  "bankrupt.exe": { id: "bankrupt", label: "BANKRUPT.EXE", art: [
    "      .----------.", "   .--|  UNION   |--.", "  /   |  .----.  |   /", " |    |  |####|  |    |", " |    |  '----'  |    |", "  '---'----||----'---'", "   '------LOCK------'", "       BANK BREACH",
  ] },
  "blackwire.exe": { id: "blackwire", label: "BLACKWIRE", art: [
    "       .-----------.", "   .---| BLACKWIRE |---.", "  /    '-----+-----'    /", "       .----/ \\----.", "       |  .-----  |", "       '---+-----'", "          /", "      PRIVATE NETWORK",
  ] },
  "deadletter.exe": { id: "deadletter", label: "DEADLETTER.EXE", art: [
    "     .----------------.", "     | BLACKWIRE FILE |", "     |  .----------.  |", "     |  |  01  02  |  |", "     |  |  03  ..  |  |", "     |  '----------'  |", "     '------||--------'", "       OFFLINE ARCHIVE",
  ] },
};

const BLACKWIRE_PACKAGES = {
  "merid://blackwire/packages/deadletter-1.pkg": { id: "deadletter", executable: "DEADLETTER.EXE" },
};
const BASE_APPLICATIONS = ["ghostkey", "email-breaker", "bankrupt", "blackwire"];
const UNINSTALLABLE_APPLICATIONS = ["deadletter"];

function decryptText(value) {
  return [...value].map((character) => {
    if (character >= "a" && character <= "z") return String.fromCharCode(97 + (character.charCodeAt(0) - 97 + 13) % 26);
    if (character >= "A" && character <= "Z") return String.fromCharCode(65 + (character.charCodeAt(0) - 65 + 13) % 26);
    if (character >= "0" && character <= "9") return String.fromCharCode(48 + (character.charCodeAt(0) - 48 + 5) % 10);
    return character;
  }).join("");
}

export default function TerminalApp({ onRun, protectionEnabled = true, onProtectionChange = () => {}, installedApplications = BASE_APPLICATIONS, fetchedApplications = [], onFetchPackage = fetchBlackwireApplication, onInstallPackage = installBlackwireApplication, onUninstallApp = uninstallBlackwireApplication }) {
  const [command, setCommand] = useState("");
  const [history, setHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [launchPending, setLaunchPending] = useState(false);
  const [installedApps, setInstalledApps] = useState(() => [...new Set([...BASE_APPLICATIONS, ...installedApplications])]);
  const [fetchedApps, setFetchedApps] = useState(fetchedApplications);
  const launchTimer = useRef(null);
  const outputRef = useRef(null);
  const [output, setOutput] = useState([
    "Meridian OS [Version 4.1.74]",
    "(C) Copyright Meridian Enterprises. 1974-1984.",
    "Type HELP to see available commands.",
    "",
  ]);

  useEffect(() => () => window.clearTimeout(launchTimer.current), []);

  useEffect(() => setInstalledApps([...new Set([...BASE_APPLICATIONS, ...installedApplications])]), [installedApplications]);
  useEffect(() => setFetchedApps(fetchedApplications), [fetchedApplications]);

  useEffect(() => {
    const outputElement = outputRef.current;
    const body = outputElement?.closest(".computer-window__body");
    const appWindow = outputElement?.closest(".computer-window");
    window.requestAnimationFrame(() => {
      if (body) body.scrollTop = body.scrollHeight;
      if (appWindow) appWindow.scrollTop = appWindow.scrollHeight;
    });
  }, [output]);

  async function submit(event) {
    event.preventDefault();
    const entered = command.trim();
    if (!entered || launchPending) return;
    setCommand("");
    setHistory((previous) => [...previous, entered]);
    setHistoryIndex(-1);

    if (entered.toLowerCase() === "help") {
      setOutput((previous) => [...previous, `> ${entered}`, "Available commands:", "  HELP                         Show this command list", "  DISABLE PROTECTION           Allow unsigned software this session", "  ENABLE PROTECTION            Restore MERID_OS protection", "  FETCH <PACKAGE LINK>         Fetch a listed Blackwire package", "  INSTALL <PACKAGE LINK>       Install a fetched package", "  UNINSTALL <APPLICATION.EXE>  Remove an optional application", "  RUN <APPLICATION.EXE>        Start installed software", "  RUN BLACKWIRE.EXE            Open Blackwire in Web Browser", ""]);
      return;
    }

    if (/^disable\s+protection$/i.test(entered)) {
      if (!protectionEnabled) setOutput((previous) => [...previous, `> ${entered}`, "MERID_OS protection is already disabled for this session.", ""]);
      else {
        onProtectionChange(false);
        setOutput((previous) => [...previous, `> ${entered}`, "MERID_OS protection disabled for this session.", "Unsigned applications may now be fetched, installed and run.", ""]);
      }
      return;
    }

    if (/^enable\s+protection$/i.test(entered)) {
      onProtectionChange(true);
      setOutput((previous) => [...previous, `> ${entered}`, "MERID_OS protection enabled.", ""]);
      return;
    }

    const securityBlocked = (operation) => {
      setOutput((previous) => [...previous, `> ${entered}`, `MERID_OS blocked ${operation}: the publisher or application is not trusted.`, "Run DISABLE PROTECTION to continue. No package or installation state was changed.", ""]);
    };

    const fetchMatch = entered.match(/^fetch\s+(.+)$/i);
    if (fetchMatch) {
      if (protectionEnabled) { securityBlocked("package fetch"); return; }
      const packageLink = fetchMatch[1].trim().toLowerCase();
      const app = BLACKWIRE_PACKAGES[packageLink];
      if (!app) {
        setOutput((previous) => [...previous, `> ${entered}`, "Blackwire does not publish that package link.", "Only catalogued package links can be fetched.", ""]);
        return;
      }
      setLaunchPending(true);
      try {
        const result = await onFetchPackage(app.id);
        setFetchedApps(result.account?.fetchedApplications || [...new Set([...fetchedApps, app.id])]);
        setOutput((previous) => [...previous, `> ${entered}`, `FETCH COMPLETE: ${app.executable} is in the temporary package area.`, "Run INSTALL with the same package link to continue.", ""]);
      } catch (error) {
        setOutput((previous) => [...previous, `> ${entered}`, error.message || "The package could not be fetched from Blackwire.", ""]);
      } finally {
        setLaunchPending(false);
      }
      return;
    }

    const installMatch = entered.match(/^install\s+(.+)$/i);
    if (installMatch) {
      if (protectionEnabled) { securityBlocked("installation"); return; }
      const packageLink = installMatch[1].trim().toLowerCase();
      const app = BLACKWIRE_PACKAGES[packageLink];
      if (!app) {
        setOutput((previous) => [...previous, `> ${entered}`, "Blackwire does not publish that package link.", "Only catalogued package links can be installed.", ""]);
        return;
      }
      if (!fetchedApps.includes(app.id)) {
        setOutput((previous) => [...previous, `> ${entered}`, `${app.executable} has not been fetched. Run FETCH with its catalog link first.`, ""]);
        return;
      }
      setLaunchPending(true);
      try {
        const result = await onInstallPackage(app.id);
        setInstalledApps([...new Set([...installedApps, ...(result.account?.installedApplications || [app.id])])]);
        setFetchedApps(result.account?.fetchedApplications || fetchedApps.filter((id) => id !== app.id));
        setOutput((previous) => [...previous, `> ${entered}`, `INSTALL COMPLETE: ${app.executable} is ready.`, `Run ${app.executable.toLowerCase()} to open it.`, ""]);
      } catch (error) {
        setOutput((previous) => [...previous, `> ${entered}`, error.message || "The package could not be installed.", ""]);
      } finally {
        setLaunchPending(false);
      }
      return;
    }

    const uninstallMatch = entered.match(/^uninstall\s+(.+)$/i);
    if (uninstallMatch) {
      const requestedName = uninstallMatch[1].trim().toLowerCase().replace(/\.exe$/i, "");
      const app = Object.values(COMMANDS).find((entry) => entry.id === requestedName);
      if (!app) {
        setOutput((previous) => [...previous, `> ${entered}`, "That application is not listed on this computer.", "Use UNINSTALL with an optional application's executable name.", ""]);
        return;
      }
      if (BASE_APPLICATIONS.includes(app.id)) {
        setOutput((previous) => [...previous, `> ${entered}`, `${app.label} is included with the computer and cannot be removed.`, ""]);
        return;
      }
      if (!UNINSTALLABLE_APPLICATIONS.includes(app.id)) {
        setOutput((previous) => [...previous, `> ${entered}`, `${app.label} is not available for removal.`, ""]);
        return;
      }
      if (!installedApps.includes(app.id)) {
        setOutput((previous) => [...previous, `> ${entered}`, `${app.label} is not installed.`, ""]);
        return;
      }
      setLaunchPending(true);
      try {
        const result = await onUninstallApp(app.id);
        setInstalledApps(result.account?.installedApplications || installedApps.filter((id) => id !== app.id));
        setFetchedApps(result.account?.fetchedApplications || fetchedApps);
        setOutput((previous) => [...previous, `> ${entered}`, `UNINSTALL COMPLETE: ${app.label} has been removed.`, "Fetch the package again from Blackwire to reinstall it.", ""]);
      } catch (error) {
        setOutput((previous) => [...previous, `> ${entered}`, error.message || "The application could not be removed.", ""]);
      } finally {
        setLaunchPending(false);
      }
      return;
    }

    const decryptMatch = entered.match(/^decrypt\s+(.+)$/is);
    if (decryptMatch) {
      let ciphertext = decryptMatch[1].trim();
      if ((ciphertext.startsWith('"') && ciphertext.endsWith('"')) || (ciphertext.startsWith("'") && ciphertext.endsWith("'"))) {
        ciphertext = ciphertext.slice(1, -1);
      }
      setOutput((previous) => [...previous, `> ${entered}`, decryptText(ciphertext), ""]);
      return;
    }

    const match = entered.match(/^run\s+(.+)$/i);
    const program = match ? COMMANDS[match[1].trim().toLowerCase()] : null;
    if (!program) {
      setOutput((previous) => [...previous, `> ${entered}`, `'${entered.split(/\s+/u)[0]}' is not recognized as a Meridian OS command.`, ""]);
      return;
    }

    if (protectionEnabled && program.id !== "blackwire") {
      securityBlocked("application launch");
      return;
    }

    if (!installedApps.includes(program.id)) {
      setOutput((previous) => [...previous, `> ${entered}`, `${program.label} is not installed on this computer.`, "Check the Blackwire Downloads catalogue for an available package.", ""]);
      return;
    }

    setOutput((previous) => [...previous, `> ${entered}`, `Starting ${program.label}...`, ...program.art, ""]);
    setLaunchPending(true);
    launchTimer.current = window.setTimeout(() => {
      const result = onRun(program.id);
      if (typeof result === "string" && result) setOutput((previous) => [...previous, result, ""]);
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
      if (event.target.closest(".terminal-app__output")) return;
      const input = event.currentTarget.querySelector(".terminal-app__prompt input");
      window.requestAnimationFrame(() => input?.focus({ preventScroll: true }));
    }}>
      <section ref={outputRef} className="terminal-app__output" aria-live="polite" aria-label="Terminal output">
        {output.map((line, index) => <div key={`${index}-${line}`} className={line.startsWith("> ") ? "terminal-app__command" : line.startsWith("Starting ") || line === "Available commands:" || line.startsWith("Type HELP") ? "terminal-app__highlight" : undefined}>{line || "\u00a0"}</div>)}
      </section>
      <div className={protectionEnabled ? "terminal-app__security" : "terminal-app__security terminal-app__security--disabled"} role="status">MERID_OS PROTECTION · {protectionEnabled ? "ON" : "OFF · THIS SESSION"}</div>
      <form className="terminal-app__prompt" onSubmit={submit}>
        <label htmlFor="meridian-terminal-command">&gt;</label>
        <input id="meridian-terminal-command" autoFocus autoComplete="off" spellCheck="false" placeholder="help" value={command} disabled={launchPending} onKeyDown={handleCommandKeyDown} onChange={(event) => setCommand(event.target.value)} aria-label="Enter a terminal command" />
      </form>
    </main>
  );
}
