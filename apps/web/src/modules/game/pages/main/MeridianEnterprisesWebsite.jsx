import InstitutionalWebsiteFrame from "./InstitutionalWebsiteFrame.jsx";

export default function MeridianEnterprisesWebsite({ location, countryName }) {
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
            <article><p>TRANSIT NETWORK</p><h3>GATE</h3><span>The General Access Transit Exchange links cities through Meridian relay stations. Access logs require an authorized terminal application.</span></article>
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
      </main>
    </InstitutionalWebsiteFrame>
  );
}
