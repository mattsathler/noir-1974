import { useCallback, useEffect, useState } from "react";
import { ComputerSharp } from "pixelarticons/react/ComputerSharp";
import { NotebookSharp } from "pixelarticons/react/NotebookSharp";
import { currentIdentity } from "../../../identity/api.js";
import NewspaperWindow from "../../../world/pages/newspaper/NewspaperWindow.jsx";
import ComputerDesktop from "./ComputerDesktop.jsx";

function formatServerDateTime(timestamp, timeZone) {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return null;

  return {
    date: new Intl.DateTimeFormat("en-US", {
      timeZone,
      weekday: "long",
      day: "2-digit",
    }).format(date),
    time: new Intl.DateTimeFormat("en-US", {
      timeZone,
      hour: "numeric",
      minute: "2-digit",
      second: "2-digit",
      hour12: true,
    }).format(date),
  };
}

export default function GamePage({ initialWorld = null, onLogout }) {
  const [world, setWorld] = useState(initialWorld);
  const [serverTime, setServerTime] = useState(initialWorld ? new Date(initialWorld.serverTime) : null);
  const [error, setError] = useState("");
  const [computerOpen, setComputerOpen] = useState(false);
  const [newspaperOpen, setNewspaperOpen] = useState(false);
  const closeComputer = useCallback(() => setComputerOpen(false), []);
  const closeNewspaper = useCallback(() => setNewspaperOpen(false), []);

  useEffect(() => {
    let active = true;
    if (initialWorld) return undefined;
    currentIdentity()
      .then((identity) => {
        if (!active) return;
        const loadedWorld = identity.world;
        if (!loadedWorld?.serverTime) throw new Error("World data is unavailable. Please restart the game server and try again.");
        setWorld(loadedWorld);
        setServerTime(new Date(loadedWorld.serverTime));
      })
      .catch((requestError) => {
        if (active) setError(requestError.message);
      });
    return () => { active = false; };
  }, [initialWorld]);

  useEffect(() => {
    if (!world) return undefined;
    const timer = window.setInterval(() => setServerTime((time) => new Date(time.getTime() + 1000)), 1000);
    return () => window.clearInterval(timer);
  }, [world]);

  const display = world && serverTime ? formatServerDateTime(serverTime, world.timeZone) : null;
  const location = world ? `${world.city}, ${world.country}` : "";

  return (
    <main className="game-screen pos-relative min-h-screen overflow-hidden bg-atmosphere text-primary" aria-label="Main game screen">
      <div className="game-backdrop pos-absolute inset-0" data-location={world?.locationId ?? "lyon"} aria-hidden="true" />
      <div className="game-vignette pos-absolute inset-0" aria-hidden="true" />
      <header className="game-hud pos-relative flex items-start justify-between gap-4 p-6">
        <div>
          <p className="text-xs tracking-widest uppercase text-accent m-0">Current location</p>
          <p className="text-lg tracking-wide uppercase m-0 game-location__name">{location}</p>
        </div>
        {display && (
          <aside className="computer-launch-group" aria-label="Local date and time">
            <button
              className="computer-launch"
              type="button"
              aria-label="Open computer desktop"
              aria-haspopup="dialog"
              aria-expanded={computerOpen}
              onClick={() => { setNewspaperOpen(false); setComputerOpen(true); }}
            >
              <ComputerSharp className="computer-launch__icon" aria-hidden="true" />
            </button>
            <button
              className="newspaper-launch"
              type="button"
              aria-label="Read the Lyon Gazette"
              aria-haspopup="dialog"
              aria-expanded={newspaperOpen}
              onClick={() => { setComputerOpen(false); setNewspaperOpen(true); }}
            >
              <NotebookSharp className="computer-launch__icon" aria-hidden="true" />
            </button>
            <div className="server-clock surface p-4 text-right">
              <p className="text-xs tracking-wide uppercase text-success m-0">Local time</p>
              <time className="server-clock__time block text-xl font-mono" dateTime={serverTime.toISOString()}>{display.time}</time>
              <p className="server-clock__date text-xs text-secondary m-0">{display.date}</p>
            </div>
          </aside>
        )}
      </header>
      {error && <p className="pos-relative m-6 text-sm text-danger" role="alert">{error}</p>}
      <footer className="game-location pos-absolute inset-x-0 bottom-0 flex items-end justify-end gap-4 p-6">
        <button className="border-0 bg-transparent text-accent underline" type="button" onClick={onLogout}>Log out</button>
      </footer>
      {computerOpen && (
        <ComputerDesktop
          location={location}
          display={display}
          serverTime={serverTime}
          onClose={closeComputer}
        />
      )}
      {newspaperOpen && <NewspaperWindow onClose={closeNewspaper} />}
    </main>
  );
}
