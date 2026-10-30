import { Fragment } from "react";

const LOCAL_ADS = [
  { name: "LECLERC TYPEWRITER CO.", headline: "A clean ribbon makes every letter count.", detail: "Fresh ribbons, repairs and office supplies." },
  { name: "THE EVENING COURIER", headline: "The city’s news before supper.", detail: "Ask your newsagent to reserve the afternoon edition." },
  { name: "MERIDIAN MUTUAL", headline: "A little foresight protects a great deal.", detail: "Household and shop cover through your local agent." },
  { name: "BELLECOUR RADIO & RECORDS", headline: "Bring the sound of the city home.", detail: "Radios, gramophone needles and this week’s records." },
  { name: "RHÔNE TRAVEL OFFICE", headline: "A change of scenery is closer than you think.", detail: "Rail timetables and weekend journeys from the central station." },
  { name: "MAISON VAUTRIN", headline: "Good shoes carry you further.", detail: "Repairs while you wait at our rue Victor-Hugo workshop." },
  { name: "CAFÉ DU PONT", headline: "The kettle is always on.", detail: "A warm lunch and a quiet table beside the river." },
  { name: "SAINT-CLAIRE PHARMACY", headline: "For the small things that keep a household well.", detail: "Open weekdays until seven; prescriptions prepared promptly." },
  { name: "LYON TAXI EXCHANGE", headline: "Across town, whatever the weather.", detail: "Telephone the exchange to arrange a dependable cab." },
  { name: "VINCENT’S PHOTOGRAPHIC SUPPLY", headline: "Keep the moment, not just the memory.", detail: "Film, camera repairs and same-day developing." },
  { name: "MERIDIAN SAVINGS & EXCHANGE", headline: "A safe place for the next good idea.", detail: "Visit your neighborhood counter for savings advice." },
  { name: "JARDIN DES FLEURS", headline: "A fresh arrangement brightens any room.", detail: "Seasonal flowers delivered throughout the city." },
];

function adsFor(site, pageId) {
  const key = `${site.company?.id || site.company?.name || "company"}:${pageId}`;
  let hash = 0;
  for (const character of key) hash = (hash * 31 + character.charCodeAt(0)) >>> 0;
  const count = { none: 0, sparse: 2, standard: 5, crowded: 12 }[site.advertising] ?? 5;
  return Array.from({ length: count }, (_, index) => LOCAL_ADS[(hash + index * 5) % LOCAL_ADS.length]);
}

function Advertisement({ ad, format = "card" }) {
  return <aside className={`company-website__ad company-website__ad--${format}`} aria-label={`Advertisement: ${ad.name}`}>
    <span className="company-website__ad-label">ADVERTISEMENT</span>
    <strong>{ad.name}</strong>
    <b>{ad.headline}</b>
    <span>{ad.detail}</span>
  </aside>;
}

function normalizedSite(site) {
  if (!site) return null;
  return {
    company: site.company || site.Company,
    address: site.address || site.Address,
    domain: site.domain || site.Domain,
    template: site.template || site.Template,
    advertising: site.advertising || site.Advertising || "standard",
    theme: site.theme || site.Theme,
    tagline: site.tagline || site.Tagline,
    pages: site.pages || site.Pages || [],
    account: site.account || site.Account,
  };
}

export default function CompanyWebsite({ site: sourceSite, pageId = "home", onNavigate, children }) {
  const site = normalizedSite(sourceSite);
  if (!site) return <p role="status">The company website could not be loaded.</p>;

  const isStaffPage = pageId === "staff";
  const page = isStaffPage ? null : site.pages.find((candidate) => candidate.id === pageId) || site.pages.find((candidate) => candidate.id === "home") || site.pages[0];
  const pages = site.pages.filter((candidate) => candidate.id !== "account").concat(site.pages.find((candidate) => candidate.id === "account") || []);
  const markBits = Array.from(site.theme?.mark || "0000000000000000", (character) => Number.parseInt(character, 16).toString(2).padStart(4, "0")).join("").slice(0, 64);
  const companyName = site.company?.name || "Company";
  const ads = adsFor(site, page?.id || pageId);
  const advertising = site.advertising;
  const inlineAds = advertising === "crowded"
    ? ads.slice(1, 1 + (page?.sections?.length || 0))
    : advertising === "sparse" ? ads.slice(0, 1) : advertising === "standard" ? ads.slice(1, 2) : [];
  const gridAds = advertising === "crowded" ? ads.slice(5) : advertising === "standard" ? ads.slice(2) : [];

  return (
    <main
      className={`company-website company-website--${site.template} company-website--layout-${site.theme?.layout}`}
      data-company-id={site.company?.id}
      data-site-mark={site.theme?.mark}
      style={{ "--company-accent": site.theme?.accent, "--company-paper": site.theme?.paper, "--company-ink": site.theme?.ink }}
      aria-label={`${companyName} public website`}
    >
      <header className={`company-website__masthead company-website__masthead--${site.theme?.header}`}>
        <span className="company-website__mark" aria-hidden="true"><svg viewBox="0 0 8 8" focusable="false">{Array.from(markBits, (bit, index) => bit === "1" ? <rect key={index} x={index % 8} y={Math.floor(index / 8)} width="1" height="1" /> : null)}</svg></span>
        <div><p>OFFICIAL COMPANY WEBSITE · {site.address?.city?.toUpperCase()}</p><h1>{companyName}</h1><span>{site.tagline}</span></div>
        <button type="button" disabled aria-label="Customer account sign in is unavailable">CUSTOMER LOGIN</button>
      </header>
      <nav className="company-website__nav" aria-label={`${companyName} website`}>
        {pages.map((item) => <a
          key={item.id}
          href={`#${site.domain}/${item.id}`}
          aria-current={item.id === page?.id ? "page" : undefined}
          onClick={(event) => { event.preventDefault(); onNavigate?.(item.id); }}
        >{item.title}</a>)}
        {isStaffPage && <span className="company-website__staff-link" aria-current="page">EMPLOYEE DIRECTORY</span>}
      </nav>
      {(advertising === "standard" || advertising === "crowded") && ads[0] && <Advertisement ad={ads[0]} format="banner" />}
      {isStaffPage
        ? <section className="company-website__staff-page" aria-label={`${companyName} employee directory`}>{children}</section>
        : <section className={`company-website__feature company-website__feature--${site.theme?.layout} company-website__feature--${site.template} company-website__feature--page-${page?.id}`} aria-labelledby="company-website-title">
        <p className="company-website__eyebrow">{page?.title} · PUBLIC INFORMATION</p>
        <h2 id="company-website-title">{page?.id === "home" ? `Welcome to ${companyName}` : page?.title}</h2>
        {page?.id !== "account" && <div className={`company-website__sections company-website__sections--${site.template}`}>
          {page?.sections?.map((section, index) => <Fragment key={`${section.heading}-${index}`}>
            <article className={`company-website__section company-website__section--${section.kind || "article"}`}>
              <h3>{section.heading}</h3>
              <p>{section.body}</p>
              {section.items?.length > 0 && <ul>{section.items.map((item) => <li key={item}>{item}</li>)}</ul>}
            </article>
            {inlineAds[index] && <Advertisement key={`ad-${index}`} ad={inlineAds[index]} format="inline" />}
          </Fragment>)}
        </div>}
        {page?.id === "contact" && <address>{site.address?.full}</address>}
        {page?.id === "account" && <div className="company-website__account">
          <h3>{page.title}</h3><p>{site.account?.message}</p><button type="button" disabled>{site.account?.label || "CUSTOMER ACCOUNT"} · UNAVAILABLE</button>
        </div>}
        </section>}
      {gridAds.length > 0 && <section className={`company-website__ad-grid company-website__ad-grid--${advertising}`} aria-label="Sponsored local offers">
        {gridAds.map((ad, index) => <Advertisement key={`${ad.name}-${index}`} ad={ad} />)}
      </section>}
      <footer className="company-website__footer"><span>{companyName} · {site.address?.full}</span><span>{site.domain} · Best viewed with Meridian Web Browser 4.1</span></footer>
    </main>
  );
}
