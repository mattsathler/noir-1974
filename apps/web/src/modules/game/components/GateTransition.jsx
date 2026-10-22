import { useEffect, useRef, useState } from "react";
import Typewriter from "../../../../../@noir-ui/components/Typewriter.jsx";

const wait = (duration) => new Promise((resolve) => window.setTimeout(resolve, duration));

export default function GateTransition({ destination, origin, onArrive, onComplete, onError }) {
  const [phase, setPhase] = useState("entering");
  const handled = useRef(false);
  const arrivalPromise = useRef(null);

  useEffect(() => {
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (!arrivalPromise.current) {
      arrivalPromise.current = Promise.resolve()
        .then(() => onArrive(destination))
        .then((value) => ({ value }), (error) => ({ error }));
    }
    const fadeTimer = window.setTimeout(() => setPhase("booting"), 420);
    const bootTimer = window.setTimeout(() => setPhase("typing"), 420 + (reducedMotion ? 700 : 2800));
    return () => {
      window.clearTimeout(fadeTimer);
      window.clearTimeout(bootTimer);
    };
  }, []);

  async function finishTypewriter() {
    if (handled.current) return;
    handled.current = true;
    try {
      const arrival = await arrivalPromise.current;
      if (arrival?.error) throw arrival.error;
    } catch (error) {
      onError(error);
      setPhase("error");
      await wait(1000);
      setPhase("leaving");
      await wait(520);
      onComplete();
      return;
    }
    setPhase("leaving");
    await wait(520);
    onComplete();
  }

  return (
    <div className={`screen-transition screen-transition--${phase}`} role="status" aria-live="polite" aria-atomic="true">
      {phase === "booting" ? (
        <div className="screen-transition__content">
          <p className="screen-transition__eyebrow">MERIDIAN GLOBAL · GATE TRANSIT RELAY</p>
          <div className="screen-transition__readout" aria-label="GATE transit status">
          <span>{(origin?.city ?? "ORIGIN").toUpperCase()} RELAY: ONLINE</span>
            <span>ROUTE LOCKED: {destination.city.name.toUpperCase()}, {destination.district.name.toUpperCase()}</span>
            <span>TRANSIT FIELD: CALIBRATING</span>
          </div>
          <div className="computer-sequence__progress" role="progressbar" aria-label="GATE transit initialization"><i /></div>
          <p className="screen-transition__notice">PLEASE REMAIN WITHIN THE FIELD</p>
        </div>
      ) : phase === "typing" ? (
        <div className="screen-transition__content">
          <p className="screen-transition__message"><Typewriter className="screen-transition__text" text={`ARRIVING IN ${destination.district.name.toUpperCase()}, ${destination.city.name.toUpperCase()}`} speed={48} onComplete={finishTypewriter} /></p>
        </div>
      ) : (
        <div className="screen-transition__content">
          <p className="screen-transition__message"><span className="screen-transition__text">{phase === "error" ? "GATE CONNECTION FAILED" : phase === "leaving" ? `ARRIVING IN ${destination.district.name.toUpperCase()}, ${destination.city.name.toUpperCase()}` : ""}</span></p>
        </div>
      )}
    </div>
  );
}
