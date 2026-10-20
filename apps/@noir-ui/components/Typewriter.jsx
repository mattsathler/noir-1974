import { useEffect, useRef, useState } from "react";

export default function Typewriter({ text, speed = 65, startDelay = 0, onComplete, className = "" }) {
  const [visibleLength, setVisibleLength] = useState(0);
  const onCompleteRef = useRef(onComplete);

  useEffect(() => {
    onCompleteRef.current = onComplete;
  }, [onComplete]);

  useEffect(() => {
    setVisibleLength(0);

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setVisibleLength(text.length);
      onCompleteRef.current?.();
      return undefined;
    }

    let characterIndex = 0;
    let timer;

    const typeNextCharacter = () => {
      characterIndex += 1;
      setVisibleLength(characterIndex);

      if (characterIndex === text.length) {
        onCompleteRef.current?.();
        return;
      }

      timer = window.setTimeout(typeNextCharacter, speed);
    };

    timer = window.setTimeout(typeNextCharacter, startDelay);
    return () => window.clearTimeout(timer);
  }, [text, speed, startDelay]);

  const complete = visibleLength >= text.length;

  return (
    <span
      className={`typewriter ${className}`.trim()}
      data-complete={complete}
      role="text"
      aria-label={text}
    >
      {text.slice(0, visibleLength)}
    </span>
  );
}
