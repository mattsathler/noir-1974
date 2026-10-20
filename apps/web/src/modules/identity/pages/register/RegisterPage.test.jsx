import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import RegisterPage from "./RegisterPage.jsx";
import { registerAccount } from "../../api.js";

vi.mock("../../api.js", () => ({ registerAccount: vi.fn() }));

afterEach(cleanup);

describe("RegisterPage", () => {
  function fillRequiredFields() {
    fireEvent.change(screen.getByLabelText("Usuário"), { target: { value: "agent-01" } });
    fireEvent.change(screen.getByLabelText("E-mail"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Senha"), { target: { value: "first-password" } });
    fireEvent.change(screen.getByLabelText("Confirmar senha"), { target: { value: "different-password" } });
  }

  it("renders required account fields and prevents navigation on submit", () => {
    render(<RegisterPage onLogin={vi.fn()} />);

    expect(screen.getByLabelText("Usuário")).toBeRequired();
    expect(screen.getByLabelText("E-mail")).toHaveAttribute("type", "email");
    expect(screen.getByLabelText("E-mail")).toBeRequired();
    expect(screen.getByLabelText("Senha")).toBeRequired();
    expect(screen.getByLabelText("Confirmar senha")).toBeRequired();
    expect(screen.getByLabelText("Usuário")).toHaveAttribute("pattern", "[A-Za-z0-9_-]{3,24}");
    expect(screen.getByText("De 12 a 72 bytes em UTF-8.")).toBeInTheDocument();

    fillRequiredFields();
    fireEvent.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(screen.getByRole("alert")).toHaveTextContent("As senhas não coincidem.");
    expect(screen.getByLabelText("Confirmar senha")).toHaveAttribute("aria-invalid", "true");
  });

  it("explains username and password constraints before calling the API", () => {
    render(<RegisterPage onLogin={vi.fn()} onAccountCreated={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Usuário"), { target: { value: "invalid name" } });
    fireEvent.change(screen.getByLabelText("E-mail"), { target: { value: "agent@example.com" } });
    fireEvent.change(screen.getByLabelText("Senha"), { target: { value: "too-short" } });
    fireEvent.change(screen.getByLabelText("Confirmar senha"), { target: { value: "too-short" } });
    fireEvent.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(screen.getByRole("alert")).toHaveTextContent("O usuário deve ter de 3 a 24 caracteres");
    expect(registerAccount).not.toHaveBeenCalled();
  });

  it("creates the account when matching passwords are submitted", async () => {
    registerAccount.mockResolvedValue({ account: { id: "1" }, character: null });
    const onAccountCreated = vi.fn();
    render(<RegisterPage onLogin={vi.fn()} onAccountCreated={onAccountCreated} />);
    fillRequiredFields();
    fireEvent.change(screen.getByLabelText("Senha"), { target: { value: "same-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Criar conta" }));

    fireEvent.change(screen.getByLabelText("Confirmar senha"), { target: { value: "same-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Criar conta" }));

    await waitFor(() => expect(onAccountCreated).toHaveBeenCalledWith({ account: { id: "1" }, character: null }));
    expect(registerAccount).toHaveBeenCalledWith({ username: "agent-01", email: "agent@example.com", password: "same-password" });
  });
});
