import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import IntroductionPage from "./IntroductionPage.jsx";

vi.mock("../../../../../../@noir-ui/components/Typewriter.jsx", () => ({
  default: ({ text, onComplete }) => <button onClick={onComplete}>{text}</button>,
}));

describe("IntroductionPage", () => {
  afterEach(() => vi.useRealTimers());

  it("runs the GATE terminal boot sequence after the location text completes", () => {
    vi.useFakeTimers();
    const onComplete = vi.fn();
    render(<IntroductionPage onComplete={onComplete} />);
    fireEvent.click(screen.getByRole("button", { name: "Lyon, France" }));
    expect(screen.getByRole("progressbar", { name: "Inicialização" })).toHaveAttribute("aria-valuenow", "0");

    act(() => vi.advanceTimersByTime(300));
    expect(screen.getByText("MERIDIAN GLOBAL / SYSTEM 74.04")).toBeInTheDocument();
    expect(onComplete).not.toHaveBeenCalled();

    act(() => vi.advanceTimersByTime(2730));
    expect(onComplete).toHaveBeenCalledOnce();
  });
});
