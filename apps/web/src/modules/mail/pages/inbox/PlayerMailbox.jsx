import { useEffect, useRef, useState } from "react";
import { MapPin } from "pixelarticons/react/MapPin";
import { Files } from "pixelarticons/react/Files";
import { Note } from "pixelarticons/react/Note";
import { Box } from "pixelarticons/react/Box";
import { deletePlayerEmail, getPlayerEmails } from "../../../world/api.js";

const ATTACHMENT_ICONS = { document: Files, paper: Note, object: Box };

function displayDate(value, timeZone) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en-US", { timeZone: timeZone || "Europe/Paris", month: "short", day: "numeric", hour: "numeric", minute: "2-digit" }).format(date);
}

export default function PlayerMailbox({ timeZone, onPin }) {
  const [messages, setMessages] = useState([]);
  const [status, setStatus] = useState("OPENING PRIVATE MAILBOX…");
  const [deletingID, setDeletingID] = useState("");
  const deletedIDs = useRef(new Set());

  useEffect(() => {
    let active = true;
    let loaded = false;
    const refresh = () => getPlayerEmails().then((items) => {
      if (!active) return;
      loaded = true;
      const visible = items.filter((item) => !deletedIDs.current.has(item.id));
      setMessages(visible);
      setStatus(visible.length ? `${visible.length} MESSAGE${visible.length === 1 ? "" : "S"} ON FILE` : "NO PERSONAL MAIL ON FILE");
    }).catch((error) => { if (active && !loaded) setStatus(error.message || "MAILBOX UNAVAILABLE"); });
    refresh();
    const interval = window.setInterval(refresh, 15000);
    return () => { active = false; window.clearInterval(interval); };
  }, []);

  async function deleteMessage(message) {
    setDeletingID(message.id);
    try {
      await deletePlayerEmail(message.id);
      deletedIDs.current.add(message.id);
      setMessages((current) => {
        const remaining = current.filter((item) => item.id !== message.id);
        setStatus(remaining.length ? `${remaining.length} MESSAGE${remaining.length === 1 ? "" : "S"} ON FILE` : "NO PERSONAL MAIL ON FILE");
        return remaining;
      });
    } catch (error) {
      setStatus(error.message || "COULD NOT DELETE MESSAGE");
    } finally {
      setDeletingID("");
    }
  }

  return <div className="email-inbox">
    <p className="computer-window__eyebrow">PRIVATE LINE · PERSONAL ACCOUNT</p>
    <h1>PERSONAL MAILBOX</h1>
    <p className="world-directory__message" role="status" aria-live="polite">{status}</p>
    {!messages.length && status === "NO PERSONAL MAIL ON FILE" && <p className="email-inbox__empty">Case notices and official reports will be delivered here.</p>}
    {messages.map((message) => <div className="email-message-row" key={message.id}>
      <details className={`email-message${message.category === "spam" ? " email-message--spam" : ""}`}>
        <summary className="email-message__summary">
          <span>{message.sender || "BILL BOUER"} → PERSONAL ACCOUNT {message.category === "spam" ? "· SPAM" : ""}</span>
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
                <footer><button className="whiteboard-pin-button" type="button" aria-label={`Pin ${item.title} to a whiteboard`} onClick={() => onPin?.({ id: item.id, source: item.source, title: item.title, content: item.content, metadata: item.metadata })}><MapPin aria-hidden="true" /></button></footer>
              </article>;
            })}</div>
          </section>}
        </div>
      </details>
      <button className="email-message__delete" type="button" disabled={deletingID === message.id} aria-label={`Delete email: ${message.subject}`} onClick={() => deleteMessage(message)}>{deletingID === message.id ? "DELETING…" : "DELETE"}</button>
    </div>)}
  </div>;
}
