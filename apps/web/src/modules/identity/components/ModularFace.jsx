import { useMemo } from "react";
import "./ModularFace.scss";

function spritePathFor(id) {
  const directory = id.startsWith("hair_") ? "hair"
    : id.startsWith("skin_") ? "skin"
      : id.startsWith("eyes_") ? "eyes"
        : id.startsWith("beard_") ? "beard"
          : id.startsWith("accessories_") ? "accessories"
            : "face";
  return `/images/characters/${directory}/${id}.png`;
}

export default function ModularFace({ visual, name = "Character" }) {
  const layers = useMemo(() => {
    if (!visual?.skinId || !visual?.eyesId || !visual?.hairId) return [];
    return [
      visual.skinId,
      visual.hairId,
      visual.eyesId,
      ...(visual.beardId ? [visual.beardId] : []),
      ...(visual.faceId ? [visual.faceId] : []),
      ...(visual.accessoryIds ?? []),
    ];
  }, [visual]);
  if (!layers.length) return null;

  return (
    <div className="modular-face" role="img" aria-label={`Pixel-art portrait of ${name}`}>
      {layers.map((id) => (
        <img
          className="modular-face__layer"
          key={id}
          src={spritePathFor(id)}
          alt=""
          draggable="false"
        />
      ))}
    </div>
  );
}
