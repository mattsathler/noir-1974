import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import ArchivePage from "./ArchivePage.jsx";
import { currentIdentity, logoutAccount } from "../../api.js";

vi.mock("../../api.js", () => ({ currentIdentity: vi.fn(), logoutAccount: vi.fn() }));
afterEach(cleanup);

describe("ArchivePage", () => {
  it("shows the saved character and ends the session", async () => {
    currentIdentity.mockResolvedValue({ account: { username: "detective" }, character: { displayName: "Ada", portraitId: 4 } });
    logoutAccount.mockResolvedValue(null);
    const onLogin = vi.fn();
    render(<ArchivePage onLogin={onLogin} />);

    expect(await screen.findByText("Ada")).toBeInTheDocument();
    expect(screen.getByText("Detetive particular · ex-colaborador da Interpol")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Encerrar sessão" }));
    await waitFor(() => expect(onLogin).toHaveBeenCalledOnce());
  });
});
