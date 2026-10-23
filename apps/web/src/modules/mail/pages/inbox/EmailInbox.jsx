import { useRef, useState } from "react";
import { MapPin } from "pixelarticons/react/MapPin";
import { getNPCEmails, searchPeople } from "../../../world/api.js";
import DatePicker from "../../../world/components/DatePicker.jsx";
import ModularFace from "../../../identity/components/ModularFace.jsx";

const FALLBACK_NPC_VISUAL = { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", faceId: "", beardId: "", accessoryIds: [] };

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
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

export default function EmailInbox({ serverTime = new Date(), timeZone = "Europe/Paris", cityId = "", facilityLocationId = "", accessAttemptId = "", runHack, onPin }) {
  const throughToday = serverTime.toISOString().slice(0, 10);
  const rangeStart = `${throughToday.slice(0, 4)}-01-01`;
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
      const cityResults = cityId
        ? results.filter(({ npc }) => npc?.id.startsWith(`npc:${cityId}/`) || (cityId === "fr/lyon" && npc?.id.startsWith("npc:lyon/")))
        : results;
      setPeople(cityResults);
      setMessage(cityResults.length ? "CORRESPONDENCE FILES LOCATED IN THIS CITY" : "NO CORRESPONDENCE FILES LOCATED IN THIS CITY");
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
      const loadMessages = () => getNPCEmails(mailbox.id, from, through, facilityLocationId, accessAttemptId);
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
        {!mailbox && <form className="world-directory__search" onSubmit={searchMailbox}>
          <label className="world-directory__label" htmlFor="mail-search-person">PERSON NAME</label>
          <div className="world-directory__searchline">
            <input id="mail-search-person" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search correspondence files" />
            <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH"}</button>
          </div>
        </form>}
        {!mailbox && <p className="world-directory__message" role="status" aria-live="polite">{message}</p>}

        {!mailbox && people.map(({ npc }) => (
          <article className="world-record" key={npc.id}>
            <header className="world-record__header"><h2>{npc.name}</h2><span>ACCOUNT LOCATED</span></header>
            <div className="world-directory__person-result">
              <figure className="world-directory__person-portrait"><ModularFace visual={npc.visual ?? FALLBACK_NPC_VISUAL} name={npc.name} /></figure>
              <div className="world-directory__person-copy"><span>IDENTIFICATION PORTRAIT</span></div>
            </div>
            <button className="world-record__activity-toggle" type="button" onClick={() => { setMailbox(npc); setMessage("Select an interval to consult this account's archived correspondence."); }}>
              OPEN CORRESPONDENCE FILE
            </button>
          </article>
        ))}

        {mailbox && <section className="email-inbox__mailbox world-person-file" aria-label={`Correspondence file for ${mailbox.name}`}>
          <header className="world-person-file__toolbar"><button type="button" onClick={() => { requestId.current++; setBusy(false); setMailbox(null); setMessages([]); setMessage("Select a correspondence file."); }}>← BACK TO RESULTS</button><span>CORRESPONDENCE FILE</span></header>
          <div className="world-person-file__identity email-inbox__person-file">
            <figure className="world-directory__person-portrait"><ModularFace visual={mailbox.visual ?? FALLBACK_NPC_VISUAL} name={mailbox.name} /></figure>
            <div><h2>{mailbox.name}</h2><p>{mailbox.status?.toUpperCase() || "STATUS UNKNOWN"} · PRIVATE ACCOUNT</p></div>
          </div>
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
    </div>
  );
}
