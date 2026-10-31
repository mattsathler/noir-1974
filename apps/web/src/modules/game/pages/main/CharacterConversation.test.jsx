import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import CharacterConversation from "./CharacterConversation.jsx";

vi.mock("../../../../../../@noir-ui/components/Typewriter.jsx", () => ({ default: ({ text }) => <span>{text}</span> }));

afterEach(() => { cleanup(); vi.restoreAllMocks(); });

describe("CharacterConversation", () => {
  it.each([["bill", "Bill Bouer"], ["crow", "The Crow"]])("starts a conversation with %s", (characterId, name) => {
    render(<CharacterConversation characterId={characterId} onClose={vi.fn()} />);
    expect(screen.getByRole("dialog", { name })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: `Pixel-art portrait of ${name}` })).toBeInTheDocument();
    if (characterId === "crow") expect(screen.getByRole("img", { name: "The Crow's skull mark" })).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveTextContent(/Ghostkey|breach|records/iu);
  });

  it("advances the dialogue while retaining the same interlocutor", () => {
    render(<CharacterConversation characterId="crow" onClose={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: "You still steal from those companies." }));
    expect(screen.getByRole("dialog", { name: "The Crow" })).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveTextContent(/Don't mistake that for innocence/u);
    expect(screen.getByText("TRANSMISSION 02 / 03")).toBeInTheDocument();
  });
});
