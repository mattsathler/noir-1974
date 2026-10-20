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

    expect(screen.getAllByRole("button", { name: /Selecionar retrato/ })).toHaveLength(6);
    fireEvent.click(screen.getByRole("button", { name: "Selecionar retrato 4" }));
    fireEvent.change(screen.getByLabelText("Nome do detetive"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Criar personagem" }));

    await waitFor(() => expect(onComplete).toHaveBeenCalledOnce());
    expect(createCharacter).toHaveBeenCalledWith({ displayName: "Ada", portraitId: 4 });
  });

  it("shows API errors and keeps the player in character setup", async () => {
    createCharacter.mockRejectedValue(new Error("Sessão expirada."));
    render(<CharacterCreationPage onComplete={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Nome do detetive"), { target: { value: "Ada" } });
    fireEvent.click(screen.getByRole("button", { name: "Criar personagem" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Sessão expirada.");
  });
});
