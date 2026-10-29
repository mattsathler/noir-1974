import { Printer } from "pixelarticons/react/Printer";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import "./HospitalPatientSearch.css";

const ROUTINE_AUTOPSY_DETAILS = {
  examiners: ["Dr. Élise Moreau", "Dr. Henri Valette", "Dr. Mireille Laurent", "Dr. André Perrin", "Dr. Colette Marchand"],
  laboratories: ["Central Municipal Pathology Laboratory", "Saint-Martin Medical Laboratory", "Meridian Regional Laboratory", "Hôpital Sainte-Claire · Pathology Unit"],
  procedures: ["Standard external and internal examination · Protocol 4B", "Routine post-mortem examination · Form PM-12", "External examination and routine tissue review · Protocol 2A"],
  assistants: ["Luc Bernard", "Jeanne Mercier", "Pauline Girard", "Étienne Faure"],
};

function stableIndex(value, length) {
  let hash = 0;
  for (const character of value) hash = (hash * 31 + character.charCodeAt(0)) >>> 0;
  return hash % length;
}

function routineAutopsyDetails(patient) {
  const key = `${patient.id}:${patient.postMortem?.reportId || "report"}`;
  return Object.fromEntries(Object.entries(ROUTINE_AUTOPSY_DETAILS).map(([field, values], fieldIndex) => [
    field,
    values[stableIndex(`${key}:${fieldIndex}`, values.length)],
  ]));
}

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

function printReport(patient, timeZone, onPrint) {
  const report = patient.postMortem;
  if (!report) return;
  onPrint?.({
    id: report.reportId,
    source: "case",
    title: `POST-MORTEM REPORT · ${patient.name}`,
    content: `ESTIMATED TIME OF DEATH · ${formatDeathTime(report.estimatedDeathAt, timeZone)}\nCAUSE · ${report.cause}${report.weapon ? ` · ${report.weapon}` : ""}\nBODY FOUND · ${report.bodyFoundAt}`,
    metadata: { ...report, npc: { id: patient.id, name: patient.name, visual: patient.visual } },
  });
}

export default function HospitalPatientRecord({ patient, timeZone, onPrint, onBack }) {
  const autopsyDetails = patient.postMortem ? routineAutopsyDetails(patient) : null;
  return <main className="hospital-search hospital-search--record">
    <header className="hospital-search__header">
      <p>HOSPITAL RECORDS · PATIENT SERVICES</p>
      <h2>{patient.status === "dead" ? "POST-MORTEM FILE" : "PATIENT FILE"}</h2>
      <span>Authorized medical index record · Fingerprint {patient.fingerprintCode}</span>
    </header>
    <section className="hospital-search__file" aria-live="polite">
      {onBack && <div className="world-person-file__toolbar"><button type="button" onClick={onBack}>← BACK TO RESULTS</button><span>HOSPITAL RECORD</span></div>}
      <div className="hospital-search__identity">
        {patient.visual && <figure className="hospital-search__portrait"><ModularFace visual={patient.visual} name={patient.name} /></figure>}
        <div><p className="hospital-search__eyebrow">{patient.status.toUpperCase()}</p><h3>{patient.name}</h3></div>
        <span className="hospital-search__fingerprint"><small>FINGERPRINT CODE</small>{patient.fingerprintCode}</span>
      </div>
      {patient.postMortem && <section className="hospital-search__report">
        <div className="hospital-search__report-heading"><h4>POST-MORTEM REPORT</h4><button className="page-print-button" type="button" aria-label={`Print post-mortem report for ${patient.name}`} title="Print report" onClick={() => printReport(patient, timeZone, onPrint)}><Printer aria-hidden="true" /></button></div>
        <dl>
          <div><dt>ESTIMATED TIME OF DEATH</dt><dd>{formatDeathTime(patient.postMortem.estimatedDeathAt, timeZone)}</dd></div>
          <div><dt>CAUSE</dt><dd>{patient.postMortem.cause}{patient.postMortem.weapon ? ` · ${patient.postMortem.weapon}` : ""}</dd></div>
          <div><dt>BODY FOUND</dt><dd>{patient.postMortem.bodyFoundAt}</dd></div>
          <div><dt>EXAMINING PATHOLOGIST</dt><dd>{autopsyDetails.examiners}</dd></div>
          <div><dt>LABORATORY</dt><dd>{autopsyDetails.laboratories}</dd></div>
          <div><dt>EXAMINATION PROCEDURE</dt><dd>{autopsyDetails.procedures}</dd></div>
          <div><dt>ASSISTING TECHNICIAN</dt><dd>{autopsyDetails.assistants}</dd></div>
        </dl>
        {patient.postMortem.evidence?.length > 0 && <div className="hospital-search__traces">
          <h4>POST-MORTEM TRACES</h4>
          {patient.postMortem.evidence.map((item) => <article className="hospital-search__trace" key={item.id}><p><strong>{item.title}</strong><span>{item.content}</span></p><button className="page-print-button" type="button" aria-label={`Print ${item.title}`} title="Print page" onClick={() => onPrint?.(item)}><Printer aria-hidden="true" /></button></article>)}
        </div>}
      </section>}
    </section>
  </main>;
}
