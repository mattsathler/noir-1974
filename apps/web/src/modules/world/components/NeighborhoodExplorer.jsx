import { useEffect, useRef, useState } from "react";
import { getGateDestinations } from "../api.js";

export default function NeighborhoodExplorer({ world, onTravel, onClose }) {
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

  const currentDistrict = destinations.find((district) => district.name === world?.district);
  const buildings = selectedDestination?.locations.filter(({ address }) => address.kind === "building") ?? [];

  return (
    <div className="computer-modal" role="dialog" aria-modal="true" aria-label="GATE neighborhood transit" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>
      <section className="computer-frame computer-frame--modal" aria-label="Neighborhood map">
        <div className="computer-bezel"><div className="computer-screen">
          <header className="computer-menubar"><p className="computer-brand">MERIDIAN GLOBAL RELAY <span className="computer-brand__divider">/</span> GATE TRANSIT</p><button ref={closeRef} className="computer-modal__close gate-transit__close" type="button" onClick={onClose} aria-label="Close GATE transit">×</button></header>
          <main className="computer-desktop__content p-6" aria-busy={busy}>
            <p className="text-xs tracking-widest uppercase text-accent">Lyon district relay</p>
            {error && <p role="alert" className="text-danger">{error}</p>}
            <section aria-label="GATE neighborhood destinations">
              <h3>GATE destinations</h3>
              {loadingDestinations ? <p className="gate-transit__loading" role="status" aria-live="polite">Loading destinations</p> : <div className="gate-transit__districts">
                {destinations.map((district) => <button className="gate-transit__district" key={district.id} type="button" aria-pressed={district.id === selectedDestination?.id} aria-current={district.id === currentDistrict?.id ? "location" : undefined} disabled={busy} onClick={() => setSelectedDestination(district)}><span>{district.name}</span><small>{district.id === currentDistrict?.id ? "CURRENT DISTRICT" : "SELECT DISTRICT"}</small></button>)}
              </div>}
            </section>
            {selectedDestination && <section aria-label={`Buildings in ${selectedDestination.name}`}>
              <h3>Buildings in {selectedDestination.name}</h3>
              {buildings.length ? <div className="gate-transit__places">
                {buildings.map(({ address }) => <div className="gate-transit__place" key={address.locationId}><span>{address.name}</span><small>{address.kind.toUpperCase()}</small></div>)}
              </div> : <p>No buildings listed in this district.</p>}
            </section>}
            <footer className="gate-transit__footer">
              <button className="gate-transit__travel" type="button" disabled={!selectedDestination || selectedDestination.name === world?.district || busy || loadingDestinations} onClick={() => travelTo(selectedDestination)}>{busy ? "Traveling…" : "Travel"}</button>
            </footer>
          </main>
        </div></div>
      </section>
    </div>
  );
}
