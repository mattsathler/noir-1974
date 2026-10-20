import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import RegisterPage from "./RegisterPage.jsx";
import { registerAccount } from "../../api.js";

vi.mock("../../api.js", () => ({ registerAccount: vi.fn() }));

afterEach(cleanup);

describe("RegisterPage", () => {
  function fillRequiredFields() {
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "agent-01" } });
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "first-password" } });
    fireEvent.change(screen.getByLabelText("Confirm password"), { target: { value: "different-password" } });
  }

  it("renders required account fields and prevents navigation on submit", () => {
    render(<RegisterPage onLogin={vi.fn()} />);

    expect(screen.getByLabelText("Username")).toBeRequired();
    expect(screen.getByLabelText("Email")).toHaveAttribute("type", "email");
    expect(screen.getByLabelText("Email")).toBeRequired();
    expect(screen.getByLabelText("Password")).toBeRequired();
    expect(screen.getByLabelText("Confirm password")).toBeRequired();
    expect(screen.getByLabelText("Username")).toHaveAttribute("pattern", "[A-Za-z0-9_-]{3,24}");

    fillRequiredFields();
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    expect(screen.getByRole("alert")).toHaveTextContent("Passwords do not match.");
    expect(screen.getByLabelText("Confirm password")).toHaveAttribute("aria-invalid", "true");
  });

  it("explains username and password constraints before calling the API", () => {
    render(<RegisterPage onLogin={vi.fn()} onAccountCreated={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "invalid name" } });
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "too-short" } });
    fireEvent.change(screen.getByLabelText("Confirm password"), { target: { value: "too-short" } });
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    expect(screen.getByRole("alert")).toHaveTextContent("Username must be 3–24 characters");
    expect(registerAccount).not.toHaveBeenCalled();
  });

  it("creates the account when matching passwords are submitted", async () => {
    registerAccount.mockResolvedValue({ account: { id: "1" }, character: null });
    const onAccountCreated = vi.fn();
    render(<RegisterPage onLogin={vi.fn()} onAccountCreated={onAccountCreated} />);
    fillRequiredFields();
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "same-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    fireEvent.change(screen.getByLabelText("Confirm password"), { target: { value: "same-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    await waitFor(() => expect(onAccountCreated).toHaveBeenCalledWith({ account: { id: "1" }, character: null }));
    expect(registerAccount).toHaveBeenCalledWith({ username: "agent-01", email: "agent@example.com", password: "same-password" });
  });
});
