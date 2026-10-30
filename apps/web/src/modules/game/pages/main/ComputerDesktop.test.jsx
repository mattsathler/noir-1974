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
    const firstOpen = render(<ComputerDesktop facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    expect(screen.getByRole("status")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Complete boot sequence" }));
    expect(await screen.findByText("MERID_OS PERSONAL COMPUTER", {}, { timeout: 1500 })).toBeInTheDocument();
    firstOpen.unmount();

    render(<ComputerDesktop facilityContext={{ kind: "personal", label: "MERID_OS PERSONAL COMPUTER" }} />);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByText("MERID_OS PERSONAL COMPUTER")).toBeInTheDocument();
  });

  it("opens the settings app and lets the player change desktop preferences", async () => {
    sessionStorage.setItem("noir:computer:personal-booted", "true");
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
