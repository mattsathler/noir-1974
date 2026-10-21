import { useEffect, useRef, useState } from "react";
import { getNPCRecords, searchCompanies, searchLocations, searchPeople } from "../../api.js";
import DatePicker from "../../components/DatePicker.jsx";

function inputDate(date, timeZone) {
  const values = Object.fromEntries(new Intl.DateTimeFormat("en-US", {
    timeZone: timeZone || "Europe/Paris",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date).map(({ type, value }) => [type, value]));
  return `${values.year}-${values.month}-${values.day}`;
}

function dateBefore(value, days) {
  const date = new Date(`${value}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() - days);
  return date.toISOString().slice(0, 10);
}

function dateDistance(first, second) {
  return (new Date(`${second}T00:00:00Z`) - new Date(`${first}T00:00:00Z`)) / 86400000;
}

function displayRecordDate(value, timeZone) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en-US", {
    timeZone: timeZone || "Europe/Paris",
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  }).format(date);
}

export default function WorldDirectory({ program, serverTime = new Date("1974-11-12T20:15:00Z"), timeZone = "Europe/Paris", runHack }) {
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState([]);
  const [companies, setCompanies] = useState([]);
  const [locations, setLocations] = useState([]);
  const [message, setMessage] = useState("Enter a name or address to search the city records.");
  const [busy, setBusy] = useState(false);
  const [recordNPC, setRecordNPC] = useState("");
  const [recordRange, setRecordRange] = useState(() => {
    const through = inputDate(serverTime, timeZone);
    return { from: dateBefore(through, 6), to: through };
  });
  const rangeStart = `${inputDate(serverTime, timeZone).slice(0, 4)}-01-01`;
  const rangeEnd = inputDate(serverTime, timeZone);
  const [activity, setActivity] = useState([]);
  const [activityMessage, setActivityMessage] = useState("");
  const [activityBusy, setActivityBusy] = useState(false);
  const [privilegedAccess, setPrivilegedAccess] = useState(false);
  const [accessReady, setAccessReady] = useState(false);
  const [accessBusy, setAccessBusy] = useState(false);
  const requestId = useRef(0);
  const activityRequestId = useRef(0);
  const isInterpol = program === "interpol";
  const isCityRecords = program === "city-records";
  const isBankRecords = program === "bank-records";
  const isGateRecords = program === "gate-records";
  const isPeopleSearch = !isCityRecords;
  const isPrivilegedSystem = isBankRecords || isGateRecords;

  useEffect(() => {
    if (!(isInterpol || isPrivilegedSystem)) {
      setAccessReady(true);
      return undefined;
    }
    setAccessReady(false);
    const timer = window.setTimeout(() => setAccessReady(true), 850);
    return () => window.clearTimeout(timer);
  }, [isInterpol, isPrivilegedSystem]);

  function updateRecordFrom(from) {
    setRecordRange((previous) => ({
      from,
      to: previous.to < from || dateDistance(from, previous.to) > 6
        ? [dateBefore(from, -6), rangeEnd].sort()[0]
        : previous.to,
    }));
  }

  function updateRecordThrough(to) {
    setRecordRange((previous) => ({
      from: previous.from > to || dateDistance(previous.from, to) > 6
        ? [dateBefore(to, 6), rangeStart].sort().reverse()[0]
        : previous.from,
      to,
    }));
  }

  async function loadNPCRecords(event, npcID) {
    event.preventDefault();
    const request = ++activityRequestId.current;
    setRecordNPC(npcID);
    setActivity([]);
    setActivityMessage(isBankRecords ? "REQUESTING FINANCIAL LEDGER…" : "REQUESTING GATE ACCESS LOG…");
    setActivityBusy(true);
    try {
      const source = isBankRecords ? "bank" : "gate";
      const loadRecords = () => getNPCRecords(npcID, recordRange.from, recordRange.to, source);
      const results = runHack ? await runHack(loadRecords) : await loadRecords();
      if (request !== activityRequestId.current) return;
      setActivity(results);
      setActivityMessage(results.length ? "ACTIVITY RECORDS RETRIEVED" : "NO ACTIVITY ON FILE FOR THIS INTERVAL");
    } catch (error) {
      if (request !== activityRequestId.current) return;
      setActivityMessage(error.message || "THE RECORDS OFFICE COULD NOT BE REACHED.");
    } finally {
      if (request === activityRequestId.current) setActivityBusy(false);
    }
  }

  async function authorizeConfidentialAccess() {
    setAccessBusy(true);
    try {
      if (runHack) await runHack(() => Promise.resolve(true));
      setPrivilegedAccess(true);
    } catch (error) {
      setActivityMessage(error.message || "THE SECURE CONNECTION FAILED.");
    } finally {
      setAccessBusy(false);
    }
  }

  async function submitSearch(event) {
    event.preventDefault();
    const term = query.trim();
    if (!term) {
      setMessage("Enter a search term to continue.");
      return;
    }
    const currentRequest = ++requestId.current;
    setBusy(true);
    setMessage("");
    setPeople([]);
    setCompanies([]);
    setLocations([]);
    try {
      if (isPeopleSearch) {
        const loadPeople = () => searchPeople(term);
        const results = (isInterpol || isPrivilegedSystem) && runHack
          ? await runHack(loadPeople)
          : await loadPeople();
        if (currentRequest !== requestId.current) return;
        setPeople(results);
        setMessage(results.length ? "PERSONNEL FILES RETRIEVED" : "NO MATCHING PERSONNEL FILES");
      } else {
        const loadRecords = () => Promise.all([searchCompanies(term), searchLocations(term)]);
        const [companyResults, locationResults] = isPrivilegedSystem && runHack
          ? await runHack(loadRecords)
          : await loadRecords();
        if (currentRequest !== requestId.current) return;
        setCompanies(companyResults);
        setLocations(locationResults);
        setMessage(companyResults.length || locationResults.length ? "CITY RECORDS RETRIEVED" : "NO MATCHING CITY RECORDS");
      }
    } catch (error) {
      if (currentRequest === requestId.current) setMessage(error.message || "The directory could not be reached.");
    } finally {
      if (currentRequest === requestId.current) setBusy(false);
    }
  }

  return (
    <div className="world-directory">
      <p className="computer-window__eyebrow">{isInterpol ? "FORMER INTERPOL AFFILIATION · LEGACY ACCESS" : isCityRecords ? "LYON MUNICIPAL ARCHIVE · PUBLIC INDEX" : isBankRecords ? `MERIDIAN GLOBAL SETTLEMENT OFFICE · ${privilegedAccess ? "PRIVILEGED SESSION" : "RESTRICTED ACCESS"}` : `MERIDIAN GLOBAL · GATE CONTROL ARCHIVE · ${privilegedAccess ? "PRIVILEGED SESSION" : "RESTRICTED ACCESS"}`}</p>
      <h1>{isInterpol ? "PERSON SEARCH" : isCityRecords ? "CITY RECORDS SEARCH" : isBankRecords ? "FINANCIAL LEDGER SEARCH" : "GATE MOVEMENT SEARCH"}</h1>
      {(isPrivilegedSystem || isInterpol) && !privilegedAccess ? (
        accessReady ? <section className="world-directory__credential-access" aria-label="Privileged system access">
          <p>CREDENTIALS REQUIRED · RESTRICTED SYSTEM</p>
          <p>Establish an unauthorized connection with the operative's legacy access tools.</p>
          <button type="button" disabled={accessBusy} onClick={authorizeConfidentialAccess}>{accessBusy ? "HACKING SYSTEM…" : "HACK SYSTEM"}</button>
        </section> : <p className="world-directory__message" role="status" aria-live="polite">ESTABLISHING SECURE CONNECTION…</p>
      ) : <>
      <form className="world-directory__search" onSubmit={submitSearch}>
        <label className="world-directory__label" htmlFor="world-directory-query">{isPeopleSearch ? "PERSON NAME" : "COMPANY OR ADDRESS"}</label>
        <div className="world-directory__searchline">
          <input id="world-directory-query" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={isPeopleSearch ? "Search personnel files" : "Search companies and locations"} />
          <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH"}</button>
        </div>
      </form>
      <p className="world-directory__message" role="status" aria-live="polite">{message}</p>

      {people.map(({ npc, workplaces, maritalStatus }) => (
        <article className="world-record" key={npc.id}>
          <header className="world-record__header"><h2>{npc.name}</h2><span>PERSONNEL FILE</span></header>
          <p>{npc.age} YEARS · {npc.appearance}</p>
          {workplaces.map(({ company, address }) => <p key={`${company}-${address}`}>WORKPLACE · {company} — {address}</p>)}
          {!workplaces.length && <p>WORKPLACE · NOT LISTED</p>}
          <p>MARITAL STATUS · {maritalStatus?.toUpperCase() || "UNKNOWN"}</p>
          {(isBankRecords || isGateRecords) && <button
            className="world-record__activity-toggle"
            type="button"
            aria-expanded={recordNPC === npc.id}
            onClick={() => { activityRequestId.current++; setActivityBusy(false); setRecordNPC(recordNPC === npc.id ? "" : npc.id); setActivity([]); setActivityMessage(""); }}
          >
            {recordNPC === npc.id ? "CLOSE RECORDS" : isBankRecords ? "OPEN FINANCIAL LEDGER" : "OPEN GATE TRAVEL FILE"}
          </button>}
          {recordNPC === npc.id && (
            <section className="world-record__activity" aria-label={`${isBankRecords ? "Financial ledger" : "GATE travel file"} for ${npc.name}`}>
              <form className="world-record__activity-filter" onSubmit={(event) => loadNPCRecords(event, npc.id)}>
                <DatePicker label="FROM" value={recordRange.from} minDate={rangeStart} maxDate={rangeEnd} onChange={updateRecordFrom} />
                <DatePicker label="THROUGH" value={recordRange.to} minDate={rangeStart} maxDate={rangeEnd} onChange={updateRecordThrough} />
                <button type="submit" disabled={activityBusy}>{activityBusy ? "READING…" : "READ RECORDS"}</button>
              </form>
              {activityMessage && <p className="world-record__activity-message" role="status" aria-live="polite">{activityMessage}</p>}
              {activity.map((record) => (
                <article className="world-record__activity-entry" key={record.id}>
                  <header><span>{isGateRecords ? "GATE ACCESS LOG" : "FINANCIAL LEDGER"}</span><time dateTime={record.occurredAt}>{displayRecordDate(record.occurredAt, timeZone)}</time></header>
                  <p>{record.description}</p>
                  <p>{record.operator}{record.reference ? ` · ${record.reference}` : ""}</p>
                  {isBankRecords && record.amount > 0 && <p className={`world-record__amount world-record__amount--${record.direction || "outgoing"}`}><span>{record.direction === "incoming" ? "CREDIT · RECEIVED" : "DEBIT · SENT"}</span><strong>{record.direction === "incoming" ? "+" : "−"}{record.amount} CREDITS</strong></p>}
                  {(record.origin || record.destination) && <div className="world-record__route" aria-label={`Recorded route from ${record.origin || "unknown location"} to ${record.destination || "unknown location"}`}>
                    <span>RECORDED PASSAGE</span>
                    <div><strong>{record.origin || "UNKNOWN ORIGIN"}</strong><b aria-hidden="true">→</b><strong>{record.destination || "UNKNOWN DESTINATION"}</strong></div>
                  </div>}
                </article>
              ))}
            </section>
          )}
        </article>
      ))}

      {isCityRecords && companies.map(({ company, address, employees }) => (
        <article className="world-record" key={company.id}>
          <header className="world-record__header"><h2>{company.name}</h2><span>COMPANY</span></header>
          <p>HEADQUARTERS · {address.full}</p>
          <p>PERSONNEL ON FILE · {employees.length}</p>
          {employees.map((employee) => <p className="world-record__detail" key={employee.id}>{employee.name} · {employee.role.toUpperCase()}</p>)}
        </article>
      ))}

      {isCityRecords && locations.map(({ address }) => (
        <article className="world-record" key={address.locationId}>
          <header className="world-record__header"><h2>{address.name}</h2><span>{address.kind.toUpperCase()}</span></header>
          <p>{address.full}</p>
        </article>
      ))}
      </>}
    </div>
  );
}
