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
    expect(screen.getByLabelText("E-mail")).toBeRequired();
    expect(screen.getByLabelText("Senha")).toBeRequired();
    expect(screen.getByText("Máximo de 72 bytes em UTF-8.")).toBeInTheDocument();

    const form = screen.getByRole("button", { name: "Entrar" }).closest("form");
    const event = new Event("submit", { bubbles: true, cancelable: true });
    form.dispatchEvent(event);
    expect(event.defaultPrevented).toBe(true);
  });

  it("opens the registration flow", () => {
    const onRegister = vi.fn();
    render(<LoginPage onRegister={onRegister} />);
    fireEvent.click(screen.getByRole("button", { name: "Criar conta" }));
    expect(onRegister).toHaveBeenCalledOnce();
  });

  it("submits credentials to the identity API", async () => {
    loginAccount.mockResolvedValue({ account: { id: "1" }, character: null });
    const onAuthenticated = vi.fn();
    render(<LoginPage onRegister={vi.fn()} onAuthenticated={onAuthenticated} />);
    fireEvent.change(screen.getByLabelText("E-mail"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Senha"), { target: { value: "secret-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));

    await vi.waitFor(() => expect(onAuthenticated).toHaveBeenCalledWith({ account: { id: "1" }, character: null }));
    expect(loginAccount).toHaveBeenCalledWith({ email: "agent@example.com", password: "secret-password" });
  });

  it("restores an existing session when the login page opens", async () => {
    const identity = { account: { id: "1" }, character: { displayName: "Ada", portraitId: 2 } };
    currentIdentity.mockResolvedValue(identity);
    const onAuthenticated = vi.fn();
    render(<LoginPage onRegister={vi.fn()} onAuthenticated={onAuthenticated} />);
    await vi.waitFor(() => expect(onAuthenticated).toHaveBeenCalledWith(identity));
  });
});
