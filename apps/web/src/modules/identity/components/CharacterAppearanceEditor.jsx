import { useEffect, useRef, useState } from "react";
import { getCharacterVisuals, updateCharacterAppearance } from "../api.js";
import ModularFace from "./ModularFace.jsx";

function labelFor(id) {
  return id.replaceAll("_", " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export default function CharacterAppearanceEditor({ character, onClose, onSaved }) {
  const [options, setOptions] = useState(null);
  const [visual, setVisual] = useState(character.visual);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const dialogRef = useRef(null);
  const closeRef = useRef(onClose);
  const savingRef = useRef(saving);
  closeRef.current = onClose;
  savingRef.current = saving;

  useEffect(() => {
    let active = true;
    getCharacterVisuals().then((result) => { if (active) setOptions(result); })
      .catch((reason) => { if (active) setError(reason.message); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    const dialog = dialogRef.current;
    const previousFocus = document.activeElement;
    dialog?.querySelector("select, button")?.focus();
    function handleKeys(event) {
      if (event.key === "Escape" && !savingRef.current) closeRef.current();
      if (event.key !== "Tab") return;
      const controls = [...dialog.querySelectorAll('button:not(:disabled), select:not(:disabled)')];
      if (!controls.length) return;
      const first = controls[0];
      const last = controls[controls.length - 1];
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
    }
    dialog?.addEventListener("keydown", handleKeys);
    return () => {
      dialog?.removeEventListener("keydown", handleKeys);
      if (previousFocus instanceof HTMLElement) previousFocus.focus();
    };
  }, []);

  function setPart(key, value) { setVisual((current) => ({ ...current, [key]: value })); }
  async function save(event) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const result = await updateCharacterAppearance(visual);
      onSaved(result.character);
    } catch (reason) {
      setError(reason.message);
      setSaving(false);
    }
  }

  const select = (label, key, ids, allowNone = false) => <label className="appearance-editor__field" key={key}>
    <span>{label}</span>
    <select value={visual[key] ?? ""} onChange={(event) => setPart(key, event.target.value)}>
      {allowNone && <option value="">None</option>}
      {ids.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}
    </select>
  </label>;

  return <div className="appearance-editor" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !saving) onClose(); }}>
    <section ref={dialogRef} className="appearance-editor__panel" role="dialog" aria-modal="true" aria-labelledby="appearance-editor-title" tabIndex={-1}>
      <header><p>PERSONNEL FILE · AMENDMENT</p><h2 id="appearance-editor-title">Change appearance</h2></header>
      <form onSubmit={save}>
        <div className="appearance-editor__content">
          <div className="appearance-editor__preview" aria-live="polite"><ModularFace visual={visual} name={character.displayName} /></div>
          <fieldset disabled={!options || saving}>
            <legend>Appearance</legend>
            {options ? <div className="appearance-editor__fields">
              {select("Hair", "hairId", options.hair)}
              {select("Skin", "skinId", options.skin)}
              {select("Eyes", "eyesId", options.eyes)}
              {select("Beard", "beardId", options.beard, true)}
              {select("Face detail", "faceId", options.face, true)}
              <label className="appearance-editor__field"><span>Accessories</span><select value={visual.accessoryIds?.[0] ?? ""} onChange={(event) => setPart("accessoryIds", event.target.value ? [event.target.value] : [])}><option value="">None</option>{options.accessories.map((id) => <option value={id} key={id}>{labelFor(id)}</option>)}</select></label>
            </div> : <p role="status">Loading appearance catalog…</p>}
          </fieldset>
        </div>
        {error && <p className="appearance-editor__error" role="alert">{error}</p>}
        <footer><button type="button" disabled={saving} onClick={onClose}>Cancel</button><button type="submit" disabled={!options || saving}>{saving ? "Saving…" : "Save appearance"}</button></footer>
      </form>
    </section>
  </div>;
}
