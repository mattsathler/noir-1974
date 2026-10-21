import { useEffect, useRef, useState } from "react";
import { getGateDestinations } from "../api.js";

export default function NeighborhoodExplorer({ world, onTravel, onVisit, onClose }) {
  const [destinations, setDestinations] = useState([]);
  const [loadingDestinations, setLoadingDestinations] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selectedDestination, setSelectedDestination] = useState(null);
  const closeRef = useRef(null);

  useEffect(() => {
    const previousFocus = document.activeElement;
    closeRef.current?.focus();
    function handleKeys(event) {
      if (event.key === "Escape") {
        onClose?.();
        return;
      }
      if (event.key !== "Tab") return;
      const focusable = event.currentTarget.querySelectorAll("button:not(:disabled)");
      if (!focusable.length) return;
      if (event.shiftKey && document.activeElement === focusable[0]) {
        event.preventDefault();
        focusable[focusable.length - 1].focus();
      } else if (!event.shiftKey && document.activeElement === focusable[focusable.length - 1]) {
        event.preventDefault();
        focusable[0].focus();
      }
    }
    const dialog = closeRef.current?.closest("[role='dialog']");
    dialog?.addEventListener("keydown", handleKeys);
    return () => {
      dialog?.removeEventListener("keydown", handleKeys);
      if (typeof previousFocus?.focus === "function") previousFocus.focus();
    };
  }, [onClose]);

  useEffect(() => {
    let active = true;
    getGateDestinations().then((items) => { if (active) setDestinations(items); })
      .catch((requestError) => { if (active) setError(requestError.message); })
      .finally(() => { if (active) setLoadingDestinations(false); });
    return () => { active = false; };
  }, []);

  async function travelTo(district) {
    setBusy(true);
    setError("");
    try {
      await onTravel(district);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setBusy(false);
    }
  }

  async function visitPlace(place) {
    setBusy(true);
    setError("");
    try {
      await onVisit?.(place);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setBusy(false);
    }
  }

  const currentDistrict = destinations.find((district) => district.id === world?.districtId)
    ?? destinations.find((district) => district.name === world?.district);
  const isCurrentDistrict = Boolean(selectedDestination && currentDistrict && selectedDestination.id === currentDistrict.id);
  const buildings = isCurrentDistrict
    ? selectedDestination?.locations.filter(({ address }) => address.kind === "building") ?? []
    : [];
  const returningToDistrict = isCurrentDistrict && world?.locationId !== currentDistrict?.id;

  return (
    <div className="modal-overlay gate-interface-overlay" role="dialog" aria-modal="true" aria-label="GATE neighborhood transit" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>
      <section className="gate-interface" aria-labelledby="gate-interface-title" aria-busy={busy}>
        <header className="gate-interface__header">
          <div className="gate-interface__identity">
            <span className="gate-interface__mark" aria-hidden="true">G</span>
            <div><p>GLOBAL ACCESS TRANSIT EXCHANGE</p><h2 id="gate-interface-title">Neighborhood transit</h2></div>
          </div>
          <div className="gate-interface__relay"><span /> LYON RELAY ONLINE</div>
          <button ref={closeRef} className="gate-interface__close" type="button" onClick={onClose} aria-label="Close GATE transit">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 5l14 14M19 5L5 19" /></svg>
          </button>
        </header>
        <div className="gate-interface__body">
          <aside className="gate-interface__destinations" aria-label="GATE neighborhood destinations">
            <div className="gate-interface__section-heading"><span>01 / DESTINATIONS</span><small>{loadingDestinations ? "SYNCING" : `${destinations.length} DISTRICTS`}</small></div>
            {loadingDestinations ? <p className="gate-interface__loading" role="status" aria-live="polite">Loading destinations<span>…</span></p> : <nav className="gate-interface__districts" aria-label="Available districts">
              {destinations.map((district) => <button className="gate-interface__district" key={district.id} type="button" aria-pressed={district.id === selectedDestination?.id} aria-current={district.id === currentDistrict?.id ? "location" : undefined} disabled={busy} onClick={() => setSelectedDestination(district)}><span>{district.name}</span><small>{district.id === currentDistrict?.id ? "CURRENT DISTRICT" : "SELECT DISTRICT"}</small></button>)}
            </nav>}
          </aside>
          <main className="gate-interface__preview">
            <div className="gate-interface__section-heading"><span>02 / DISTRICT PREVIEW</span><small>{selectedDestination ? "ROUTE DATA" : "AWAITING INPUT"}</small></div>
            {error && <p className="gate-interface__error" role="alert">{error}</p>}
            {selectedDestination ? <section aria-label={`Buildings in ${selectedDestination.name}`}>
              <h3 className="gate-interface__district-name">{selectedDestination.name}</h3>
              <p className="gate-interface__caption">BUILDINGS IN THIS DISTRICT</p>
              {!isCurrentDistrict ? <p className="gate-interface__empty">Building records are available after arriving in this district.</p> : buildings.length ? <div className="gate-interface__places">
                {buildings.map(({ address }) => <div className="gate-interface__place" key={address.locationId}>
                  <span>{address.name}</span>
                  {world?.locationId === address.locationId
                    ? <small>CURRENT LOCATION</small>
                    : <button className="gate-interface__visit" type="button" disabled={busy} onClick={() => visitPlace(address)}>ENTER</button>}
                </div>)}
              </div> : <p className="gate-interface__empty">No buildings listed in this district.</p>}
            </section> : <div className="gate-interface__empty-state"><span aria-hidden="true">⌖</span><p>Select a district to inspect its buildings and prepare a route.</p></div>}
          </main>
        </div>
        <footer className="gate-interface__footer">
          <p><span>{returningToDistrict ? "LOCATION" : "ROUTE"}</span>{returningToDistrict ? `RETURN TO ${currentDistrict.name.toUpperCase()}` : selectedDestination ? `CENTRAL RELAY  →  ${selectedDestination.name.toUpperCase()}` : "SELECT A DESTINATION"}</p>
          {returningToDistrict
            ? <button className="gate-interface__travel" type="button" disabled={busy} onClick={() => visitPlace({ locationId: currentDistrict.id })}>{busy ? "CONNECTING…" : "RETURN TO DISTRICT"}<span aria-hidden="true">↗</span></button>
            : <button className="gate-interface__travel" type="button" disabled={!selectedDestination || selectedDestination.id === currentDistrict?.id || busy || loadingDestinations} onClick={() => travelTo(selectedDestination)}>{busy ? "CONNECTING…" : "TRAVEL"}<span aria-hidden="true">→</span></button>}
        </footer>
      </section>
    </div>
  );
}
