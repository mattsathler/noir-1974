import { useCallback, useEffect, useState } from "react";
import { Cross } from "pixelarticons/react/Cross";
import { MonitorSharp } from "pixelarticons/react/MonitorSharp";
import { Earth } from "pixelarticons/react/Earth";
import { NotebookSharp } from "pixelarticons/react/NotebookSharp";
import { Phone } from "pixelarticons/react/Phone";
import moneyIcon from "pixelarticons/svg/money.svg";
import starIcon from "pixelarticons/svg/star.svg";
import buildingIcon from "pixelarticons/svg/building.svg";
import { currentIdentity } from "../../../identity/api.js";
import NewspaperWindow from "../../../world/pages/newspaper/NewspaperWindow.jsx";
import ComputerDesktop from "./ComputerDesktop.jsx";
import GateCityExplorer from "../../../world/components/GateCityExplorer.jsx";
import { getCompanyAtLocation, getCurrentWorld, getNeighborhoodDestinations, movePlayer } from "../../../world/api.js";
import GateTransition from "../../components/GateTransition.jsx";
import LocalTerminal from "./LocalTerminal.jsx";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import Telephone from "./Telephone.jsx";

function buildingVisual(building) {
  const identity = `${building.locationId} ${building.name}`.toLowerCase();
  if (identity.includes("hospital")) return { Icon: Cross, label: "HOSPITAL" };
  if (identity.includes("police")) return { icon: starIcon, label: "POLICE STATION" };
  if (identity.includes("bank")) return { icon: moneyIcon, label: "BANK" };
  if (identity.includes("gate")) return { Icon: Earth, label: "GATE CENTRAL" };
  return { icon: buildingIcon, label: "LOCAL BUILDING" };
}

function interiorProfile(place) {
  const identity = `${place.locationId} ${place.name}`.toLowerCase();
  if (identity.includes("police")) return { type: "police", label: "CORPORATE SECURITY OFFICE", room: "SECURITY DESK", detail: "A contracted security office serving this city.", facility: "police" };
  if (identity.includes("bank")) return { type: "bank", label: "MERIDIAN FINANCIAL SERVICES", room: "CUSTOMER HALL", detail: "Accounts and transactions are handled under corporate service terms.", facility: "bank" };
  if (identity.includes("hospital")) return { type: "hospital", label: "CITY HOSPITAL", room: "EMERGENCY WARD", detail: "A public medical facility serving the surrounding neighborhood." };
  if (identity.includes("gate")) return { type: "gate", label: "MERIDIAN GLOBAL · GATE NETWORK", room: "TRANSIT CONCOURSE", detail: "Passenger access is monitored by the local GATE relay.", facility: "gate" };
  if (identity.includes("home") || identity.includes("residence") || identity.includes("apartment")) return { label: "PRIVATE RESIDENCE", room: "ENTRY ROOM", detail: "A private interior at the selected address." };
  return { type: "building", label: "LOCAL COMMERCIAL PREMISES", room: "RECEPTION AREA", detail: "A privately operated business at the selected address.", facility: "company" };
}

async function neighborhoodCompanies(district) {
  const addresses = (district?.locations ?? []).map(({ address }) => address).filter((address) => address?.kind === "building");
  const companyAddresses = addresses.filter((address) => !address.locationId?.startsWith("service:") && address.status !== "closed");
  const companyRecords = await Promise.all(companyAddresses.map((address) => getCompanyAtLocation(address.locationId)));
  const seenCompanies = new Set();
  const companies = companyRecords.flat().flatMap((record) => {
    if (!record?.company?.id || seenCompanies.has(record.company.id)) return [];
    seenCompanies.add(record.company.id);
    return [{ kind: "company", company: record.company, address: record.address, record }];
  });
  const services = addresses.filter((address) => address.locationId?.startsWith("service:")).map((address) => ({
    kind: "service",
    company: { id: address.locationId, name: address.name },
    address,
    serviceKind: interiorProfile(address).facility,
    status: address.status,
  }));
  return [...companies, ...services].sort((first, second) => first.company.name.localeCompare(second.company.name));
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
  const [computerEntry, setComputerEntry] = useState(null);
  const [computerContext, setComputerContext] = useState(null);
  const [newspaperOpen, setNewspaperOpen] = useState(false);
  const [telephoneOpen, setTelephoneOpen] = useState(false);
  const [transitOpen, setTransitOpen] = useState(false);
  const [gateTransition, setGateTransition] = useState(null);
  const [arrival, setArrival] = useState(null);
  const [selectedCompany, setSelectedCompany] = useState(null);
  const [companyDirectoryLoading, setCompanyDirectoryLoading] = useState(false);
  const interiorPlace = selectedCompany?.address ?? null;
  const closeComputer = useCallback(() => { setComputerOpen(false); setComputerContext(null); }, []);
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
    const companies = await neighborhoodCompanies(response.district);
    setArrival({ city: destination.city, district: response.district, companies });
    setSelectedCompany(null);
    closeTransit();
  }, [closeTransit]);
  const visitCompany = useCallback(async (company) => {
    setError("");
    setSelectedCompany(company);
    setCompanyDirectoryLoading(company.kind === "company");
    if (company.kind !== "company") return;
    try {
      const [record] = await getCompanyAtLocation(company.address.locationId, company.company.id);
      if (record) setSelectedCompany((current) => current?.company.id === company.company.id ? { ...current, record } : current);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setCompanyDirectoryLoading(false);
    }
  }, []);
  const returnToDistrict = useCallback(() => {
    setSelectedCompany(null);
  }, []);
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

  useEffect(() => {
    if (!world || arrival || !world.cityId) return undefined;
    let active = true;
    const districtId = world.districtId || `${world.cityId}/district/central`;
    getNeighborhoodDestinations(world.cityId).then((districts) => {
      if (!active) return;
      const district = districts.find((item) => item.id === districtId)
        ?? districts.find((item) => item.name.toLowerCase() === "central");
      if (!district) return;
      return neighborhoodCompanies(district).then((companies) => {
        if (!active) return;
        setArrival({
          city: { id: world.cityId, name: world.city, country: world.country },
          district,
          companies,
        });
      });
    }).catch((requestError) => {
      if (active) setError(requestError.message);
    });
    return () => { active = false; };
  }, [world?.cityId, world?.districtId, world?.city, world?.country, arrival]);

  const display = world && serverTime ? formatServerDateTime(serverTime, world.timeZone) : null;
  const issueDay = localDateKey(serverTime, world?.timeZone);
  const location = world ? [world.city, world.country].filter(Boolean).join(", ") : "";
  const cityBackdrop = world?.cityId === "fr/lyon" ? 'url("/images/cities/lyon.png")' : "none";
  const activeInteriorProfile = interiorPlace ? interiorProfile(interiorPlace) : null;

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
              aria-expanded={computerOpen || Boolean(computerEntry)}
              onClick={() => { setNewspaperOpen(false); setComputerEntry(null); setComputerContext({ kind: "personal", label: "NOIR PERSONAL COMPUTER", city: world?.city }); setComputerOpen(true); }}
            >
              <MonitorSharp className="computer-launch__icon" aria-hidden="true" />
            </button>
            <button
              className="computer-launch"
              type="button"
              aria-label="Open telephone"
              aria-haspopup="dialog"
              aria-expanded={telephoneOpen}
              onClick={() => setTelephoneOpen(true)}
            >
              <Phone className="computer-launch__icon" aria-hidden="true" />
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
          <div><p>NEIGHBORHOOD DIRECTORY · {arrival.city.name.toUpperCase()}, {arrival.city.country.toUpperCase()}</p><h1 id="district-arrival-title">{arrival.district?.name}</h1></div>
          <span>{arrival.companies.length} COMPANIES &amp; SERVICES</span>
        </header>
        <div className="district-arrival__cards">
          {arrival.companies.length === 0 && <p className="district-arrival__empty">NO COMPANIES OR PUBLIC SERVICES ARE LISTED IN THIS NEIGHBORHOOD.</p>}
          {arrival.companies.map((company) => {
            const { Icon, icon, label } = company.kind === "service"
              ? buildingVisual(company.address)
              : { icon: buildingIcon, label: "COMPANY" };
            return <article className="district-arrival__card" key={company.company.id}>
              <div className="district-arrival__icon" aria-hidden="true">
                {icon ? <span className="district-arrival__icon-image" style={{ "--building-icon-image": `url("${icon}")` }} /> : <Icon />}
                <span>{label}</span>
              </div>
              <p>{company.status === "closed" ? "PUBLIC SERVICE · CLOSED" : company.kind === "service" ? "PUBLIC SERVICE" : "LOCAL COMPANY"}</p>
              <h2>{company.company.name}</h2>
              {company.status === "closed"
                ? <span className="district-arrival__closed">CLOSED · FUTURE SERVICE</span>
                : <button type="button" onClick={() => visitCompany(company)}>ENTER <span aria-hidden="true">↗</span></button>}
            </article>;
          })}
        </div>
      </section>}
      {interiorPlace && selectedCompany && <section className="building-interior building-interior--company" aria-labelledby="building-interior-title">
        <header className="building-interior__header">
          <div><p>{selectedCompany.kind === "service" ? "PUBLIC SERVICE ACCESS" : "COMPANY WORKSPACE"} · {world?.city?.toUpperCase()}</p><h1 id="building-interior-title">{selectedCompany.company.name}</h1><span>{interiorPlace.full}</span></div>
          <button type="button" onClick={returnToDistrict}>RETURN TO NEIGHBORHOOD</button>
        </header>
        <div className="building-interior__layout">
          <div className="building-interior__room" aria-label={`${selectedCompany.company.name} personnel`}>
            <div className="building-interior__personnel-heading"><div><span className="building-interior__room-label">PERSONNEL ON SITE</span><h2>Staff directory</h2></div><span>{companyDirectoryLoading ? "RETRIEVING FILES…" : `${selectedCompany.record?.employees?.length ?? 0} EMPLOYEES`}</span></div>
            {companyDirectoryLoading
              ? <p className="building-interior__personnel-status">ACCESSING COMPANY PERSONNEL FILES…</p>
              : selectedCompany.kind === "company" && (selectedCompany.record?.employees?.length ?? 0) > 0
                ? <div className="building-interior__personnel-list">
                  {selectedCompany.record.employees.map((employee) => <article className="building-interior__employee" key={employee.id}>
                    <div className="building-interior__employee-portrait">
                      {employee.visual
                        ? <ModularFace visual={employee.visual} name={employee.name} />
                        : <span aria-hidden="true">{employee.name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase()}</span>}
                    </div>
                    <div className="building-interior__employee-copy"><h3>{employee.name}</h3><p>{employee.role}</p><span>INTERROGATION SUBJECT</span></div>
                  </article>)}
                </div>
                : <p className="building-interior__personnel-status">NO PERSONNEL FILES AVAILABLE AT THIS LOCATION.</p>}
          </div>
          <aside className="building-interior__record">
            <p>{selectedCompany.kind === "service" ? "PUBLIC SERVICE RECORD" : "COMPANY RECORD"}</p><h2>{selectedCompany.company.name}</h2><span>{activeInteriorProfile?.detail ?? "A private company operating at this address."}</span>
            <dl><div><dt>BUILDING</dt><dd>{interiorPlace.name}</dd></div><div><dt>ADDRESS</dt><dd>{interiorPlace.full}</dd></div><div><dt>ACCESS</dt><dd>VISIT CONTEXT · NOT SAVED</dd></div></dl>
            <button className="building-interior__computer-access" type="button" onClick={() => { setComputerContext({ kind: selectedCompany.kind === "service" ? selectedCompany.serviceKind : "company", label: selectedCompany.company.name, city: world?.city, locationId: interiorPlace.locationId, companyId: selectedCompany.kind === "company" ? selectedCompany.company.id : undefined, district: interiorPlace.district }); setComputerEntry(null); setComputerOpen(true); }}>
              <span className="building-interior__option-icon"><MonitorSharp className="computer-program__glyph" aria-hidden="true" /></span>
              <span><strong>{selectedCompany.kind === "service" ? `OPEN ${activeInteriorProfile?.facility?.toUpperCase() ?? "LOCAL"} TERMINAL` : "ACCESS COMPANY COMPUTER"}</strong><small>ACCESS LOCAL SYSTEMS</small></span>
              <span aria-hidden="true">↗</span>
            </button>
          </aside>
        </div>
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
          facilityContext={computerContext}
          onClose={closeComputer}
        />
      )}
      {computerEntry && <LocalTerminal facility={computerEntry} onAccessGranted={() => { setComputerContext(computerEntry); setComputerEntry(null); setComputerOpen(true); }} onClose={() => setComputerEntry(null)} />}
      {newspaperOpen && <NewspaperWindow cityId={world?.cityId} issueDay={issueDay} onClose={closeNewspaper} />}
      {telephoneOpen && <Telephone onClose={() => setTelephoneOpen(false)} />}
      {transitOpen && <GateCityExplorer world={world} onTravel={travelViaGate} onClose={closeTransit} />}
      {gateTransition && <GateTransition destination={gateTransition} origin={world} onArrive={arriveAtCity} onComplete={completeGateTransition} onError={failGateTransition} />}
    </main>
  );
}
