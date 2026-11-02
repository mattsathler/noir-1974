import { useEffect, useState } from "react";
import { Users } from "pixelarticons/react/Users";
import { getCompanyWebsite } from "../../../world/api.js";
import CompanyWebsite from "./CompanyWebsite.jsx";
import WorldDirectory from "../../../world/pages/directory/WorldDirectory.jsx";

export default function EmployeeDirectoryApp({ companyRecord, onOpenCompanyStaff, onPrint }) {
  const [directory, setDirectory] = useState(null);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [securityStatus, setSecurityStatus] = useState("");
  const company = companyRecord?.company || companyRecord?.Company;
  const address = companyRecord?.address || companyRecord?.Address;

  useEffect(() => {
    let active = true;
    if (!company?.id || !address?.locationId) {
      setError("THIS EMPLOYEE DIRECTORY HAS NO LINKED COMPANY SITE.");
      setBusy(false);
      return () => { active = false; };
    }
    setBusy(true);
    setError("");
    setSecurityStatus("");
    getCompanyWebsite(company.id)
      .then(async (website) => {
        const protection = website?.cyberSecurity;
        if (protection?.level > 0) setSecurityStatus(`${protection.provider.toUpperCase()} IS INSPECTING THE REMOTE REQUEST…`);
        const attempt = await onOpenCompanyStaff?.({ company, address, onStatus: (message) => { if (active) setSecurityStatus(message); } });
        if (!active) return;
        if (!website || !attempt?.attemptId) throw new Error("EMPLOYEE DIRECTORY ACCESS WAS DENIED.");
        setDirectory({ website, company, address, attemptId: String(attempt.attemptId), sessionRestored: Boolean(attempt.sessionRestored), securityProvider: attempt.securityProvider || website.cyberSecurity?.provider });
      })
      .catch((cause) => { if (active) setError(cause.message || "THE EMPLOYEE DIRECTORY COULD NOT BE OPENED."); })
      .finally(() => { if (active) setBusy(false); });
    return () => { active = false; };
  }, [address, company, onOpenCompanyStaff]);

  if (directory) return (
    <CompanyWebsite site={directory.website} pageId="staff">
      {directory.sessionRestored && <p className="company-website__security-event" role="status">
        {directory.securityProvider?.toUpperCase() || "CORPORATE FIREWALL"} INTERRUPTED THE FIRST CONNECTION. A NEW SESSION IS ACTIVE.
      </p>}
      <WorldDirectory
        key={`${directory.company.id}:${directory.attemptId}`}
        program="staff-directory"
        companyLocationId={directory.address.locationId}
        companyId={directory.company.id}
        accessAttemptId={directory.attemptId}
        onPrint={onPrint}
      />
    </CompanyWebsite>
  );

  return (
    <main className="employee-directory-app" aria-busy={busy}>
      <header className="employee-directory-app__header">
        <Users aria-hidden="true" />
        <div><p>GHOSTKEY.EXE · REMOTE PERSONNEL FILES</p><h1>Employee Directory</h1><span>{company?.name || "Company site unavailable"}</span></div>
      </header>
      <p className="employee-directory-app__target">TARGET SITE · {companyRecord?.siteAddress || "HOST ADDRESS UNAVAILABLE"}</p>
      {busy && <p className="employee-directory-app__status" role="status">{securityStatus || `CONNECTING TO ${company?.name?.toUpperCase() || "TARGET SITE"} PERSONNEL FILES…`}</p>}
      {error && <p className="employee-directory-app__error" role="alert">{error}</p>}
    </main>
  );
}
