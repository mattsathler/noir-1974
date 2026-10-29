import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import ModularFace from "./ModularFace.jsx";

afterEach(cleanup);

describe("ModularFace", () => {
  it("composes a face in the shared sprite layer order", () => {
    const visual = {
      skinId: "skin_olive",
      hairId: "hair_short_ginger",
      eyesId: "eyes_green",
      beardId: "beard_moustache_ginger",
      faceId: "face_scar_short",
      accessoryIds: ["accessories_glasses_round_brass"],
    };

    render(<ModularFace visual={visual} name="Ada" />);

    expect(screen.getByRole("img", { name: "Pixel-art portrait of Ada" })).toBeInTheDocument();
    expect([...document.querySelectorAll(".modular-face__layer")].map((image) => image.getAttribute("src"))).toEqual([
      "/images/characters/skin/skin_olive.png",
      "/images/characters/hair/hair_short_ginger.png",
      "/images/characters/eyes/eyes_green.png",
      "/images/characters/beard/beard_moustache_ginger.png",
      "/images/characters/face/face_scar_short.png",
      "/images/characters/accessories/accessories_glasses_round_brass.png",
    ]);
  });

  it("omits optional layers and refuses incomplete identities", () => {
    const { rerender } = render(<ModularFace visual={{ skinId: "skin_white", hairId: "hair_short_brown", eyesId: "eyes_brown" }} />);
    expect(document.querySelectorAll(".modular-face__layer")).toHaveLength(3);

    rerender(<ModularFace visual={{ skinId: "skin_white", hairId: "hair_short_brown" }} />);
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("loads added facial details from crisp transparent SVG layers", () => {
    render(<ModularFace visual={{ skinId: "skin_medium", hairId: "hair_long_curly_brown", eyesId: "eyes_hazel", faceId: "face_freckles_light" }} />);
    expect(document.querySelector(".modular-face__layer[src='/images/characters/face/face_freckles_light.svg']")).toBeInTheDocument();
  });
});
