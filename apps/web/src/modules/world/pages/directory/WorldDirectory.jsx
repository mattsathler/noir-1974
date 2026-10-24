import { useEffect, useRef, useState } from "react";
import { MapPin } from "pixelarticons/react/MapPin";
import { getCompanyAtLocation, getNPCRecords, searchCompanies, searchLocations, searchPeople } from "../../api.js";
import DatePicker from "../../components/DatePicker.jsx";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import { phoneNumberForNPC } from "../../phoneDirectory.js";

const FALLBACK_NPC_VISUAL = {
  hairId: "hair_short_brown",
  skinId: "skin_white",
  eyesId: "eyes_brown",
  faceId: "",
  beardId: "",
  accessoryIds: [],
};

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

export default function WorldDirectory({ program, companyLocationId = "", companyId = "", facilityLocationId = "", accessAttemptId = "", cityId = "", countryName = "", serverTime = new Date(), timeZone = "Europe/Paris", runHack, onPin }) {
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState([]);
  const [selectedPerson, setSelectedPerson] = useState(null);
  const [companies, setCompanies] = useState([]);
  const [locations, setLocations] = useState([]);
  const isStaffDirectory = program === "staff-directory";
  const [message, setMessage] = useState(isStaffDirectory ? "Enter a company name to search its staff directory." : program === "interpol" ? "Enter a name to search countrywide personnel files." : "Enter a name or address to search the city records.");
  const [busy, setBusy] = useState(false);
  const [recordRange, setRecordRange] = useState(() => {
    const through = inputDate(serverTime, timeZone);
    return { from: dateBefore(through, 6), to: through };
  });
  const rangeStart = `${inputDate(serverTime, timeZone).slice(0, 4)}-01-01`;
  const rangeEnd = inputDate(serverTime, timeZone);
  const [activity, setActivity] = useState([]);
  const [activityMessage, setActivityMessage] = useState("");
  const [activityBusy, setActivityBusy] = useState(false);
  const requestId = useRef(0);
  const activityRequestId = useRef(0);
  const isInterpol = program === "interpol";
  const isCityRecords = program === "city-records";
  const isBankRecords = program === "bank-records";
  const isGateRecords = program === "gate-records";
  const isPeopleSearch = !isCityRecords && !isStaffDirectory;
  const isPrivilegedSystem = isBankRecords || isGateRecords;
  const staffCompanies = companyId ? companies.filter(({ company }) => company.id === companyId) : companies;

  useEffect(() => {
    if (!isStaffDirectory) return undefined;
    let active = true;
    if (!companyLocationId) {
      setMessage("COMPANY IDENTIFIER NOT AVAILABLE FOR THIS WORKSTATION");
      return () => { active = false; };
    }

    setBusy(true);
    setMessage("LOADING EMPLOYEE DIRECTORY…");
    const loadDirectory = () => getCompanyAtLocation(companyLocationId, companyId, accessAttemptId);
    const request = runHack ? runHack(loadDirectory) : loadDirectory();
    Promise.resolve(request)
      .then((results) => {
        if (!active) return;
        setCompanies(results);
        const companyExists = companyId ? results.some(({ company }) => company.id === companyId) : results.length > 0;
        setMessage(companyExists ? "COMPANY EMPLOYEE DIRECTORY LOADED" : companyId ? "NO EMPLOYEE RECORDS FOUND FOR THIS COMPANY" : "NO COMPANY REGISTERED IN THIS BUILDING");
      })
      .catch((error) => {
        if (active) setMessage(error.message || "THE EMPLOYEE DIRECTORY COULD NOT BE REACHED.");
      })
      .finally(() => { if (active) setBusy(false); });

    return () => { active = false; };
  }, [accessAttemptId, companyId, companyLocationId, isStaffDirectory, runHack]);

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
    setActivity([]);
    setActivityMessage(isBankRecords ? "REQUESTING FINANCIAL LEDGER…" : "REQUESTING GATE ACCESS LOG…");
    setActivityBusy(true);
    try {
      const source = isBankRecords ? "bank" : "gate";
      const loadRecords = () => getNPCRecords(npcID, recordRange.from, recordRange.to, source, facilityLocationId, accessAttemptId);
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
    setSelectedPerson(null);
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
        const loadRecords = () => Promise.all([searchCompanies(term, isCityRecords ? cityId : ""), searchLocations(term)]);
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
      <p className="computer-window__eyebrow">{isInterpol ? `CORPORATE SECURITY OFFICE · ${countryName ? `${countryName.toUpperCase()} ` : "COUNTRYWIDE "}PERSONNEL SEARCH` : isStaffDirectory ? "MERIDIAN GLOBAL · EMPLOYEE DIRECTORY" : isCityRecords ? "LYON MUNICIPAL ARCHIVE · PUBLIC INDEX" : isBankRecords ? "MERIDIAN GLOBAL SETTLEMENT OFFICE · PRIVILEGED SESSION" : "MERIDIAN GLOBAL · GATE CONTROL ARCHIVE · PRIVILEGED SESSION"}</p>
      <h1>{isInterpol ? "COUNTRYWIDE PERSON SEARCH" : isStaffDirectory ? "STAFF DIRECTORY" : isCityRecords ? "CITY RECORDS SEARCH" : isBankRecords ? "FINANCIAL LEDGER SEARCH" : "GATE MOVEMENT SEARCH"}</h1>
      {isInterpol && <p className="world-directory__scope">SEARCH JURISDICTION · {countryName ? countryName.toUpperCase() : "CURRENT COUNTRY"}</p>}
      {!isStaffDirectory && !selectedPerson && <form className="world-directory__search" onSubmit={submitSearch}>
        <label className="world-directory__label" htmlFor="world-directory-query">{isStaffDirectory ? "COMPANY NAME" : isPeopleSearch ? "PERSON NAME" : "COMPANY OR ADDRESS"}</label>
        <div className="world-directory__searchline">
          <input id="world-directory-query" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={isStaffDirectory ? "Search a company directory" : isPeopleSearch ? "Search personnel files" : "Search companies and locations"} />
          <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH"}</button>
        </div>
      </form>}
      {!selectedPerson && <p className="world-directory__message" role="status" aria-live="polite">{message}</p>}

      {isPeopleSearch && !selectedPerson && people.map((record) => {
        const { npc } = record;
        return <button className="world-person-result" type="button" key={npc.id} onClick={() => {
          activityRequestId.current++;
          setActivity([]);
          setActivityMessage("");
          setActivityBusy(false);
          setSelectedPerson(record);
        }}>
          <figure className="world-directory__person-portrait"><ModularFace visual={npc.visual ?? FALLBACK_NPC_VISUAL} name={npc.name} /></figure>
          <span className="world-person-result__copy"><strong>{npc.name}</strong><small>{isInterpol ? "PERSONNEL FILE" : isBankRecords ? "ACCOUNT HOLDER" : "TRAVEL SUBJECT"} · {npc.status?.toUpperCase() || "STATUS UNKNOWN"}</small></span>
          <span className="world-person-result__open">OPEN FILE →</span>
        </button>;
      })}

      {selectedPerson && (() => {
        const { npc, workplaces = [], maritalStatus, reportedScene } = selectedPerson;
        const employment = workplaces.length ? workplaces.map(({ company, address }) => `${company} — ${address}`).join("; ") : "NOT LISTED";
        return <section className="world-person-file" aria-label={`${npc.name} details`}>
          <header className="world-person-file__toolbar"><button type="button" onClick={() => { activityRequestId.current++; setActivity([]); setActivityMessage(""); setSelectedPerson(null); }}>← BACK TO RESULTS</button><span>{isInterpol ? "PERSONNEL FILE" : isBankRecords ? "ACCOUNT HOLDER" : "TRAVEL SUBJECT"}</span></header>
          <div className="world-person-file__identity">
            <figure className="world-directory__person-portrait"><ModularFace visual={npc.visual ?? FALLBACK_NPC_VISUAL} name={npc.name} /></figure>
            <div><h2>{npc.name}</h2><p>{npc.status?.toUpperCase() || "STATUS UNKNOWN"} · {npc.age} YEARS</p></div>
          </div>

          {isInterpol && <>
            <dl className="world-record__personnel-facts world-person-file__facts">
              <div><dt>DESCRIPTION</dt><dd>{npc.appearance || "NOT RECORDED"}</dd></div>
              <div><dt>EMPLOYMENT</dt><dd>{employment}</dd></div>
              <div><dt>MARITAL STATUS</dt><dd>{maritalStatus?.toUpperCase() || "UNKNOWN"}</dd></div>
              {reportedScene && <div><dt>REPORTED SCENE</dt><dd>{reportedScene.address}</dd></div>}
            </dl>
            <button className="whiteboard-pin-button" type="button" aria-label={`Pin personnel file for ${npc.name}`} onClick={() => onPin?.({
              id: `person:${npc.id}`, source: "person", title: npc.name,
              content: [`AGE · ${npc.age} YEARS`, `STATUS · ${npc.status?.toUpperCase() || "UNKNOWN"}`, `DESCRIPTION · ${npc.appearance || "NOT RECORDED"}`, `EMPLOYMENT · ${employment}`, `MARITAL STATUS · ${maritalStatus?.toUpperCase() || "UNKNOWN"}`, ...(reportedScene ? [`REPORTED SCENE · ${reportedScene.address}`] : [])].join("\n"),
              metadata: selectedPerson,
            })}><MapPin aria-hidden="true" /></button>
          </>}

          {(isBankRecords || isGateRecords) && <section className="world-record__activity" aria-label={`${isBankRecords ? "Financial ledger" : "GATE travel file"} for ${npc.name}`}>
            <h3>{isBankRecords ? "FINANCIAL LEDGER SEARCH" : "GATE TRAVEL HISTORY"}</h3>
            <form className="world-record__activity-filter" onSubmit={(event) => loadNPCRecords(event, npc.id)}>
              <DatePicker label="FROM" value={recordRange.from} minDate={rangeStart} maxDate={rangeEnd} onChange={updateRecordFrom} />
              <DatePicker label="THROUGH" value={recordRange.to} minDate={rangeStart} maxDate={rangeEnd} onChange={updateRecordThrough} />
              <button type="submit" disabled={activityBusy}>{activityBusy ? "READING…" : isBankRecords ? "READ FINANCIAL RECORDS" : "READ GATE LOGS"}</button>
            </form>
            {activityMessage && <p className="world-record__activity-message" role="status" aria-live="polite">{activityMessage}</p>}
            {activity.map((record) => <article className="world-record__activity-entry" key={record.id}>
              <header><span>{isGateRecords ? "GATE ACCESS LOG" : "FINANCIAL LEDGER"}</span><time dateTime={record.occurredAt}>{displayRecordDate(record.occurredAt, timeZone)}</time></header>
              <p>{record.description}</p><p>{record.operator}{record.reference ? ` · ${record.reference}` : ""}</p>
              {isBankRecords && record.amount > 0 && <p className={`world-record__amount world-record__amount--${record.direction || "outgoing"}`}><span>{record.direction === "incoming" ? "CREDIT · RECEIVED" : "DEBIT · SENT"}</span><strong>{record.direction === "incoming" ? "+" : "−"}{record.amount} CREDITS</strong></p>}
              {(record.origin || record.destination) && <div className="world-record__route" aria-label={`Recorded route from ${record.origin || "unknown location"} to ${record.destination || "unknown location"}`}><span>RECORDED PASSAGE</span><div><strong>{record.origin || "UNKNOWN ORIGIN"}</strong><b aria-hidden="true">→</b><strong>{record.destination || "UNKNOWN DESTINATION"}</strong></div></div>}
              <button className="whiteboard-pin-button" type="button" aria-label={`Pin ${isGateRecords ? "GATE log" : "financial record"}`} onClick={() => onPin?.({ id: record.id, source: record.source, title: isGateRecords ? "GATE ACCESS LOG" : "FINANCIAL RECORD", content: `${record.description}${record.amount > 0 ? ` · ${record.amount} Credits` : ""}${record.origin || record.destination ? ` · ${record.origin || "UNKNOWN"} → ${record.destination || "UNKNOWN"}` : ""}`, metadata: record })}><MapPin aria-hidden="true" /></button>
            </article>)}
          </section>}
        </section>;
      })()}

      {isCityRecords && companies.map(({ company, address }) => (
        <article className="world-record" key={company.id}>
          <header className="world-record__header"><h2>{company.name}</h2><span>COMPANY</span></header>
          <p>HEADQUARTERS · {address.full}</p>
          <button
            className="whiteboard-pin-button"
            type="button"
            aria-label={`Pin company file for ${company.name}`}
            onClick={() => onPin?.({
              id: company.id,
              source: "company",
              title: company.name,
              content: `HEADQUARTERS · ${address.full}`,
              metadata: { company, address },
            })}
          >
            <MapPin aria-hidden="true" />
          </button>
        </article>
      ))}

      {isStaffDirectory && staffCompanies.map(({ company, address, employees }) => (
        <article className="world-record" key={company.id}>
          <header className="world-record__header"><h2>{company.name}</h2><span>EMPLOYER FILE</span></header>
          <p>HEADQUARTERS · {address.full}</p>
          <p>EMPLOYEES · {employees.length}</p>
          <div className="world-record__staff-list" aria-label={`Employees at ${company.name}`}>
            {employees.map((employee) => <article className="world-record__staff-member" key={employee.id}>
              <figure className="world-record__staff-portrait">
                <ModularFace visual={employee.visual ?? FALLBACK_NPC_VISUAL} name={employee.name} />
              </figure>
              <div><strong>{employee.name}</strong><span>{employee.role?.toUpperCase() || "ROLE NOT LISTED"}</span><small>TELEPHONE · {phoneNumberForNPC(employee.id)}</small></div>
            </article>)}
          </div>
        </article>
      ))}

      {isCityRecords && locations.map(({ address }) => (
        <article className="world-record" key={address.locationId}>
          <header className="world-record__header"><h2>{address.name}</h2><span>{address.kind.toUpperCase()}</span></header>
          <p>{address.full}</p>
        </article>
      ))}
    </div>
  );
}
