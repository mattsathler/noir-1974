import { useEffect, useRef } from "react";
import { createPortal } from "react-dom";

export default function NotificationCenter({ notifications = [], open = false, onToggle, onClose, snackbar, onDismissSnackbar }) {
  const dialogRef = useRef(null);
  useEffect(() => {
    if (!open) return undefined;
    const previousFocus = document.activeElement;
    dialogRef.current?.querySelector("button")?.focus();
    function onKeyDown(event) {
      if (event.key === "Escape") onClose?.();
      if (event.key !== "Tab" || !dialogRef.current) return;
      const buttons = [...dialogRef.current.querySelectorAll("button:not(:disabled)")];
      const first = buttons[0];
      const last = buttons.at(-1);
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
    }
    document.addEventListener("keydown", onKeyDown);
    return () => { document.removeEventListener("keydown", onKeyDown); previousFocus?.focus?.(); };
  }, [open, onClose]);

  return <>
    <button className="notification-center__toggle" type="button" aria-label={`Open notification history, ${notifications.length} total`} aria-expanded={open} aria-controls="notification-center-panel" onClick={() => { if (!open) onToggle?.(); }}>
      <svg viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 1a4 4 0 0 0-4 4v2.2c0 .8-.3 1.5-.8 2.1L2 10.5V12h12v-1.5l-1.2-1.2a3 3 0 0 1-.8-2.1V5a4 4 0 0 0-4-4Zm-1.5 12a1.5 1.5 0 0 0 3 0h-3Z" /></svg>
      {notifications.length > 0 && <span aria-hidden="true">{notifications.length > 99 ? "99+" : notifications.length}</span>}
    </button>
    {snackbar && createPortal(<div className="notification-center__snackbar" role="status" aria-live="polite" aria-atomic="true">
      <span><strong>{snackbar.title}</strong><small>{snackbar.message}</small></span>
      <button type="button" onClick={onDismissSnackbar} aria-label="Dismiss notification">×</button>
    </div>, document.body)}
    {open && <section id="notification-center-panel" className="computer-window notification-center__window" ref={dialogRef} role="dialog" aria-modal="false" aria-labelledby="notification-center-title">
      <header className="computer-window__titlebar"><span><span className="computer-status-light" /> NOTIFICATION CENTER</span><button className="computer-window__close" type="button" onClick={onClose} aria-label="Close notification history">×</button></header>
      <div className="computer-window__body notification-center__content">
        <div className="notification-center__heading"><span>MERIDIAN OS · SYSTEM MESSAGES</span><h2 id="notification-center-title">Notification history</h2></div>
        {!notifications.length ? <p className="notification-center__empty">No new information.</p> : <ol>
          {notifications.map((item) => <li key={item.id}><small>{item.type} · {item.time}</small><strong>{item.title}</strong><p>{item.message}</p></li>)}
        </ol>}
      </div>
    </section>}
  </>;
}
