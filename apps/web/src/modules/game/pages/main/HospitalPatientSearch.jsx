import { useState } from "react";
import { MapPin } from "pixelarticons/react/MapPin";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import { searchHospitalPatients } from "../../../world/api.js";
import "./HospitalPatientSearch.css";

function formatDeathTime(value, timeZone) {
  if (!value) return "UNKNOWN";
  return new Intl.DateTimeFormat("en-US", {
    timeZone,
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

export default function HospitalPatientSearch({ accessAttemptId, timeZone, onPin }) {
  const [query, setQuery] = useState("");
  const [patients, setPatients] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function submit(event) {
    event.preventDefault();
    if (!query.trim() || busy) return;
    setBusy(true);
    setError("");
    setSelectedPatient(null);
    try {
      setPatients(await searchHospitalPatients(query.trim(), accessAttemptId));
    } catch (cause) {
      setPatients([]);
      setError(cause.message || "PATIENT SEARCH IS UNAVAILABLE");
    } finally {
      setBusy(false);
    }
  }

  function pinReport(patient) {
    const report = patient.postMortem;
    if (!report) return;
    onPin?.({
      id: report.reportId,
      source: "case",
      title: `POST-MORTEM REPORT · ${patient.name}`,
      content: `ESTIMATED TIME OF DEATH · ${formatDeathTime(report.estimatedDeathAt, timeZone)}\nCAUSE · ${report.cause}${report.weapon ? ` · ${report.weapon}` : ""}\nBODY FOUND · ${report.bodyFoundAt}`,
      metadata: { ...report, npc: { id: patient.id, name: patient.name, visual: patient.visual } },
    });
  }

  return <main className="hospital-search">
    <header className="hospital-search__header">
      <p>CITY HOSPITAL · AUTHORIZED LOCAL SYSTEM</p>
      <h2>PATIENT SEARCH</h2>
      <span>Search the current city’s medical index by patient name.</span>
    </header>
    <form className="hospital-search__form" onSubmit={submit}>
      <label htmlFor="hospital-patient-query">PATIENT NAME</label>
      <div className="hospital-search__searchline">
        <input id="hospital-patient-query" autoComplete="off" maxLength={80} value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Enter a full or partial name" />
        <button type="submit" disabled={busy || !query.trim()}>{busy ? "SEARCHING…" : "SEARCH"}</button>
      </div>
    </form>
    {error && <p className="hospital-search__error" role="alert">{error}</p>}

    {selectedPatient ? <section className="hospital-search__file" aria-live="polite">
      <div className="world-person-file__toolbar">
        <button type="button" onClick={() => setSelectedPatient(null)}>← BACK TO RESULTS</button>
        <span>{selectedPatient.status === "dead" ? "POST-MORTEM FILE" : "PATIENT FILE"}</span>
      </div>
      <div className="hospital-search__identity">
        {selectedPatient.visual && <figure className="hospital-search__portrait"><ModularFace visual={selectedPatient.visual} name={selectedPatient.name} /></figure>}
        <div><p className="hospital-search__eyebrow">{selectedPatient.status.toUpperCase()}</p><h3>{selectedPatient.name}</h3></div>
        <span className="hospital-search__fingerprint"><small>FINGERPRINT CODE</small>{selectedPatient.fingerprintCode}</span>
      </div>
      {selectedPatient.postMortem && <section className="hospital-search__report">
        <div className="hospital-search__report-heading"><h4>POST-MORTEM REPORT</h4><button className="whiteboard-pin-button" type="button" aria-label={`Pin post-mortem report for ${selectedPatient.name}`} onClick={() => pinReport(selectedPatient)}><MapPin aria-hidden="true" /></button></div>
        <dl>
          <div><dt>ESTIMATED TIME OF DEATH</dt><dd>{formatDeathTime(selectedPatient.postMortem.estimatedDeathAt, timeZone)}</dd></div>
          <div><dt>CAUSE</dt><dd>{selectedPatient.postMortem.cause}{selectedPatient.postMortem.weapon ? ` · ${selectedPatient.postMortem.weapon}` : ""}</dd></div>
          <div><dt>BODY FOUND</dt><dd>{selectedPatient.postMortem.bodyFoundAt}</dd></div>
        </dl>
        {selectedPatient.postMortem.evidence?.length > 0 && <div className="hospital-search__traces">
          <h4>POST-MORTEM TRACES</h4>
          {selectedPatient.postMortem.evidence.map((item) => <article className="hospital-search__trace" key={item.id}><p><strong>{item.title}</strong><span>{item.content}</span></p><button className="whiteboard-pin-button" type="button" aria-label={`Pin ${item.title}`} onClick={() => onPin?.(item)}><MapPin aria-hidden="true" /></button></article>)}
        </div>}
      </section>}
    </section> : <>
      {!busy && query && patients.length === 0 && !error && <p className="hospital-search__empty">NO PATIENT RECORDS MATCH THIS SEARCH.</p>}
      {patients.length > 0 && <section className="hospital-search__results" aria-label="Patient search results" aria-live="polite">
        <p className="hospital-search__count">{patients.length} PATIENT RECORD{patients.length === 1 ? "" : "S"} FOUND · SELECT A FILE</p>
        {patients.map((patient) => <button className="hospital-search__result" key={patient.id} type="button" onClick={() => setSelectedPatient(patient)}>
          <span className="hospital-search__result-copy"><strong>{patient.name}</strong><small>{patient.status === "dead" ? "POST-MORTEM FILE AVAILABLE" : "PATIENT FILE"} · FINGERPRINT {patient.fingerprintCode}</small></span>
          <span className="hospital-search__result-open">OPEN FILE →</span>
        </button>)}
      </section>}
    </>}
  </main>;
}
