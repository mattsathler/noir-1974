import { useEffect, useState } from "react";
import { currentIdentity } from "../../api.js";
import ModularFace from "../../components/ModularFace.jsx";

const DOSSIER_LOAD_DELAY = 1800;

export default function ArchivePage({ initialIdentity = null, onLoaded }) {
  const [identity, setIdentity] = useState(initialIdentity);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (initialIdentity?.character && initialIdentity?.world?.serverTime) {
      setIdentity(initialIdentity);
      setLoading(false);
      return undefined;
    }

    let active = true;
    currentIdentity()
      .then((current) => {
        if (!active) return;
        setIdentity(current);
        if (!current.character) setError("No detective record found for this session.");
      })
      .catch((requestError) => {
        if (active) setError(requestError.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => { active = false; };
  }, [initialIdentity]);

  useEffect(() => {
    if (loading || error || !identity?.character || !identity?.world?.serverTime) return undefined;
    const timer = window.setTimeout(() => onLoaded?.(identity.world), DOSSIER_LOAD_DELAY);
    return () => window.clearTimeout(timer);
  }, [error, identity, loading, onLoaded]);

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-deep text-primary" aria-label="Loading game" aria-busy={loading}>
      <div className="texture-scanlines pos-absolute inset-0 pointer-events-none opacity-10" aria-hidden="true" />
      <section className="world-loader pos-relative w-full max-w-sm p-6 surface shadow-panel" aria-labelledby="dossier-title">
        <p className="text-xs tracking-wide uppercase text-success m-0 mb-3">PRIVATE FILE · LYON</p>
        <h1 id="dossier-title" className="text-2xl m-0">Case file opened</h1>
        {identity?.character && <>
          <div className="character-portrait-preview grid place-items-center py-3"><ModularFace visual={identity.character.visual} name={identity.character.displayName} /></div>
          <p className="world-loader__agent text-sm text-secondary">AGENT · {identity.character.displayName}</p>
        </>}
        <p className="text-xs tracking-wide uppercase text-accent m-0">{loading ? "Reading case files" : "Synchronizing world"}</p>
        <div className={`world-loader__track${!loading && !error ? " world-loader__track--ready" : ""}`} role="progressbar" aria-label="Loading world" aria-valuemin="0" aria-valuemax="100" aria-valuenow={loading ? 35 : error ? 0 : 70}>
          <span />
        </div>
        {error && <p className="text-sm text-danger m-0" role="alert">{error}</p>}
      </section>
    </main>
  );
}
