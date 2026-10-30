import { useEffect, useMemo, useState } from "react";
import { connectProxy, getProxyNodes } from "../api.js";
import APIErrorNotice from "../../../shared/APIErrorNotice.jsx";

export default function ProxyConnection({ world, onConnected, hasOpenBrowser = false, onRouteChanged, onBusyChange }) {
  const [nodes, setNodes] = useState([]);
  const [countryId, setCountryId] = useState("");
  const [cityId, setCityId] = useState(world?.cityId ?? "");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(null);
  const [retry, setRetry] = useState(0);
  const [pendingNode, setPendingNode] = useState(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    getProxyNodes().then((items) => {
      if (!active) return;
      setNodes(items);
      const current = items.find(({ id }) => id === world?.cityId);
      setCountryId(current?.countryId ?? items[0]?.countryId ?? "");
      setCityId(current?.id ?? items[0]?.id ?? "");
    }).catch((requestError) => {
      if (active) setError({ cause: requestError, retry: () => setRetry((value) => value + 1) });
    }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [world?.cityId, retry]);

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

  async function applyRoute(node) {
    if (!node || busy) return;
    setBusy(true);
    onBusyChange?.(true, node);
    setError(null);
    try {
      const [response] = await Promise.all([
        connectProxy(node.id),
        new Promise((resolve) => window.setTimeout(resolve, 2400)),
      ]);
      await onConnected?.(response);
      if (node.id !== world?.cityId) onRouteChanged?.();
      setPendingNode(null);
    } catch (requestError) {
      setError({ cause: requestError });
    } finally {
      setBusy(false);
      onBusyChange?.(false, null);
    }
  }

  function connect() {
    if (!selectedNode || busy) return;
    if (selectedNode.id === world?.cityId) {
      void applyRoute(selectedNode);
      return;
    }
    setPendingNode(selectedNode);
  }

  return (
    <section className="proxy-interface proxy-interface--app" aria-labelledby="proxy-title" aria-busy={busy}>
      <header className="proxy-interface__header">
        <div><p>UNREGISTERED RESOLVER · SHADOW ROUTING</p><h2 id="proxy-title">Rogue DNS</h2></div>
        <span className="proxy-interface__status"><i aria-hidden="true" /> {world?.city?.toUpperCase() ?? "NODE OFFLINE"} · CURRENT EXIT</span>
      </header>
      <div className="proxy-interface__body">
        <section aria-labelledby="proxy-countries-title">
          <header><span id="proxy-countries-title">RESOLVER REGIONS</span><small>{loading ? "SCANNING" : `${countries.length} REGIONS`}</small></header>
          {loading ? <p className="proxy-interface__loading" role="status">Loading proxy nodes…</p> : <nav aria-label="Countries">
            {countries.map((country) => <button key={country.id} type="button" aria-pressed={country.id === countryId} disabled={busy} onClick={() => { setCountryId(country.id); setCityId(""); }}>
              <span>{country.name}</span><small>{country.cities.length} CITY NODES</small>
            </button>)}
          </nav>}
        </section>
        <section aria-labelledby="proxy-cities-title">
          <header><span id="proxy-cities-title">EXIT NODES</span><small>{selectedCountry?.name.toUpperCase() ?? "SELECT A REGION"}</small></header>
          {selectedCountry ? <nav aria-label={`Proxy nodes in ${selectedCountry.name}`}>
            {selectedCountry.cities.map((node) => <button key={node.id} type="button" aria-pressed={node.id === cityId} disabled={busy} onClick={() => setCityId(node.id)}>
              <span>{node.name}</span><small>{node.id === world?.cityId ? "ACTIVE EXIT NODE" : node.timeZone}</small>
            </button>)}
          </nav> : <p className="proxy-interface__empty">Choose a country to view its city nodes.</p>}
        </section>
      </div>
      {error && <APIErrorNotice error={error.cause ?? error} onRetry={error.retry} retrying={loading} retryLabel="RETRY NODE CATALOG" className="proxy-interface__error" />}
      <footer className="proxy-interface__footer">
        <p><span>ROUTE TRAFFIC THROUGH</span>{selectedNode ? `${selectedNode.name.toUpperCase()}, ${selectedNode.country.toUpperCase()}` : "SELECT AN EXIT NODE"}</p>
        <button type="button" onClick={connect} disabled={!selectedNode || busy || loading}>
          {busy ? "SWITCHING ROUTE…" : selectedNode?.id === world?.cityId ? "RECONNECT DNS" : "USE THIS NODE"}<span aria-hidden="true">→</span>
        </button>
      </footer>
      {pendingNode && <div className="proxy-route-confirm__backdrop">
        <section className="proxy-route-confirm" role="alertdialog" aria-modal="true" aria-labelledby="proxy-route-confirm-title" aria-describedby="proxy-route-confirm-message" onKeyDown={(event) => { if (event.key === "Escape" && !busy) setPendingNode(null); }}>
          <p className="proxy-route-confirm__eyebrow">UNLISTED ROUTE CHANGE</p>
          <h3 id="proxy-route-confirm-title">Switch exit node?</h3>
          <p id="proxy-route-confirm-message">Route traffic through {pendingNode.name}, {pendingNode.country}? {hasOpenBrowser ? "All open browser windows will close when the switch completes." : "The active city node will change when the switch completes."}</p>
          <footer>
            <button type="button" autoFocus disabled={busy} onClick={() => setPendingNode(null)}>CANCEL</button>
            <button type="button" disabled={busy} onClick={() => { void applyRoute(pendingNode); }}>{busy ? "SWITCHING…" : "SWITCH ROUTE"}</button>
          </footer>
        </section>
      </div>}
    </section>
  );
}
