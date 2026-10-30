import { useEffect, useMemo, useState } from "react";
import { getCrimeReports } from "../../../world/api.js";
import DatePicker from "../../../world/components/DatePicker.jsx";
import ModularFace from "../../../identity/components/ModularFace.jsx";

const FALLBACK_VISUAL = { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", faceId: "", beardId: "", accessoryIds: [] };

function dateDaysBefore(dateKey, days) {
  const date = new Date(`${dateKey}T12:00:00Z`);
  date.setUTCDate(date.getUTCDate() - days);
  return date.toISOString().slice(0, 10);
}

function formatReportDate(value, timeZone) {
  if (!value) return "DATE NOT RECORDED";
  return new Intl.DateTimeFormat("en-US", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    ...(timeZone ? { timeZone } : {}),
  }).format(new Date(value)).toUpperCase();
}

function todayKey() {
  return new Date().toISOString().slice(0, 10);
}

export default function CaseFiles({ cityId, issueDay, timeZone, initialSearch = "" }) {
  const initialTo = issueDay || todayKey();
  const [cases, setCases] = useState([]);
  const [missing, setMissing] = useState([]);
  const [selected, setSelected] = useState(null);
  const [search, setSearch] = useState(initialSearch);
  const [appliedSearch, setAppliedSearch] = useState(initialSearch);
  const [fromDate, setFromDate] = useState(dateDaysBefore(initialTo, 30));
  const [toDate, setToDate] = useState(initialTo);
  const [appliedRange, setAppliedRange] = useState({ from: dateDaysBefore(initialTo, 30), to: initialTo });
  const [message, setMessage] = useState("CONNECTING TO CITY REPORTS…");
  const [reportError, setReportError] = useState(false);
  const [rangeError, setRangeError] = useState("");
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    const latestDay = issueDay || todayKey();
    const previousMonth = dateDaysBefore(latestDay, 30);
    setFromDate(previousMonth);
    setToDate(latestDay);
    setAppliedRange({ from: previousMonth, to: latestDay });
    setSelected(null);
  }, [cityId, issueDay]);

  useEffect(() => {
    let active = true;
    setCases([]);
    setMissing([]);
    setSelected(null);
    setReportError(false);
    setMessage("SEARCHING CITY RECORDS…");
    getCrimeReports(appliedRange)
      .then((reports) => {
        if (!active) return;
        setCases(reports.cases ?? []);
        setMissing(reports.missing ?? []);
        setMessage("CITY RECORDS SEARCH COMPLETE");
      })
      .catch((error) => { if (active) { setMessage(error.message || "CITY REPORTS ARE UNAVAILABLE"); setReportError(true); } });
    return () => { active = false; };
  }, [cityId, appliedRange, retryCount]);

  const needle = appliedSearch.trim().toLocaleLowerCase();
  const visibleCases = useMemo(() => cases.filter((record) => !needle || (record.victimName ?? "").toLocaleLowerCase().includes(needle)), [cases, needle]);
  const visibleMissing = useMemo(() => missing.filter((person) => !needle || person.name.toLocaleLowerCase().includes(needle)), [missing, needle]);
  const missingIds = useMemo(() => new Set(missing.map((person) => person.npcId)), [missing]);
  const openCaseVictims = useMemo(() => new Set(visibleCases.map((record) => record.victimId)), [visibleCases]);
  const additionalMissing = visibleMissing.filter((person) => !openCaseVictims.has(person.npcId));

  function searchDateRange(event) {
    event.preventDefault();
    setAppliedSearch(search);
    if (document.activeElement?.type === "search") return;
    if (!fromDate || !toDate) { setRangeError("SELECT BOTH DATES TO SEARCH THE CITY REGISTER."); return; }
    if (fromDate > toDate) { setRangeError("THE START DATE MUST COME BEFORE THE END DATE."); return; }
    setRangeError("");
    setAppliedRange({ from: fromDate, to: toDate });
  }

  function openRecord(kind, record) {
    setSelected({ kind, record });
  }

  if (selected) {
    const { kind, record } = selected;
    const name = kind === "case" ? record.victimName : record.name;
    const visual = kind === "case" ? record.victimVisual : record.visual;
    return (
      <section className="case-files case-files--detail" aria-label="Public case file">
        <button className="case-files__back" type="button" onClick={() => setSelected(null)}>← BACK TO CITY REPORTS</button>
        <p className="computer-window__eyebrow">CITY POLICE · PUBLIC CASE FILE</p>
        <div className="case-files__identity">
          <figure className="case-files__portrait"><ModularFace visual={visual ?? FALLBACK_VISUAL} name={name || "Unidentified victim"} /></figure>
          <div>
            <p className="case-files__record-type">{kind === "case" ? "OPEN HOMICIDE INQUIRY" : "MISSING PERSON"}</p>
            <h1>{name || "UNIDENTIFIED VICTIM"}</h1>
          </div>
        </div>
        <p className="case-files__detail-note">The public file contains no further investigative findings.</p>
      </section>
    );
  }

  return (
    <section className="case-files" aria-label="City crime reports">
      <p className="computer-window__eyebrow">CITY POLICE · PUBLIC REPORTS</p>
      <h1>CRIME REPORTS</h1>

      <form className="case-files__filters" onSubmit={searchDateRange}>
        <label className="case-files__name-search">
          <span>SEARCH VICTIM OR MISSING PERSON</span>
          <input type="search" maxLength="80" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="TYPE A NAME · PRESS ENTER" aria-keyshortcuts="Enter" />
        </label>
        <div className="world-record__activity-filter">
          <DatePicker label="FROM" value={fromDate} maxDate={toDate} onChange={setFromDate} />
          <DatePicker label="TO" value={toDate} minDate={fromDate} maxDate={issueDay || todayKey()} onChange={setToDate} />
          <button className="search-enter-submit" type="submit" tabIndex={-1} aria-label="Press Enter to search open cases">ENTER ↵</button>
        </div>
      </form>
      {rangeError && <p className="case-files__range-error" role="alert">{rangeError}</p>}
      <p className="case-files__message" role="status" aria-live="polite">{message}</p>
      {reportError && <button className="case-files__retry" type="button" onClick={() => setRetryCount((count) => count + 1)}>RETRY CITY REPORTS</button>}

      <section className="case-files__section" aria-labelledby="open-cases-title">
        <h2 id="open-cases-title">OPEN INQUIRIES <span>{visibleCases.length}</span></h2>
        {visibleCases.length ? (
          <div className="case-files__list">
            {visibleCases.map((record) => (
              <article className="case-files__record" key={record.id}>
                <span className="case-files__select-status">{missingIds.has(record.victimId) ? "MISSING PERSON · OPEN INQUIRY" : "OPEN HOMICIDE INQUIRY"}</span>
                <button className="case-files__name-link" type="button" onClick={() => openRecord("case", record)}>{record.victimName || "UNIDENTIFIED VICTIM"}</button>
                <time dateTime={record.crimeAt}>REPORTED · {formatReportDate(record.crimeAt, timeZone)}</time>
              </article>
            ))}
          </div>
        ) : <p className="case-files__empty">{needle ? "NO MATCHING OPEN INQUIRIES." : "NO OPEN INQUIRIES IN THIS DATE RANGE."}</p>}
      </section>

      <section className="case-files__section case-files__missing" aria-labelledby="missing-persons-title">
        <h2 id="missing-persons-title">MISSING PERSONS <span>{additionalMissing.length}</span></h2>
        {additionalMissing.length ? (
          <div className="case-files__list">
            {additionalMissing.map((person) => (
              <article className="case-files__record case-files__record--missing" key={person.npcId}>
                <span className="case-files__select-status">MISSING PERSON</span>
                <button className="case-files__name-link" type="button" onClick={() => openRecord("missing", person)}>{person.name}</button>
                <time dateTime={person.missingAt}>MISSING SINCE · {formatReportDate(person.missingAt, timeZone)}</time>
              </article>
            ))}
          </div>
        ) : <p className="case-files__empty">{visibleMissing.length ? "ALL MISSING PERSONS ARE LISTED WITH THEIR OPEN INQUIRY ABOVE." : needle ? "NO MATCHING MISSING PERSONS." : "NO MISSING PERSONS REPORTED IN THIS CITY."}</p>}
      </section>
    </section>
  );
}
