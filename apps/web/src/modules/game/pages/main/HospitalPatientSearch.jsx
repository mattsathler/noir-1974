import { useState } from "react";
import { searchHospitalPatients } from "../../../world/api.js";
import HospitalPatientRecord from "./HospitalPatientRecord.jsx";
import "./HospitalPatientSearch.css";

export default function HospitalPatientSearch({ accessAttemptId, timeZone, onPrint, onOpenPatient }) {
  const [query, setQuery] = useState("");
  const [patients, setPatients] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [busy, setBusy] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [error, setError] = useState("");

  async function submit(event) {
    event.preventDefault();
    if (!query.trim() || busy) return;
    setBusy(true);
    setError("");
    setSelectedPatient(null);
    try {
      setPatients(await searchHospitalPatients(query.trim(), accessAttemptId));
      setHasSearched(true);
    } catch (cause) {
      setPatients([]);
      setHasSearched(false);
      setError(cause.message || "PATIENT SEARCH IS UNAVAILABLE");
    } finally {
      setBusy(false);
    }
  }

  function openPatient(patient) {
    if (onOpenPatient) onOpenPatient(patient);
    else setSelectedPatient(patient);
  }

  return <main className="hospital-search">
    {!selectedPatient && <>
      <header className="hospital-search__header">
        <p>HOSPITAL RECORDS · PATIENT SERVICES</p>
        <h2>PATIENT SEARCH</h2>
        <span>Search the current city’s medical index by patient name.</span>
      </header>
      <form className="hospital-search__form" onSubmit={submit}>
        <label htmlFor="hospital-patient-query">PATIENT NAME</label>
        <div className="hospital-search__searchline">
          <input id="hospital-patient-query" autoComplete="off" maxLength={80} value={query} onChange={(event) => { setQuery(event.target.value); setHasSearched(false); setPatients([]); setSelectedPatient(null); setError(""); }} placeholder="Enter a full or partial name" />
          <button type="submit" disabled={busy || !query.trim()}>{busy ? "SEARCHING…" : "SEARCH"}</button>
        </div>
      </form>
      {error && <p className="hospital-search__error" role="alert">{error}</p>}
      {!busy && hasSearched && patients.length === 0 && !error && <p className="hospital-search__empty">NO PATIENT RECORDS MATCH THIS SEARCH.</p>}
      {patients.length > 0 && <section className="hospital-search__results" aria-label="Patient search results" aria-live="polite">
        <p className="hospital-search__count">{patients.length} PATIENT RECORD{patients.length === 1 ? "" : "S"} FOUND · SELECT A FILE</p>
        {patients.map((patient) => <article className="hospital-search__result" key={patient.id}>
          <button className="hospital-search__name-link" type="button" aria-label={`${patient.name} · ${patient.status === "dead" ? "POST-MORTEM FILE AVAILABLE" : "PATIENT FILE"} · FINGERPRINT ${patient.fingerprintCode}`} onClick={() => openPatient(patient)}>{patient.name}</button>
          <small className="hospital-search__result-meta">{patient.status === "dead" ? "POST-MORTEM FILE AVAILABLE" : "PATIENT FILE"} · FINGERPRINT {patient.fingerprintCode}</small>
        </article>)}
      </section>}
    </>}
    {selectedPatient && <HospitalPatientRecord patient={selectedPatient} timeZone={timeZone} onPrint={onPrint} onBack={() => setSelectedPatient(null)} />}
  </main>;
}
