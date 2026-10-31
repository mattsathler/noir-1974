import { useEffect, useRef, useState } from "react";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import Typewriter from "../../../../../../@noir-ui/components/Typewriter.jsx";

const CHARACTERS = {
  bill: {
    name: "Bill Bouer",
    role: "STATE INVESTIGATOR · CITY POLICE",
    visual: { hairId: "hair_slicked_grey", skinId: "skin_light", eyesId: "eyes_gray", beardId: "beard_full_grey", accessoryIds: [] },
    lines: [
      { text: "The Crow calls every breach an act of resistance. The people whose records he steals never get to choose whether they become part of his cause.", replies: ["You sound certain the law protects them.", "What are you doing to catch him?"] },
      { text: "The law is the only barrier between a powerful institution and whoever it decides to watch. I know how often that barrier is bent. I still have to work inside it.", replies: ["And Ghostkey? The state calls that a crime too.", "Then tell me what the investigation has found."] },
      { text: "Ghostkey opens doors without permission. That can expose a neighbor as easily as a ministry. I want The Crow found before another person pays for his war.", replies: ["I will keep my own eyes open.", "I need to decide for myself who is right."] },
    ],
  },
  crow: {
    name: "The Crow",
    role: "WANTED · SIGNAL SOURCE UNVERIFIED",
    visual: { hairId: "hair_messy_black", skinId: "skin_brown", eyesId: "eyes_green", faceId: "face_wrinkles_fine", accessoryIds: [] },
    lines: [
      { text: "Bouer says I steal from ordinary people. The companies keep their ledgers behind locked doors, then call any hand that opens them a criminal's hand. They scrub my skull mark from every terminal they can reach.", replies: ["You still steal from those companies.", "Bouer says your breaches put residents at risk."] },
      { text: "I do. I take what they hoard and publish what they bury. Don't mistake that for innocence. A breach can hurt people. So can a quiet order signed in a clean office. The skull is a warning that someone got inside; it does not tell you who deserved it.", replies: ["Why leave the choice to you?", "What do you know about Ghostkey?"] },
      { text: "You should not leave it to me. Or Bouer. Ghostkey is a key, not a conscience. The state forbids it because it can open their files, and because they want you to believe those files belong to them.", replies: ["I will decide what to do with what I learn.", "Bouer is still hunting you."] },
    ],
  },
};

export function requestFixedNPCContact(characterId) {
  if (!(characterId in CHARACTERS)) return;
  window.dispatchEvent(new CustomEvent("noir:fixed-npc-contact", { detail: { characterId } }));
}

export default function CharacterConversation({ characterId, onClose }) {
  const character = CHARACTERS[characterId];
  const [lineIndex, setLineIndex] = useState(0);
  const dialogRef = useRef(null);
  const line = character.lines[lineIndex];

  useEffect(() => {
    const previousFocus = document.activeElement;
    dialogRef.current?.querySelector("button")?.focus();
    function onKeyDown(event) {
      if (event.key === "Escape") onClose();
      if (event.key !== "Tab" || !dialogRef.current) return;
      const buttons = [...dialogRef.current.querySelectorAll("button:not(:disabled)")];
      const first = buttons[0];
      const last = buttons.at(-1);
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
    }
    document.addEventListener("keydown", onKeyDown);
    return () => { document.removeEventListener("keydown", onKeyDown); previousFocus?.focus?.(); };
  }, [onClose]);

  return <div className="character-conversation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
    <section className="character-conversation__dialog" ref={dialogRef} role="dialog" aria-modal="true" aria-labelledby="character-conversation-name">
      <header className="character-conversation__header"><span>SECURE LINE · SOURCE UNVERIFIED</span><button type="button" onClick={onClose} aria-label="End conversation">×</button></header>
      <div className={`character-conversation__portrait${characterId === "crow" ? " character-conversation__portrait--crow" : ""}`}>
        <ModularFace visual={character.visual} name={character.name} />
        {characterId === "crow" && <svg className="character-conversation__skull-mark" viewBox="0 0 16 16" role="img" aria-label="The Crow's skull mark">
          <path fill="currentColor" d="M4 1h8v1h2v2h1v6h-2v2h-2v2H5v-2H3v-2H1V4h1V2h2z" />
          <path fill="#111511" d="M4 5h3v3H4zm5 0h3v3H9zm-2 4h2v2H7zM5 12h2v2H5zm4 0h2v2H9z" />
        </svg>}
      </div>
      <h1 id="character-conversation-name">{character.name}</h1>
      <p className="character-conversation__role">{character.role}</p>
      <div className="character-conversation__speech" aria-live="polite" aria-atomic="true"><span className="sr-only">{line.text}</span><span aria-hidden="true"><Typewriter key={`${characterId}-${lineIndex}`} text={line.text} speed={28} /></span></div>
      <div className="character-conversation__replies" aria-label="Your responses">
        {line.replies.map((reply) => <button key={reply} type="button" onClick={() => setLineIndex((index) => index < character.lines.length - 1 ? index + 1 : 0)}>{reply}</button>)}
      </div>
      <footer>TRANSMISSION {String(lineIndex + 1).padStart(2, "0")} / {String(character.lines.length).padStart(2, "0")}</footer>
    </section>
  </div>;
}
