import { useEffect, useRef, useState } from "react";
import { Phone } from "pixelarticons/react/Phone";
import { getInterrogationAnswers, lookupPhoneContact } from "../../../world/api.js";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";

const KEYS = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "*", "0", "#"];
const KEY_LEGENDS = { 2: "ABC", 3: "DEF", 4: "GHI", 5: "JKL", 6: "MNO", 7: "PQRS", 8: "TUV", 9: "WXYZ" };
const PHONE_LENGTH = 7;
const FALLBACK_NPC_VISUAL = { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", faceId: "", beardId: "", accessoryIds: [] };
const INTERVIEW_QUESTIONS = [
  { id: "routine", text: "How has your routine been lately?" },
  { id: "associates", text: "Who do you usually spend time with?" },
  { id: "unusual", text: "Has anything unusual happened around here recently?" },
  { id: "work", text: "How are things at work?" },
  { id: "worries", text: "Is anything worrying you?" },
];

function formatNumber(value) {
  const digits = value.replace(/\D/gu, "").slice(0, PHONE_LENGTH);
  if (digits.length <= 3) return digits;
  return `${digits.slice(0, 3)}-${digits.slice(3)}`;
}

export default function Telephone({ onClose, onNotification }) {
  const [number, setNumber] = useState("");
  const [callState, setCallState] = useState("idle");
  const [contact, setContact] = useState(null);
  const [selectedQuestion, setSelectedQuestion] = useState(-1);
  const [answer, setAnswer] = useState("");
  const [interviewAnswers, setInterviewAnswers] = useState(null);
  const [usedAnswers, setUsedAnswers] = useState({});
  const [interviewLoading, setInterviewLoading] = useState(false);
  const [interviewError, setInterviewError] = useState(false);
  const callTimer = useRef(null);
  const callRequest = useRef(0);

  useEffect(() => () => window.clearTimeout(callTimer.current), []);

  useEffect(() => {
    if (callState === "connected" && contact) onNotification?.({ type: "CALL", title: `Connected to ${contact.name}`, message: "Outgoing line is open." });
  }, [callState, contact, onNotification]);

  useEffect(() => {
    if (callState !== "connected" || !contact) return undefined;
    let active = true;
    setInterviewLoading(true);
    getInterrogationAnswers(contact.id)
      .then((result) => { if (active) setInterviewAnswers(result.answers ?? {}); })
      .catch(() => { if (active) { setInterviewAnswers({}); setInterviewError(true); } })
      .finally(() => { if (active) setInterviewLoading(false); });
    return () => { active = false; };
  }, [callState, contact]);

  function resetCall() {
    callRequest.current += 1;
    window.clearTimeout(callTimer.current);
    setCallState("idle");
    setContact(null);
    setSelectedQuestion(-1);
    setAnswer("");
    setInterviewAnswers(null);
    setInterviewLoading(false);
    setInterviewError(false);
  }

  function pressKey(key) {
    if (!/\d/u.test(key)) return;
    resetCall();
    setNumber((current) => current.length < PHONE_LENGTH ? `${current}${key}` : current);
  }

  async function handleCallButton() {
    if (callState === "calling" || callState === "connected") {
      resetCall();
      return;
    }
    const requestId = ++callRequest.current;
    setContact(null);
    setInterviewError(false);
    setCallState("calling");
    try {
      const { contact: match } = await lookupPhoneContact(number);
      if (requestId !== callRequest.current) return;
      if (match.status !== "alive") {
        setCallState("dead");
        return;
      }
      setContact(match);
      const answerDelay = 900 + Math.floor(Math.random() * 2101);
      callTimer.current = window.setTimeout(() => {
        if (requestId === callRequest.current) setCallState("connected");
      }, answerDelay);
    } catch (error) {
      if (requestId !== callRequest.current) return;
      setCallState(error.status === 404 ? "not-found" : "directory-error");
    }
  }

  return (
    <div className="telephone-overlay" role="dialog" aria-modal="true" aria-label="Telephone">
      <section className="telephone-device">
        <header className="telephone-device__header"><span>MERIDIAN · PRIVATE LINE</span><button type="button" aria-label="Close telephone" onClick={onClose}>×</button></header>
        <div className="telephone-device__body">
          <div className={`telephone-device__handset${callState === "calling" || callState === "connected" ? " telephone-device__handset--active" : ""}`} aria-hidden="true">
            <svg className="telephone-device__receiver" viewBox="0 0 24 24">
              <path fill="#e0bf68" d="M4 1h5v2H4zm5 2h2v4H9zM7 7h2v4H7zm-3 5h2v2H4zM2 3h2v9H2zm7 8h2v2H9zm2 2h2v2h-2zm2 2h4v2h-4zm4-2h4v2h-4zm4 2h2v5h-2zM6 14h2v2H6zm2 2h2v2H8zm2 2h2v2h-2zm2 2h9v2h-9z" />
            </svg>
          </div>
          <div className="telephone-app">
            {callState === "connected" && contact ? <section className="telephone-interrogation" aria-label={`Telephone interview with ${contact.name}`}>
              <header className="telephone-interrogation__header"><strong>{contact.name}</strong><span>{contact.company}</span></header>
              <div className="telephone-interrogation__portrait"><ModularFace visual={contact.visual ?? FALLBACK_NPC_VISUAL} name={contact.name} /></div>
              <div className="telephone-interrogation__answer" role="status" aria-live="polite">
                {answer
                  ? <Typewriter key={`${contact.id}:${selectedQuestion}:${answer}`} text={answer} speed={24} />
                  : interviewLoading ? "The line is open. The caller is gathering their thoughts." : interviewError ? "The line is open, but no testimony could be retrieved." : "The line is open. Choose a question."}
              </div>
              <div className="telephone-interrogation__questions" aria-label="Interview questions">
                {INTERVIEW_QUESTIONS.map((question, index) => <button key={question.id} type="button" disabled={interviewLoading || !(interviewAnswers?.[question.id]?.length)} aria-pressed={selectedQuestion === index} onClick={() => {
                  const pool = interviewAnswers?.[question.id] ?? [];
                  const usageKey = `${contact.id}:${question.id}`;
                  const used = usedAnswers[usageKey] ?? [];
                  const available = pool.filter((line) => !used.includes(line));
                  const choices = available.length ? available : pool;
                  if (!choices.length) return;
                  const line = choices[Math.floor(Math.random() * choices.length)];
                  setSelectedQuestion(index);
                  setAnswer(line);
                  setUsedAnswers((previous) => ({ ...previous, [usageKey]: available.length ? [...used, line] : [line] }));
                }}>{question.text}</button>)}
              </div>
              <button className="telephone-interrogation__hangup" type="button" onClick={resetCall} aria-label="End call"><Phone aria-hidden="true" /></button>
            </section> : <>
              <div className="telephone-app__display">
                <input
                  className="telephone-app__number-input"
                  type="tel"
                  inputMode="numeric"
                  autoComplete="off"
                  aria-label="Telephone number"
                  value={formatNumber(number)}
                  onChange={(event) => {
                    resetCall();
                    setNumber(event.target.value.replace(/\D/gu, "").slice(0, PHONE_LENGTH));
                  }}
                />
              </div>
              <p className={`telephone-app__call-status telephone-app__call-status--${callState}`} role="status" aria-live="polite" aria-hidden={callState === "idle"}>
                {callState === "not-found" ? "NUMBER NOT IN DIRECTORY" : callState === "dead" ? "LINE DISCONNECTED" : callState === "directory-error" ? "DIRECTORY UNAVAILABLE" : callState === "calling" ? contact ? `CALLING · ${contact.name}` : "LOOKING UP DIRECTORY…" : ""}
              </p>
              <div className="telephone-app__decorative-actions" aria-hidden="true">
                {["HELP", "HOLD", "CONF.", "OPTION", "DIRECTORY", "CALLERS", "SERVICES", "COPY", "PRINT", "REDIAL"].map((label) => (
                  <button className="telephone-app__decorative-action" key={label} type="button" disabled tabIndex={-1}>{label}</button>
                ))}
              </div>
              <div className="telephone-app__keypad" aria-label="Keypad">
                {KEYS.map((key) => <button key={key} type="button" onClick={() => pressKey(key)} aria-label={`Dial ${key}`}>
                  <span>{key}</span>{KEY_LEGENDS[key] && <small aria-hidden="true">{KEY_LEGENDS[key]}</small>}
                </button>)}
              </div>
              <div className="telephone-app__controls">
                <button className="telephone-app__clear" type="button" onClick={() => { resetCall(); setNumber(""); }} disabled={!number}>CLEAR</button>
                <button className={`telephone-app__dial${callState === "calling" ? " telephone-app__dial--active" : ""}`} type="button" onClick={handleCallButton} disabled={number.length !== PHONE_LENGTH} aria-label={callState === "calling" ? "End call" : "Dial number"}><Phone aria-hidden="true" /></button>
              </div>
            </>}
            <div className="telephone-app__grille" aria-hidden="true"><span /><span /><span /></div>
          </div>
        </div>
      </section>
    </div>
  );
}
