import { useEffect, useRef, useState } from "react";

export default function Typewriter({ text, texts, speed = 65, startDelay = 0, lineDelay = 180, onComplete, className = "" }) {
  const sequence = texts ?? (Array.isArray(text) ? text : [text ?? ""]);
  const sequenceKey = JSON.stringify(sequence);
  const multiple = sequence.length > 1 || texts !== undefined || Array.isArray(text);
  const [lineIndex, setLineIndex] = useState(0);
  const [visibleLength, setVisibleLength] = useState(0);
  const onCompleteRef = useRef(onComplete);

  useEffect(() => {
    onCompleteRef.current = onComplete;
  }, [onComplete]);

  useEffect(() => {
    setLineIndex(0);
    setVisibleLength(0);

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setLineIndex(Math.max(0, sequence.length - 1));
      setVisibleLength(sequence.at(-1)?.length ?? 0);
      onCompleteRef.current?.();
      return undefined;
    }

    if (!sequence.length) {
      onCompleteRef.current?.();
      return undefined;
    }

    let activeLine = 0;
    let characterIndex = 0;
    let timer;

    const advanceLine = () => {
      if (activeLine === sequence.length - 1) {
        onCompleteRef.current?.();
        return;
      }
      timer = window.setTimeout(() => {
        activeLine += 1;
        characterIndex = 0;
        setLineIndex(activeLine);
        setVisibleLength(0);
        timer = window.setTimeout(typeNextCharacter, speed);
      }, lineDelay);
    };

    const typeNextCharacter = () => {
      const currentText = sequence[activeLine] ?? "";
      if (characterIndex >= currentText.length) {
        advanceLine();
        return;
      }

      characterIndex += 1;
      setVisibleLength(characterIndex);
      if (characterIndex === currentText.length) advanceLine();
      else timer = window.setTimeout(typeNextCharacter, speed);
    };

    timer = window.setTimeout(typeNextCharacter, startDelay);
    return () => window.clearTimeout(timer);
  }, [sequenceKey, speed, startDelay, lineDelay]);

  const complete = lineIndex >= sequence.length - 1 && visibleLength >= (sequence.at(-1)?.length ?? 0);
  const label = sequence.join("\n");

  return (
    <span
      className={`typewriter ${className}`.trim()}
      data-complete={complete}
      data-multiple={multiple || undefined}
      role="text"
      aria-label={label}
    >
      {multiple ? sequence.slice(0, lineIndex + 1).map((line, index) => (
        <span className="typewriter__line" data-last-line={index === sequence.length - 1 || undefined} data-complete={index < lineIndex || (index === lineIndex && visibleLength >= line.length)} key={`${index}-${line}`}>
          {index < lineIndex ? line : line.slice(0, visibleLength)}
        </span>
      )) : sequence[0]?.slice(0, visibleLength)}
    </span>
  );
}
