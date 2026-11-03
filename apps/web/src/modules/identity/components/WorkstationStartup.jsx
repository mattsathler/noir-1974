import { useEffect, useState } from "react";
import { COMPUTER_BOOT_LINES, markBootedThisLogin } from "../../game/pages/main/computerBoot.js";
import { playRetroSound, readRetroSoundEnabled } from "../../game/pages/main/retroSoundEffects.js";

export default function WorkstationStartup({ onComplete }) {
  const [phase, setPhase] = useState("boot");
  const [visibleLines, setVisibleLines] = useState(0);

  useEffect(() => {
    let timer = 0;
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reducedMotion) {
      setVisibleLines(COMPUTER_BOOT_LINES.length);
      timer = window.setTimeout(() => setPhase("intro"), 300);
      return () => window.clearTimeout(timer);
    }

    let lineIndex = 0;
    const revealNextLine = () => {
      lineIndex += 1;
      setVisibleLines(lineIndex);
      if (lineIndex < COMPUTER_BOOT_LINES.length) timer = window.setTimeout(revealNextLine, 95);
      else timer = window.setTimeout(() => setPhase("intro"), 450);
    };
    timer = window.setTimeout(revealNextLine, 150);
    return () => window.clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (phase !== "intro") return undefined;
    playRetroSound("boot", readRetroSoundEnabled());
    const timer = window.setTimeout(() => {
      markBootedThisLogin();
      onComplete();
    }, 1800);
    return () => window.clearTimeout(timer);
  }, [onComplete, phase]);

  return (
    <main className="computer-startup" aria-label="Computer startup" aria-live="polite">
      <section className={`computer-startup__screen${phase === "intro" ? " computer-startup__screen--intro" : ""}`}>
        {phase === "boot"
          ? <div className="computer-sequence computer-sequence--boot computer-startup__bios" role="status">
              <div className="computer-sequence__bios">{COMPUTER_BOOT_LINES.slice(0, visibleLines).map((line, index) => <span key={`${line}-${index}`}>{line}</span>)}</div>
            </div>
          : <div className="computer-startup__brand" role="img" aria-label="M_OS">
              <strong className="computer-sequence__brand"><span>M</span><span>OS</span></strong>
              <p>PERSONAL COMPUTER</p>
            </div>}
      </section>
    </main>
  );
}
