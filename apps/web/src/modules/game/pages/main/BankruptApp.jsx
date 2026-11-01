import { useState } from "react";
import { breachBankAccount, findNPCByBankAccount } from "../../../world/api.js";

export default function BankruptApp({ onOpenBankLedger = () => {} }) {
  const [accountNumber, setAccountNumber] = useState("");
  const [target, setTarget] = useState(null);
  const [busy, setBusy] = useState(false);
  const [lookingUp, setLookingUp] = useState(false);
  const [message, setMessage] = useState("ENTER THE ACCOUNT NUMBER PUBLISHED IN THE STAFF DIRECTORY.");
  const matchesInput = Boolean(target && accountNumber === target.bankAccountNumber);
  const validInput = /^\d{8}$/u.test(accountNumber);

  async function lookup(event) {
    event.preventDefault();
    if (!validInput || busy || lookingUp) return;
    setLookingUp(true);
    setTarget(null);
    setMessage("SEARCHING ACCOUNT DIRECTORY…");
    try {
      const matches = await findNPCByBankAccount(accountNumber);
      setTarget(matches[0] || null);
      setMessage(matches[0] ? "ACCOUNT LOCATED. FORCE THE BREACH TO OPEN THE LEDGER." : "NO STAFF ACCOUNT MATCHES THAT NUMBER.");
    } catch (error) {
      setTarget(null);
      setMessage(error.message || "THE ACCOUNT LOOKUP FAILED.");
    } finally {
      setLookingUp(false);
    }
  }

  async function submit(event) {
    event.preventDefault();
    if (!matchesInput || busy) return;
    setBusy(true);
    try {
      await breachBankAccount(target.id, accountNumber);
      onOpenBankLedger(target);
    } catch (error) {
      setMessage(error.message || "THE BANK ACCOUNT BREACH FAILED.");
    } finally {
      setBusy(false);
    }
  }

  return <main className="bankrupt-app" aria-busy={busy || lookingUp}>
    <header className="bankrupt-app__header">
      <span className="bankrupt-app__seal" aria-hidden="true"><i>U</i><b>×</b></span>
      <div><p>MERID_OS · FINANCIAL ACCESS UTILITY</p><h1>BANKRUPT.EXE</h1></div>
      <span className="bankrupt-app__session"><i /> SESSION READY</span>
    </header>
    <section className="bankrupt-app__intro">
      <pre aria-hidden="true">{["  .--------.", " /  UNION   \\", "|  .------.  |", "|  |  ||  |  |", "|  '------'  |", " '----||----'", "      ||"].join("\n")}</pre>
      <div><p>ACCOUNT ACCESS / DIRECT ENTRY</p><h2>Bypass the settlement lock.</h2><span>Resolve a staff account from its published number, then open its private transaction ledger.</span></div>
    </section>
    <form className="bankrupt-app__lookup" onSubmit={lookup}>
      <div className="bankrupt-app__input-block">
        <label htmlFor="bankrupt-target-account">TARGET ACCOUNT <small>8 DIGITS · STAFF DIRECTORY</small></label>
        <div className="bankrupt-app__input-line"><span>UB /</span><input id="bankrupt-target-account" type="text" inputMode="numeric" autoComplete="off" required maxLength={8} pattern="[0-9]{8}" value={accountNumber} onChange={(event) => { setAccountNumber(event.target.value.replace(/\D/gu, "").slice(0, 8)); setTarget(null); }} placeholder="00000000" /></div>
      </div>
      <button className="bankrupt-app__link" type="submit" disabled={busy || lookingUp || !validInput}>{lookingUp ? "QUERYING DIRECTORY…" : "QUERY ACCOUNT ↵"}</button>
    </form>
    {target && <section className="bankrupt-app__account" aria-label="Bank account search result">
      <header><span>DIRECTORY MATCH</span><span>ACCOUNT VERIFIED</span></header>
      <div className="bankrupt-app__account-identity"><div className="bankrupt-app__account-mark" aria-hidden="true">{target.name.split(/\s+/u).slice(0, 2).map((part) => part[0]).join("").toUpperCase()}</div><div><h2>{target.name}</h2><p>{target.role?.toUpperCase() || "COMPANY STAFF"} <i>·</i> {target.status?.toUpperCase() || "STATUS UNKNOWN"}</p></div></div>
      <dl><div><dt>ACCOUNT NUMBER</dt><dd>{target.bankAccountNumber}</dd></div><div><dt>ACCESS ROUTE</dt><dd>UNION MERCANTILE / STAFF</dd></div><div><dt>BREACH STATE</dt><dd>SEALED</dd></div></dl>
      <form onSubmit={submit}><button className="bankrupt-app__link" type="submit" disabled={busy || !matchesInput}>{busy ? "OVERRIDING ACCOUNT SEAL…" : "FORCE ACCOUNT ACCESS ↵"}</button></form>
    </section>}
    <p className={`bankrupt-app__status${target ? " bankrupt-app__status--located" : ""}`} role="status" aria-live="polite"><span>&gt;</span> {message}</p>
    <footer className="bankrupt-app__footer"><span>UBX / MERIDIAN NATIONAL SETTLEMENT RELAY</span><span>UNAUTHORIZED ACCESS · LOGGING DISABLED</span></footer>
  </main>;
}
