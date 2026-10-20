import { useEffect, useState } from "react";
import { getCaseFile, getNewspaperIssue, getOpenCases, saveCaseProgress } from "../../../world/api.js";

export default function CaseFiles() {
  const [cases, setCases] = useState([]);
  const [issue, setIssue] = useState(null);
  const [activeCase, setActiveCase] = useState(null);
  const [detail, setDetail] = useState(null);
  const [message, setMessage] = useState("CONNECTING TO CASE ARCHIVE…");
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [savingEvidence, setSavingEvidence] = useState("");

  useEffect(() => {
    let active = true;
    Promise.all([getOpenCases(), getNewspaperIssue()])
      .then(([openCases, newspaper]) => {
        if (!active) return;
        setCases(openCases);
        setIssue(newspaper);
        setMessage(openCases.length ? "OPEN CASES RETRIEVED" : "NO OPEN CASES ON FILE");
        if (openCases.length) setActiveCase(openCases[0]);
      })
      .catch((error) => { if (active) setMessage(error.message || "THE CASE ARCHIVE IS UNAVAILABLE"); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (!activeCase) { setDetail(null); return undefined; }
    let active = true;
    setLoadingDetail(true);
    getCaseFile(activeCase.id)
      .then((record) => { if (active) setDetail(record); })
      .catch((error) => { if (active) setMessage(error.message || "THE CASE FILE COULD NOT BE READ"); })
      .finally(() => { if (active) setLoadingDetail(false); });
    return () => { active = false; };
  }, [activeCase]);

  async function toggleEvidence(evidenceID) {
    if (!detail || savingEvidence) return;
    const current = detail.progress.discoveredEvidenceIds ?? [];
    const next = current.includes(evidenceID)
      ? current.filter((id) => id !== evidenceID)
      : [...current, evidenceID];
    setSavingEvidence(evidenceID);
    try {
      const saved = await saveCaseProgress(detail.case.id, next);
      setDetail((previous) => previous ? { ...previous, progress: saved.progress } : previous);
    } catch (error) {
      setMessage(error.message || "CASE NOTES COULD NOT BE SAVED");
    } finally {
      setSavingEvidence("");
    }
  }

  const lead = issue?.lead;
  const selectedHeadline = lead?.caseId === activeCase?.id ? lead.headline : "OPEN HOMICIDE INQUIRY";
  const selectedSummary = lead?.caseId === activeCase?.id ? lead.summary : "The investigation remains open.";
  const filedIDs = detail?.progress?.discoveredEvidenceIds ?? [];

  return (
    <section className="case-files" aria-label="Open case files">
      <p className="computer-window__eyebrow">PRIVATE INVESTIGATION · SERVER RECORDS</p>
      <h1>OPEN CASE FILES</h1>
      <p className="case-files__message" role="status" aria-live="polite">{message}</p>

      {cases.length > 0 && (
        <nav className="case-files__list" aria-label="Open investigations">
          {cases.map((record) => (
            <button
              className={`case-files__select${activeCase?.id === record.id ? " case-files__select--active" : ""}`}
              key={record.id}
              type="button"
              aria-pressed={activeCase?.id === record.id}
              onClick={() => setActiveCase(record)}
            >
              <span>HOMICIDE · {record.status.toUpperCase()}</span>
              <span>{record.id}</span>
            </button>
          ))}
        </nav>
      )}

      {loadingDetail && <p className="case-files__message">READING CASE FILE…</p>}
      {detail && !loadingDetail && (
        <article className="case-files__detail">
          <header className="world-record__header">
            <h2>{selectedHeadline}</h2>
            <span>{detail.case.status.toUpperCase()}</span>
          </header>
          <p>{selectedSummary}</p>
          <p className="case-files__reference">VICTIM RECORD · {detail.case.victimId}</p>
          <p className="case-files__reference">SCENE RECORD · {detail.case.locationId}</p>
          <h2 className="case-files__section-heading">RECORDS FILED</h2>
          {detail.evidence.map((item) => {
            const filed = filedIDs.includes(item.id);
            return (
              <article className="case-files__evidence" key={item.id}>
                <p className="case-files__evidence-title">{item.presentation}</p>
                <p>{item.content}</p>
                <button type="button" disabled={Boolean(savingEvidence)} onClick={() => toggleEvidence(item.id)}>
                  {savingEvidence === item.id ? "SAVING…" : filed ? "REMOVE FROM NOTES" : "FILE IN PERSONAL NOTES"}
                </button>
              </article>
            );
          })}
        </article>
      )}
    </section>
  );
}
