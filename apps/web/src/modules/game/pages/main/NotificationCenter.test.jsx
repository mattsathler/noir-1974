import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import NotificationCenter from "./NotificationCenter.jsx";

afterEach(cleanup);

describe("NotificationCenter", () => {
  it("shows a snackbar and keeps notifications in an accessible history overlay", () => {
    const notification = { id: "mail-1", type: "EMAIL", title: "New case report", message: "From Bill Bouer.", time: "9:30 AM" };
    const onToggle = vi.fn();
    render(<NotificationCenter notifications={[notification]} open={false} onToggle={onToggle} onClose={vi.fn()} snackbar={notification} onDismissSnackbar={vi.fn()} />);

    expect(screen.getByRole("status")).toHaveTextContent("New case report");
    fireEvent.click(screen.getByRole("button", { name: /NOTICES/ }));
    expect(onToggle).toHaveBeenCalledOnce();
    cleanup();

    render(<NotificationCenter notifications={[notification]} open onToggle={vi.fn()} onClose={vi.fn()} snackbar={null} onDismissSnackbar={vi.fn()} />);
    expect(screen.getByRole("dialog", { name: "Notification history" })).toBeInTheDocument();
    expect(screen.getByText("From Bill Bouer.")).toBeInTheDocument();
  });
});
