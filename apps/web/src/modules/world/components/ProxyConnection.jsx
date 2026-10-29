import { useEffect, useMemo, useRef, useState } from "react";
import { connectProxy, getProxyNodes } from "../api.js";
import APIErrorNotice from "../../../shared/APIErrorNotice.jsx";

export default function ProxyConnection({ world, onConnected, onClose }) {
  const [nodes, setNodes] = useState([]);
  const [countryId, setCountryId] = useState("");
  const [cityId, setCityId] = useState(world?.cityId ?? "");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(null);
  const [retry, setRetry] = useState(0);
  const closeRef = useRef(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    getProxyNodes().then((items) => {
      if (!active) return;
      setNodes(items);
      const current = items.find(({ id }) => id === world?.cityId);
      setCountryId(current?.countryId ?? items[0]?.countryId ?? "");
      if (!items.some(({ id }) => id === cityId)) setCityId(current?.id ?? "");
    }).catch((requestError) => {
      if (active) setError({ cause: requestError, retry: () => setRetry((value) => value + 1) });
    }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [world?.cityId, retry]);

  useEffect(() => {
    const previousFocus = document.activeElement;
    closeRef.current?.focus();
    function handleKeys(event) {
      if (event.key === "Escape" && !busy) onClose?.();
    }
    window.addEventListener("keydown", handleKeys);
    return () => {
      window.removeEventListener("keydown", handleKeys);
      if (typeof previousFocus?.focus === "function") previousFocus.focus();
    };
  }, [busy, onClose]);

  const countries = useMemo(() => {
    const grouped = new Map();
    for (const node of nodes) {
      if (!grouped.has(node.countryId)) grouped.set(node.countryId, { id: node.countryId, name: node.country, cities: [] });
      grouped.get(node.countryId).cities.push(node);
    }
    return [...grouped.values()];
  }, [nodes]);
  const selectedCountry = countries.find(({ id }) => id === countryId);
  const selectedNode = nodes.find(({ id }) => id === cityId);

  async function connect() {
    if (!selectedNode || busy) return;
    setBusy(true);
    setError(null);
    try {
      const response = await connectProxy(selectedNode.id);
      await onConnected?.(response);
      onClose?.();
    } catch (requestError) {
      setError({ cause: requestError });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="proxy-interface-overlay" role="dialog" aria-modal="true" aria-label="Proxy city connection" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) onClose?.(); }}>
      <section className="proxy-interface" aria-labelledby="proxy-title" aria-busy={busy}>
        <header className="proxy-interface__header">
          <div><p>NOIR NETWORK · JURISDICTION RELAY</p><h2 id="proxy-title">Proxy connection</h2></div>
          <span className="proxy-interface__status"><i aria-hidden="true" /> {world?.city?.toUpperCase() ?? "NODE OFFLINE"} · ACTIVE NODE</span>
          <button ref={closeRef} type="button" onClick={onClose} disabled={busy} aria-label="Close Proxy">×</button>
        </header>
        <div className="proxy-interface__body">
          <section aria-labelledby="proxy-countries-title">
            <header><span id="proxy-countries-title">COUNTRY NODES</span><small>{loading ? "SYNCING" : `${countries.length} COUNTRIES`}</small></header>
            {loading ? <p className="proxy-interface__loading" role="status">Loading proxy nodes…</p> : <nav aria-label="Countries">
              {countries.map((country) => <button key={country.id} type="button" aria-pressed={country.id === countryId} disabled={busy} onClick={() => { setCountryId(country.id); setCityId(""); }}>
                <span>{country.name}</span><small>{country.cities.length} CITY NODES</small>
              </button>)}
            </nav>}
          </section>
          <section aria-labelledby="proxy-cities-title">
            <header><span id="proxy-cities-title">CITY NODES</span><small>{selectedCountry?.name.toUpperCase() ?? "SELECT A COUNTRY"}</small></header>
            {selectedCountry ? <nav aria-label={`Proxy nodes in ${selectedCountry.name}`}>
              {selectedCountry.cities.map((node) => <button key={node.id} type="button" aria-pressed={node.id === cityId} disabled={busy} onClick={() => setCityId(node.id)}>
                <span>{node.name}</span><small>{node.id === world?.cityId ? "CURRENT JURISDICTION" : node.timeZone}</small>
              </button>)}
            </nav> : <p className="proxy-interface__empty">Choose a country to view its city nodes.</p>}
          </section>
        </div>
        {error && <APIErrorNotice error={error.cause ?? error} onRetry={error.retry} retrying={loading} retryLabel="RETRY NODE CATALOG" className="proxy-interface__error" />}
        <footer className="proxy-interface__footer">
          <p><span>ACTIVE JURISDICTION</span>{selectedNode ? `${selectedNode.name.toUpperCase()}, ${selectedNode.country.toUpperCase()}` : "SELECT A CITY NODE"}</p>
          <button type="button" onClick={connect} disabled={!selectedNode || busy || loading}>
            {busy ? "CONNECTING…" : selectedNode?.id === world?.cityId ? "RECONNECT NODE" : "CONNECT PROXY"}<span aria-hidden="true">→</span>
          </button>
        </footer>
      </section>
    </div>
  );
}
