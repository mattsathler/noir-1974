import { useState } from "react";
import { currentIdentity, loginAccount } from "../../api.js";

export default function LoginPage({ onRegister, onAuthenticated }) {
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    const fields = new FormData(event.currentTarget);
    const email = String(fields.get("email")).trim();
    const password = String(fields.get("password"));
    const emailInput = event.currentTarget.elements.namedItem("email");
    if (!emailInput.checkValidity() || new TextEncoder().encode(email).length > 254) {
      setError("Enter a valid email address.");
      return;
    }
    if (!password || new TextEncoder().encode(password).length > 72) {
      setError("Enter your password.");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      const loginIdentity = await loginAccount({ email, password });
      const identity = loginIdentity.character || loginIdentity.world?.serverTime
        ? loginIdentity
        : await currentIdentity();
      onAuthenticated(identity);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-deep text-primary">
      <div className="pd-backdrop pos-absolute inset-0" aria-hidden="true" />
      <div className="texture-scanlines pos-absolute inset-0 pointer-events-none opacity-10" aria-hidden="true" />
      <div className="pos-relative w-full max-w-sm">
        <img className="brand-mark" src="/images/logo.png" alt="NOIR 1974" />
        <section className="w-full p-6 surface shadow-panel" aria-labelledby="login-title">
          <p className="text-xs tracking-wide uppercase text-success m-0 mb-3">INTERPOL · LYON</p>
          <h1 id="login-title" className="text-2xl m-0 mb-4">Access the files</h1>
          <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
            <label className="text-sm text-secondary mt-2" htmlFor="agent-email">Email</label>
            <input className="w-full min-h-11 p-3 border bg-base text-primary focus-ring" id="agent-email" name="email" type="email" autoComplete="email" maxLength={254} aria-describedby="login-email-hint" required />
            <label className="text-sm text-secondary mt-2" htmlFor="agent-password">Password</label>
            <input className="w-full min-h-11 p-3 border bg-base text-primary focus-ring" id="agent-password" name="password" type="password" autoComplete="current-password" aria-describedby="login-password-hint" required />
            {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
            <button className="min-h-11 mt-4 justify-self-end border-0 bg-transparent text-accent tracking-wide uppercase transition" type="submit" disabled={submitting}>{submitting ? "Verifying…" : "Sign in"}</button>
          </form>
          <p className="text-sm text-secondary text-center mt-6 mb-0">
            Need access?{" "}
            <button className="border-0 bg-transparent text-accent underline" type="button" onClick={onRegister}>
              Create account
            </button>
          </p>
        </section>
      </div>
    </main>
  );
}
