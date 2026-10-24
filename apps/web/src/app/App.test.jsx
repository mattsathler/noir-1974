import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App.jsx";

vi.mock("../modules/identity/api.js", () => ({
  currentIdentity: vi.fn(),
  logoutAccount: vi.fn().mockResolvedValue(null),
}));
import { currentIdentity } from "../modules/identity/api.js";

vi.mock("../modules/prologue/pages/introduction/IntroductionPage.jsx", () => ({
  default: ({ onComplete }) => <button onClick={onComplete}>Continue</button>,
}));
vi.mock("../modules/identity/pages/login/LoginPage.jsx", () => ({
  default: ({ onRegister, onAuthenticated }) => (
    <main>
      Login screen
      <button onClick={onRegister}>Open registration</button>
      <button onClick={() => onAuthenticated({ character: null })}>Authenticate new account</button>
      <button onClick={() => onAuthenticated({ character: { displayName: "Ada" } })}>Authenticate existing account</button>
    </main>
  ),
}));
vi.mock("../modules/identity/pages/register/RegisterPage.jsx", () => ({
  default: ({ onLogin, onAccountCreated }) => (
    <main>
      Register screen
      <button onClick={onLogin}>Back to login</button>
      <button onClick={() => onAccountCreated({ account: { id: "1" }, character: null })}>Create account</button>
    </main>
  ),
}));
vi.mock("../modules/identity/pages/character-creation/CharacterCreationPage.jsx", () => ({
  default: ({ onComplete }) => <button onClick={onComplete}>Finish character</button>,
}));
vi.mock("../modules/identity/pages/archive/ArchivePage.jsx", () => ({
  default: () => <main>Archive screen</main>,
}));
vi.mock("../modules/cases/pages/files/Whiteboard.jsx", () => ({
  default: ({ initialBoardId, standalone }) => <main data-board-id={initialBoardId} data-standalone={standalone}>Standalone whiteboard</main>,
}));

describe("App routing", () => {
  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    currentIdentity.mockRejectedValue(new Error("no session"));
  });
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("shows the introduction and navigates to login", async () => {
    render(<App />);
    await screen.findByRole("button", { name: "Continue" });
    fireEvent.click(screen.getByRole("button", { name: "Continue" }));
    expect(await screen.findByText("Login screen", {}, { timeout: 3000 })).toBeInTheDocument();
    expect(window.location.pathname).toBe("/login");
  });

  it("checks the session before showing login", async () => {
    window.history.replaceState({}, "", "/login");
    render(<App />);
    expect(await screen.findByText("Login screen")).toBeInTheDocument();
  });

  it("opens a board directly in the standalone route after checking the session", async () => {
    window.history.replaceState({}, "", "/whiteboard/board_case-notes");
    currentIdentity.mockResolvedValue({ character: { displayName: "Ada" }, world: { city: "Lyon" } });

    render(<App />);

    expect(await screen.findByText("Standalone whiteboard")).toBeInTheDocument();
    expect(document.querySelector("[data-board-id]")).toHaveAttribute("data-board-id", "board_case-notes");
    expect(document.querySelector("[data-standalone]")).toHaveAttribute("data-standalone", "true");
    expect(window.location.pathname).toBe("/whiteboard/board_case-notes");
  });

  it("opens the board picker route without requiring an existing board id", async () => {
    window.history.replaceState({}, "", "/whiteboard");
    currentIdentity.mockResolvedValue({ character: { displayName: "Ada" } });
    render(<App />);

    expect(await screen.findByText("Standalone whiteboard")).toBeInTheDocument();
    expect(document.querySelector("[data-standalone]")).toHaveAttribute("data-standalone", "true");
    expect(window.location.pathname).toBe("/whiteboard");
  });

  it("returns an authenticated player to a requested board after login", async () => {
    window.history.replaceState({}, "", "/whiteboard/board_case-notes");
    localStorage.setItem("noir:world-data:v1:gate-cities", JSON.stringify({ items: [{ id: "stale" }] }));
    localStorage.setItem("noir:world-data:v1:neighborhoods:fr%2Flyon", JSON.stringify({ items: [{ id: "stale" }] }));
    localStorage.setItem("noir:whiteboard:v1:board_case-notes", JSON.stringify({ boardId: "board_case-notes" }));
    render(<App />);
    expect(await screen.findByText("Login screen")).toBeInTheDocument();
    expect(window.location.search).toContain("returnTo=%2Fwhiteboard%2Fboard_case-notes");
    fireEvent.click(screen.getByRole("button", { name: "Authenticate existing account" }));
    expect(await screen.findByText("Standalone whiteboard")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/whiteboard/board_case-notes");
    expect(localStorage.getItem("noir:world-data:v1:gate-cities")).toBeNull();
    expect(localStorage.getItem("noir:world-data:v1:neighborhoods:fr%2Flyon")).toBeNull();
    expect(localStorage.getItem("noir:whiteboard:v1:board_case-notes")).not.toBeNull();
  });

  it("starts on registration when its path is selected", () => {
    window.history.replaceState({}, "", "/register");
    render(<App />);
    expect(screen.getByText("Register screen")).toBeInTheDocument();
  });

  it("navigates between login and registration", async () => {
    window.history.replaceState({}, "", "/login");
    render(<App />);
    await screen.findByText("Login screen");
    fireEvent.click(screen.getByRole("button", { name: "Open registration" }));
    expect(screen.getByText("Register screen")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/register");

    fireEvent.click(screen.getByRole("button", { name: "Back to login" }));
    expect(await screen.findByText("Login screen")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/login");
  });

  it("opens first-login character creation after account registration", async () => {
    window.history.replaceState({}, "", "/register");
    render(<App />);
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));
    expect(await screen.findByRole("button", { name: "Finish character" }, { timeout: 3000 })).toBeInTheDocument();
    expect(window.location.pathname).toBe("/character/new");
  });

  it("sends first-time logins to character setup and returning players to the archive", async () => {
    window.history.replaceState({}, "", "/login");
    const firstRender = render(<App />);
    await screen.findByText("Login screen");
    fireEvent.click(screen.getByRole("button", { name: "Authenticate new account" }));
    expect(window.location.pathname).toBe("/character/new");

    firstRender.unmount();
    window.history.replaceState({}, "", "/login");
    render(<App />);
    await screen.findByText("Login screen");
    fireEvent.click(screen.getByRole("button", { name: "Authenticate existing account" }));
    expect(window.location.pathname).toBe("/archive");
  });

  it("follows browser history and removes its listener on unmount", async () => {
    const { unmount } = render(<App />);
    await screen.findByRole("button", { name: "Continue" });
    fireEvent.click(screen.getByRole("button", { name: "Continue" }));
    await screen.findByText("Login screen", {}, { timeout: 3000 });
    act(() => {
      window.history.replaceState({}, "", "/");
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    expect(await screen.findByRole("button", { name: "Continue" })).toBeInTheDocument();
    unmount();
  });
});
