import { useState } from "react";
import { searchNationalCompanies } from "../../../world/api.js";

const PORTAL_INTRO = "NATIONAL DIRECTORY · MERIDIAN RELAY";

export default function LegacyCountryPortal({ countryName, location, cityId = "", programs, navigate, onOpenCompany, onNavigateCompanyWebsite, initialSearch = "", initialCompanies = [] }) {
  const [search, setSearch] = useState(initialSearch);
  const [appliedSearch, setAppliedSearch] = useState(initialSearch);
  const [companyResults, setCompanyResults] = useState(initialCompanies);
  const [searchMessage, setSearchMessage] = useState(initialCompanies.length ? "COMPANY LISTINGS RESTORED" : "");
  const [searching, setSearching] = useState(false);
  const [openingCompanyId, setOpeningCompanyId] = useState("");
  const cityName = location?.split(",")[0]?.trim() || "Connected City";
  const serviceCityId = cityId === "fr/lyon" ? "lyon" : cityId;
  const policeProgram = programs.find((site) => site.id === "police-website") || { id: "police-website", Icon: null, state: "ONLINE", available: true };
  const emailProgram = programs.find((site) => site.id === "personal-mail") || { id: "personal-mail", Icon: null, state: "ONLINE", available: true };
  const hospitalSite = {
    id: "hospital-website",
    label: "NATIONAL HOSPITAL",
    Icon: null,
    state: "ONLINE",
    available: true,
    service: { kind: "hospital", address: { locationId: `service:${serviceCityId}/hospital`, name: `${cityName} Hospital`, city: cityName, full: location || countryName } },
  };
  const policeSite = { ...policeProgram, label: "POLICE DEPARTMENT", available: true };
  const enterpriseSite = {
    id: "meridian-enterprises",
    label: "MERIDIAN ENTERPRISES",
    Icon: null,
    state: "ONLINE",
    available: true,
  };
  const bankSite = {
    id: "bank-website",
    label: "UNION MERCANTILE BANK",
    Icon: null,
    state: "ONLINE",
    available: true,
  };
  const emailSite = { ...emailProgram, label: "MERIDIAN EMAIL", available: true };
  const directorySites = [hospitalSite, policeSite, bankSite, enterpriseSite, emailSite];
  const openSite = (site) => navigate(site);
  const visibleDirectorySites = directorySites.filter((site) => `${site.label} ${portalDescription(site.id)}`.toLocaleLowerCase().includes(appliedSearch.trim().toLocaleLowerCase()));

  async function submitSearch(event) {
    event.preventDefault();
    const term = search.trim();
    setAppliedSearch(term);
    if (!term) {
      setCompanyResults([]);
      setSearchMessage("ENTER A COMPANY, WEBSITE, SERVICE OR ADDRESS TO SEARCH.");
      return;
    }
    setSearching(true);
    setCompanyResults([]);
    setSearchMessage("SEARCHING NATIONAL COMPANY LISTINGS…");
    try {
      const results = await searchNationalCompanies(term);
      setCompanyResults(results);
      setSearchMessage(results.length ? `${results.length} NATIONAL COMPANY LISTING${results.length === 1 ? "" : "S"} FOUND.` : "NO MATCHING COMPANY WEBSITES.");
    } catch (error) {
      setSearchMessage(error?.message || "THE NATIONAL DIRECTORY COULD NOT BE REACHED.");
    } finally {
      setSearching(false);
    }
  }

  async function openCompany(record) {
    if (!onOpenCompany || openingCompanyId) return;
    setOpeningCompanyId(record.company.id);
    setSearchMessage(`CONNECTING TO ${record.company.name.toUpperCase()}…`);
    try {
      const site = await onOpenCompany(record);
      if (site) onNavigateCompanyWebsite?.(site, { query: appliedSearch, companies: companyResults });
    } catch (error) {
      setSearchMessage(error?.message || "THE COMPANY WEBSITE COULD NOT BE REACHED.");
    } finally {
      setOpeningCompanyId("");
    }
  }
  return (
    <main className="legacy-site legacy-portal">
      <header className="legacy-site__masthead legacy-portal__masthead">
        <div><span className="legacy-site__brand">MERIDIAN</span><span className="legacy-site__brand-sub">NATIONAL DIRECTORY</span></div>
        <form className="legacy-portal__search" onSubmit={submitSearch}><input type="search" aria-label="Search the National Directory" placeholder="Type a company, then press Enter" aria-keyshortcuts="Enter" value={search} onChange={(event) => setSearch(event.target.value)} /><button className="search-enter-submit" type="submit" disabled={searching} aria-label={searching ? "Searching" : "Press Enter to search"}>{searching ? "SEARCHING…" : "ENTER ↵"}</button></form>
        <div className="legacy-site__masthead-note"><strong>{countryName || "COUNTRY PORTAL"}</strong><br />{location || "RELAY REGION NOT SET"}<br />Updated Thursday, 29 October</div>
      </header>
      <nav className="legacy-site__toplinks" aria-label="Portal sections">
        <span>HOME</span><a href="#sites">NATIONAL DIRECTORY</a><a href="#health">HOSPITALS &amp; CARE</a><a href="#notices">NOTICES</a><a href="#help">HELP</a>
      </nav>
      <div className="legacy-site__layout">
        <aside className="legacy-site__sidebar">
          <section><h2>QUICK LINKS</h2><ul>{directorySites.map((site) => <li key={site.id}><button type="button" onClick={() => openSite(site)}>{site.label}</button></li>)}</ul></section>
          <section className="legacy-site__ad legacy-site__ad--small"><b>LECLERC<br />TYPEWRITER CO.</b><span>New ribbons in stock.<br />Ask for the detective's discount.</span><small>ADVERTISEMENT · 14</small></section>
          <section><h2>ON THIS PAGE</h2><ul><li><a href="#notices">Notices</a></li><li><a href="#services">Government</a></li><li><a href="#help">Contact the directory</a></li></ul></section>
        </aside>
        <div className="legacy-site__main-column">
          <section className="legacy-site__welcome">
            <div><p className="legacy-site__eyebrow">{PORTAL_INTRO}</p><h1>Welcome to {countryName || "the national network"}</h1><p>Find national companies, city institutions, records and services connected through the Meridian relay.</p></div>
            <div className="legacy-site__welcome-links"><b>NATIONAL WEBSITES</b>{directorySites.map((site) => <button key={site.id} type="button" onClick={() => openSite(site)}>{site.label}</button>)}</div>
          </section>
          <section className="legacy-site__section" id="sites"><h2>National directory</h2><p className="legacy-site__section-lead">Browse institutions and businesses by category. Protected records require an authorized connection.</p>
            <div className="legacy-portal__directory">{visibleDirectorySites.map((site) => <article id={site.id === "hospital-website" ? "health" : undefined} key={site.id} className="legacy-portal__listing"><h3><button type="button" onClick={() => openSite(site)}>{site.label} »</button></h3><p>{portalDescription(site.id)}</p></article>)}</div>
            {searchMessage && <p className="legacy-portal__search-status" role="status" aria-live="polite">{searchMessage}</p>}
            {companyResults.length > 0 && <div className="legacy-portal__company-results" aria-label="National company websites">
              {companyResults.map((record) => <article className="legacy-portal__company-result" key={record.company.id}>
                <h3><button type="button" disabled={Boolean(openingCompanyId)} onClick={() => openCompany(record)}>{record.company.name} »</button></h3>
                <p>{record.websiteDescription || "PUBLIC COMPANY INFORMATION"}</p>
                <p><b>WEBSITE</b> · {record.websiteDomain}</p>
                <p><b>HEAD OFFICE</b> · {record.address.full}</p>
                {openingCompanyId === record.company.id && <small role="status">CONNECTING…</small>}
              </article>)}
            </div>}
            {!searching && appliedSearch.trim() && !companyResults.length && !visibleDirectorySites.length && searchMessage.startsWith("NO MATCHING") && <p className="legacy-portal__no-results">No directory listings match “{appliedSearch}”.</p>}
          </section>
          <div className="legacy-site__columns">
            <section id="notices" className="legacy-site__notice"><h2>Bulletin board</h2><ul><li><b>29 Oct</b> — Regional relay maintenance is scheduled after midnight.</li><li><b>28 Oct</b> — Municipal offices will observe the Armistice Day schedule.</li><li><b>24 Oct</b> — Police department publishes updated public records guidance.</li></ul><button type="button" onClick={() => policeSite && openSite(policeSite)}>Read department announcements »</button></section>
            <aside id="help" className="legacy-site__ad"><b>VINCENT'S<br />PHOTOGRAPHIC SUPPLY</b><span>Film developed while you wait.<br />12 rue de la République, Lyon</span><small>SPONSORED LISTING · MERIDIAN ADVERTISING EXCHANGE</small></aside>
          </div>
          <section id="services" className="legacy-site__directory-footer"><b>Directory office</b><span>Corrections and listing requests: Meridian National Directory, National Desk.</span><a href="#sites">Return to directory ↑</a></section>
        </div>
      </div>
      <footer className="legacy-site__footer">© Meridian National Directory · Best viewed at 800 × 600 · Information service</footer>
    </main>
  );
}

function portalDescription(id) {
  return ({
    "police-website": "Department news, public safety information and access to published incident reports.",
    "personal-mail": "Private correspondence delivered to your personal account.",
    "company-portal": "Search the national register of companies and visit their public listings.",
    "hospital-website": "Hospital information, departments, visiting hours and admissions guidance for the connected city.",
    "bank-website": "Branch information, savings services and public notices from Union Mercantile Bank.",
    "meridian-enterprises": "Corporate information, Meridian OS services and the GATE network. Access logs are available inside the GATE section.",
  })[id] || "Public information and records available through the connected relay.";
}
