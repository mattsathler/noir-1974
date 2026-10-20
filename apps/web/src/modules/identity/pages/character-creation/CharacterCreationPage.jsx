import { useState } from "react";
import { createCharacter } from "../../api.js";

const portraits = [1, 2, 3, 4, 5, 6];

export default function CharacterCreationPage({ onComplete }) {
  const [portraitId, setPortraitId] = useState(1);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    const fields = new FormData(event.currentTarget);
    const displayName = String(fields.get("displayName")).trim();
    const displayNameLength = Array.from(displayName).length;
    if (displayNameLength < 2 || displayNameLength > 40) {
      setError("Detective name must be 2–40 characters long.");
      return;
    }
    setSubmitting(true);
    setError("");
    try {
      await createCharacter({ displayName, portraitId });
      onComplete();
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-atmosphere text-primary">
      <div className="register-backdrop pos-absolute inset-0" aria-hidden="true" />
      <div className="pos-relative w-full max-w-sm">
        <img className="brand-mark" src="/images/logo.png" alt="NOIR 1974" />
        <section className="w-full p-6 surface shadow-panel" aria-labelledby="character-title">
          <p className="text-xs tracking-wide uppercase text-success m-0 mb-3">DOSSIER · INTERPOL</p>
          <h1 id="character-title" className="text-2xl m-0 mb-4">Your identity</h1>
          <p className="text-sm text-secondary mt-0 mb-5">You work independently. Your past with Interpol still opens a few doors.</p>
          <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
            <label className="text-sm text-secondary mt-2" htmlFor="character-name">Detective name</label>
            <input className="w-full min-h-11 p-3 border bg-base text-primary focus-ring" id="character-name" name="displayName" autoComplete="off" minLength={2} maxLength={40} required />
            <p className="text-xs text-secondary m-0">2–40 characters.</p>
            <fieldset className="grid gap-2 border-0 p-0 mt-3" aria-describedby="portrait-note">
              <legend className="text-sm text-secondary mb-2">Portrait</legend>
              <div className="portrait-grid">
                {portraits.map((id) => (
                  <button
                    className={`portrait-slot${portraitId === id ? " portrait-slot--selected" : ""}`}
                    type="button"
                    key={id}
                    aria-pressed={portraitId === id}
                    aria-label={`Select portrait ${id}`}
                    onClick={() => setPortraitId(id)}
                  >
                    <span className="portrait-slot__number">{String(id).padStart(2, "0")}</span>
                    <span className="portrait-slot__caption">In development</span>
                  </button>
                ))}
              </div>
              <span className="text-xs text-secondary" id="portrait-note">Choose 1–6. Final artwork will follow Noir's pixel art direction.</span>
            </fieldset>
            {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
            <button className="min-h-11 mt-4 border-0 bg-accent text-ink tracking-wide uppercase transition hover-bg-brass" type="submit" disabled={submitting}>{submitting ? "Filing…" : "Create detective"}</button>
          </form>
        </section>
      </div>
    </main>
  );
}
