import { useState } from "react";
import InstitutionalWebsiteFrame from "./InstitutionalWebsiteFrame.jsx";

export default function MeridianEnterprisesWebsite({ location, cityId = "", countryName, onOpenService }) {
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState("");
  const city = location?.split(",")[0]?.trim() || "Connected City";
  const serviceCityId = cityId === "fr/lyon" ? "lyon" : cityId;
  const gateService = {
    kind: "gate",
    address: {
      locationId: `service:${serviceCityId}/gate-station`,
      name: "Meridian Enterprises",
      city,
      full: location || countryName || "National Relay",
    },
  };

  async function openGateLogs() {
    if (!onOpenService || connecting) return;
    setError("");
    setConnecting(true);
    try {
      await onOpenService(gateService);
    } catch (cause) {
      setError(cause?.message || "The GATE records office could not be reached.");
    } finally {
      setConnecting(false);
    }
  }

  return (
    <InstitutionalWebsiteFrame variant="gate" footerContent={<>Meridian Enterprises · {location || countryName || "National Relay"} · Corporate information service</>}>
      <main className="meridian-enterprises">
        <section className="meridian-enterprises__hero" id="enterprise-home">
          <div><p>MERIDIAN ENTERPRISES · GLOBAL SYSTEMS</p><h1>We built the connections the modern world depends on.</h1><span>Movement, communication and personal computing, connected through one Meridian network.</span></div>
          <aside><b>MERIDIAN NETWORK</b><strong>CONNECTED</strong><span>Operating worldwide since 1948</span></aside>
        </section>
        <div className="meridian-enterprises__columns">
          <section className="meridian-enterprises__divisions" aria-labelledby="enterprise-divisions-title">
            <h2 id="enterprise-divisions-title">Our divisions</h2>
            <article><p>TRANSIT NETWORK</p><h3>GATE</h3><span>The General Access Transit Exchange links cities through Meridian relay stations.</span><button type="button" disabled={connecting} onClick={openGateLogs}>{connecting ? "CONNECTING TO GATE…" : "GATE ACCESS LOGS »"}</button></article>
            <article><p>PERSONAL COMPUTING</p><h3>Meridian OS</h3><span>The operating environment used by homes, offices and public terminals across the connected world.</span></article>
            <article><p>GLOBAL INFRASTRUCTURE</p><h3>Meridian Relay</h3><span>Telephone, records and information services delivered through a dependable national network.</span></article>
          </section>
          <aside className="meridian-enterprises__bulletin">
            <h2>Company bulletin</h2>
            <article><time>29 OCTOBER</time><b>GATE regional maintenance</b><p>Scheduled relay inspection will begin after midnight. Passenger service is expected to continue.</p></article>
            <article><time>18 OCTOBER</time><b>Meridian OS service notice</b><p>Updated terminal documentation is available through authorized city offices.</p></article>
            <div><small>MERIDIAN ADVERTISING EXCHANGE</small><strong>THE WORLD, WITHIN REACH.</strong><span>A Meridian system for every office.</span></div>
          </aside>
        </div>
        {error && <p className="meridian-enterprises__error" role="alert">{error}</p>}
      </main>
    </InstitutionalWebsiteFrame>
  );
}
