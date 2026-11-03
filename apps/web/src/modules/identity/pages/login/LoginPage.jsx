import { useRef, useState } from "react";
import { currentIdentity, loginAccount } from "../../api.js";
import ComputerAccessShell from "../../components/ComputerAccessShell.jsx";
import { unlockRetroAudio } from "../../../game/pages/main/retroSoundEffects.js";

export default function LoginPage({ onRegister, onAuthenticated }) {
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const submissionLocked = useRef(false);

  async function handleSubmit(event) {
    event.preventDefault();
    if (submissionLocked.current) return;
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
    submissionLocked.current = true;
    setSubmitting(true);
    unlockRetroAudio();
    try {
      const loginIdentity = await loginAccount({ email, password });
      const identity = loginIdentity.character || loginIdentity.world?.serverTime
        ? loginIdentity
        : await currentIdentity();
      onAuthenticated(identity);
    } catch (requestError) {
      submissionLocked.current = false;
      setSubmitting(false);
      setError(requestError.message);
    }
  }

  return (
    <ComputerAccessShell title="User sign-in" labelledBy="login-title" variant="login">
      <section className="w-full max-w-sm m-auto text-primary" aria-labelledby="login-title">
        <div className="computer-access-shell__app-icon" aria-hidden="true"><strong className="computer-sequence__brand"><span>M</span><span>OS</span></strong></div>
        <h1 id="login-title" className="text-2xl m-0 mb-5 text-center">Sign in</h1>
        <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
          <label className="text-sm text-secondary mt-2" htmlFor="agent-email">Email</label>
          <input className="computer-access-shell__input" id="agent-email" name="email" type="email" autoComplete="email" maxLength={254} required />
          <label className="text-sm text-secondary mt-2" htmlFor="agent-password">Password</label>
          <input className="computer-access-shell__input" id="agent-password" name="password" type="password" autoComplete="current-password" required />
          {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
          <button className="min-h-11 mt-4 justify-self-end border-0 bg-transparent text-accent tracking-wide uppercase transition" type="submit" disabled={submitting}>{submitting ? "Verifying…" : "Sign in"}</button>
        </form>
        <p className="text-sm text-secondary text-center mt-6 mb-0">
          Need a user account?{" "}
          <button className="border-0 bg-transparent text-accent underline" type="button" onClick={onRegister}>
            Create account
          </button>
        </p>
      </section>
    </ComputerAccessShell>
  );
}
