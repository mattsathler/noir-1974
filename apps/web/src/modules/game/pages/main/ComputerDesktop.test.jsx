import { cleanup, fireEvent, render, screen } from "@testing-library/react";
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
    const firstOpen = render(<ComputerDesktop facilityContext={{ kind: "personal", label: "NOIR PERSONAL COMPUTER" }} />);
    expect(screen.getByRole("status")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Complete boot sequence" }));
    expect(await screen.findByText("NOIR PERSONAL COMPUTER", {}, { timeout: 1500 })).toBeInTheDocument();
    firstOpen.unmount();

    render(<ComputerDesktop facilityContext={{ kind: "personal", label: "NOIR PERSONAL COMPUTER" }} />);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByText("NOIR PERSONAL COMPUTER")).toBeInTheDocument();
  });
});
