import { useEffect, useState } from "react";
import { getNeighborhoodDestinations } from "../../api.js";
import APIErrorNotice from "../../../../shared/APIErrorNotice.jsx";

function serviceKind(address) {
  const identity = `${address.locationId} ${address.name}`.toLowerCase();
  if (identity.includes("police")) return "police";
  if (identity.includes("hospital")) return "hospital";
  if (identity.includes("bank")) return "bank";
  if (identity.includes("gate")) return "gate";
  return "";
}

export default function PublicServices({ cityId = "", countryName = "", onOpenService, onOpenHospitalWebsite }) {
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [busyId, setBusyId] = useState("");
  const [retry, setRetry] = useState(0);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    getNeighborhoodDestinations(cityId).then((districts) => {
      if (!active) return;
      const records = new Map();
      for (const district of districts) {
        for (const { address } of district.locations ?? []) {
          if (!address.locationId?.startsWith("service:")) continue;
          const kind = serviceKind(address);
          if (kind) records.set(address.locationId, { address, kind });
        }
      }
      setServices([...records.values()].sort((left, right) => left.address.name.localeCompare(right.address.name)));
    }).catch((requestError) => {
      if (active) setError(requestError);
    }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [cityId, retry]);

  async function openService(service) {
    if ((!onOpenService && !onOpenHospitalWebsite) || busyId) return;
    setBusyId(service.address.locationId);
    try {
      if (service.kind === "hospital" && onOpenHospitalWebsite) await onOpenHospitalWebsite(service);
      else await onOpenService?.(service);
    } catch (requestError) {
      setError(requestError);
    } finally {
      setBusyId("");
    }
  }

  return <div className="world-directory public-services">
    <p className="computer-window__eyebrow">MERIDIAN NETWORK · CITY INSTITUTION DIRECTORY</p>
    <h1>CITY SERVICES DIRECTORY</h1>
    <p className="world-directory__scope">CITY NODE · {cityId.split("/").at(-1)?.replaceAll("-", " ").toUpperCase()} · {countryName.toUpperCase()}</p>
    {loading && <p className="world-directory__message" role="status">LOADING CITY SERVICE SITES…</p>}
    {error && <APIErrorNotice error={error} onRetry={() => setRetry((value) => value + 1)} className="public-services__error" />}
    {!loading && !error && services.length === 0 && <p className="world-directory__message">NO PUBLIC SERVICE SITES ARE LISTED FOR THIS CITY.</p>}
    {services.map((service) => <article className="world-record public-services__record" key={service.address.locationId}>
      <header className="world-record__header"><h2>{service.address.name}</h2><span>{service.kind.toUpperCase()} SITE</span></header>
      <p>{service.address.full}</p>
      <button type="button" className="world-company-site__open" disabled={Boolean(busyId)} onClick={() => openService(service)}>
        {busyId === service.address.locationId ? "OPENING…" : service.kind === "hospital" && onOpenHospitalWebsite ? "VISIT HOSPITAL WEBSITE" : "CONNECT TO SITE"}<span aria-hidden="true">↗</span>
      </button>
    </article>)}
  </div>;
}
