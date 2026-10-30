import { useState } from "react";
import { searchPeople, requestResidentialSearch } from "../../../world/api.js";
import ModularFace from "../../../identity/components/ModularFace.jsx";

const FALLBACK_VISUAL = { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", faceId: "", beardId: "", accessoryIds: [] };

export default function ResidentialSearchWindow({ cityId, onClose, standalone = false }) {
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState([]);
  const [selected, setSelected] = useState(null);
  const [message, setMessage] = useState("Search the personnel register for a subject.");
  const [busy, setBusy] = useState(false);
  const [requestStatus, setRequestStatus] = useState("");

  async function findPeople(event) {
    event.preventDefault();
    if (!query.trim()) { setMessage("Enter a name to search."); return; }
    setBusy(true); setSelected(null); setPeople([]); setMessage("SEARCHING PERSONNEL REGISTER…");
    try {
      const results = await searchPeople(query.trim());
      setPeople(results);
      setMessage(results.length ? `${results.length} PERSONNEL FILE${results.length === 1 ? "" : "S"} LOCATED COUNTRYWIDE` : "NO MATCHING FILES FOUND COUNTRYWIDE");
    } catch (error) { setMessage(error.message || "PERSONNEL REGISTER UNAVAILABLE"); }
    finally { setBusy(false); }
  }

  async function submitRequest() {
    if (!selected || requestStatus === "pending" || requestStatus === "processing") return;
    setBusy(true); setMessage("SUBMITTING RESIDENTIAL SEARCH REQUEST…");
    try {
      const cityKey = cityId === "fr/lyon" ? "lyon" : cityId;
      const result = await requestResidentialSearch(selected.npc.id, `service:${cityKey}/police-station`);
      setRequestStatus(result.status);
      setMessage(result.status === "granted"
        ? "SEARCH COMPLETE. THE REPORT IS AVAILABLE IN YOUR PERSONAL MAILBOX."
        : "REQUEST RECEIVED. AUTHORIZATION REVIEW IS IN PROGRESS. THE REPORT WILL ARRIVE BY PERSONAL MAIL IN ABOUT 3 MINUTES.");
    } catch (error) { setMessage(error.message || "REQUEST COULD NOT BE COMPLETED"); }
    finally { setBusy(false); }
  }

  const panel = <section className={`warrant-window__panel${standalone ? " police-restricted-app__panel" : ""}`}>
      <header className="warrant-window__header"><div><p className="computer-window__eyebrow">CORPORATE SECURITY OFFICE · FORM 74-B</p><h1 id="warrant-title">RESIDENTIAL SEARCH REQUEST</h1></div><button type="button" aria-label="Close request" onClick={onClose}>×</button></header>
      <p className="warrant-window__copy">Identify a person to request authorization for a search of their residence. The subject may be living or deceased.</p>
      {!selected && <form className="world-directory__search" onSubmit={findPeople}>
        <label className="world-directory__label" htmlFor="warrant-person-search">PERSON NAME</label>
        <div className="world-directory__searchline"><input id="warrant-person-search" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Type a name, then press Enter" aria-keyshortcuts="Enter" /><button className="search-enter-submit" type="submit" disabled={busy} aria-label={busy ? "Searching" : "Press Enter to search"}>{busy ? "SEARCHING…" : "ENTER ↵"}</button></div>
      </form>}
      <p className="world-directory__message" role="status" aria-live="polite">{message}</p>
      {!selected && <div className="warrant-window__results">
        {people.map((record) => {
          const { npc } = record;
          return <article className="warrant-person" key={npc.id}>
            <span className="warrant-person__portrait"><ModularFace visual={npc.visual ?? FALLBACK_VISUAL} name={npc.name} /></span><span className="warrant-person__details"><strong><button className="warrant-person__name" type="button" onClick={() => { setSelected(record); setMessage(""); }}>{npc.name}</button></strong><small>{npc.status === "dead" ? "DECEASED · FILE AVAILABLE" : "PERSONNEL FILE LOCATED"}</small></span>
          </article>;
        })}
      </div>}
      {selected && <section className="warrant-person-file">
        <header className="world-person-file__toolbar"><button type="button" disabled={busy} onClick={() => { setSelected(null); setMessage("PERSONNEL FILE LOCATED"); }}>← BACK TO RESULTS</button><span>SUBJECT DETAILS</span></header>
        <div className="world-person-file__identity">
          <figure className="world-directory__person-portrait"><ModularFace visual={selected.npc.visual ?? FALLBACK_VISUAL} name={selected.npc.name} /></figure>
          <div><h2>{selected.npc.name}</h2><p>{selected.npc.status?.toUpperCase() || "STATUS UNKNOWN"} · {selected.npc.age} YEARS</p></div>
        </div>
        <dl className="world-record__personnel-facts world-person-file__facts">
          <div><dt>DESCRIPTION</dt><dd>{selected.npc.appearance || "NOT RECORDED"}</dd></div>
          <div><dt>RESIDENTIAL SEARCH</dt><dd>HOME ADDRESS WITHHELD FROM THIS FILE</dd></div>
        </dl>
        <footer className="warrant-window__actions"><button className="warrant-window__deeplink" type="button" disabled={busy || requestStatus === "pending" || requestStatus === "processing"} onClick={submitRequest}>{busy ? "PROCESSING…" : requestStatus === "pending" || requestStatus === "processing" ? "REQUEST PENDING" : "REQUEST SEARCH"}</button></footer>
      </section>}
      {!selected && <footer className="warrant-window__actions"></footer>}
    </section>;
  return standalone
    ? <section className="police-restricted-app">{panel}</section>
    : <div className="warrant-window" role="dialog" aria-modal="true" aria-labelledby="warrant-title">{panel}</div>;
}
