import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import CharacterCreationPage from "./CharacterCreationPage.jsx";
import { createCharacter } from "../../api.js";

vi.mock("../../api.js", () => ({ createCharacter: vi.fn() }));
afterEach(cleanup);

describe("CharacterCreationPage", () => {
  it("creates a private detective with one of six portrait selections", async () => {
    createCharacter.mockResolvedValue({ character: { displayName: "Ada", portraitId: 4 } });
    const onComplete = vi.fn();
    render(<CharacterCreationPage onComplete={onComplete} />);

    expect(screen.getAllByRole("button", { name: /Select portrait/ })).toHaveLength(6);
    fireEvent.click(screen.getByRole("button", { name: "Select portrait 4" }));
    fireEvent.change(screen.getByLabelText("Detective name"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Create detective" }));

    await waitFor(() => expect(onComplete).toHaveBeenCalledOnce());
    expect(createCharacter).toHaveBeenCalledWith({ displayName: "Ada", portraitId: 4 });
  });

  it("shows API errors and keeps the player in character setup", async () => {
    createCharacter.mockRejectedValue(new Error("Session expired."));
    render(<CharacterCreationPage onComplete={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Detective name"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Create detective" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Session expired.");
  });
});
