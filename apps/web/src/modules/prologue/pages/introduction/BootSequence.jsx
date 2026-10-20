import { useEffect, useState } from "react";

const BOOT_LINES = [
  "MERIDIAN GLOBAL / SYSTEM 74.04",
  "MEMORY CHECK ............... 640K OK",
  "GATE RELAY: LYON-01 ONLINE",
  "PRIVATE NETWORK: INTERPOL",
  "SYSTEM READY.",
];

const BOOT_END_DELAY = 1600;

export default function BootSequence({ onComplete }) {
  const [visibleLines, setVisibleLines] = useState(0);

  useEffect(() => {
    let lineIndex = 0;
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    let timer;

    if (reducedMotion) {
      setVisibleLines(BOOT_LINES.length);
      timer = window.setTimeout(onComplete, BOOT_END_DELAY);
      return () => window.clearTimeout(timer);
    }

    const revealNextLine = () => {
      lineIndex += 1;
      setVisibleLines(lineIndex);

      if (lineIndex === BOOT_LINES.length) {
        timer = window.setTimeout(onComplete, BOOT_END_DELAY);
        return;
      }

      timer = window.setTimeout(revealNextLine, 520);
    };

    timer = window.setTimeout(revealNextLine, 300);
    return () => window.clearTimeout(timer);
  }, [onComplete]);

  return (
    <section className="boot-console" aria-label="Meridian system startup" aria-live="polite">
      <div className="boot-console__titlebar">
        <span>MERIDIAN GLOBAL</span>
        <span>GATE TERMINAL / 1974</span>
      </div>
      <div className="boot-console__output">
        {BOOT_LINES.slice(0, visibleLines).map((line) => (
          <div className="boot-console__line" key={line}>
            <span className="boot-console__prompt">&gt; </span>{line}
          </div>
        ))}
      </div>
      <div
        className="boot-console__progress"
        role="progressbar"
        aria-label="System startup"
        aria-valuemin="0"
        aria-valuemax={BOOT_LINES.length}
        aria-valuenow={visibleLines}
      >
        <span style={{ width: `${(visibleLines / BOOT_LINES.length) * 100}%` }} />
      </div>
    </section>
  );
}
