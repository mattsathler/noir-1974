import { useEffect, useRef, useState } from "react";
import { Printer } from "pixelarticons/react/Printer";
import { Files } from "pixelarticons/react/Files";
import { Note } from "pixelarticons/react/Note";
import { Box } from "pixelarticons/react/Box";
import { Trash } from "pixelarticons/react/Trash";
import { deletePlayerEmail, getNPCEmails, getPlayerEmails } from "../../../world/api.js";
import DatePicker from "../../../world/components/DatePicker.jsx";

const ATTACHMENT_ICONS = { document: Files, paper: Note, object: Box };

function displayDate(value, timeZone) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en-US", { timeZone: timeZone || "Europe/Paris", month: "short", day: "numeric", hour: "numeric", minute: "2-digit" }).format(date);
}

function dateBefore(value, days) {
  const date = new Date(`${value}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() - days);
  return date.toISOString().slice(0, 10);
}

export default function PlayerMailbox({ serverTime = new Date(), timeZone, onPrint, onMessageReceived, hackedAccount = null, from: initialFrom = "", through: initialThrough = "", onExit }) {
  const throughToday = new Date(serverTime).toISOString().slice(0, 10);
  const [messages, setMessages] = useState([]);
  const [status, setStatus] = useState("OPENING MAILBOX…");
  const [from, setFrom] = useState(() => initialFrom || dateBefore(throughToday, 6));
  const [through, setThrough] = useState(() => initialThrough || throughToday);
  const deletedIDs = useRef(new Set());
  const pendingDeleteIDs = useRef(new Set());
  const previousMessageIDs = useRef(null);

  useEffect(() => {
    if (!hackedAccount) return undefined;
    let active = true;
    setMessages([]);
    setStatus("OPENING MAILBOX…");
    getNPCEmails(hackedAccount.id, from, through).then((items) => {
      if (!active) return;
      setMessages(items);
      setStatus(items.length ? `${items.length} MESSAGE${items.length === 1 ? "" : "S"} ON FILE` : "NO MESSAGES IN THIS DATE RANGE");
    }).catch((error) => { if (active) setStatus(error.message || "MAILBOX UNAVAILABLE"); });
    return () => { active = false; };
  }, [hackedAccount?.id, from, through]);

  useEffect(() => {
    if (hackedAccount) return undefined;
    let active = true;
    let loaded = false;
    const refresh = () => getPlayerEmails().then((items) => {
      if (!active) return;
      loaded = true;
      const currentIDs = new Set(items.map((item) => item.id));
      if (previousMessageIDs.current) {
        items.filter((item) => !previousMessageIDs.current.has(item.id)).forEach((item) => onMessageReceived?.(item));
      }
      previousMessageIDs.current = currentIDs;
      const visible = items.filter((item) => !deletedIDs.current.has(item.id) && !pendingDeleteIDs.current.has(item.id));
      setMessages(visible);
      setStatus("MAILBOX_LOADED");
    }).catch((error) => { if (active && !loaded) setStatus(error.message || "MAILBOX UNAVAILABLE"); });
    refresh();
    const interval = window.setInterval(refresh, 15000);
    return () => { active = false; window.clearInterval(interval); };
  }, [hackedAccount, onMessageReceived]);

  function updateFrom(nextFrom) {
    setFrom(nextFrom);
    const latestThrough = dateBefore(nextFrom, -6);
    if (through < nextFrom) setThrough(nextFrom);
    else if (through > latestThrough) setThrough(latestThrough);
  }

  function updateThrough(nextThrough) {
    setThrough(nextThrough);
    const earliestFrom = dateBefore(nextThrough, 6);
    if (from > nextThrough) setFrom(nextThrough);
    else if (from < earliestFrom) setFrom(earliestFrom);
  }

  const visibleMessages = messages.filter((item) => {
    const sentDate = item.sentAt?.slice(0, 10);
    return sentDate >= from && sentDate <= through;
  });
  const displayStatus = !hackedAccount && status === "MAILBOX_LOADED"
    ? visibleMessages.length ? `${visibleMessages.length} MESSAGE${visibleMessages.length === 1 ? "" : "S"} ON FILE` : "NO MESSAGES IN THIS DATE RANGE"
    : status;

  async function deleteMessage(message) {
    const originalIndex = messages.findIndex((item) => item.id === message.id);
    pendingDeleteIDs.current.add(message.id);
    const remaining = messages.filter((item) => item.id !== message.id);
    setMessages(remaining);
    setStatus("MAILBOX_LOADED");
    try {
      await deletePlayerEmail(message.id);
      deletedIDs.current.add(message.id);
      pendingDeleteIDs.current.delete(message.id);
    } catch (error) {
      pendingDeleteIDs.current.delete(message.id);
      setMessages((current) => {
        if (current.some((item) => item.id === message.id)) return current;
        const restored = [...current];
        restored.splice(Math.min(Math.max(originalIndex, 0), restored.length), 0, message);
        return restored;
      });
      setStatus(error.message || "COULD NOT DELETE MESSAGE");
    }
  }

  function printMessage(message) {
    onPrint?.({
      id: message.id,
      source: "email",
      title: message.subject || "EMAIL MESSAGE",
      content: `FROM: ${message.sender || "UNKNOWN"}\nTO: ${message.recipient || (hackedAccount ? hackedAccount.email : "PERSONAL ACCOUNT")}\nDATE: ${displayDate(message.sentAt, timeZone)}\nSUBJECT: ${message.subject || "NO SUBJECT"}\n\n${message.body}`,
      metadata: message,
    });
  }

  return <div className={`email-inbox${hackedAccount ? " email-inbox--hacked" : ""}`}>
    {hackedAccount && onExit && <button className="email-inbox__hacked-back" type="button" onClick={onExit}>← RETURN TO EMAIL_BREAKER</button>}
    <p className="computer-window__eyebrow">PRIVATE LINE · PERSONAL ACCOUNT</p>
    <h1>PERSONAL MAILBOX</h1>
    <section className="player-mail__date-range" aria-label="Email date range">
      <span className="world-record__date-label">DATE RANGE</span>
      <div className="world-record__activity-filter">
        <DatePicker label="FROM" value={from} minDate="1974-01-01" maxDate={throughToday} onChange={updateFrom} />
        <DatePicker label="THROUGH" value={through} minDate="1974-01-01" maxDate={throughToday} onChange={updateThrough} />
      </div>
    </section>
    <p className="world-directory__message" role="status" aria-live="polite">{displayStatus}</p>
    {!visibleMessages.length && displayStatus === "NO MESSAGES IN THIS DATE RANGE" && <p className="email-inbox__empty">No messages are available for this date range.</p>}
    {visibleMessages.map((message) => <div className="email-message-row" key={message.id}>
      <details className={`email-message${message.category === "spam" ? " email-message--spam" : ""}`}>
        <summary className="email-message__summary">
          <span>{message.sender || "BILL BOUER"} → {hackedAccount ? message.recipient || hackedAccount.email : "PERSONAL ACCOUNT"} {message.category === "spam" ? "· SPAM" : ""}</span>
          <span>{message.subject}</span>
          <time dateTime={message.sentAt}>{displayDate(message.sentAt, timeZone)}</time>
        </summary>
        <div className="email-message__body">
          <div className="email-message__header"><span>{message.category === "spam" ? "UNSOLICITED CIRCULAR · SPAM" : "PERSONAL MAIL NOTICE"}</span><time dateTime={message.sentAt}>{displayDate(message.sentAt, timeZone)}</time></div>
          <p>{message.body}</p>
          {!!message.attachments?.length && <section className="player-mail__attachments" aria-label="Recovered items">
            <h3 className="case-files__section-heading">RECOVERED AND SECURED · {message.attachments.length} ITEMS</h3>
            <div className="player-mail__attachment-grid">{message.attachments.map((item) => {
              const AttachmentIcon = ATTACHMENT_ICONS[item.kind] ?? Box;
              const bodyFound = item.kind === "body_found";
              return <article className={`search-attachment${bodyFound ? " search-attachment--body-found" : ""}`} key={item.id}>
                <span className="search-attachment__tack" aria-hidden="true" />
                <header><span>{bodyFound ? "BODY DISCOVERY · CONFIRMED" : "CASE · ATTACHMENT"}</span></header>
                <span className={`search-attachment__icon search-attachment__icon--${item.kind}`} aria-hidden="true"><AttachmentIcon /></span>
                <h4>{item.title}</h4><p>{item.content}</p>
                <footer><button className="page-print-button" type="button" aria-label={`Print ${item.title}`} title="Print page" onClick={() => onPrint?.({ id: item.id, source: item.source, title: item.title, content: item.content, metadata: item.metadata })}><Printer aria-hidden="true" /></button></footer>
              </article>;
            })}</div>
          </section>}
        </div>
      </details>
      {hackedAccount && <button className="email-message__print page-print-button" type="button" aria-label={`Add email to whiteboard: ${message.subject}`} title="Add email to a whiteboard" onClick={() => printMessage(message)}><Printer aria-hidden="true" /></button>}
      {hackedAccount
        ? <button className="email-message__delete" type="button" disabled aria-label={`Cannot delete email: ${message.subject}`} title="Mailbox is read-only"><Trash aria-hidden="true" /></button>
        : <button className="email-message__delete" type="button" aria-label={`Delete email: ${message.subject}`} title="Delete email" onClick={() => deleteMessage(message)}><Trash aria-hidden="true" /></button>}
    </div>)}
  </div>;
}
