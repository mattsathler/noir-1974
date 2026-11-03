import { useRef, useState } from "react";
import { registerAccount } from "../../api.js";
import ComputerAccessShell from "../../components/ComputerAccessShell.jsx";

export default function RegisterPage({ onLogin, onAccountCreated }) {
  const [passwordMismatch, setPasswordMismatch] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const submissionLocked = useRef(false);

  async function handleSubmit(event) {
    event.preventDefault();
    if (submissionLocked.current) return;
    const fields = new FormData(event.currentTarget);
    setPasswordMismatch(false);
    const username = String(fields.get("username")).trim();
    const email = String(fields.get("email")).trim();
    const password = String(fields.get("password"));
    const confirmation = String(fields.get("confirmPassword"));

    if (!/^[a-zA-Z0-9_-]{3,24}$/.test(username)) {
      setError("Username must be 3–24 characters: letters, numbers, _ or -.");
      return;
    }
    const emailInput = event.currentTarget.elements.namedItem("email");
    if (!emailInput.checkValidity() || new TextEncoder().encode(email).length > 254) {
      setError("Enter a valid email address.");
      return;
    }
    const passwordBytes = new TextEncoder().encode(password).length;
    if (passwordBytes < 12 || passwordBytes > 72) {
      setError("Password must be 12–72 bytes in UTF-8.");
      return;
    }
    if (password !== confirmation) {
      setPasswordMismatch(true);
      setError("");
      return;
    }
    setError("");
    submissionLocked.current = true;
    setSubmitting(true);
    try {
      const account = await registerAccount({ username, email, password });
      onAccountCreated(account);
    } catch (requestError) {
      submissionLocked.current = false;
      setSubmitting(false);
      setError(requestError.message);
    }
  }

  return (
    <ComputerAccessShell title="Create account" labelledBy="register-title" variant="login">
      <section className="w-full max-w-sm m-auto text-primary" aria-labelledby="register-title">
          <div className="computer-access-shell__app-icon" aria-hidden="true"><strong className="computer-sequence__brand"><span>M</span><span>OS</span></strong></div>
          <h1 id="register-title" className="text-2xl m-0 mb-5 text-center">Create account</h1>
          <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
            <label className="text-sm text-secondary mt-2" htmlFor="register-username">Username</label>
            <input className="computer-access-shell__input" id="register-username" name="username" type="text" autoComplete="username" minLength={3} maxLength={24} pattern="[A-Za-z0-9_-]{3,24}" aria-describedby="register-username-hint" required />
            <label className="text-sm text-secondary mt-2" htmlFor="register-email">Email</label>
            <input className="computer-access-shell__input" id="register-email" name="email" type="email" autoComplete="email" maxLength={254} aria-describedby="register-email-hint" required />
            <label className="text-sm text-secondary mt-2" htmlFor="register-password">Password</label>
            <input className="computer-access-shell__input" id="register-password" name="password" type="password" autoComplete="new-password" aria-describedby="register-password-hint" required />
            <label className="text-sm text-secondary mt-2" htmlFor="register-confirm-password">Confirm password</label>
            <input
              className="computer-access-shell__input"
              id="register-confirm-password"
              name="confirmPassword"
              type="password"
              autoComplete="new-password"
              aria-invalid={passwordMismatch}
              aria-describedby={`register-confirm-password-hint${passwordMismatch ? " register-password-error" : ""}`}
              required
            />
            {passwordMismatch && <p className="text-sm text-danger m-0" id="register-password-error" role="alert">Passwords do not match.</p>}
            {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
            <button className="min-h-11 mt-4 justify-self-end border-0 bg-transparent text-accent tracking-wide uppercase transition" type="submit" disabled={submitting}>{submitting ? "Registering…" : "Create account"}</button>
          </form>
          <p className="text-sm text-secondary text-center mt-6 mb-0">
            Already have access?{" "}
            <button className="border-0 bg-transparent text-accent underline" type="button" onClick={onLogin}>
              Sign in
            </button>
          </p>
      </section>
    </ComputerAccessShell>
  );
}
