import { useState } from "react";
import { findNPCByEmail, getEmailBreakerAccess, requestEmailBreaker } from "../../../world/api.js";

export default function EmailBreakerApp({ onOpenHackedMailbox = () => {} }) {
  const [email, setEmail] = useState("");
  const [target, setTarget] = useState(null);
  const [busy, setBusy] = useState(false);
  const [lookingUp, setLookingUp] = useState(false);
  const [message, setMessage] = useState("ENTER A STAFF EMAIL AND FORCE THE ACCOUNT BREACH.");
  const targetMatchesInput = Boolean(target && email.trim().toLowerCase() === target.email?.trim().toLowerCase());

  async function lookup(event) {
    event.preventDefault();
    if (!email.trim() || busy || lookingUp) return;
    setLookingUp(true);
    setTarget(null);
    setMessage("SEARCHING COMPANY STAFF DIRECTORIES…");
    try {
      const matches = await findNPCByEmail(email.trim());
      const normalizedEmail = email.trim().toLowerCase();
      if (matches[0]) {
        setTarget(matches[0]);
        setMessage("STAFF ACCOUNT LOCATED. FORCE THE BREACH TO OPEN THEIR MAILBOX.");
      } else {
        const accesses = await getEmailBreakerAccess().catch(() => []);
        const previousAccess = accesses.find((item) => item.email?.trim().toLowerCase() === normalizedEmail);
        if (previousAccess) {
          setTarget({ id: previousAccess.npcId, name: previousAccess.name, email: previousAccess.email, status: "compromised" });
          setMessage("PREVIOUSLY ACCESSED ACCOUNT LOCATED. FORCE THE BREACH TO OPEN THEIR MAILBOX.");
        } else {
          setTarget(null);
          setMessage("NO STAFF ACCOUNT MATCHES THAT EMAIL ADDRESS.");
        }
      }
    } catch (error) { setMessage(error.message || "THE DIRECTORY SEARCH FAILED."); }
    finally { setLookingUp(false); }
  }

  async function submit(event) {
    event.preventDefault();
    if (!targetMatchesInput || busy) return;
    setBusy(true);
    try {
      await requestEmailBreaker(target.id, target.email);
      onOpenHackedMailbox(target);
    } catch (error) { setMessage(error.message || "THE ACCOUNT BREACH FAILED."); }
    finally { setBusy(false); }
  }

  return <main className="email-breaker" aria-busy={busy}>
    <header className="email-breaker__header"><span aria-hidden="true">✉</span><div><p>MERID_OS · UNSIGNED EXECUTABLE</p><h1>EMAIL_BREAKER.EXE</h1></div></header>
    <p className="email-breaker__copy">Force a direct breach of a staff mailbox. You need the recipient's published company email address.</p>
    <form className="email-breaker__form" onSubmit={lookup}>
      <label className="world-directory__label" htmlFor="email-breaker-target">TARGET EMAIL</label>
      <input id="email-breaker-target" type="email" autoComplete="off" required maxLength={160} value={email} onChange={(event) => setEmail(event.target.value)} placeholder="name@meridian-mail.net" />
      <button className="email-breaker__deeplink" type="submit" disabled={busy || lookingUp}>{lookingUp ? "SEARCHING…" : "LOOK UP ACCOUNT ↵"}</button>
    </form>
    {target && <section className="email-breaker__target" aria-label="Email search result">
      <p className="world-directory__label">{target.status === "compromised" ? "PREVIOUSLY ACCESSED ACCOUNT" : "ACCOUNT FOUND"}</p>
      <strong>{target.name}</strong><span>{target.email}</span>
      <small>{target.role?.toUpperCase() || "COMPANY STAFF"} · {target.status?.toUpperCase() || "STATUS UNKNOWN"}</small>
      <form className="email-breaker__form" onSubmit={submit}>
        <button className="email-breaker__deeplink" type="submit" disabled={busy || !targetMatchesInput}>{busy ? "FORCING BREACH…" : "FORCE BREACH ↵"}</button>
      </form>
    </section>}
    <p className="world-directory__message" role="status" aria-live="polite">{message}</p>
  </main>;
}
