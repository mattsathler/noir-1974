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
