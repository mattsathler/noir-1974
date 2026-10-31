import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import ComputerDesktop from "./ComputerDesktop.jsx";

vi.mock("../../../../../../@noir-ui/components/Typewriter.jsx", () => ({
  default: ({ onComplete }) => <button type="button" onClick={onComplete}>Complete boot sequence</button>,
}));

vi.mock("../../../world/api.js", () => ({
  cachePinnedWhiteboardItem: vi.fn(),
  cacheWhiteboardDraft: vi.fn(),
  createWhiteboard: vi.fn(),
  createWhiteboardDraft: vi.fn(),
  getCachedWhiteboard: vi.fn(() => null),
  getWhiteboards: vi.fn().mockResolvedValue([]),
  pinWhiteboardItem: vi.fn(),
  setWhiteboardSaveState: vi.fn(),
}));

afterEach(() => {
  cleanup();
  sessionStorage.clear();
  vi.restoreAllMocks();
});

describe("ComputerDesktop boot sequence", () => {
  it("boots the personal computer only once per browser session", async () => {
    vi.spyOn(window, "matchMedia").mockReturnValue({ matches: true, addListener: vi.fn(), removeListener: vi.fn() });
    const firstOpen = render(<ComputerDesktop facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    expect(screen.getByRole("status")).toBeInTheDocument();

    await waitFor(() => expect(screen.queryByRole("status")).not.toBeInTheDocument(), { timeout: 5000 });
    firstOpen.unmount();

    render(<ComputerDesktop facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "Detective computer desktop" })).toBeInTheDocument();
  }, 10000);

  it("opens the settings app and lets the player change desktop preferences", async () => {
    sessionStorage.setItem("noir:computer:booted-this-login", "true");
    const onDesktopPreferencesChange = vi.fn();
    render(<ComputerDesktop persistent facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} desktopPreferences={{ wallpaper: "meridian-night", theme: "dark" }} onDesktopPreferencesChange={onDesktopPreferencesChange} onAppearance={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "SETTINGS application, READY" }));
    expect(await screen.findByRole("heading", { name: "Desktop settings" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Relay blue/i }));
    fireEvent.click(screen.getByRole("button", { name: /Light A brighter desktop/i }));
    await waitFor(() => expect(onDesktopPreferencesChange).toHaveBeenNthCalledWith(2, { theme: "light" }));
    expect(onDesktopPreferencesChange).toHaveBeenNthCalledWith(1, { wallpaper: "relay-blue" });
  });
});

describe("ComputerDesktop window minimization", () => {
  it("keeps a window mounted and restores its content and position; closing stays in the window", async () => {
    sessionStorage.setItem("noir:computer:booted-this-login", "true");
    sessionStorage.setItem("noir:computer:windows", JSON.stringify({
      savedAt: Date.now(),
      windows: [{ id: "noir-browser:12", programId: "terminal", zIndex: 12, position: { left: 88, top: 46 } }],
    }));
    const { container } = render(<ComputerDesktop persistent facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    const terminal = container.querySelector('[data-window-id="noir-browser:12"]');
    const savedPosition = { left: Number.parseFloat(terminal.style.left), top: Number.parseFloat(terminal.style.top) };
    const prompt = terminal.querySelector(".terminal-app__prompt input");
    fireEvent.change(prompt, { target: { value: "HELP" } });
    fireEvent.submit(prompt.closest("form"));
    expect(await screen.findByText("Available commands:")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Minimize TERMINAL" }));
    expect(terminal).toHaveAttribute("hidden");
    expect(terminal).toHaveTextContent("Available commands:");
    const restore = screen.getByRole("button", { name: "Restore TERMINAL, minimized" });
    expect(restore).toHaveClass("computer-taskbar__active--minimized");
    await waitFor(() => expect(JSON.parse(sessionStorage.getItem("noir:computer:windows")).windows[0]).toMatchObject({ minimized: true, position: savedPosition }));

    fireEvent.click(restore);
    expect(terminal).not.toHaveAttribute("hidden");
    expect(terminal).toHaveTextContent("Available commands:");
    expect(terminal).toHaveStyle({ left: `${savedPosition.left}px`, top: `${savedPosition.top}px` });
    expect(restore).toHaveAttribute("aria-pressed", "true");

    fireEvent.click(screen.getByRole("button", { name: "Minimize TERMINAL" }));
    expect(screen.queryByRole("button", { name: "Close minimized TERMINAL" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Restore TERMINAL, minimized" }));
    fireEvent.click(screen.getByRole("button", { name: "Close TERMINAL" }));
    expect(container.querySelector('[data-window-id="noir-browser:12"]')).not.toBeInTheDocument();
  });

  it("restores minimized status after remounting in the same session", async () => {
    sessionStorage.setItem("noir:computer:booted-this-login", "true");
    const { container, unmount } = render(<ComputerDesktop persistent facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    fireEvent.click(screen.getByRole("button", { name: "TERMINAL application, READY" }));
    fireEvent.click(await screen.findByRole("button", { name: "Minimize TERMINAL" }));
    unmount();

    const reopened = render(<ComputerDesktop persistent facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    expect(reopened.container.querySelector('[data-window-id^="noir-browser:"]')).toHaveAttribute("hidden");
    fireEvent.click(screen.getByRole("button", { name: "Restore TERMINAL, minimized" }));
    expect(reopened.container.querySelector('[data-window-id^="noir-browser:"]')).not.toHaveAttribute("hidden");
    expect(container).toBeTruthy();
  });
});
