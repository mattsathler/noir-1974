import { useRef, useState } from "react";
import { searchCompanies, searchLocations, searchPeople } from "../../api.js";

export default function WorldDirectory({ program }) {
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState([]);
  const [companies, setCompanies] = useState([]);
  const [locations, setLocations] = useState([]);
  const [message, setMessage] = useState("Enter a name or address to search the city records.");
  const [busy, setBusy] = useState(false);
  const requestId = useRef(0);
  const isInterpol = program === "interpol";

  async function submitSearch(event) {
    event.preventDefault();
    const term = query.trim();
    if (!term) {
      setMessage("Enter a search term to continue.");
      return;
    }
    const currentRequest = ++requestId.current;
    setBusy(true);
    setMessage("");
    setPeople([]);
    setCompanies([]);
    setLocations([]);
    try {
      if (isInterpol) {
        const results = await searchPeople(term);
        if (currentRequest !== requestId.current) return;
        setPeople(results);
        setMessage(results.length ? "PERSONNEL FILES RETRIEVED" : "NO MATCHING PERSONNEL FILES");
      } else {
        const [companyResults, locationResults] = await Promise.all([searchCompanies(term), searchLocations(term)]);
        if (currentRequest !== requestId.current) return;
        setCompanies(companyResults);
        setLocations(locationResults);
        setMessage(companyResults.length || locationResults.length ? "CITY RECORDS RETRIEVED" : "NO MATCHING CITY RECORDS");
      }
    } catch (error) {
      if (currentRequest === requestId.current) setMessage(error.message || "The directory could not be reached.");
    } finally {
      if (currentRequest === requestId.current) setBusy(false);
    }
  }

  return (
    <div className="world-directory">
      <p className="computer-window__eyebrow">{isInterpol ? "FORMER INTERPOL AFFILIATION · LEGACY ACCESS" : "LYON MUNICIPAL ARCHIVE · PUBLIC INDEX"}</p>
      <h1>{isInterpol ? "INTERPOL PERSONNEL SEARCH" : "CITY RECORDS SEARCH"}</h1>
      <form className="world-directory__search" onSubmit={submitSearch}>
        <label className="world-directory__label" htmlFor="world-directory-query">{isInterpol ? "PERSON NAME" : "COMPANY OR ADDRESS"}</label>
        <div className="world-directory__searchline">
          <input id="world-directory-query" autoComplete="off" maxLength="80" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={isInterpol ? "Search personnel files" : "Search companies and locations"} />
          <button type="submit" disabled={busy}>{busy ? "SEARCHING…" : "SEARCH"}</button>
        </div>
      </form>
      <p className="world-directory__message" role="status" aria-live="polite">{message}</p>

      {people.map(({ npc, employment, personalRelationships }) => (
        <article className="world-record" key={npc.id}>
          <header className="world-record__header"><h2>{npc.name}</h2><span>{npc.status?.toUpperCase() || "PERSON OF INTEREST"}</span></header>
          <p>{npc.age} YEARS · {npc.appearance}</p>
          {employment.map(({ company, role }) => <p key={company.id}>EMPLOYMENT · {role.toUpperCase()} AT {company.name}</p>)}
          {personalRelationships.map(({ npc: related, relationshipType }) => <p key={`${related.id}-${relationshipType}`}>ASSOCIATION · {relationshipType.replaceAll("_", " ").toUpperCase()} — {related.name}</p>)}
        </article>
      ))}

      {companies.map(({ company, address, employees }) => (
        <article className="world-record" key={company.id}>
          <header className="world-record__header"><h2>{company.name}</h2><span>COMPANY</span></header>
          <p>HEADQUARTERS · {address.full}</p>
          <p>PERSONNEL ON FILE · {employees.length}</p>
          {employees.map((employee) => <p className="world-record__detail" key={employee.id}>{employee.name} · {employee.status.toUpperCase()}</p>)}
        </article>
      ))}

      {locations.map(({ address }) => (
        <article className="world-record" key={address.locationId}>
          <header className="world-record__header"><h2>{address.name}</h2><span>{address.kind.toUpperCase()}</span></header>
          <p>{address.full}</p>
        </article>
      ))}
    </div>
  );
}
