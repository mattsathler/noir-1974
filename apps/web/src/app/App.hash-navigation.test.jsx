import { act, cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import App from "./App.jsx";
import { currentIdentity } from "../modules/identity/api.js";

vi.mock("../modules/identity/api.js", () => ({
  currentIdentity: vi.fn(),
  logoutAccount: vi.fn(),
}));
vi.mock("../modules/game/pages/main/GamePage.jsx", () => ({
  default: () => <main>Game remains open</main>,
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

it("keeps the game mounted without rechecking the session on fragment navigation", async () => {
  window.history.replaceState({}, "", "/game");
  currentIdentity.mockResolvedValue({ character: { displayName: "Ada" }, world: {} });
  render(<App />);
  const game = await screen.findByText("Game remains open");
  expect(currentIdentity).toHaveBeenCalledTimes(1);

  act(() => {
    window.history.pushState({}, "", "/game#sites");
    window.dispatchEvent(new PopStateEvent("popstate"));
  });

  expect(screen.getByText("Game remains open")).toBe(game);
  expect(currentIdentity).toHaveBeenCalledTimes(1);
  expect(window.location.hash).toBe("#sites");
});

it("still checks the session when history enters a different protected route", async () => {
  window.history.replaceState({}, "", "/login");
  currentIdentity.mockResolvedValue({ character: { displayName: "Ada" }, world: {} });
  render(<App />);

  act(() => {
    window.history.pushState({}, "", "/game");
    window.dispatchEvent(new PopStateEvent("popstate"));
  });

  expect(await screen.findByText("Game remains open")).toBeInTheDocument();
  expect(currentIdentity).toHaveBeenCalledTimes(1);
});
