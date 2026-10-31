import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import Telephone from "./Telephone.jsx";
import { getInterrogationAnswers, lookupPhoneContact } from "../../../world/api.js";

vi.mock("../../../world/api.js", () => ({ getInterrogationAnswers: vi.fn(), lookupPhoneContact: vi.fn() }));
vi.mock("../../../../../../@noir-ui/components/Typewriter.jsx", () => ({ default: ({ text }) => <span>{text}</span> }));

afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.clearAllMocks(); });

describe("Telephone notifications", () => {
  it("records a notice when an outgoing call connects", async () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const onNotification = vi.fn();
    lookupPhoneContact.mockResolvedValue({ contact: { id: "npc:ada", name: "Ada Martin", company: "Meridian", status: "alive" } });
    getInterrogationAnswers.mockResolvedValue({ answers: { unusual: ["The papers say Bill Bouer has a new lead on The Crow. Someone drew a skull on the station door."] } });
    render(<Telephone onClose={vi.fn()} onNotification={onNotification} />);

    fireEvent.change(screen.getByRole("textbox", { name: "Telephone number" }), { target: { value: "1234567" } });
    fireEvent.click(screen.getByRole("button", { name: "Dial number" }));

    expect(await screen.findByLabelText("Telephone interview with Ada Martin", {}, { timeout: 2500 })).toBeInTheDocument();
    expect(onNotification).toHaveBeenCalledWith({ type: "CALL", title: "Connected to Ada Martin", message: "Outgoing line is open." });
    const unusualQuestion = screen.getByRole("button", { name: "Has anything unusual happened around here recently?" });
    await waitFor(() => expect(unusualQuestion).toBeEnabled());
    fireEvent.click(unusualQuestion);
    expect(screen.getByRole("status")).toHaveTextContent("Bill Bouer has a new lead on The Crow");
  });
});
