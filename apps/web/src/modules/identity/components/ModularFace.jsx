import { useMemo } from "react";
import "./ModularFace.scss";

function spritePathFor(id) {
  const directory = id.startsWith("hair_") ? "hair"
    : id.startsWith("skin_") ? "skin"
      : id.startsWith("eyes_") || id.startsWith("face_eyes_") ? "eyes"
        : id.startsWith("accessory_") || id.startsWith("face_glasses_") ? "accessories"
          : "face";
  return `/images/characters/${directory}/${id}.png`;
}

export default function ModularFace({ visual, name = "Character" }) {
  const layers = useMemo(() => {
    if (!visual?.skinColorId || !visual?.eyeColorId || !visual?.hairId) return [];
    const accessories = visual.accessoryIds ?? [];
    return [
      visual.skinColorId,
      visual.hairId,
      ...accessories.filter((id) => id.startsWith("accessory_hat_")),
      visual.eyeColorId,
      ...(visual.faceDetailId ? [visual.faceDetailId] : []),
      ...accessories.filter((id) => !id.startsWith("accessory_hat_")),
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
