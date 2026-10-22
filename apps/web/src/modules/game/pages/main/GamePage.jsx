import { useCallback, useEffect, useState } from "react";
import { Cross } from "pixelarticons/react/Cross";
import { MonitorSharp } from "pixelarticons/react/MonitorSharp";
import { Earth } from "pixelarticons/react/Earth";
import { NotebookSharp } from "pixelarticons/react/NotebookSharp";
import moneyIcon from "pixelarticons/svg/money.svg";
import starIcon from "pixelarticons/svg/star.svg";
import buildingIcon from "pixelarticons/svg/building.svg";
import { currentIdentity } from "../../../identity/api.js";
import NewspaperWindow from "../../../world/pages/newspaper/NewspaperWindow.jsx";
import ComputerDesktop from "./ComputerDesktop.jsx";
import GateCityExplorer from "../../../world/components/GateCityExplorer.jsx";
import { getCurrentWorld, movePlayer } from "../../../world/api.js";
import GateTransition from "../../components/GateTransition.jsx";

function buildingVisual(building) {
  const identity = `${building.locationId} ${building.name}`.toLowerCase();
  if (identity.includes("hospital")) return { Icon: Cross, label: "HOSPITAL" };
  if (identity.includes("police")) return { icon: starIcon, label: "POLICE STATION" };
  if (identity.includes("bank")) return { icon: moneyIcon, label: "BANK" };
  if (identity.includes("gate")) return { Icon: Earth, label: "GATE CENTRAL" };
  return { icon: buildingIcon, label: "LOCAL BUILDING" };
}

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

function localDateKey(timestamp, timeZone) {
  if (!timestamp || !timeZone) return "";
  const parts = Object.fromEntries(new Intl.DateTimeFormat("en-US", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(timestamp).map(({ type, value }) => [type, value]));
  return `${parts.year}-${parts.month}-${parts.day}`;
}

export default function GamePage({ initialWorld = null, onLogout }) {
  const [world, setWorld] = useState(initialWorld);
  const [serverTime, setServerTime] = useState(initialWorld ? new Date(initialWorld.serverTime) : null);
  const [error, setError] = useState("");
  const [computerOpen, setComputerOpen] = useState(false);
  const [newspaperOpen, setNewspaperOpen] = useState(false);
  const [transitOpen, setTransitOpen] = useState(false);
  const [gateTransition, setGateTransition] = useState(null);
  const [arrival, setArrival] = useState(null);
  const [interiorPlace, setInteriorPlace] = useState(null);
  const closeComputer = useCallback(() => setComputerOpen(false), []);
  const closeNewspaper = useCallback(() => setNewspaperOpen(false), []);
  const closeTransit = useCallback(() => setTransitOpen(false), []);

  const travelViaGate = useCallback((destination) => {
    setError("");
    setGateTransition(destination);
  }, []);
  const arriveAtCity = useCallback(async (destination) => {
    const response = await movePlayer("/api/world/gate/travel", {
      cityId: destination.city.id,
      districtId: destination.district.id,
    });
    const nextWorld = response.world ?? await getCurrentWorld();
    setWorld(nextWorld);
    setServerTime(new Date(nextWorld.serverTime));
    setArrival({ city: destination.city, district: response.district, buildings: response.buildings ?? [] });
    setInteriorPlace(null);
    closeTransit();
  }, [closeTransit]);
  const visitPlace = useCallback((place) => setInteriorPlace(place), []);
  const returnToDistrict = useCallback(() => setInteriorPlace(null), []);
  const completeGateTransition = useCallback(() => setGateTransition(null), []);
  const failGateTransition = useCallback((requestError) => {
    setError([
      requestError.message,
      requestError.code && `CODE ${requestError.code}`,
      requestError.requestId && `REFERENCE ${requestError.requestId}`,
      requestError.details,
    ].filter(Boolean).join(" · "));
  }, []);

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
  const issueDay = localDateKey(serverTime, world?.timeZone);
  const location = world ? [world.city, world.country].filter(Boolean).join(", ") : "";
  const cityBackdrop = world?.cityId === "fr/lyon" ? 'url("/images/cities/lyon.png")' : "none";

  return (
    <main className="game-screen pos-relative min-h-screen overflow-hidden bg-atmosphere text-primary" aria-label={interiorPlace ? `${interiorPlace.name} interior` : "Main game screen"}>
      <div className="game-backdrop pos-absolute inset-0" data-location={world?.cityId ?? "fr/lyon"} style={{ "--city-backdrop": cityBackdrop }} aria-hidden="true" />
      <div className="game-vignette pos-absolute inset-0" aria-hidden="true" />
      {!interiorPlace && <header className="game-hud pos-relative flex items-start justify-between gap-4 p-6">
        <div className="game-hud__location">
          <button
            className="gate-transit-launch"
            type="button"
            aria-label="Open GATE travel"
            aria-haspopup="dialog"
            aria-expanded={transitOpen}
            onClick={() => { setComputerOpen(false); setNewspaperOpen(false); setTransitOpen(true); }}
          >
            <Earth className="computer-launch__icon" aria-hidden="true" />
          </button>
          <div className="game-hud__location-copy">
            <p className="text-xs tracking-widest uppercase text-accent m-0">Current location</p>
            <p className="text-lg tracking-wide uppercase m-0 game-location__name">{location}</p>
          </div>
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
              <MonitorSharp className="computer-launch__icon" aria-hidden="true" />
            </button>
            <button
              className="newspaper-launch"
              type="button"
              aria-label="Read the local newspaper"
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
      </header>}
      {error && <p className="pos-relative m-6 text-sm text-danger" role="alert">{error}</p>}
      {arrival && !interiorPlace && <section className="district-arrival" aria-labelledby="district-arrival-title">
        <header className="district-arrival__header">
          <div><p>ARRIVAL RECORD · {arrival.city.name.toUpperCase()}, {arrival.city.country.toUpperCase()}</p><h1 id="district-arrival-title">{arrival.district?.name}</h1></div>
          <span>{arrival.buildings.length} CONSTRUCTIONS</span>
        </header>
        <div className="district-arrival__cards">
          {arrival.buildings.map((building) => {
            const { Icon, icon, label } = buildingVisual(building);
            return <article className="district-arrival__card" key={building.locationId}>
              <div className="district-arrival__icon" aria-hidden="true">
                {icon ? <span className="district-arrival__icon-image" style={{ "--building-icon-image": `url("${icon}")` }} /> : <Icon />}
                <span>{label}</span>
              </div>
              <p>{building.status === "closed" ? "PUBLIC SERVICE · CLOSED" : "LOCAL CONSTRUCTION"}</p>
              <h2>{building.name}</h2>
              {building.status === "closed"
                ? <span className="district-arrival__closed">CLOSED · FUTURE SERVICE</span>
                : <button type="button" onClick={() => visitPlace(building)}>VISIT <span aria-hidden="true">↗</span></button>}
            </article>;
          })}
        </div>
      </section>}
      {interiorPlace && <section className="building-interior" aria-labelledby="building-interior-title">
        <header className="building-interior__header"><div><p>INTERIOR ACCESS · {world?.city?.toUpperCase()}</p><h1 id="building-interior-title">{interiorPlace.name}</h1></div><button type="button" onClick={returnToDistrict}>RETURN TO NEIGHBORHOOD</button></header>
        <div className="building-interior__room"><span>INTERIOR VIEW</span><p>{interiorPlace.name.toUpperCase()}</p></div>
      </section>}
      {!interiorPlace && <footer className="game-location pos-absolute inset-x-0 bottom-0 flex items-end justify-end gap-4 p-6">
        <button className="border-0 bg-transparent text-accent underline" type="button" onClick={onLogout}>Log out</button>
      </footer>}
      {computerOpen && (
        <ComputerDesktop
          location={location}
          display={display}
          serverTime={serverTime}
          timeZone={world?.timeZone}
          cityId={world?.cityId}
          issueDay={issueDay}
          onClose={closeComputer}
        />
      )}
      {newspaperOpen && <NewspaperWindow cityId={world?.cityId} issueDay={issueDay} onClose={closeNewspaper} />}
      {transitOpen && <GateCityExplorer world={world} onTravel={travelViaGate} onClose={closeTransit} />}
      {gateTransition && <GateTransition destination={gateTransition} origin={world} onArrive={arriveAtCity} onComplete={completeGateTransition} onError={failGateTransition} />}
    </main>
  );
}
