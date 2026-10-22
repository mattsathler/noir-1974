import { useEffect, useMemo, useRef, useState } from "react";
import { getGateDestinations } from "../api.js";

export default function GateCityExplorer({ world, onTravel, onClose }) {
  const [cities, setCities] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selectedCity, setSelectedCity] = useState(null);
  const closeRef = useRef(null);

  useEffect(() => {
    const previousFocus = document.activeElement;
    closeRef.current?.focus();
    function handleKeys(event) {
      if (event.key === "Escape") onClose?.();
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
    getGateDestinations().then((items) => {
      if (!active) return;
      setCities(items);
      setSelectedCity(items.find(({ id }) => id === world?.cityId) ?? null);
    }).catch((requestError) => {
      if (active) setError(requestError);
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [world?.cityId]);

  const countries = useMemo(() => {
    const grouped = new Map();
    for (const city of cities) {
      if (!grouped.has(city.countryId)) grouped.set(city.countryId, { id: city.countryId, name: city.country, cities: [] });
      grouped.get(city.countryId).cities.push(city);
    }
    return [...grouped.values()];
  }, [cities]);

  async function travel() {
    if (!selectedCity || selectedCity.id === world?.cityId) return;
    setBusy(true);
    setError("");
    try {
      await onTravel?.(selectedCity);
    } catch (requestError) {
      setError(requestError);
      setBusy(false);
    }
  }

  return (
    <div className="modal-overlay gate-interface-overlay" role="dialog" aria-modal="true" aria-label="GATE city transit" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>
      <section className="gate-interface" aria-labelledby="gate-city-title" aria-busy={busy}>
        <header className="gate-interface__header">
          <div className="gate-interface__identity">
            <span className="gate-interface__mark" aria-hidden="true">G</span>
            <div><p>MERIDIAN GLOBAL · GATE NETWORK</p><h2 id="gate-city-title">City transit</h2></div>
          </div>
          <div className="gate-interface__relay"><span /> {(world?.city ?? "CITY").toUpperCase()} RELAY ONLINE</div>
          <button ref={closeRef} className="gate-interface__close" type="button" onClick={onClose} aria-label="Close GATE city transit">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 5l14 14M19 5L5 19" /></svg>
          </button>
        </header>
        <div className="gate-interface__body">
          <aside className="gate-interface__destinations" aria-label="City destinations">
            <div className="gate-interface__section-heading"><span>01 / COUNTRIES</span><small>{loading ? "SYNCING" : `${countries.length} COUNTRIES · ${cities.length} CITIES`}</small></div>
            {loading ? <p className="gate-interface__loading" role="status">Loading city registry<span>…</span></p> : <nav className="gate-interface__districts" aria-label="Cities grouped by country">
              {countries.map((country) => <section className="gate-interface__country" key={country.id} aria-label={country.name}>
                <h3 className="gate-interface__caption">{country.name}</h3>
                {country.cities.map((city) => <button className="gate-interface__district" key={city.id} type="button" aria-pressed={city.id === selectedCity?.id} aria-current={city.id === world?.cityId ? "location" : undefined} disabled={busy} onClick={() => setSelectedCity(city)}>
                  <span>{city.name}</span><small>{city.id === world?.cityId ? "CURRENT CITY" : "SELECT CITY"}</small>
                </button>)}
              </section>)}
            </nav>}
          </aside>
          <main className="gate-interface__preview">
            <div className="gate-interface__section-heading"><span>02 / TRANSIT FILE</span><small>{selectedCity ? "DESTINATION VERIFIED" : "AWAITING INPUT"}</small></div>
            {error && <div className="gate-interface__error" role="alert"><p>{error.message}</p>{error.requestId && <small>REFERENCE {error.requestId}{error.code ? ` · ${error.code}` : ""}</small>}{error.details && <small>{error.details}</small>}</div>}
            {selectedCity ? <section aria-label={`Selected city ${selectedCity.name}`}>
              <h3 className="gate-interface__district-name">{selectedCity.name}</h3>
              <p className="gate-interface__caption">{selectedCity.country.toUpperCase()}</p>
              <p className="gate-interface__empty">GATE transit connects cities. Neighborhood access and building visits are handled through local interfaces.</p>
              <dl className="gate-interface__city-record"><div><dt>ORIGIN</dt><dd>{world?.city}, {world?.country}</dd></div><div><dt>DESTINATION</dt><dd>{selectedCity.name}, {selectedCity.country}</dd></div><div><dt>RELAY TIME ZONE</dt><dd>{selectedCity.timeZone}</dd></div></dl>
            </section> : <div className="gate-interface__empty-state"><span aria-hidden="true">⌖</span><p>Select a city from the Meridian Global network.</p></div>}
          </main>
        </div>
        <footer className="gate-interface__footer">
          <p><span>CITY ROUTE</span>{selectedCity ? `${world?.city?.toUpperCase()}  →  ${selectedCity.name.toUpperCase()}` : "SELECT A CITY"}</p>
          <button className="gate-interface__travel" type="button" disabled={!selectedCity || selectedCity.id === world?.cityId || busy || loading} onClick={travel}>{busy ? "CONNECTING…" : "TRAVEL BY GATE"}<span aria-hidden="true">→</span></button>
        </footer>
      </section>
    </div>
  );
}
