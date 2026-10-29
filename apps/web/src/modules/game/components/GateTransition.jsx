import { useEffect, useRef, useState } from "react";
import Typewriter from "../../../../../@noir-ui/components/Typewriter.jsx";

export default function GateTransition({ destination, origin, onArrive, onComplete, onError }) {
  const [phase, setPhase] = useState("entering");
  const [bootReady, setBootReady] = useState(false);
  const [arrivalConfirmed, setArrivalConfirmed] = useState(false);
  const handled = useRef(false);
  const arrivalPromise = useRef(null);

  useEffect(() => {
    let active = true;
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const fadeTimer = window.setTimeout(() => setPhase("booting"), 420);
    const bootTimer = window.setTimeout(() => setBootReady(true), 420 + (reducedMotion ? 700 : 2800));
    let errorTimer;
    let exitTimer;

    if (!arrivalPromise.current) {
      arrivalPromise.current = Promise.resolve().then(() => onArrive(destination));
    }
    arrivalPromise.current
      .then(() => {
        if (active) setArrivalConfirmed(true);
      })
      .catch((error) => {
        if (!active) return;
        onError(error);
        setPhase("error");
        errorTimer = window.setTimeout(() => {
          if (!active) return;
          setPhase("leaving");
          exitTimer = window.setTimeout(onComplete, 520);
        }, 1000);
      });

    return () => {
      active = false;
      window.clearTimeout(fadeTimer);
      window.clearTimeout(bootTimer);
      window.clearTimeout(errorTimer);
      window.clearTimeout(exitTimer);
    };
  }, []);

  useEffect(() => {
    if (phase === "booting" && bootReady && arrivalConfirmed) setPhase("typing");
  }, [phase, bootReady, arrivalConfirmed]);

  function finishTypewriter() {
    if (handled.current) return;
    handled.current = true;
    setPhase("leaving");
    window.setTimeout(onComplete, 520);
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
