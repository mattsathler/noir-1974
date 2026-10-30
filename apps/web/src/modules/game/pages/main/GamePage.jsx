import { useCallback, useEffect, useState } from "react";
import { currentIdentity } from "../../../identity/api.js";
import ComputerDesktop from "./ComputerDesktop.jsx";
import { connectProxy, getCompanyWebsite, getCurrentWorld, requestFacilitySiteAccess } from "../../../world/api.js";
import Telephone from "./Telephone.jsx";
import APIErrorNotice from "../../../../shared/APIErrorNotice.jsx";
import { readDesktopPreferences, saveDesktopPreferences } from "./desktopPreferences.js";

function cityIdFromLocation(locationId = "") {
  const path = locationId.replace(/^service:/u, "").replace(/^location:/u, "");
  if (path.startsWith("lyon/")) return "fr/lyon";
  const [countryId, cityId] = path.split("/");
  return countryId && cityId ? `${countryId}/${cityId}` : "";
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
  const [error, setError] = useState(null);
  const [identityAttempt, setIdentityAttempt] = useState(0);
  const [accountUsername, setAccountUsername] = useState("");
  const [playerCharacter, setPlayerCharacter] = useState(null);
  const [desktopPreferences, setDesktopPreferences] = useState(readDesktopPreferences);
  const [computerOpen, setComputerOpen] = useState(false);
  const [computerContext, setComputerContext] = useState(null);
  const [companyAccessAttemptId, setCompanyAccessAttemptId] = useState("");
  const [telephoneOpen, setTelephoneOpen] = useState(false);
  const closeComputer = useCallback(() => { setComputerOpen(false); setComputerContext(null); }, []);

  useEffect(() => {
    let active = true;
    currentIdentity().then((identity) => {
      if (!active) return;
      setAccountUsername(identity.account?.username || "");
      setPlayerCharacter(identity.character ?? null);
    }).catch(() => {});
    return () => { active = false; };
  }, []);

  const connectToProxy = useCallback((response) => {
    setError(null);
    const nextWorld = response.world;
    setWorld(nextWorld);
    setServerTime(new Date(nextWorld.serverTime));
  }, []);
  const openCompanyWebsite = useCallback(async ({ company, address }) => {
    const cityId = cityIdFromLocation(address.locationId);
    setError(null);
    const response = await connectProxy(cityId);
    setWorld(response.world);
    setServerTime(new Date(response.world.serverTime));
    return getCompanyWebsite(company.id);
  }, []);
  const openCompanyStaffDirectory = useCallback(async (site) => {
    const company = site.company || site.Company;
    const address = site.address || site.Address;
    const response = await connectProxy(cityIdFromLocation(address.locationId));
    setWorld(response.world);
    setServerTime(new Date(response.world.serverTime));
    const attempt = await requestFacilitySiteAccess({
      kind: "company",
      companyId: company.id,
      locationId: address.locationId,
    });
    if (attempt.status !== "granted") throw new Error("Company directory access was denied.");
    return attempt;
  }, []);
  const openPublicServiceWebsite = useCallback(async ({ address, kind }) => {
    const cityId = cityIdFromLocation(address.locationId);
    const response = await connectProxy(cityId);
    setWorld(response.world);
    setServerTime(new Date(response.world.serverTime));
    const facility = { kind, label: address.name, city: address.city, locationId: address.locationId, district: address.district };
    const attempt = await requestFacilitySiteAccess(facility);
    if (attempt.status !== "granted") throw new Error("Institution website access was denied.");
    setCompanyAccessAttemptId(String(attempt.attemptId));
    setComputerContext(facility);
    setComputerOpen(true);
  }, []);

  useEffect(() => {
    let active = true;
    if (initialWorld) return undefined;
    setError(null);
    currentIdentity()
      .then((identity) => {
        if (!active) return;
        const loadedWorld = identity.world;
        if (!loadedWorld?.serverTime) throw new Error("World data is unavailable. Please restart the game server and try again.");
        setWorld(loadedWorld);
        setServerTime(new Date(loadedWorld.serverTime));
      })
      .catch((requestError) => {
        if (active) setError({ cause: requestError, retry: () => setIdentityAttempt((attempt) => attempt + 1) });
      });
    return () => { active = false; };
  }, [initialWorld, identityAttempt]);

  useEffect(() => {
    if (!world) return undefined;
    const timer = window.setInterval(() => setServerTime((time) => new Date(time.getTime() + 1000)), 1000);
    return () => window.clearInterval(timer);
  }, [world]);

  const display = world && serverTime ? formatServerDateTime(serverTime, world.timeZone) : null;
  const issueDay = localDateKey(serverTime, world?.timeZone);
  const location = world ? [world.city, world.country].filter(Boolean).join(", ") : "";
  const updateDesktopPreferences = useCallback((patch) => {
    setDesktopPreferences((current) => ({ ...current, ...patch }));
  }, []);

  useEffect(() => { saveDesktopPreferences(desktopPreferences); }, [desktopPreferences]);

  return (
    <main className="game-screen pos-relative min-h-screen overflow-hidden bg-atmosphere text-primary" aria-label="Detective computer desktop">
      <p className="sr-only" role="status" aria-live="polite" aria-atomic="true">Proxy jurisdiction: {location || "not connected"}</p>
      {error && <APIErrorNotice error={error.cause ?? error} onRetry={error.retry} retryLabel={error.retryLabel} className="game-api-error pos-relative m-6 text-sm text-danger" />}
      {world && (
        <ComputerDesktop
          location={location}
          display={display}
          serverTime={serverTime}
          timeZone={world?.timeZone}
          cityId={world?.cityId}
          countryName={world?.country}
          issueDay={issueDay}
          facilityContext={computerOpen ? computerContext : { kind: "personal", label: "MERID_OS PERSONAL COMPUTER", city: world?.city }}
          initialAccessAttemptId={companyAccessAttemptId}
          initialFacilityAccessGranted={Boolean(companyAccessAttemptId)}
          onClose={closeComputer}
          onOpenCompany={openCompanyWebsite}
          onOpenCompanyStaff={openCompanyStaffDirectory}
          onOpenService={openPublicServiceWebsite}
          persistent
          onProxyConnected={connectToProxy}
          onPrepareProxy={() => setTelephoneOpen(false)}
          onOpenTelephone={() => setTelephoneOpen(true)}
          onOpenWhiteboard={() => window.open("/whiteboard", "_blank", "noopener,noreferrer")}
          accountUsername={accountUsername}
          character={playerCharacter}
          onCharacterSaved={setPlayerCharacter}
          desktopPreferences={desktopPreferences}
          onDesktopPreferencesChange={updateDesktopPreferences}
          onLogout={onLogout}
        />
      )}
      {telephoneOpen && <Telephone onClose={() => setTelephoneOpen(false)} />}
    </main>
  );
}
