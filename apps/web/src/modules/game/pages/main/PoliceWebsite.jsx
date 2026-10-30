import { useState } from "react";
import { deepLinkPage } from "./DeepLinkWebsite.jsx";

const CRIME_REPORTS_PAGE = { id: "crime-reports", label: "CRIME REPORTS", available: true };
const POLICE_NOTICES_PAGE = { id: "police-notices", label: "PUBLIC NOTICES", available: true };
const POLICE_ABOUT_PAGE = { id: "police-about", label: "ABOUT THE DEPARTMENT", available: true };
const POLICE_BUREAUS_PAGE = { id: "police-bureaus", label: "DISTRICT BUREAUS", available: true };
const POLICE_POLICIES_PAGE = { id: "police-policies", label: "PUBLICATION POLICIES", available: true };
const POLICE_CONTACT_PAGE = { id: "police-contact", label: "CONTACT THE POLICE SERVICE", available: true };

export { CRIME_REPORTS_PAGE, POLICE_NOTICES_PAGE, POLICE_ABOUT_PAGE, POLICE_BUREAUS_PAGE, POLICE_POLICIES_PAGE, POLICE_CONTACT_PAGE };

export default function PoliceWebsite({ page = "police-website", navigate, children }) {
  const [search, setSearch] = useState("");
  const title = page === "police-notices" ? "Department notices" : page === "police-about" ? "About the department" : "Meridian Police Service";
  function searchReports(event) {
    event.preventDefault();
    navigate({ ...CRIME_REPORTS_PAGE, searchTerm: search.trim() });
  }
  function openDeepLink(linkId) { navigate(deepLinkPage(linkId)); }
  return (
    <main className="legacy-site legacy-police">
      <header className="legacy-site__masthead legacy-police__masthead">
        <div><span className="legacy-police__crest" aria-hidden="true">M</span><span className="legacy-site__brand">MERIDIAN POLICE SERVICE</span><span className="legacy-site__brand-sub">PUBLIC INFORMATION OFFICE · NATIONAL RELAY</span></div>
        <div className="legacy-site__masthead-note"><strong>PUBLIC ACCESS</strong><br />Emergency calls: 17<br />Office hours: Monday—Saturday</div>
      </header>
      <nav className="legacy-site__toplinks" aria-label="Police website sections">
        <button type="button" onClick={() => navigate({ id: "police-website", label: "MERIDIAN POLICE", available: true })}>HOME</button>
        <button type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}>CRIME REPORTS</button>
        <button type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}>PUBLIC NOTICES</button>
        <button type="button" onClick={() => navigate(POLICE_ABOUT_PAGE)}>ABOUT</button>
      </nav>
      <form className="legacy-police__site-search" onSubmit={searchReports}><label htmlFor="police-site-search">SEARCH POLICE INFORMATION</label><input id="police-site-search" type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Type a name, then press Enter" aria-keyshortcuts="Enter" /><button className="search-enter-submit" type="submit" aria-label="Press Enter to search reports">ENTER ↵</button></form>
      <div className="legacy-site__layout legacy-police__layout">
        <aside className="legacy-site__sidebar">
          <section className="legacy-police__date"><b>THURSDAY</b><strong>29</strong><span>OCTOBER</span></section>
          <section><h2>POLICE DESK</h2><ul>
            <li><button type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}>Published crime reports</button></li>
            <li><button type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}>Department notices</button></li>
            <li><button type="button" onClick={() => navigate(POLICE_ABOUT_PAGE)}>About this service</button></li>
            <li><button type="button" onClick={() => navigate(POLICE_CONTACT_PAGE)}>Contact information</button></li>
          </ul></section>
          <section className="legacy-site__ad legacy-police__ad"><small>PUBLIC SERVICE ANNOUNCEMENT</small><b>LOCK YOUR DOOR<br />BEFORE YOU LEAVE</b><span>Report suspicious activity to your local station.</span></section>
          <section><h2>RELATED SITES</h2><ul><li><button type="button" onClick={() => openDeepLink("municipal-directory")}>Municipal directory</button></li><li><button type="button" onClick={() => openDeepLink("telephone-service")}>Emergency telephone service</button></li></ul></section>
        </aside>
        <div className="legacy-site__main-column">
          {page === "police-website" && <>
            <section className="legacy-police__hero"><div><p>MERIDIAN POLICE · PUBLIC INFORMATION</p><h1>Working together for a safer Meridian.</h1><span>News, services and public records from your city police department.</span><button type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}>SEARCH CRIME REPORTS »</button></div><aside><b>ON THE BEAT</b><span>PUBLIC SAFETY<br />CITY RECORDS<br />COMMUNITY NEWS</span><small>29 OCT</small></aside></section>
            <div className="legacy-site__columns">
              <section className="legacy-site__section legacy-police__reports"><h2>Records and services</h2><ul className="legacy-police__link-list"><li><div><button type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}><b>Crime reports »</b></button><span>Search incident reports released for public review.</span></div></li><li><div><button type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}><b>Department notices »</b></button><span>Read recent announcements and public safety notices.</span></div></li><li><div><button type="button" onClick={() => navigate(POLICE_ABOUT_PAGE)}><b>About the department »</b></button><span>Learn how this office maintains public records.</span></div></li></ul></section>
              <aside className="legacy-police__news"><h2>Latest announcements</h2><article><time>29 OCT</time><h3>Public records office hours</h3><p>Requests are received at the central desk from 9:00 AM to 5:00 PM.</p><button type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}>More notices »</button></article><article><time>24 OCT</time><h3>Winter road safety</h3><p>Drivers are asked to keep the river approaches clear after dusk.</p></article></aside>
            </div>
            <section className="legacy-police__categories"><h2>Explore the department</h2><div><button type="button" onClick={() => navigate(POLICE_BUREAUS_PAGE)}><b>BUREAUS</b><span>District offices and department divisions »</span></button><button type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}><b>SERVICES</b><span>Public records and community assistance »</span></button><button type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}><b>MEDIA</b><span>News releases and department notices »</span></button><button type="button" onClick={() => navigate(POLICE_POLICIES_PAGE)}><b>POLICIES</b><span>Public information and records guidance »</span></button></div></section>
            <section className="legacy-police__sponsors"><p>SPONSORED ADVERTISEMENTS</p><article><button type="button" onClick={() => openDeepLink("leclerc-locksmith")}><b>LECLERC LOCK & KEY</b><span>Reliable locks for home and office »</span></button></article><article><button type="button" onClick={() => openDeepLink("evening-courier")}><b>THE EVENING COURIER</b><span>Local news delivered each afternoon »</span></button></article><article><button type="button" onClick={() => openDeepLink("meridian-mutual")}><b>MERIDIAN MUTUAL</b><span>Household cover from a local insurer »</span></button></article></section>
          </>}
          {page === "police-notices" && <section className="legacy-site__section legacy-police__article"><p className="legacy-site__eyebrow">POLICE DESK / PUBLIC NOTICES</p><h1>Department notices</h1><p>Announcements issued by the Meridian Police Service public information office.</p><ul className="legacy-police__notice-list"><li><time>29 October</time><b>Public records office hours</b><p>The central desk accepts public record requests Monday through Saturday, 9:00 AM to 5:00 PM.</p></li><li><time>24 October</time><b>Winter road safety</b><p>Drivers are asked to keep the river approaches clear after dusk. Patrols will assist stranded motorists where possible.</p></li><li><time>18 October</time><b>Telephone directory corrections</b><p>Residents may submit corrections to their public listing through the municipal directory office.</p></li></ul><button className="legacy-site__text-link" type="button" onClick={() => navigate({ id: "police-website", label: "MERIDIAN POLICE", available: true })}>« Return to police service</button></section>}
          {page === "police-about" && <section className="legacy-site__section legacy-police__article"><p className="legacy-site__eyebrow">POLICE DESK / ABOUT THIS SERVICE</p><h1>About the department</h1><p>The Meridian Police Service maintains public safety and publishes selected records for community reference. Each municipal office is responsible for the accuracy and release status of its local records.</p><h2>Public records</h2><p>Published crime reports are reviewed before release. Personal, sealed and active investigation material is not included in public listings.</p><h2>Contact</h2><p>For an emergency, telephone 17. For directory corrections or questions about a published report, contact the local records office during business hours.</p><button className="legacy-site__text-link" type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}>Browse published crime reports »</button></section>}
          {page === "police-bureaus" && <section className="legacy-site__section legacy-police__article"><p className="legacy-site__eyebrow">POLICE DESK / DISTRICT BUREAUS</p><h1>District offices</h1><p>Local bureaus handle neighborhood patrols, lost property and routine reports. The central desk can direct visitors to the correct district office.</p><h2>Central records bureau</h2><p>Public records requests and copies of released incident reports are handled at the central desk, Monday through Saturday.</p><h2>Community safety bureau</h2><p>Residents may ask about street lighting, neighborhood patrols and safety notices for public gatherings.</p><button className="legacy-site__text-link" type="button" onClick={() => navigate(POLICE_CONTACT_PAGE)}>Find office contact information »</button></section>}
          {page === "police-policies" && <section className="legacy-site__section legacy-police__article"><p className="legacy-site__eyebrow">POLICE DESK / PUBLICATION POLICIES</p><h1>Public information and records</h1><p>Published records are reviewed by the local office before release. Active investigations, sealed files and private identifying details are withheld.</p><h2>Requesting a record</h2><p>Include the incident date and district when writing to the records bureau. Requests are answered in the order received during office hours.</p><button className="legacy-site__text-link" type="button" onClick={() => navigate(CRIME_REPORTS_PAGE)}>Browse published crime reports »</button></section>}
          {page === "police-contact" && <section className="legacy-site__section legacy-police__article"><p className="legacy-site__eyebrow">POLICE DESK / CONTACT</p><h1>Contact the Meridian Police Service</h1><p>For an emergency, telephone 17. For routine matters, call the local exchange and ask for the Meridian Police central desk.</p><h2>Public information office</h2><p>Records and directory enquiries are received Monday through Saturday, 9:00 AM to 5:00 PM. Please include your district and a return address with written requests.</p><button className="legacy-site__text-link" type="button" onClick={() => navigate(POLICE_NOTICES_PAGE)}>View office hours and notices »</button></section>}
          {page === "crime-reports" && <section className="legacy-police__crime-reports"><p className="legacy-site__eyebrow">POLICE DESK / PUBLIC RECORDS</p><h1>Published crime reports</h1><p>Search incidents released by the connected city police office.</p>{children}</section>}
          <footer id="contact" className="legacy-site__directory-footer"><b>Meridian Police Service</b><span>Public Information Office · National relay directory</span><span>Emergency: 17 · Non-emergency: consult your local exchange</span></footer>
        </div>
      </div>
      <footer className="legacy-site__footer">© Meridian Police Service · Information subject to local publication rules</footer>
    </main>
  );
}
