import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import CharacterAppearanceEditor from "./CharacterAppearanceEditor.jsx";
import { getCharacterVisuals, updateCharacterAppearance } from "../api.js";

vi.mock("../api.js", () => ({ getCharacterVisuals: vi.fn(), updateCharacterAppearance: vi.fn() }));

const visual = {
  hairId: "hair_short_brown",
  skinId: "skin_white",
  eyesId: "eyes_brown",
  faceId: "",
  beardId: "",
  accessoryIds: [],
};
const options = {
  hair: ["hair_short_brown", "hair_long_curly_brown"],
  skin: ["skin_white", "skin_medium"],
  eyes: ["eyes_brown", "eyes_hazel"],
  face: ["face_scar_short", "face_freckles_light", "face_wrinkles_fine"],
  beard: ["beard_moustache_brown"],
  accessories: ["accessories_glasses_round_black"],
};

beforeEach(() => {
  vi.clearAllMocks();
  getCharacterVisuals.mockResolvedValue(options);
});
afterEach(cleanup);

describe("CharacterAppearanceEditor", () => {
  it("previews selections and cancels without persisting them", async () => {
    const onClose = vi.fn();
    render(<CharacterAppearanceEditor character={{ displayName: "Ada", visual }} onClose={onClose} onSaved={vi.fn()} />);

    const hair = await screen.findByLabelText("Hair");
    fireEvent.change(hair, { target: { value: "hair_long_curly_brown" } });
    expect(screen.getByRole("img", { name: "Pixel-art portrait of Ada" })).toBeInTheDocument();
    expect(updateCharacterAppearance).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledOnce();
    expect(updateCharacterAppearance).not.toHaveBeenCalled();
  });

  it("saves the complete selection and returns the persisted character", async () => {
    const updated = { displayName: "Ada", visual: { ...visual, hairId: "hair_long_curly_brown" } };
    updateCharacterAppearance.mockResolvedValue({ character: updated });
    const onSaved = vi.fn();
    render(<CharacterAppearanceEditor character={{ displayName: "Ada", visual }} onClose={vi.fn()} onSaved={onSaved} />);

    fireEvent.change(await screen.findByLabelText("Hair"), { target: { value: "hair_long_curly_brown" } });
    fireEvent.click(screen.getByRole("button", { name: "Save appearance" }));
    await waitFor(() => expect(onSaved).toHaveBeenCalledWith(updated));
    expect(updateCharacterAppearance).toHaveBeenCalledWith({ ...visual, hairId: "hair_long_curly_brown" });
  });

  it("does not steal focus from the selected control on parent rerenders", async () => {
    const character = { displayName: "Ada", visual };
    const { rerender } = render(<CharacterAppearanceEditor character={character} onClose={() => {}} onSaved={vi.fn()} />);
    const eyes = await screen.findByLabelText("Eyes");
    eyes.focus();
    rerender(<CharacterAppearanceEditor character={character} onClose={() => {}} onSaved={vi.fn()} />);
    expect(document.activeElement).toBe(eyes);
  });
});
