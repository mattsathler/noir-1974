import { act, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import Typewriter from "./Typewriter.jsx";

describe("Typewriter", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("reveals characters and calls onComplete", () => {
    vi.useFakeTimers();
    const onComplete = vi.fn();
    const { container } = render(<Typewriter text="Gate" speed={10} onComplete={onComplete} className="custom" />);
    const text = container.querySelector("span");
    expect(text.dataset.complete).toBe("false");
    expect(text.textContent).toBe("");

    act(() => vi.advanceTimersByTime(40));
    expect(text.textContent).toBe("Gate");
    expect(text.dataset.complete).toBe("true");
    expect(onComplete).toHaveBeenCalledOnce();
    expect(text).toHaveClass("custom");
  });

  it("skips animation when reduced motion is requested", () => {
    vi.spyOn(window, "matchMedia").mockReturnValue({ matches: true });
    const onComplete = vi.fn();
    const { container } = render(<Typewriter text="Lyon" onComplete={onComplete} />);
    expect(container.querySelector("span").textContent).toBe("Lyon");
    expect(onComplete).toHaveBeenCalledOnce();
  });

  it("clears a pending timer when unmounted", () => {
    vi.useFakeTimers();
    const clearTimeout = vi.spyOn(window, "clearTimeout");
    const { unmount } = render(<Typewriter text="GATE" startDelay={50} />);
    unmount();
    expect(clearTimeout).toHaveBeenCalled();
  });
});
