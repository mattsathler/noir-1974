import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import CharacterCreationPage from "./CharacterCreationPage.jsx";
import { createCharacter, getCharacterVisuals } from "../../api.js";

const visualOptions = {
  hair: ["hair_short_brown", "hair_short_ginger"],
  skinColors: ["skin_light"],
  eyeColors: ["eyes_brown"],
  faceDetails: ["face_moustache_black"],
  accessories: ["face_glasses_round_black"],
};

vi.mock("../../api.js", () => ({ createCharacter: vi.fn(), getCharacterVisuals: vi.fn() }));
afterEach(cleanup);

describe("CharacterCreationPage", () => {
  it("creates a private detective with a modular appearance", async () => {
    getCharacterVisuals.mockResolvedValue(visualOptions);
    createCharacter.mockResolvedValue({ character: { displayName: "Ada" } });
    const onComplete = vi.fn();
    render(<CharacterCreationPage onComplete={onComplete} />);

    fireEvent.change(await screen.findByLabelText("Hair"), { target: { value: "hair_short_ginger" } });
    fireEvent.change(screen.getByLabelText("Detective name"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Create detective" }));

    await waitFor(() => expect(onComplete).toHaveBeenCalledOnce());
    expect(createCharacter).toHaveBeenCalledWith({ displayName: "Ada", visual: { hairId: "hair_short_ginger", skinColorId: "skin_light", eyeColorId: "eyes_brown", faceDetailId: "", accessoryIds: [] } });
  });

  it("shows API errors and keeps the player in character setup", async () => {
    getCharacterVisuals.mockResolvedValue(visualOptions);
    createCharacter.mockRejectedValue(new Error("Session expired."));
    render(<CharacterCreationPage onComplete={vi.fn()} />);
    await screen.findByLabelText("Hair");
    fireEvent.change(screen.getByLabelText("Detective name"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Create detective" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Session expired.");
  });

  it("ignores repeated submits and stays locked after character creation", async () => {
    getCharacterVisuals.mockResolvedValue(visualOptions);
    let resolveCharacter;
    createCharacter.mockReturnValue(new Promise((resolve) => { resolveCharacter = resolve; }));
    const onComplete = vi.fn();
    render(<CharacterCreationPage onComplete={onComplete} />);
    await screen.findByLabelText("Hair");
    fireEvent.change(screen.getByLabelText("Detective name"), { target: { value: "Ada" } });
    const form = screen.getByRole("button", { name: "Create detective" }).closest("form");

    fireEvent.submit(form);
    fireEvent.submit(form);
    expect(createCharacter).toHaveBeenCalledOnce();
    expect(screen.getByRole("button", { name: "Filing…" })).toBeDisabled();

    resolveCharacter({ character: { displayName: "Ada", portraitId: 1 } });
    await waitFor(() => expect(onComplete).toHaveBeenCalledOnce());
    expect(screen.getByRole("button", { name: "Filing…" })).toBeDisabled();
  });
});
