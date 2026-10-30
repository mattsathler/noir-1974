import { useEffect, useMemo, useState } from "react";
import { getWhiteboard, getWhiteboards, markWhiteboardSubmitted, submitCaseAccusation } from "../../../world/api.js";

function boardReadiness(board) {
  const pins = board?.pins ?? [];
  const hasProof = pins.some((pin) => pin.designation === "proof" && pin.source !== "note");
  const hasAccused = pins.some((pin) => pin.designation === "accused" && pin.source === "person");
  const hasVictim = pins.some((pin) => pin.designation === "victim" && pin.source === "person");
  return { hasProof, hasAccused, hasVictim, ready: hasProof && hasAccused && hasVictim };
}

export default function AccusationWindow({ stationLocationId, onClose, standalone = false }) {
  const [boards, setBoards] = useState([]);
  const [selectedId, setSelectedId] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [message, setMessage] = useState("SELECT AN INVESTIGATION BOARD TO SUBMIT TO CORPORATE REVIEW.");
  const selectedBoard = boards.find((board) => board.id === selectedId);
  const readiness = useMemo(() => boardReadiness(selectedBoard), [selectedBoard]);

  useEffect(() => {
    let active = true;
    getWhiteboards().then(async (items) => {
      const loaded = await Promise.all(items.map(async (item) => {
        try { return await getWhiteboard(item.id); }
        catch { return { ...item, pins: [] }; }
      }));
      if (!active) return;
      setBoards(loaded);
      setSelectedId(loaded.find((board) => !board.lockedAt)?.id ?? loaded[0]?.id ?? "");
      setMessage(loaded.length ? "SELECT A BOARD FOR THE ACCUSATION REPORT." : "NO INVESTIGATION BOARDS ARE AVAILABLE.");
    }).catch((error) => { if (active) setMessage(error.message || "INVESTIGATION BOARDS COULD NOT BE LOADED."); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (!submitted || !selectedId) return undefined;
    let active = true;
    const refreshBoardStatus = async () => {
      try {
        const latest = await getWhiteboards();
        if (!active) return;
        setBoards((previous) => latest.map((board) => ({
          ...board,
          pins: previous.find((item) => item.id === board.id)?.pins ?? [],
        })));
        const selected = latest.find((board) => board.id === selectedId);
        if (selected && !selected.lockedAt) {
          setSubmitted(false);
          setMessage("REVIEW COMPLETE. THIS BOARD IS UNLOCKED AND READY FOR ANOTHER REVIEW.");
        }
      } catch { /* Keep the submitted state while the API is temporarily unavailable. */ }
    };
    const interval = window.setInterval(refreshBoardStatus, 5000);
    return () => {
      active = false;
      window.clearInterval(interval);
    };
  }, [submitted, selectedId]);

  async function submit(event) {
    event.preventDefault();
    if (!selectedBoard || !readiness.ready || busy || submitted) return;
    setBusy(true);
    setMessage("FORWARDING ACCUSATION FOR REVIEW…");
    try {
      const result = await submitCaseAccusation(selectedBoard.id, stationLocationId);
      markWhiteboardSubmitted(selectedBoard.id);
      setBoards((previous) => previous.map((board) => board.id === selectedBoard.id ? { ...board, lockedAt: new Date().toISOString() } : board));
      setSubmitted(true);
      setMessage(result.status === "completed"
        ? "THIS BOARD WAS ALREADY REVIEWED. NO NEW REVIEW OR EMAIL WAS CREATED. CHECK PERSONAL MAIL FOR THE SAVED DECISION; USE A NEW BOARD FOR ANOTHER REVIEW."
        : "ACCUSED BOARD RECEIVED. REVIEW IS IN PROGRESS. THE DECISION WILL ARRIVE BY PERSONAL MAIL IN ABOUT 3 MINUTES.");
    } catch (error) {
      setMessage(error.message || "THE ACCUSATION COULD NOT BE SUBMITTED.");
    } finally { setBusy(false); }
  }

  const panel = <section className={`warrant-window__panel accusation-window__panel${standalone ? " police-restricted-app__panel" : ""}`}>
      <header className="warrant-window__header"><div><p className="computer-window__eyebrow">CORPORATE SECURITY OFFICE · FORM 74-C</p><h1 id="accusation-title">MURDER ACCUSATION</h1></div><button type="button" aria-label="Close accusation form" onClick={onClose}>×</button></header>
      <p className="warrant-window__copy">Submit an investigation board naming the accused person, the victim, and supporting proof. The review decision will be delivered to personal mail.</p>
      <form className="accusation-window__form" onSubmit={submit}>
        <label className="world-directory__label" htmlFor="accusation-board">INVESTIGATION BOARD</label>
        <select id="accusation-board" value={selectedId} disabled={loading || busy || submitted || boards.length === 0} onChange={(event) => setSelectedId(event.target.value)}>
          {boards.length === 0 ? <option value="">NO BOARDS AVAILABLE</option> : boards.map((board) => {
            const status = boardReadiness(board);
            return <option key={board.id} value={board.id} disabled={Boolean(board.lockedAt)}>{board.title} · {board.lockedAt ? "SUBMITTED · LOCKED" : status.ready ? "READY" : "INCOMPLETE"}</option>;
          })}
        </select>
        {selectedBoard && <ul className="accusation-window__checklist" aria-label="Required board contents">
          <li className={readiness.hasProof ? "is-ready" : "is-missing"}>{readiness.hasProof ? "✓" : "!"} AT LEAST ONE PIN MARKED PROOF</li>
          <li className={readiness.hasAccused ? "is-ready" : "is-missing"}>{readiness.hasAccused ? "✓" : "!"} ONE PERSON MARKED ACCUSED</li>
          <li className={readiness.hasVictim ? "is-ready" : "is-missing"}>{readiness.hasVictim ? "✓" : "!"} ONE PERSON MARKED VICTIM</li>
        </ul>}
        {selectedBoard && !readiness.ready && <p className="accusation-window__warning" role="alert">This board is missing required pins. Complete all three items above before submitting.</p>}
        {selectedBoard?.lockedAt && <p className="accusation-window__warning" role="alert">This board is locked while its review is in progress. It will unlock when the decision is delivered.</p>}
        <p className="world-directory__message" role="status" aria-live="polite">{loading ? "LOADING INVESTIGATION BOARDS…" : message}</p>
        <footer className="warrant-window__actions"><button type="submit" disabled={loading || busy || submitted || selectedBoard?.lockedAt || !readiness.ready}>{busy ? "SUBMITTING…" : submitted ? "SUBMITTED" : "SUBMIT ACCUSATION"}</button></footer>
      </form>
    </section>;
  return standalone
    ? <section className="police-restricted-app">{panel}</section>
    : <div className="warrant-window" role="dialog" aria-modal="true" aria-labelledby="accusation-title">{panel}</div>;
}
