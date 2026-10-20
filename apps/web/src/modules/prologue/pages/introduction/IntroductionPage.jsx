import { useCallback, useEffect, useRef, useState } from "react";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";
import { currentIdentity } from "../../../identity/api.js";
import BootSequence from "./BootSequence.jsx";

export default function IntroductionPage({ onComplete }) {
  const [booting, setBooting] = useState(false);
  const beginBoot = useCallback(() => setBooting(true), []);

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-atmosphere text-primary" aria-label="Introduction">
      <div className="lyon-backdrop pos-absolute inset-0" aria-hidden="true" />
      <div className="texture-scanlines pos-absolute inset-0 pointer-events-none opacity-10" aria-hidden="true" />
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
  const [showTime, setShowTime] = useState(false);
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

  function completeArrival() {
    if (completed.current) return;
    completed.current = true;
    onComplete?.(world);
  }

  const timeZone = world?.timeZone ?? "Europe/Paris";
  const serverTime = world ? new Date(world.serverTime) : null;
  const location = world ? `${world.city}, ${world.country}` : "";
  const formattedTime = serverTime && !Number.isNaN(serverTime.getTime())
    ? new Intl.DateTimeFormat("en-US", { timeZone, hour: "numeric", minute: "2-digit", hour12: true }).format(serverTime)
    : "";

  return (
    <main className="grid min-h-screen place-items-center overflow-hidden pos-relative p-6 bg-atmosphere text-primary" aria-label="Introduction to current location" aria-live="polite" aria-atomic="true">
      <div className="game-backdrop pos-absolute inset-0" data-location={world?.locationId ?? "lyon"} aria-hidden="true" />
      <div className="game-vignette pos-absolute inset-0" aria-hidden="true" />
      <section className="pos-relative w-full max-w-2xl text-center">
        <p className="text-xs tracking-widest uppercase text-success m-0">CURRENT LOCATION</p>
        <h1 className="text-display tracking-widest leading-tight shadow-glow m-0 mt-4">
          {location ? <Typewriter text={location} speed={105} onComplete={() => setShowTime(true)} /> : "Locating player"}
        </h1>
        {showTime && formattedTime && (
          <p className="min-h-line mt-4 text-sm tracking-widest uppercase m-0">
            LOCAL TIME · <Typewriter text={formattedTime} speed={90} onComplete={completeArrival} />
          </p>
        )}
        {error && <p className="text-sm text-danger mt-4" role="alert">{error}</p>}
      </section>
    </main>
  );
}
