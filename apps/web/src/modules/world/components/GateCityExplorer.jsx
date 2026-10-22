import { useEffect, useMemo, useRef, useState } from "react";
import { getGateDestinations, getNeighborhoodDestinations } from "../api.js";

export default function GateCityExplorer({ world, onTravel, onClose }) {
  const [cities, setCities] = useState([]);
  const [districts, setDistricts] = useState([]);
  const [countryId, setCountryId] = useState("");
  const [cityId, setCityId] = useState("");
  const [districtId, setDistrictId] = useState("");
  const [districtSearch, setDistrictSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadingDistricts, setLoadingDistricts] = useState(false);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const closeRef = useRef(null);
  const districtRequest = useRef(0);

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
      const currentCity = items.find(({ id }) => id === world?.cityId);
      if (!currentCity) return;
      setCountryId(currentCity.countryId);
      setCityId(currentCity.id);
      setLoadingDistricts(true);
      getNeighborhoodDestinations(currentCity.id).then((destinations) => {
        if (active) setDistricts(destinations);
      }).catch((requestError) => {
        if (active) setError(requestError);
      }).finally(() => {
        if (active) setLoadingDistricts(false);
      });
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
  const selectedCountry = countries.find((country) => country.id === countryId);
  const availableCities = selectedCountry?.cities ?? [];
  const selectedCity = cities.find((city) => city.id === cityId);
  const selectedDistrict = districts.find((district) => district.id === districtId);
  const filteredDistricts = useMemo(() => {
    const query = districtSearch.trim().normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLocaleLowerCase();
    return query ? districts.filter(({ name }) => name.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLocaleLowerCase().includes(query)) : districts;
  }, [districts, districtSearch]);

  async function selectCity(city) {
    const request = ++districtRequest.current;
    setCityId(city.id);
    setDistrictId("");
    setDistrictSearch("");
    setDistricts([]);
    setError(null);
    setLoadingDistricts(true);
    try {
      const items = await getNeighborhoodDestinations(city.id);
      if (request === districtRequest.current) setDistricts(items);
    } catch (requestError) {
      if (request === districtRequest.current) setError(requestError);
    } finally {
      if (request === districtRequest.current) setLoadingDistricts(false);
    }
  }

  async function travel() {
    if (!selectedCity || !selectedDistrict) return;
    setBusy(true);
    setError(null);
    try {
      await onTravel?.({ city: selectedCity, district: selectedDistrict });
    } catch (requestError) {
      setError(requestError);
      setBusy(false);
    }
  }

  return (
    <div className="modal-overlay gate-interface-overlay" role="dialog" aria-modal="true" aria-label="GATE travel selection" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>
      <section className="gate-interface" aria-labelledby="gate-city-title" aria-busy={busy}>
        <header className="gate-interface__header">
          <div className="gate-interface__identity">
            <span className="gate-interface__mark" aria-hidden="true">G</span>
            <div><p>MERIDIAN GLOBAL · GATE NETWORK</p><h2 id="gate-city-title">Travel registry</h2></div>
          </div>
          <div className="gate-interface__relay"><span /> {(world?.city ?? "CITY").toUpperCase()} RELAY ONLINE</div>
          <button ref={closeRef} className="gate-interface__close" type="button" onClick={onClose} aria-label="Close GATE travel selection">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 5l14 14M19 5L5 19" /></svg>
          </button>
        </header>
        <div className="gate-interface__body gate-interface__body--gate-steps">
          <section className="gate-interface__column" aria-labelledby="gate-countries-heading">
            <div className="gate-interface__section-heading"><span id="gate-countries-heading">01 / COUNTRY</span><small>{loading ? "SYNCING" : `${countries.length} COUNTRIES`}</small></div>
            {loading ? <p className="gate-interface__loading" role="status">Loading countries<span>…</span></p> : <nav className="gate-interface__districts" aria-label="Countries">
              {countries.map((country) => <button className="gate-interface__district" key={country.id} type="button" aria-pressed={country.id === countryId} disabled={busy} onClick={() => { districtRequest.current += 1; setLoadingDistricts(false); setCountryId(country.id); setCityId(""); setDistrictId(""); setDistricts([]); setError(null); }}>
                <span>{country.name}</span><small>{country.cities.length} CITIES</small>
              </button>)}
            </nav>}
          </section>
          <section className="gate-interface__column" aria-labelledby="gate-cities-heading">
            <div className="gate-interface__section-heading"><span id="gate-cities-heading">02 / CITY</span><small>{selectedCountry ? selectedCountry.name.toUpperCase() : "SELECT A COUNTRY"}</small></div>
            {selectedCountry ? <nav className="gate-interface__districts" aria-label={`Cities in ${selectedCountry.name}`}>
              {availableCities.map((city) => <button className="gate-interface__district" key={city.id} type="button" aria-pressed={city.id === cityId} disabled={busy} onClick={() => selectCity(city)}>
                <span>{city.name}</span><small>{city.id === world?.cityId ? "CURRENT CITY" : "SELECT CITY"}</small>
              </button>)}
            </nav> : <p className="gate-interface__column-empty">Choose a country to view its cities.</p>}
          </section>
          <section className="gate-interface__column" aria-labelledby="gate-neighborhoods-heading">
            <div className="gate-interface__section-heading"><span id="gate-neighborhoods-heading">03 / NEIGHBORHOOD</span><small>{loadingDistricts ? "LOADING" : selectedCity ? `${districts.length} AVAILABLE` : "SELECT A CITY"}</small></div>
            {error && <div className="gate-interface__error" role="alert"><p>{error.message}</p>{error.requestId && <small>REFERENCE {error.requestId}{error.code ? ` · ${error.code}` : ""}</small>}{error.details && <small>{error.details}</small>}</div>}
            {loadingDistricts ? <p className="gate-interface__loading" role="status">Loading neighborhoods<span>…</span></p> : selectedCity ? <>
              <label className="gate-interface__search"><span>SEARCH NEIGHBORHOODS</span><input type="search" value={districtSearch} onChange={(event) => setDistrictSearch(event.target.value)} placeholder="Filter neighborhoods" /></label>
              {filteredDistricts.length ? <nav className="gate-interface__districts" aria-label={`Neighborhoods in ${selectedCity.name}`}>
              {filteredDistricts.map((district) => <button className="gate-interface__district" key={district.id} type="button" aria-pressed={district.id === districtId} disabled={busy} onClick={() => setDistrictId(district.id)}>
                <span>{district.name}</span><small>{district.id.endsWith("/district/central") ? "PUBLIC SERVICES" : "NEIGHBORHOOD"}</small>
              </button>)}
              </nav> : <p className="gate-interface__column-empty">No neighborhoods match that search.</p>}
            </> : <p className="gate-interface__column-empty">Choose a city to view its neighborhoods.</p>}
          </section>
        </div>
        <footer className="gate-interface__footer">
          <p><span>GATE ROUTE</span>{selectedCountry && selectedCity && selectedDistrict ? `${world?.city?.toUpperCase()}  →  ${selectedDistrict.name.toUpperCase()}, ${selectedCity.name.toUpperCase()}` : "SELECT COUNTRY · CITY · NEIGHBORHOOD"}</p>
          <button className="gate-interface__travel" type="button" disabled={!selectedCity || !selectedDistrict || busy || loading || loadingDistricts} onClick={travel}>{busy ? "CONNECTING…" : "TRAVEL BY GATE"}<span aria-hidden="true">→</span></button>
        </footer>
      </section>
    </div>
  );
}
