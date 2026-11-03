import { useCallback, useEffect, useRef, useState } from "react";
import { currentIdentity } from "../../../identity/api.js";
import ComputerDesktop from "./ComputerDesktop.jsx";
import { connectProxy, getCompanyWebsite, getCurrentWorld, requestFacilitySiteAccess } from "../../../world/api.js";
import APIErrorNotice from "../../../../shared/APIErrorNotice.jsx";
import ComputerAccessShell from "../../../identity/components/ComputerAccessShell.jsx";
import { readDesktopPreferences, saveDesktopPreferences } from "./desktopPreferences.js";
import { playRetroSound, readRetroSoundEnabled, setRetroSoundEnabled } from "./retroSoundEffects.js";

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
  const [playerProgression, setPlayerProgression] = useState(null);
  const [desktopPreferences, setDesktopPreferences] = useState(readDesktopPreferences);
  const [soundEffectsEnabled, setSoundEffectsEnabled] = useState(readRetroSoundEnabled);
  const [computerOpen, setComputerOpen] = useState(false);
  const [computerContext, setComputerContext] = useState(null);
  const [companyAccessAttemptId, setCompanyAccessAttemptId] = useState("");
  const [conversationCharacter, setConversationCharacter] = useState("");
  const [notifications, setNotifications] = useState([]);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [snackbar, setSnackbar] = useState(null);
  const notificationSequence = useRef(0);
  const snackbarTimer = useRef(null);
  const closeComputer = useCallback(() => { setComputerOpen(false); setComputerContext(null); }, []);
  const closeConversation = useCallback(() => setConversationCharacter(""), []);
  const closeNotifications = useCallback(() => setNotificationsOpen(false), []);
  const toggleNotifications = useCallback(() => setNotificationsOpen((open) => !open), []);
  const deleteNotification = useCallback((notificationId) => setNotifications((current) => current.filter((item) => item.id !== notificationId)), []);
  const notify = useCallback((notification) => {
    const id = `${Date.now()}-${notificationSequence.current++}`;
    const item = { ...notification, id, time: new Intl.DateTimeFormat("en-US", { hour: "numeric", minute: "2-digit" }).format(new Date()) };
    setNotifications((current) => [item, ...current].slice(0, 40));
    setSnackbar(item);
    playRetroSound("notification");
    window.clearTimeout(snackbarTimer.current);
    snackbarTimer.current = window.setTimeout(() => setSnackbar(null), 6000);
  }, []);

  useEffect(() => {
    function handlePointerDown(event) {
      if (event.button === 0 && event.isPrimary !== false && !event.target.closest?.("[data-silent-click-sound]")) playRetroSound("click");
    }
    function handleTyping(event) {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      const target = event.target;
      const acceptsText = target instanceof Element && (target.matches("input:not([type=button]):not([type=submit]):not([type=checkbox]):not([type=radio]), textarea, [contenteditable='true']") || target.isContentEditable);
      if (acceptsText && (event.key.length === 1 || ["Backspace", "Delete", "Enter", " "].includes(event.key))) playRetroSound("key");
    }
    document.addEventListener("pointerdown", handlePointerDown, true);
    document.addEventListener("keydown", handleTyping, true);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown, true);
      document.removeEventListener("keydown", handleTyping, true);
    };
  }, []);

  useEffect(() => {
    const messageSelector = '[role="alert"], [role="alertdialog"]';
    const dialogSelector = '[role="dialog"]';
    const lastMessages = new WeakMap();
    const notifiedDialogs = new WeakSet();
    const readMessage = (element) => {
      if (element.getAttribute("role") === "alertdialog") {
        const describedBy = element.getAttribute("aria-describedby");
        const description = describedBy && document.getElementById(describedBy);
        if (description) return description.textContent.trim();
      }
      const copy = element.cloneNode(true);
      copy.querySelectorAll("button, [role='button'], [aria-hidden='true']").forEach((control) => control.remove());
      return copy.textContent.trim();
    };
    document.querySelectorAll(messageSelector).forEach((element) => lastMessages.set(element, readMessage(element)));
    document.querySelectorAll(dialogSelector).forEach((element) => notifiedDialogs.add(element));
    const observer = new MutationObserver((records) => {
      const changedMessages = new Set();
      const newDialogs = new Set();
      const collect = (node) => {
        if (!(node instanceof Element)) return;
        if (node.matches(messageSelector)) changedMessages.add(node);
        node.querySelectorAll(messageSelector).forEach((element) => changedMessages.add(element));
        if (node.matches(dialogSelector)) newDialogs.add(node);
        node.querySelectorAll(dialogSelector).forEach((element) => newDialogs.add(element));
      };
      records.forEach((record) => {
        const target = record.target instanceof Element ? record.target : record.target.parentElement;
        const alertContainer = target?.closest(messageSelector);
        if (alertContainer) changedMessages.add(alertContainer);
        record.addedNodes.forEach(collect);
        if (record.type === "attributes") collect(record.target);
      });
      changedMessages.forEach((element) => {
        const message = readMessage(element);
        if (message && lastMessages.get(element) !== message && !element.closest(".stock-exchange")) playRetroSound("warning");
        lastMessages.set(element, message);
      });
      newDialogs.forEach((element) => {
        if (notifiedDialogs.has(element)) return;
        notifiedDialogs.add(element);
        if (element.hasAttribute("data-notification-sound-manual")) return;
        playRetroSound("notification");
      });
    });
    if (document.body) observer.observe(document.body, { subtree: true, childList: true, characterData: true, attributes: true, attributeFilter: ["role"] });
    return () => observer.disconnect();
  }, []);

  useEffect(() => setRetroSoundEnabled(soundEffectsEnabled), [soundEffectsEnabled]);

  useEffect(() => () => window.clearTimeout(snackbarTimer.current), []);

  useEffect(() => {
    function handleFixedNPCContact(event) {
      const characterId = event.detail?.characterId;
      if (characterId === "bill" || characterId === "crow") setConversationCharacter(characterId);
    }
    window.addEventListener("noir:fixed-npc-contact", handleFixedNPCContact);
    return () => window.removeEventListener("noir:fixed-npc-contact", handleFixedNPCContact);
  }, []);

  useEffect(() => {
    let active = true;
    currentIdentity().then((identity) => {
      if (!active) return;
      setAccountUsername(identity.account?.username || "");
      setPlayerCharacter(identity.character ?? null);
      setPlayerProgression(identity.progression ?? null);
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
    const onStatus = site.onStatus;
    const response = await connectProxy(cityIdFromLocation(address.locationId));
    setWorld(response.world);
    setServerTime(new Date(response.world.serverTime));
    let attempt = await requestFacilitySiteAccess({
      kind: "company",
      companyId: company.id,
      locationId: address.locationId,
    });
    if (attempt.status !== "granted") throw new Error("Company directory access was denied.");
    const inspectionDelay = Math.max(0, Number(attempt.inspectionDelayMs) || 0);
    if (inspectionDelay) {
      onStatus?.(`${attempt.securityProvider?.toUpperCase() || "CORPORATE SECURITY"} IS INSPECTING THE REMOTE REQUEST…`);
      await new Promise((resolve) => window.setTimeout(resolve, inspectionDelay));
    }
    let sessionRestored = false;
    if (attempt.sessionInterrupted) {
      sessionRestored = true;
      onStatus?.(`${attempt.securityProvider?.toUpperCase() || "CORPORATE SECURITY"} INTERRUPTED THE SESSION · RECONNECTING…`);
      await new Promise((resolve) => window.setTimeout(resolve, 900));
      attempt = await requestFacilitySiteAccess({
        kind: "company",
        companyId: company.id,
        locationId: address.locationId,
        reconnect: true,
      });
      if (attempt.status !== "granted") throw new Error("The company firewall closed the session. Retry the employee directory.");
      const reconnectDelay = Math.max(0, Number(attempt.inspectionDelayMs) || 0);
      if (reconnectDelay) {
        onStatus?.(`${attempt.securityProvider?.toUpperCase() || "CORPORATE SECURITY"} IS CHECKING THE RECOVERED SESSION…`);
        await new Promise((resolve) => window.setTimeout(resolve, reconnectDelay));
      }
    }
    if (sessionRestored) attempt.sessionRestored = true;
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

  if (!world && error) {
    return <ComputerAccessShell title="Workstation connection error" labelledBy="workstation-error-title">
      <section className="w-full max-w-xl m-auto text-primary" aria-labelledby="workstation-error-title">
        <h1 id="workstation-error-title" className="text-xl m-0 mb-4">Workstation unavailable</h1>
        <APIErrorNotice error={error.cause ?? error} onRetry={error.retry} retryLabel={error.retryLabel} className="text-sm text-danger" />
      </section>
    </ComputerAccessShell>;
  }

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
          onNotification={notify}
          notifications={notifications}
          notificationsOpen={notificationsOpen}
          onToggleNotifications={toggleNotifications}
          onCloseNotifications={closeNotifications}
          onDeleteNotification={deleteNotification}
          snackbar={snackbar}
          onDismissSnackbar={() => setSnackbar(null)}
          onOpenWhiteboard={() => window.open("/whiteboard", "_blank", "noopener,noreferrer")}
          accountUsername={accountUsername}
          character={playerCharacter}
          onCharacterSaved={setPlayerCharacter}
          playerProgression={playerProgression}
          onProgressionChange={setPlayerProgression}
          desktopPreferences={desktopPreferences}
          onDesktopPreferencesChange={updateDesktopPreferences}
          soundEffectsEnabled={soundEffectsEnabled}
          onSoundEffectsChange={setSoundEffectsEnabled}
          onLogout={onLogout}
          conversationCharacter={conversationCharacter}
          onCloseConversation={closeConversation}
        />
      )}
    </main>
  );
}
