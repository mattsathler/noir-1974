import { useEffect, useMemo, useRef, useState } from "react";
import { Folder } from "pixelarticons/react/Folder";
import { Globe } from "pixelarticons/react/Globe";
import { Bug } from "pixelarticons/react/Bug";
import { Note } from "pixelarticons/react/Note";
import { User } from "pixelarticons/react/User";
import { Users } from "pixelarticons/react/Users";
import { Cross } from "pixelarticons/react/Cross";
import { Map as MapIcon } from "pixelarticons/react/Map";
import { getHostingProviders } from "../../../world/api.js";

const ROOT_FOLDERS = [
  { id: "bin", name: "bin", children: ["tools"] },
  { id: "www", name: "www", children: ["cgi-bin", "html"] },
  { id: "lib", name: "lib", children: ["core"] },
];
const HOST_PAGE_SIZE = 20;
const APP_BOOT_LINES = [
  "MERID_OS APPLICATION LOADER 4.1",
  "CHECKING EXECUTABLE SIGNATURE...",
  "GHOSTKEY.EXE · PUBLISHER UNKNOWN",
  "ACCESS POLICY... OVERRIDDEN",
  "ISOLATING PROCESS... READY",
];
const NETWORK_BOOT_LINES = [
  "INITIALIZING NETWORK EXPLORER...",
  "RESOLVING COUNTRY ROUTE...",
  "REQUESTING PROVIDER INDEX...",
  "VERIFYING REMOTE HOST DIRECTORY...",
];
const PROVIDER_BOOT_LINES = [
  "OPENING PROVIDER NODE...",
  "ENUMERATING HOST DIRECTORIES...",
  "READING SEEDED SITE INDEX...",
  "MAPPING REMOTE IP ADDRESSES...",
];

function BootTerminal({ lines, onComplete }) {
  const [visibleLines, setVisibleLines] = useState(0);
  const onCompleteRef = useRef(onComplete);
  onCompleteRef.current = onComplete;
  useEffect(() => {
    let active = true;
    let timer;
    let count = 0;
    function reveal() {
      if (!active) return;
      count += 1;
      setVisibleLines(count);
      if (count < lines.length) timer = window.setTimeout(reveal, 190);
      else if (onCompleteRef.current) timer = window.setTimeout(() => onCompleteRef.current?.(), 450);
    }
    timer = window.setTimeout(reveal, 150);
    return () => { active = false; window.clearTimeout(timer); };
  }, [lines]);
  return (
    <section className="ghostkey__boot" role="status" aria-label="Ghostkey startup sequence">
      <div className="ghostkey__boot-log" aria-live="polite">
        {lines.slice(0, visibleLines).map((line, index) => <span key={`${index}:${line}`}>{line}</span>)}
      </div>
    </section>
  );
}

function genericFiles(prefix, count = 3) {
  return ["cache_01.dat", "readme.txt", "index.old"].slice(0, count).map((name) => ({
    id: `${prefix}-${name}`,
    name,
    kind: "unsupported",
  }));
}

function folderEntries(directory, path, selectedCompany = null) {
  const current = path.at(-1);
  const executables = directory.files || [];
  const hasExecutable = executables.length > 0 || Boolean(selectedCompany);
  if (!current) return [
    ...ROOT_FOLDERS,
  ].map((folder) => ({
    ...folder,
    kind: "folder",
    fileCount: folder.id === "www" ? 9 + executables.length : 8,
    exeCount: folder.id === "www" ? executables.length : 0,
  }));
  if (path.length === 1) {
    const children = ROOT_FOLDERS.find((folder) => folder.id === current)?.children || [];
    return children.map((name) => ({
      id: name,
      name,
      kind: "folder",
      fileCount: hasExecutable && current === "www" && name === "cgi-bin" ? 2 + executables.length : name === "html" ? 8 : 3,
      exeCount: hasExecutable && current === "www" && name === "cgi-bin" ? executables.length : 0,
    }));
  }
  if (current === "cgi-bin" && executables.length) {
    return [
      ...genericFiles("cgi-bin", 2),
      ...executables.map((file) => ({ id: file.id, name: file.name, kind: file.kind })),
    ];
  }
  if (path.length >= 3) return genericFiles(current);
  return [
    { id: `${current}-archive`, name: "archive", kind: "folder", fileCount: 3, exeCount: 0 },
    { id: `${current}-data`, name: "data", kind: "folder", fileCount: 3, exeCount: 0 },
    ...genericFiles(current, 2),
  ];
}

function ProviderIPSearch({ onSearch }) {
  const [query, setQuery] = useState("");
  function submit(event) {
    event.preventDefault();
    onSearch(query.trim());
  }
  return (
    <form className="ghostkey__directory-search" onSubmit={submit}>
      <label htmlFor="ghostkey-ip-search">FILTER BY IP ADDRESS</label>
      <div><input id="ghostkey-ip-search" value={query} onChange={(event) => setQuery(event.target.value)} inputMode="decimal" autoComplete="off" spellCheck="false" placeholder="123.224.72.118" /><button type="submit">SEARCH ↵</button></div>
    </form>
  );
}

export default function GhostkeyExplorer({ countryId, countryName, onOpenPersonSearch, onOpenPatientSearch, onOpenGateRecords, onOpenEmployeeDirectory, onOpenWarrantRequest }) {
  const [appView, setAppView] = useState("boot");
  const [providers, setProviders] = useState([]);
  const [providerId, setProviderId] = useState("");
  const [directoryId, setDirectoryId] = useState("");
  const [selectedCompany, setSelectedCompany] = useState(null);
  const [path, setPath] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingStage, setLoadingStage] = useState("network");
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const [unsupportedMessage, setUnsupportedMessage] = useState("");
  const [hostFilter, setHostFilter] = useState("");
  const [visibleHostCount, setVisibleHostCount] = useState(HOST_PAGE_SIZE);

  useEffect(() => {
    if (appView !== "network") return undefined;
    if (!countryId) {
      setProviders([]);
      setLoading(false);
      setError("NO COUNTRY ROUTE IS AVAILABLE.");
      return undefined;
    }
    let active = true;
    let completionTimer;
    const startedAt = Date.now();
    setLoading(true);
    setError("");
    setProviderId("");
    setDirectoryId("");
    setSelectedCompany(null);
    setPath([]);
    getHostingProviders(countryId)
      .then((items) => { if (active) setProviders(items); })
      .catch((cause) => { if (active) setError(cause.message || "THE PROVIDER INDEX COULD NOT BE READ."); })
      .finally(() => {
        const remainingBootTime = Math.max(0, 1170 - (Date.now() - startedAt));
        completionTimer = window.setTimeout(() => { if (active) setLoading(false); }, remainingBootTime);
      });
    return () => { active = false; window.clearTimeout(completionTimer); };
  }, [appView, countryId, retry]);

  const selectedProvider = providers.find((provider) => provider.id === providerId);
  const selectedDirectory = selectedProvider?.directories.find((directory) => directory.id === directoryId);
  const siteAddress = selectedCompany?.siteAddress || selectedDirectory?.siteAddress || selectedDirectory?.address;
  const entries = useMemo(() => selectedDirectory ? folderEntries(selectedDirectory, path, selectedCompany) : [], [selectedDirectory, path, selectedCompany]);
  const normalizedHostFilter = hostFilter.replaceAll(".", "");
  const filteredProviderDirectories = useMemo(
    () => selectedProvider?.directories.filter((directory) => directory.ip.replaceAll(".", "").includes(normalizedHostFilter)) ?? [],
    [selectedProvider, normalizedHostFilter],
  );
  const visibleProviderDirectories = useMemo(
    () => filteredProviderDirectories.slice(0, visibleHostCount),
    [filteredProviderDirectories, visibleHostCount],
  );

  async function openProvider(provider) {
    const startedAt = Date.now();
    setLoading(true);
    setLoadingStage("provider");
    setError("");
    setHostFilter("");
    setVisibleHostCount(HOST_PAGE_SIZE);
    setDirectoryId("");
    setSelectedCompany(null);
    setPath([]);
    try {
      const loadedProviders = await getHostingProviders(countryId, provider.id === "northstar" ? provider.id : "");
      setProviders(loadedProviders);
      setProviderId(provider.id);
    } catch (cause) {
      setError(cause.message || "THE PROVIDER DIRECTORY COULD NOT BE READ.");
    } finally {
      await new Promise((resolve) => window.setTimeout(resolve, Math.max(0, 1170 - (Date.now() - startedAt))));
      setLoading(false);
    }
  }
  function openDirectory(directory) {
    setDirectoryId(directory.id);
    setSelectedCompany(directory.company ? { company: directory.company, address: directory.companyAddress, siteAddress: directory.address } : null);
    setPath([]);
  }
  function goTo(level, index = -1) {
    if (level === "home") {
      setAppView("home");
      setLoading(false);
      setError("");
      setProviderId("");
      setDirectoryId("");
      setSelectedCompany(null);
      setPath([]);
    } else if (level === "network") {
      setProviderId("");
      setDirectoryId("");
      setSelectedCompany(null);
      setPath([]);
    } else if (level === "provider") {
      setDirectoryId("");
      setSelectedCompany(null);
      setPath([]);
    } else if (level === "host") setPath([]);
    else setPath((current) => current.slice(0, index + 1));
  }
  function openNetworkExplorer() {
    setProviders([]);
    setProviderId("");
    setDirectoryId("");
    setSelectedCompany(null);
    setPath([]);
    setHostFilter("");
    setVisibleHostCount(HOST_PAGE_SIZE);
    setError("");
    setLoadingStage("network");
    setAppView("network");
  }
  function openEntry(entry) {
    setUnsupportedMessage("");
    if (entry.kind === "person-search") {
      onOpenPersonSearch?.(selectedDirectory.address);
      return;
    }
    if (entry.kind === "patient-search") {
      onOpenPatientSearch?.(selectedDirectory.siteAddress || selectedDirectory.address);
      return;
    }
    if (entry.kind === "gate-records") {
      Promise.resolve(onOpenGateRecords?.(selectedDirectory.siteAddress || selectedDirectory.address))
        .catch((cause) => setError(cause.message || "THE GATE ACCESS LOG APPLICATION COULD NOT BE OPENED."));
      return;
    }
    if (entry.kind === "employee-directory") {
      onOpenEmployeeDirectory?.({ ...selectedCompany, siteAddress });
      return;
    }
    if (entry.kind === "warrant-request") {
      onOpenWarrantRequest?.(selectedDirectory.siteAddress || selectedDirectory.address);
      return;
    }
    if (entry.kind === "unsupported") {
      setUnsupportedMessage(entry.name);
      return;
    }
    setPath((current) => [...current, entry.id]);
  }

  const breadcrumbs = path.map((id, index) => ({ id, label: index === 0 ? id : ROOT_FOLDERS.find((folder) => folder.id === path[0])?.children.find((child) => child === id) || id }));
  const pageTitle = appView === "home" ? "Tool Console" : appView === "network" ? "Network Explorer" : "Application Startup";
  return (
    <main className="ghostkey" aria-label="Ghostkey remote access tool" aria-busy={loading || appView === "boot"}>
      <header className="ghostkey__header">
        <div className="ghostkey__brand">
          <span className="ghostkey__mark" aria-hidden="true"><Bug /></span>
          <div className="ghostkey__identity"><p><strong>GHOSTKEY.EXE</strong><span> / ACCESS BYPASS</span></p><h1>{pageTitle}</h1><div className="ghostkey__channel"><i aria-hidden="true" /> {appView === "boot" ? "PROCESS INITIALIZING" : appView === "network" ? "BREACH CHANNEL ACTIVE" : "BREACH CHANNEL READY"} <b>v2.2</b></div></div>
        </div>
        <div className="ghostkey__country"><small>ROUTE NODE</small><strong>{countryName || countryId || "NO ROUTE"}</strong></div>
      </header>
      {appView === "home" && <nav className="ghostkey__breadcrumb" aria-label="Current folder"><span aria-current="page">TOOLS / HOME</span></nav>}
      {appView === "network" && <nav className="ghostkey__breadcrumb" aria-label="Current folder">
        <button type="button" onClick={() => goTo("home")}>TOOLS</button><span>/</span><button type="button" onClick={() => goTo("network")} aria-current={!selectedProvider ? "page" : undefined}>NETWORK</button>
        {selectedProvider && <><span>/</span><button type="button" onClick={() => goTo("provider")} aria-current={!selectedDirectory ? "page" : undefined}>{selectedProvider.host.toUpperCase()}</button></>}
        {selectedDirectory && <><span>/</span><button type="button" onClick={() => goTo("host")} aria-current={!path.length ? "page" : undefined}>{selectedCompany?.company?.name || selectedDirectory.ip}</button></>}
        {breadcrumbs.map((item, index) => <span key={`${item.id}-${index}`} className="ghostkey__crumb-segment"><span>/</span><button type="button" onClick={() => goTo("path", index)} aria-current={index === path.length - 1 ? "page" : undefined}>{item.label.toUpperCase()}</button></span>)}
        {!selectedProvider && <><span>/</span><span aria-current="page">{(countryId || "NO ROUTE").toUpperCase()}</span></>}
      </nav>}
      {appView === "boot" && <BootTerminal lines={APP_BOOT_LINES} onComplete={() => setAppView("home")} />}
      {appView === "home" && <>
        <div className="ghostkey__section-title"><div><h2>Available tools</h2><span>SELECT A MODULE TO CONTINUE</span></div><span>01 READY</span></div>
        <div className="ghostkey__tool-grid">
          <button type="button" className="ghostkey__tool" onClick={openNetworkExplorer}>
            <span className="ghostkey__tool-icon"><Globe aria-hidden="true" /></span>
            <span className="ghostkey__tool-copy"><small>MODULE 01 · NETWORK ACCESS</small><strong>Network Explorer</strong><span>Inspect provider nodes and remote host directories.</span></span>
            <b>OPEN TOOL →</b>
          </button>
        </div>
      </>}
      {appView === "network" && (loading
        ? <BootTerminal lines={loadingStage === "provider" ? PROVIDER_BOOT_LINES : NETWORK_BOOT_LINES} />
        : error ? <section className="ghostkey__empty"><p role="alert">{error}</p><button type="button" onClick={() => setRetry((attempt) => attempt + 1)}>RETRY CONNECTION</button></section>
          : !selectedProvider ? <>
            <div className="ghostkey__section-title"><h2>Network providers</h2><span>{providers.length} FOUND</span></div>
            <div className="ghostkey__entries ghostkey__entries--icons" role="list" aria-label="Country providers">
              {providers.map((provider) => <button className="ghostkey__entry ghostkey__entry--icon" type="button" role="listitem" key={provider.id} onClick={() => openProvider(provider)}>
                <Folder className="ghostkey__folder-icon" aria-hidden="true" /><span className="ghostkey__entry-copy"><strong>{provider.name}</strong><small>{provider.host}.{provider.providerIp}:4000 · {provider.directories.length} HOSTS</small></span>
              </button>)}
            </div>
          </> : !selectedDirectory ? <>
            <div className="ghostkey__section-title"><h2>{selectedProvider.name}</h2><span>{hostFilter ? `${filteredProviderDirectories.length} / ` : ""}{selectedProvider.directories.length} HOST DIRECTORIES</span></div>
            <ProviderIPSearch onSearch={(query) => { setHostFilter(query); setVisibleHostCount(HOST_PAGE_SIZE); }} />
            <div className="ghostkey__entries ghostkey__entries--icons" role="list" aria-label={`${selectedProvider.name} host folders`}>
              {visibleProviderDirectories.map((directory) => <button className="ghostkey__entry ghostkey__entry--icon" type="button" role="listitem" key={directory.company?.id || directory.id} onClick={() => openDirectory(directory)}>
                <Folder className="ghostkey__folder-icon" aria-hidden="true" /><span className="ghostkey__entry-copy"><strong>{directory.ip}</strong><small>{directory.siteName || "HOST DIRECTORY"} · 3 FOLDERS</small></span>
              </button>)}
            </div>
            {visibleHostCount < filteredProviderDirectories.length && <button className="ghostkey__load-more" type="button" onClick={() => setVisibleHostCount((count) => count + HOST_PAGE_SIZE)}>LOAD MORE · 20</button>}
          </> : <>
            <div className="ghostkey__section-title"><div><h2>{path.length ? path.at(-1).toUpperCase() : selectedCompany?.company?.name || selectedDirectory.siteName || "REMOTE HOST"}</h2><span>{siteAddress}</span></div><span>{entries.length} ITEMS</span></div>
            <div className="ghostkey__entries ghostkey__entries--icons" role="list" aria-label="Remote folders and files">
              {entries.map((entry) => entry.kind === "folder" ? <button className="ghostkey__entry ghostkey__entry--icon" type="button" role="listitem" key={entry.id} onClick={() => openEntry(entry)}>
                <Folder className="ghostkey__folder-icon" aria-hidden="true" /><span className="ghostkey__entry-copy"><strong>{entry.name}</strong><small>{entry.fileCount} FILES · {entry.exeCount} EXES</small></span>
              </button> : <button className={`ghostkey__entry ghostkey__entry--icon${["person-search", "warrant-request"].includes(entry.kind) ? " ghostkey__entry--executable" : " ghostkey__entry--file"}`} type="button" role="listitem" key={entry.id} onClick={() => openEntry(entry)}>
                {entry.kind === "person-search" ? <User className="ghostkey__file-icon ghostkey__file-icon--person" aria-hidden="true" /> : entry.kind === "patient-search" ? <Cross className="ghostkey__file-icon ghostkey__file-icon--person" aria-hidden="true" /> : entry.kind === "gate-records" ? <MapIcon className="ghostkey__file-icon" aria-hidden="true" /> : entry.kind === "employee-directory" ? <Users className="ghostkey__file-icon ghostkey__file-icon--person" aria-hidden="true" /> : <Note className="ghostkey__file-icon" aria-hidden="true" />}<span className="ghostkey__entry-copy"><strong>{entry.name}</strong><small>{entry.kind === "person-search" ? "POLICE RECORDS APPLICATION" : entry.kind === "patient-search" ? "HOSPITAL PATIENT RECORDS APPLICATION" : entry.kind === "gate-records" ? "GATE TRANSIT ACCESS LOG APPLICATION" : entry.kind === "employee-directory" ? "CORPORATE APPLICATION" : entry.kind === "warrant-request" ? "RESIDENTIAL SEARCH" : "DATA FILE"}</small></span>
              </button>)}
            </div>
            {unsupportedMessage && <div className="ghostkey__error-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setUnsupportedMessage(""); }}>
              <section className="ghostkey__error-dialog" role="alertdialog" aria-modal="true" aria-labelledby="ghostkey-error-title" aria-describedby="ghostkey-error-message" onKeyDown={(event) => { if (event.key === "Escape") setUnsupportedMessage(""); }}>
                <header className="ghostkey__error-titlebar"><span><i aria-hidden="true">!</i> MERID_OS — SYSTEM ERROR</span><button type="button" aria-label="Close error" onClick={() => setUnsupportedMessage("")}>×</button></header>
                <div className="ghostkey__error-content"><strong id="ghostkey-error-title">Unsupported file type</strong><p id="ghostkey-error-message">MERID_OS cannot open “{unsupportedMessage}”.</p></div>
                <footer><button type="button" autoFocus onClick={() => setUnsupportedMessage("")}>OK</button></footer>
              </section>
            </div>}
          </>)}
      <footer className="ghostkey__footer"><span><i aria-hidden="true" /> GHOSTKEY SESSION · {selectedProvider ? "REMOTE HOST" : "LOCAL NETWORK"}</span><span>COUNTRY NODE {(countryId || "NO ROUTE").toUpperCase()}</span></footer>
    </main>
  );
}
