import { useRef, useState } from "react";
import { MapPin } from "pixelarticons/react/MapPin";
import { getNPCEmails, searchPeople } from "../../../world/api.js";
import DatePicker from "../../../world/components/DatePicker.jsx";

function dateBefore(value, days) {
  const date = new Date(`${value}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() - days);
  return date.toISOString().slice(0, 10);
}

function displayDate(value, timeZone) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en-US", {
    timeZone: timeZone || "Europe/Paris",
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export default function EmailInbox({ serverTime = new Date("1974-11-12T20:15:00Z"), timeZone = "Europe/Paris", runHack, onPin }) {
  const throughToday = serverTime.toISOString().slice(0, 10);
  const rangeStart = `${throughToday.slice(0, 4)}-01-01`;
  const [authorized, setAuthorized] = useState(false);
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState([]);
  const [mailbox, setMailbox] = useState(null);
  const [messages, setMessages] = useState([]);
  const [from, setFrom] = useState(() => dateBefore(throughToday, 6));
  const [through, setThrough] = useState(throughToday);
  const [message, setMessage] = useState("Search for a name to consult a correspondence file.");
  const [busy, setBusy] = useState(false);
  const requestId = useRef(0);

  async function hackRequest(loadRequest) {
    return runHack ? runHack(loadRequest) : loadRequest();
  }

  async function authorizeMailbox() {
    setBusy(true);
    try {
      if (runHack) await runHack(() => Promise.resolve(true));
      setAuthorized(true);
    } catch (error) {
      setMessage(error.message || "THE SECURE CONNECTION FAILED.");
    } finally {
      setBusy(false);
    }
  }

  function updateFrom(nextFrom) {
    setFrom(nextFrom);
    if (through < nextFrom || (new Date(`${through}T00:00:00Z`) - new Date(`${nextFrom}T00:00:00Z`)) / 86400000 > 6) {
      const candidate = dateBefore(nextFrom, -6);
      setThrough(candidate > throughToday ? throughToday : candidate);
    }
  }

  function updateThrough(nextThrough) {
    setThrough(nextThrough);
    if (from > nextThrough || (new Date(`${nextThrough}T00:00:00Z`) - new Date(`${from}T00:00:00Z`)) / 86400000 > 6) {
      const candidate = dateBefore(nextThrough, 6);
      setFrom(candidate < rangeStart ? rangeStart : candidate);
    }
  }

  async function searchMailbox(event) {
    event.preventDefault();
    const term = query.trim();
    if (!term) {
      setMessage("Enter a name to continue.");
      return;
    }
    const request = ++requestId.current;
    setBusy(true);
    setPeople([]);
    setMailbox(null);
    setMessages([]);
    setMessage("");
    try {
      const results = await hackRequest(() => searchPeople(term));
      if (request !== requestId.current) return;
      setPeople(results);
      setMessage(results.length ? "CORRESPONDENCE FILES LOCATED" : "NO CORRESPONDENCE FILES LOCATED");
    } catch (error) {
      if (request === requestId.current) setMessage(error.message || "The exchange could not be reached.");
    } finally {
      if (request === requestId.current) setBusy(false);
    }
  }

  async function readMailbox(event) {
    event.preventDefault();
    if (!mailbox) return;
    const request = ++requestId.current;
    setBusy(true);
    setMessages([]);
    setMessage("REQUESTING PRIVATE CORRESPONDENCE…");
    try {
      const loadMessages = () => getNPCEmails(mailbox.id, from, through);
      const results = await hackRequest(loadMessages);
      if (request !== requestId.current) return;
      setMessages(results);
      setMessage(results.length ? "CORRESPONDENCE RETRIEVED" : "NO MESSAGES ON FILE FOR THIS INTERVAL");
    } catch (error) {
      if (request === requestId.current) setMessage(error.message || "The exchange could not be reached.");
    } finally {
      if (request === requestId.current) setBusy(false);
    }
  }

  return (
    <div className="email-inbox">
      <p className="computer-window__eyebrow">MERIDIAN GLOBAL · PRIVATE CORRESPONDENCE EXCHANGE</p>
      <h1>MAILROOM ARCHIVE</h1>
      {!authorized ? (
        <section className="email-inbox__access" aria-label="Correspondence access">
          <p>CREDENTIALS REQUIRED · RESTRICTED EXCHANGE</p>
          <p>Legacy operative credentials can establish a connection through this Meridian relay.</p>
          <button type="button" disabled={busy} onClick={authorizeMailbox}>{busy ? "HACKING SYSTEM…" : "HACK SYSTEM"}</button>
        </section>
      ) : <>
        <form className="world-directory__search" onSubmit={searchMailbox}>
          <label className="world-directory__label" htmlFor="mail-search-person">PERSON NAME</label>
          <div className="world-directory__searchline">
            <input id="mail-search-person" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search correspondence files" />
            <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH"}</button>
          </div>
        </form>
        <p className="world-directory__message" role="status" aria-live="polite">{message}</p>

        {!mailbox && people.map(({ npc }) => (
          <article className="world-record" key={npc.id}>
            <header className="world-record__header"><h2>{npc.name}</h2><span>ACCOUNT LOCATED</span></header>
            <button className="world-record__activity-toggle" type="button" onClick={() => { setMailbox(npc); setMessage("Select an interval to consult this account's archived correspondence."); }}>
              OPEN CORRESPONDENCE FILE
            </button>
          </article>
        ))}

        {mailbox && <section className="email-inbox__mailbox" aria-label={`Correspondence file for ${mailbox.name}`}>
          <header className="world-record__header">
            <h2>{mailbox.name}</h2>
            <button className="email-inbox__back" type="button" onClick={() => { requestId.current++; setBusy(false); setMailbox(null); setMessages([]); setMessage("Select a correspondence file."); }}>BACK TO FILES</button>
          </header>
          <form className="world-record__activity-filter" onSubmit={readMailbox}>
            <DatePicker label="FROM" value={from} minDate={rangeStart} maxDate={throughToday} onChange={updateFrom} />
            <DatePicker label="THROUGH" value={through} minDate={rangeStart} maxDate={throughToday} onChange={updateThrough} />
            <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH MAIL"}</button>
          </form>
          {messages.map((item) => (
            <details className="email-message" key={item.id}>
              <summary className="email-message__summary">
                <span>{item.sender} → {item.recipient}</span>
                <span>{item.subject}</span>
                <time dateTime={item.sentAt}>{displayDate(item.sentAt, timeZone)}</time>
              </summary>
              <div className="email-message__body">
                <p className="email-message__service">{item.service}</p>
                <p>{item.body}</p>
                <button className="whiteboard-pin-button" type="button" aria-label={`Pin email: ${item.subject}`} onClick={() => onPin?.({ id: item.id, source: "email", title: item.subject || "EMAIL", content: `${item.sender} → ${item.recipient}\n${item.body}`, metadata: item })}><MapPin aria-hidden="true" /></button>
              </div>
            </details>
          ))}
        </section>}
      </>}
    </div>
  );
}
