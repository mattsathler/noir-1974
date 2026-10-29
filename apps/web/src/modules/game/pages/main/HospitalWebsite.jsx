import InstitutionalWebsiteFrame from "./InstitutionalWebsiteFrame.jsx";

export default function HospitalWebsite({ service, onConnect }) {
  const name = service?.address?.name || "Hôpital Sainte-Claire";
  const address = service?.address?.full || "City Medical Network · Patient Services";
  return (
    <InstitutionalWebsiteFrame variant="hospital" organizationName={name} footerContent={<><strong>{name}</strong> · {address} · Emergency: 15 · General enquiries: contact the city exchange</>}>
      <main className="hospital-website">
      <section className="hospital-website__hero" id="hospital-home"><div><small>WELCOME TO {name.toUpperCase()}</small><h1>Care for every member of our community.</h1><p>Hospital information, visiting guidance and patient services for residents of the connected city.</p></div><div className="hospital-website__hours"><b>VISITING HOURS</b><strong>2:00 — 6:00 PM</strong><span>All visitors report to the main desk.</span></div></section>
      <div className="hospital-website__content">
        <section id="hospital-services"><h2>Patient and family services</h2><div className="hospital-website__service-grid"><article><h3>Patient search</h3><p>Authorized lookup for admissions and hospital records.</p><button type="button" onClick={() => onConnect?.(service)}>OPEN SECURE PATIENT SEARCH »</button></article><article><h3>Admissions desk</h3><p>Information about arrivals, visiting rooms and family inquiries.</p><a href="#hospital-contact">Read admission guidance »</a></article><article><h3>Medical records office</h3><p>Records are released only to verified and authorized accounts.</p><button type="button" onClick={() => onConnect?.(service)}>REQUEST RECORDS ACCESS »</button></article></div></section>
        <aside className="hospital-website__notice"><b>PUBLIC NOTICE · 29 OCTOBER</b><p>The west entrance is closed for repairs. Visitors should use the main avenue entrance until further notice.</p><a href="#hospital-contact">All hospital notices »</a></aside>
        <section id="hospital-departments" className="hospital-website__departments"><h2>Departments</h2><a href="#hospital-services">General medicine</a><a href="#hospital-services">Maternity ward</a><a href="#hospital-services">Surgical services</a><a href="#hospital-services">Records and admissions</a></section>
      </div>
      </main>
    </InstitutionalWebsiteFrame>
  );
}
