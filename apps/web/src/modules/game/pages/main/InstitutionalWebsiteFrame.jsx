const SITE_IDENTITIES = {
  police: { title: "Meridian Police Service", short: "POLICE SERVICE", subtitle: "PUBLIC SAFETY · RECORDS · COMMUNITY", symbol: "M", status: "PUBLIC INFORMATION OFFICE" },
  hospital: { title: "Hôpital Sainte-Claire", short: "HOSPITAL", subtitle: "PATIENT SERVICES · MEDICAL RECORDS", symbol: "+", status: "CITY MEDICAL NETWORK" },
  bank: { title: "Union Mercantile Bank", short: "UNION MERCANTILE", subtitle: "ACCOUNTS · SETTLEMENTS · CUSTOMER SERVICE", symbol: "U", status: "NATIONAL BANKING NETWORK" },
  gate: { title: "Meridian Enterprises", short: "MERIDIAN ENTERPRISES", subtitle: "GLOBAL SYSTEMS · TRANSIT · COMPUTING", symbol: "ME", status: "NATIONAL TECHNOLOGY NETWORK" },
  civic: { title: "Office of Civic Services", short: "CITY SERVICES", subtitle: "RESIDENT INFORMATION · MUNICIPAL OFFICES", symbol: "C", status: "MUNICIPAL INFORMATION EXCHANGE" },
  company: { title: "Meridian Global", short: "CORPORATE DIRECTORY", subtitle: "COMPANY INFORMATION · EMPLOYMENT", symbol: "MG", status: "NATIONAL BUSINESS NETWORK" },
  mail: { title: "Meridian Personal Mail", short: "MAIL SERVICE", subtitle: "PRIVATE CORRESPONDENCE · ACCOUNT SERVICES", symbol: "✉", status: "PRIVATE MAIL SERVICE" },
};

export function identityForWebsite(variant, organizationName = "") {
  const identity = SITE_IDENTITIES[variant] || SITE_IDENTITIES.company;
  if (variant === "hospital" && organizationName) {
    return { ...identity, title: organizationName, short: "HOSPITAL", status: "CITY MEDICAL NETWORK" };
  }
  return variant === "company" && organizationName
    ? { ...identity, title: organizationName, short: "COMPANY WEBSITE", status: "PUBLIC COMPANY INFORMATION" }
    : identity;
}

export default function InstitutionalWebsiteFrame({ variant, organizationName = "", onHome, footerContent, children }) {
  const identity = identityForWebsite(variant, organizationName);
  return (
    <main className={`institutional-site institutional-site--${variant}`}>
      <header className="institutional-site__masthead">
        <span className="institutional-site__symbol" aria-hidden="true">{identity.symbol}</span>
        <span className="institutional-site__identity"><strong>{identity.title}</strong><small>{identity.subtitle}</small></span>
        <span className="institutional-site__status">{identity.status}<br />Established 1948 · National Relay</span>
      </header>
      <nav className="institutional-site__nav" aria-label={`${identity.title} website sections`}>
        <a href={variant === "hospital" ? "#hospital-home" : "#institutional-content"} onClick={onHome ? (event) => { event.preventDefault(); onHome(); } : undefined}>HOME</a>
        <a href="#institutional-contact">CONTACT</a>
        <span>INFORMATION INDEX: A—Z</span>
      </nav>
      <div className="institutional-site__notice"><b>{identity.short}</b><span>Connected through the secure national relay</span><a href="#institutional-contact">Contact the office »</a></div>
      <div className="institutional-site__content" id="institutional-content">{children}</div>
      <footer className="institutional-site__footer" id="institutional-contact">{footerContent || `${identity.title} · Public website · Best viewed with Meridian Web Browser 4.1`}</footer>
    </main>
  );
}
