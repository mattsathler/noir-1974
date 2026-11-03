import { useEffect, useRef, useState } from "react";
import { createCharacter, getCharacterVisuals } from "../../api.js";
import ModularFace from "../../components/ModularFace.jsx";
import ComputerAccessShell from "../../components/ComputerAccessShell.jsx";

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
    <ComputerAccessShell title="Profile setup" labelledBy="character-title" variant="login">
        <section className="w-full max-w-xl m-auto p-2 text-primary" aria-labelledby="character-title">
          <div className="computer-access-shell__app-icon" aria-hidden="true"><strong className="computer-sequence__brand"><span>M</span><span>OS</span></strong></div>
          <h1 id="character-title" className="text-2xl m-0 mb-5 text-center">Set up your profile</h1>
          <form className="grid gap-3" onSubmit={handleSubmit} noValidate>
            <label className="text-sm text-secondary mt-2" htmlFor="character-name">Display name</label>
            <input className="computer-access-shell__input" id="character-name" name="displayName" autoComplete="off" minLength={2} maxLength={40} required />

            <div className="character-portrait-preview grid place-items-center py-3" aria-live="polite">
              <ModularFace visual={visual} name="Profile preview" />
            </div>

            <fieldset className="grid gap-3 border-0 p-0 mt-2" disabled={!visualOptions || submitting}>
              <legend className="text-sm text-secondary mb-2">Appearance</legend>
              {visualOptions ? <>
                <label className="grid gap-1 text-xs uppercase text-secondary">Hair
                  <select className="computer-access-shell__input" value={visual.hairId} onChange={(event) => setVisualPart("hairId", event.target.value)}>
                    {visualOptions.hair.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Skin
                  <select className="computer-access-shell__input" value={visual.skinId} onChange={(event) => setVisualPart("skinId", event.target.value)}>
                    {visualOptions.skin.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Eyes
                  <select className="computer-access-shell__input" value={visual.eyesId} onChange={(event) => setVisualPart("eyesId", event.target.value)}>
                    {visualOptions.eyes.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Beard
                  <select className="computer-access-shell__input" value={visual.beardId} onChange={(event) => setVisualPart("beardId", event.target.value)}>
                    <option value="">None</option>
                    {visualOptions.beard.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Face
                  <select className="computer-access-shell__input" value={visual.faceId} onChange={(event) => setVisualPart("faceId", event.target.value)}>
                    <option value="">None</option>
                    {visualOptions.face.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
                <label className="grid gap-1 text-xs uppercase text-secondary">Accessories
                  <select className="computer-access-shell__input" value={visual.accessoryIds[0] ?? ""} onChange={(event) => setVisualPart("accessoryIds", event.target.value ? [event.target.value] : [])}>
                    <option value="">None</option>
                    {visualOptions.accessories.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
                  </select>
                </label>
              </> : <p className="text-xs text-secondary" role="status">Loading the appearance catalog…</p>}
            </fieldset>
            {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
            <button className="min-h-11 mt-4 border-0 bg-transparent text-accent tracking-wide uppercase transition" type="submit" disabled={submitting || !visualOptions}>{submitting ? "Saving…" : "Save profile"}</button>
          </form>
        </section>
    </ComputerAccessShell>
  );
}
