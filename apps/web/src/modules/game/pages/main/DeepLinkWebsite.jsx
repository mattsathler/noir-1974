const PAGES = {
  "municipal-directory": {
    title: "Municipal Telephone Directory",
    eyebrow: "CITY SERVICES · DIRECTORY OFFICE",
    headline: "Find the right public office.",
    paragraphs: ["The Meridian municipal directory lists public counters, district offices and civic services connected to the national relay.", "For corrections to a household or business listing, write to the Directory Office, 12 rue de la République, Lyon. Changes are reviewed during regular office hours."],
    notice: "AUTUMN EDITION",
  },
  "telephone-service": {
    title: "Meridian Telephone Exchange",
    eyebrow: "PUBLIC TELEPHONE SERVICE",
    headline: "An operator is ready to assist.",
    paragraphs: ["For emergency services, dial 17 for police, 15 for medical assistance, or 18 for the fire brigade.", "For other numbers, ask the city exchange for a local connection. Long-distance calls may be placed through the national operator."],
    notice: "EMERGENCY LINES REMAIN OPEN AT ALL HOURS",
  },
  "evening-courier": {
    title: "The Evening Courier",
    eyebrow: "LYON · AFTERNOON EDITION",
    headline: "The city, before the evening bell.",
    paragraphs: ["The Evening Courier has served local readers since 1896, with reporting from the courthouse, the river wards and the municipal hall.", "Today’s paper includes the council’s revised winter schedule, a notice from the central police desk and Saturday’s fixtures."],
    notice: "SINGLE COPY · 35 CENTIMES · HOME DELIVERY AVAILABLE",
  },
  "meridian-mutual": {
    title: "Meridian Mutual Assurance",
    eyebrow: "HOME · LIFE · COMMERCIAL COVER",
    headline: "Steady protection for ordinary days.",
    paragraphs: ["Meridian Mutual provides household, shop and accident policies through local agents across the connected cities.", "Policy holders may visit the nearest branch to update a beneficiary, report damage or request a copy of their agreement."],
    notice: "LOCAL AGENTS AVAILABLE MONDAY—SATURDAY",
  },
  "leclerc-locksmith": {
    title: "Leclerc Lock & Key",
    eyebrow: "LOCKSMITHS · ESTABLISHED 1932",
    headline: "A sound lock is a quiet comfort.",
    paragraphs: ["Leclerc Lock & Key supplies mortice locks, brass keys and reinforced door fittings for homes and offices.", "Repairs are available by appointment. Please bring the original key when requesting a duplicate."],
    notice: "17 PLACE BELLECOUR · OPEN 8:30 AM—6:00 PM",
  },
};

export function deepLinkPage(linkId) {
  const page = PAGES[linkId];
  return page ? { id: "deep-link", linkId, label: page.title.toUpperCase(), available: true } : null;
}

export default function DeepLinkWebsite({ linkId }) {
  const page = PAGES[linkId] || PAGES["municipal-directory"];
  return <article className="legacy-site__section legacy-police__article">
    <p className="legacy-site__eyebrow">{page.eyebrow}</p>
    <h1>{page.headline}</h1>
    {page.paragraphs.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
    <aside className="hospital-website__notice"><b>INFORMATION DESK</b><p>{page.notice}</p></aside>
  </article>;
}
