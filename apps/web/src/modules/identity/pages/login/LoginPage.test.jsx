import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import LoginPage from "./LoginPage.jsx";
import { currentIdentity, loginAccount } from "../../api.js";

vi.mock("../../api.js", () => ({ currentIdentity: vi.fn(), loginAccount: vi.fn() }));

beforeEach(() => currentIdentity.mockRejectedValue(new Error("no session")));
afterEach(() => {
  cleanup();
  currentIdentity.mockResolvedValue({ character: null });
});

describe("LoginPage", () => {
  it("renders required credentials and prevents form navigation", () => {
    render(<LoginPage onRegister={vi.fn()} />);
    expect(screen.getByRole("img", { name: "NOIR 1974" })).toHaveAttribute("src", "/images/logo.png");
    expect(screen.getByLabelText("Email")).toBeRequired();
    expect(screen.getByLabelText("Password")).toBeRequired();

    const form = screen.getByRole("button", { name: "Sign in" }).closest("form");
    const event = new Event("submit", { bubbles: true, cancelable: true });
    form.dispatchEvent(event);
    expect(event.defaultPrevented).toBe(true);
  });

  it("opens the registration flow", () => {
    const onRegister = vi.fn();
    render(<LoginPage onRegister={onRegister} />);
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));
    expect(onRegister).toHaveBeenCalledOnce();
  });

  it("submits credentials to the identity API", async () => {
    loginAccount.mockResolvedValue({ account: { id: "1" }, character: null, world: { serverTime: "2026-11-12T20:15:00.000Z" } });
    const onAuthenticated = vi.fn();
    render(<LoginPage onRegister={vi.fn()} onAuthenticated={onAuthenticated} />);
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "secret-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await vi.waitFor(() => expect(onAuthenticated).toHaveBeenCalledWith({ account: { id: "1" }, character: null, world: { serverTime: "2026-11-12T20:15:00.000Z" } }));
    expect(loginAccount).toHaveBeenCalledWith({ email: "agent@example.com", password: "secret-password" });
  });

  it("ignores duplicate submits and stays locked until navigation", async () => {
    let resolveLogin;
    loginAccount.mockReturnValue(new Promise((resolve) => { resolveLogin = resolve; }));
    const onAuthenticated = vi.fn();
    render(<LoginPage onRegister={vi.fn()} onAuthenticated={onAuthenticated} />);
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "secret-password" } });
    const form = screen.getByRole("button", { name: "Sign in" }).closest("form");

    fireEvent.submit(form);
    fireEvent.submit(form);
    expect(loginAccount).toHaveBeenCalledOnce();
    expect(screen.getByRole("button", { name: "Verifying…" })).toBeDisabled();

    const identity = { account: { id: "1" }, character: { displayName: "Ada" }, world: { serverTime: "2026-11-12T20:15:00.000Z" } };
    resolveLogin(identity);
    await vi.waitFor(() => expect(onAuthenticated).toHaveBeenCalledWith(identity));
    expect(screen.getByRole("button", { name: "Verifying…" })).toBeDisabled();
  });

  it("continues an existing player session after successful sign in", async () => {
    const identity = { account: { id: "1" }, character: { displayName: "Ada", portraitId: 2 }, world: { serverTime: "2026-11-12T20:15:00.000Z" } };
    loginAccount.mockResolvedValue(identity);
    const onAuthenticated = vi.fn();
    render(<LoginPage onRegister={vi.fn()} onAuthenticated={onAuthenticated} />);
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "secret-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
    await vi.waitFor(() => expect(onAuthenticated).toHaveBeenCalledWith(identity));
  });
});
