import { useEffect, useRef, useState } from "react";
import { createCharacter, getCharacterVisuals } from "../../api.js";
import ModularFace from "../../components/ModularFace.jsx";

const defaultVisual = {
  hairId: "hair_short_brown",
  skinId: "skin_white",
  eyesId: "eyes_brown",
  faceId: "",
  beardId: "",
  accessoryIds: [],
};

function labelFor(id) {
  return id.replaceAll("_", " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export default function CharacterCreationPage({ onComplete }) {
  const [visualOptions, setVisualOptions] = useState(null);
  const [visual, setVisual] = useState(defaultVisual);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const submissionLocked = useRef(false);

  useEffect(() => {
    let active = true;
    getCharacterVisuals().then((options) => { if (active) setVisualOptions(options); })
      .catch((requestError) => { if (active) setError(requestError.message); });
    return () => { active = false; };
  }, []);

  async function handleSubmit(event) {
    event.preventDefault();
    if (submissionLocked.current || !visualOptions) return;
    const fields = new FormData(event.currentTarget);
    const displayName = String(fields.get("displayName")).trim();
    const displayNameLength = Array.from(displayName).length;
    if (displayNameLength < 2 || displayNameLength > 40) {
      setError("Detective name must be 2–40 characters long.");
      return;
    }
    submissionLocked.current = true;
    setSubmitting(true);
    setError("");
    try {
      await createCharacter({ displayName, visual });
      onComplete();
    } catch (requestError) {
      submissionLocked.current = false;
      setSubmitting(false);
      setError(requestError.message);
    }
  }

  function setVisualPart(key, value) {
    setVisual((current) => ({ ...current, [key]: value }));
  }

  return (
    <main className="grid min-h-screen place-items-center overflow-x-hidden overflow-y-auto pos-relative p-6 bg-atmosphere text-primary">
      <div className="pd-backdrop pos-absolute inset-0" aria-hidden="true" />
      <div className="pos-relative w-full max-w-sm">
        <span className="brand-mark" role="img" aria-label="NOIR">NOIR</span>
        <section className="w-full p-6 surface shadow-panel" aria-labelledby="character-title">
          <p className="text-xs tracking-wide uppercase text-success m-0 mb-3">DOSSIER · INTERPOL</p>
          <h1 id="character-title" className="text-2xl m-0 mb-4">Your identity</h1>
          <p className="text-sm text-secondary mt-0 mb-5">You work independently. Your past with Interpol still opens a few doors.</p>
          <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
            <label className="text-sm text-secondary mt-2" htmlFor="character-name">Detective name</label>
            <input className="w-full min-h-11 p-3 border bg-base text-primary focus-ring" id="character-name" name="displayName" autoComplete="off" minLength={2} maxLength={40} required />
            <p className="text-xs text-secondary m-0">2–40 characters.</p>

            <div className="character-portrait-preview grid place-items-center py-3" aria-live="polite">
              <ModularFace visual={visual} name="Your detective" />
            </div>

            <fieldset className="grid gap-3 border-0 p-0 mt-2" disabled={!visualOptions || submitting}>
              <legend className="text-sm text-secondary mb-2">Appearance</legend>
              {visualOptions ? <>
                <label className="grid gap-1 text-xs uppercase text-secondary">Hair
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.hairId} onChange={(event) => setVisualPart("hairId", event.target.value)}>
                    {visualOptions.hair.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Skin
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.skinId} onChange={(event) => setVisualPart("skinId", event.target.value)}>
                    {visualOptions.skin.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Eyes
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.eyesId} onChange={(event) => setVisualPart("eyesId", event.target.value)}>
                    {visualOptions.eyes.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Beard
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.beardId} onChange={(event) => setVisualPart("beardId", event.target.value)}>
                    <option value="">None</option>
                    {visualOptions.beard.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Face
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.faceId} onChange={(event) => setVisualPart("faceId", event.target.value)}>
                    <option value="">None</option>
                    {visualOptions.face.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Accessories
                  <select className="w-full min-h-11 p-3 border bg-base text-primary" value={visual.accessoryIds[0] ?? ""} onChange={(event) => setVisualPart("accessoryIds", event.target.value ? [event.target.value] : [])}>
                    <option value="">None</option>
                    {visualOptions.accessories.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
              </> : <p className="text-xs text-secondary" role="status">Loading the appearance catalog…</p>}
            </fieldset>
            {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
            <button className="min-h-11 mt-4 border-0 bg-transparent text-accent tracking-wide uppercase transition" type="submit" disabled={submitting || !visualOptions}>{submitting ? "Filing…" : "Create detective"}</button>
          </form>
        </section>
      </div>
    </main>
  );
}
