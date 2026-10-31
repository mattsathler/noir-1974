import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

function dateKey(year, month, day) {
  return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

function monthLabel(month) {
  return ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"][month - 1];
}

export default function DatePicker({ label, value, onChange, minDate, maxDate }) {
  const [open, setOpen] = useState(false);
  const [visibleMonth, setVisibleMonth] = useState(() => `${value.slice(0, 7)}-01`);
  const [position, setPosition] = useState(null);
  const triggerRef = useRef(null);
  const calendarRef = useRef(null);
  const selected = new Date(`${value}T00:00:00Z`);
  const visible = new Date(`${visibleMonth}T00:00:00Z`);
  const year = visible.getUTCFullYear();
  const month = visible.getUTCMonth();
  const monthDays = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
  const offset = (new Date(Date.UTC(year, month, 1)).getUTCDay() + 6) % 7;
  const days = [...Array(offset).fill(null), ...Array.from({ length: monthDays }, (_, index) => index + 1)];
  const parentWindow = triggerRef.current?.closest(".computer-window");
  const siteVariant = parentWindow?.dataset.siteVariant;
  const calendarClassName = `world-calendar${parentWindow?.classList.contains("computer-window--website") ? " world-calendar--website" : ""}`;

  useEffect(() => setVisibleMonth(`${value.slice(0, 7)}-01`), [value]);

  useEffect(() => {
    if (!open) return undefined;
    function closeOnOutsidePointer(event) {
      if (triggerRef.current?.contains(event.target) || calendarRef.current?.contains(event.target)) return;
      setOpen(false);
    }
    function closeOnEscape(event) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("pointerdown", closeOnOutsidePointer);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsidePointer);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  function toggleCalendar() {
    if (open) {
      setOpen(false);
      return;
    }
    const trigger = triggerRef.current;
    if (trigger) {
      const triggerRect = trigger.getBoundingClientRect();
      const popupWidth = Math.min(256, window.innerWidth - 16);
      const popupHeight = 260;
      const below = triggerRect.bottom + 4;
      const above = triggerRect.top - popupHeight - 4;
      const top = below + popupHeight <= window.innerHeight - 8 ? below : Math.max(8, above);
      const left = Math.max(8, Math.min(window.innerWidth - popupWidth - 8, triggerRect.left));
      setPosition({ position: "fixed", left, top, width: popupWidth });
    }
    setOpen(true);
  }

  function moveMonth(amount) {
    const next = new Date(Date.UTC(year, month + amount, 1));
    setVisibleMonth(dateKey(next.getUTCFullYear(), next.getUTCMonth() + 1, 1));
  }

  function chooseDate(candidate) {
    onChange(candidate);
    setOpen(false);
    const form = triggerRef.current?.closest("form");
    const pickers = form ? [...form.querySelectorAll(".world-record__date-trigger")] : [];
    const index = pickers.indexOf(triggerRef.current);
    const target = index >= 0 && index < pickers.length - 1
      ? pickers[index + 1]
      : form?.querySelector(".search-enter-submit, .email-breaker__deeplink");
    window.setTimeout(() => (target || triggerRef.current)?.focus(), 0);
  }

  const calendar = open && position ? createPortal(
    <section ref={calendarRef} className={calendarClassName} data-site-variant={siteVariant} style={position} aria-label={`Choose ${label} date`}>
      <header className="world-calendar__header">
        <button type="button" aria-label="Previous month" onClick={() => moveMonth(-1)}>‹</button>
        <strong>{monthLabel(month + 1)}</strong>
        <button type="button" aria-label="Next month" onClick={() => moveMonth(1)}>›</button>
      </header>
      <div className="world-calendar__grid" role="group" aria-label={monthLabel(month + 1)}>
        {["M", "T", "W", "T", "F", "S", "S"].map((day, index) => <span className="world-calendar__weekday" key={`${day}-${index}`}>{day}</span>)}
        {days.map((day, index) => {
          if (day === null) return <span className="world-calendar__blank" key={`blank-${index}`} />;
          const candidate = dateKey(year, month + 1, day);
          return <button key={candidate} type="button" aria-label={`${monthLabel(month + 1)} ${day}`} aria-pressed={candidate === value} disabled={candidate < minDate || candidate > maxDate} className={candidate === value ? "world-calendar__day world-calendar__day--selected" : "world-calendar__day"} onClick={() => chooseDate(candidate)}>{day}</button>;
        })}
      </div>
    </section>, document.body,
  ) : null;

  return (
    <div className="world-record__date-picker">
      <span className="world-record__date-label">{label}</span>
      <button ref={triggerRef} className="world-record__date-trigger" type="button" aria-label={`Select date ${label}`} aria-expanded={open} onClick={toggleCalendar}>
        {new Intl.DateTimeFormat("en-US", { month: "short", day: "2-digit", timeZone: "UTC" }).format(selected)} <span aria-hidden="true">▾</span>
      </button>
      {calendar}
    </div>
  );
}
