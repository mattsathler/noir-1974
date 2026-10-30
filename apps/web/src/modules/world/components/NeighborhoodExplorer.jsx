import { useEffect, useRef, useState } from "react";
import { getNeighborhoodDestinations } from "../api.js";

export default function NeighborhoodExplorer({ world, onTravel, onVisit, onClose }) {
  const [destinations, setDestinations] = useState([]);
  const [loadingDestinations, setLoadingDestinations] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selectedDestination, setSelectedDestination] = useState(null);
  const [neighborhoodSearch, setNeighborhoodSearch] = useState("");
  const [buildingSearch, setBuildingSearch] = useState("");
  const [appliedNeighborhoodSearch, setAppliedNeighborhoodSearch] = useState("");
  const [appliedBuildingSearch, setAppliedBuildingSearch] = useState("");
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
      const focusable = event.currentTarget.querySelectorAll("button:not(:disabled), input:not(:disabled)");
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
    setLoadingDestinations(true);
    setDestinations([]);
    setError("");
    getNeighborhoodDestinations(world?.cityId).then((items) => { if (active) setDestinations(items); })
      .catch((requestError) => { if (active) setError(requestError.message); })
      .finally(() => { if (active) setLoadingDestinations(false); });
    return () => { active = false; };
  }, [world?.cityId]);

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
  const normalizeSearch = (value) => value.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLocaleLowerCase();
  const neighborhoodQuery = normalizeSearch(appliedNeighborhoodSearch.trim());
  const filteredDestinations = neighborhoodQuery
    ? destinations.filter(({ name }) => normalizeSearch(name).includes(neighborhoodQuery))
    : destinations;
  const buildingQuery = normalizeSearch(appliedBuildingSearch.trim());
  const filteredBuildings = buildingQuery
    ? buildings.filter(({ address }) => normalizeSearch(address.name).includes(buildingQuery))
    : buildings;
  const returningToDistrict = isCurrentDistrict && world?.locationId !== currentDistrict?.id;

  return (
    <div className="modal-overlay gate-interface-overlay" role="dialog" aria-modal="true" aria-label="Local neighborhood directory" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>
      <section className="gate-interface" aria-labelledby="gate-interface-title" aria-busy={busy}>
        <header className="gate-interface__header">
          <div className="gate-interface__identity">
            <span className="gate-interface__mark" aria-hidden="true">G</span>
            <div><p>LOCAL CITY DIRECTORY</p><h2 id="gate-interface-title">Neighborhood access</h2></div>
          </div>
          <div className="gate-interface__relay"><span /> {(world?.city ?? "CITY").toUpperCase()} LOCAL NETWORK</div>
          <button ref={closeRef} className="gate-interface__close" type="button" onClick={onClose} aria-label="Close neighborhood access">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 5l14 14M19 5L5 19" /></svg>
          </button>
        </header>
        <div className="gate-interface__body">
          <aside className="gate-interface__destinations" aria-label="Neighborhood destinations">
            <div className="gate-interface__section-heading"><span>01 / NEIGHBORHOODS</span><small>{loadingDestinations ? "SYNCING" : `${destinations.length} DISTRICTS`}</small></div>
            {loadingDestinations ? <p className="gate-interface__loading" role="status" aria-live="polite">Loading neighborhoods<span>…</span></p> : <>
              <label className="gate-interface__search"><span>SEARCH NEIGHBORHOODS · PRESS ENTER</span><input type="search" value={neighborhoodSearch} onChange={(event) => setNeighborhoodSearch(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); setAppliedNeighborhoodSearch(neighborhoodSearch); } }} placeholder="Type, then press Enter" aria-keyshortcuts="Enter" /></label>
              {filteredDestinations.length ? <nav className="gate-interface__districts" aria-label="Available neighborhoods">
                {filteredDestinations.map((district) => <button className="gate-interface__district" key={district.id} type="button" aria-pressed={district.id === selectedDestination?.id} aria-current={district.id === currentDistrict?.id ? "location" : undefined} disabled={busy} onClick={() => setSelectedDestination(district)}><span>{district.name}</span><small>{district.id === currentDistrict?.id ? "CURRENT DISTRICT" : "OPEN NEIGHBORHOOD"}</small></button>)}
              </nav> : <p className="gate-interface__empty">No neighborhoods match that search.</p>}
            </>}
          </aside>
          <main className="gate-interface__preview">
            <div className="gate-interface__section-heading"><span>02 / LOCAL RECORDS</span><small>{selectedDestination ? "NEIGHBORHOOD FILE" : "AWAITING INPUT"}</small></div>
            {error && <p className="gate-interface__error" role="alert">{error}</p>}
            {selectedDestination ? <section aria-label={`Buildings in ${selectedDestination.name}`}>
              <h3 className="gate-interface__district-name">{selectedDestination.name}</h3>
              <p className="gate-interface__caption">BUILDINGS IN THIS NEIGHBORHOOD</p>
              {!isCurrentDistrict ? <p className="gate-interface__empty">Building records are available after opening this neighborhood.</p> : buildings.length ? <>
                <label className="gate-interface__search"><span>SEARCH BUILDINGS · PRESS ENTER</span><input type="search" value={buildingSearch} onChange={(event) => setBuildingSearch(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); setAppliedBuildingSearch(buildingSearch); } }} placeholder="Type, then press Enter" aria-keyshortcuts="Enter" /></label>
                {filteredBuildings.length ? <div className="gate-interface__places">
                {filteredBuildings.map(({ address }) => <div className="gate-interface__place" key={address.locationId}>
                  <span>{address.name}</span>
                  {world?.locationId === address.locationId
                    ? <small>CURRENT LOCATION</small>
                    : <button className="gate-interface__visit" type="button" disabled={busy} onClick={() => visitPlace(address)}>ENTER</button>}
                </div>)}
                </div> : <p className="gate-interface__empty">No buildings match that search.</p>}
              </> : <p className="gate-interface__empty">No buildings listed in this district.</p>}
            </section> : <div className="gate-interface__empty-state"><span aria-hidden="true">⌖</span><p>Select a neighborhood to inspect its local buildings.</p></div>}
          </main>
        </div>
        <footer className="gate-interface__footer">
          <p><span>{returningToDistrict ? "LOCATION" : "LOCAL ACCESS"}</span>{returningToDistrict ? `RETURN TO ${currentDistrict.name.toUpperCase()}` : selectedDestination ? `CURRENT NEIGHBORHOOD  →  ${selectedDestination.name.toUpperCase()}` : "SELECT A NEIGHBORHOOD"}</p>
          {returningToDistrict
            ? <button className="gate-interface__travel" type="button" disabled={busy} onClick={() => visitPlace({ locationId: currentDistrict.id })}>{busy ? "CONNECTING…" : "RETURN TO DISTRICT"}<span aria-hidden="true">↗</span></button>
            : <button className="gate-interface__travel" type="button" disabled={!selectedDestination || selectedDestination.id === currentDistrict?.id || busy || loadingDestinations} onClick={() => travelTo(selectedDestination)}>{busy ? "CONNECTING…" : "TRAVEL"}<span aria-hidden="true">→</span></button>}
        </footer>
      </section>
    </div>
  );
}
