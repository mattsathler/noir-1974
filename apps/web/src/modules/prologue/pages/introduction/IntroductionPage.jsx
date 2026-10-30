import { useCallback, useEffect, useRef, useState } from "react";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";
import { currentIdentity } from "../../../identity/api.js";
import BootSequence from "./BootSequence.jsx";

export default function IntroductionPage({ onComplete }) {
  const [booting, setBooting] = useState(false);
  const beginBoot = useCallback(() => setBooting(true), []);

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-deep text-primary" aria-label="Introduction">
      <section className="pos-relative w-full max-w-2xl text-center" aria-live="polite" aria-atomic="true">
        <h1 className="text-display tracking-widest leading-tight shadow-glow m-0">
          <Typewriter text="Interpol HQ" speed={105} startDelay={1500} />
        </h1>
        <p className="min-h-line mt-4 text-sm tracking-widest uppercase m-0">
          <Typewriter text="Lyon, France" speed={105} startDelay={3000} onComplete={beginBoot} />
        </p>
        {booting && <BootSequence onComplete={onComplete} />}
      </section>
    </main>
  );
}

export function LocationIntroductionPage({ world: initialWorld = null, onComplete }) {
  const [world, setWorld] = useState(initialWorld);
  const [error, setError] = useState("");
  const completed = useRef(false);

  useEffect(() => {
    if (world) return undefined;
    let active = true;
    currentIdentity()
      .then((identity) => {
        if (!identity.world?.serverTime) throw new Error("Current location is unavailable.");
        if (active) setWorld(identity.world);
      })
      .catch((requestError) => {
        if (active) setError(requestError.message);
      });
    return () => { active = false; };
  }, [world]);

  useEffect(() => {
    if (world?.serverTime && !completed.current) {
      completed.current = true;
      onComplete?.(world);
    }
  }, [onComplete, world]);

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-deep text-primary" aria-label="Introduction to current location" aria-live="polite" aria-atomic="true">
      <section className="pos-relative w-full max-w-2xl text-center" aria-busy={!world && !error}>
        {!world && <p className="text-xs tracking-widest uppercase text-success m-0">CONNECTING TO CURRENT LOCATION…</p>}
        {error && <p className="text-sm text-danger mt-4" role="alert">{error}</p>}
      </section>
    </main>
  );
}
