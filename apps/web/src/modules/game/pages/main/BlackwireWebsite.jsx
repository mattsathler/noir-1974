import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { getBlackwireMarket, markWhiteboardItemSold, registerBlackwireProfile, registerNewBlackwireIdentity, sellBlackwireItem } from "../../../world/api.js";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import "./BlackwireWebsite.css";

export const BLACKWIRE_PORTRAITS = [
  { id: "portrait-01", name: "Relay Standard", visual: { hairId: "hair_short_brown", skinId: "skin_white", eyesId: "eyes_brown", beardId: "", faceId: "", accessoryIds: [] } },
  { id: "portrait-02", name: "Night Shift", visual: { hairId: "hair_long_black", skinId: "skin_dark", eyesId: "eyes_green", beardId: "", faceId: "", accessoryIds: [] } },
  { id: "portrait-03", name: "Operator", visual: { hairId: "hair_slicked_black", skinId: "skin_olive", eyesId: "eyes_gray", beardId: "beard_moustache_black", faceId: "", accessoryIds: [] } },
  { id: "portrait-04", name: "Courier", visual: { hairId: "hair_wavy_ginger", skinId: "skin_medium", eyesId: "eyes_hazel", beardId: "", faceId: "", accessoryIds: [] } },
  { id: "portrait-05", name: "Switchboard", visual: { hairId: "hair_bob_blonde", skinId: "skin_pale", eyesId: "eyes_blue", beardId: "", faceId: "", accessoryIds: [] } },
  { id: "portrait-06", name: "Unknown", visual: { hairId: "hair_mohawk_black", skinId: "skin_brown", eyesId: "eyes_amber", beardId: "", faceId: "", accessoryIds: [] } },
];

const CROW_DISPATCHES = [
  {
    id: "footprints",
    category: "FIELD NOTES · NETWORKS",
    date: "29 OCTOBER 1974",
    title: "Every door leaves a mark",
    standfirst: "A route is not a disguise. It is a record someone else may keep.",
    paragraphs: [
      "Directories tell you where a service lives. They also remember who came asking. Learn what a system publishes before you touch it, and notice which requests leave a name, a time or a return address behind.",
      "Change your route when the work calls for it. Never mistake a different exit for permission, and never assume a quiet connection is an invisible one.",
    ],
  },
  {
    id: "corroboration",
    category: "PRACTICE · RECORDS",
    date: "27 OCTOBER 1974",
    title: "Read the margin before the headline",
    standfirst: "One record can point. Two independent records can begin to tell a story.",
    paragraphs: [
      "Write down who made a record, when it was made and what the source could actually know. A company ledger, a witness and a newspaper do not carry the same weight, even when their dates match.",
      "Keep the gaps visible. A blank line is a question, not an invitation to invent an answer.",
    ],
  },
  {
    id: "infrastructure",
    category: "EXPOSÉ · MERIDIAN",
    date: "24 OCTOBER 1974",
    title: "The wire has an owner",
    standfirst: "Meridian calls the relay a public service. Its switches still answer to private offices.",
    paragraphs: [
      "A network can connect a city and still decide which doors open easily. The same infrastructure carries public notices, company correspondence and the quiet logs that make movement legible to whoever holds the keys.",
      "Ask who maintains the line, who pays for it and who can read its traffic. That is where the map becomes politics.",
    ],
  },
];

export default function BlackwireWebsite({ profile = null, onAccountChange = () => {} }) {
  const siteRef = useRef(null);
  const [market, setMarket] = useState({ offers: [] });
  const [boardFilter, setBoardFilter] = useState("all");
  const [windowHost, setWindowHost] = useState(null);
  const [selected, setSelected] = useState(null);
  const [loading, setLoading] = useState(true);
  const [selling, setSelling] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [section, setSection] = useState("hub");
  const [selectedAppId, setSelectedAppId] = useState("");
  const [username, setUsername] = useState("");
  const [avatarId, setAvatarId] = useState("portrait-01");
  const [profileBusy, setProfileBusy] = useState(false);

  const boardGroups = Array.from(market.offers.reduce((groups, offer) => {
    const group = groups.get(offer.boardId) || { id: offer.boardId, title: offer.boardTitle, offers: [] };
    group.offers.push(offer);
    groups.set(offer.boardId, group);
    return groups;
  }, new Map()).values());
  const visibleBoardGroups = boardFilter === "all" ? boardGroups : boardGroups.filter((board) => board.id === boardFilter);
  const selectedApplication = market.applications?.find((app) => app.id === selectedAppId) || null;
  const selectedApplicationCommands = selectedApplication
    ? [
      ...(selectedApplication.commands || []),
      ...(selectedApplication.id !== "blackwire" ? [`uninstall ${selectedApplication.id}.exe`] : []),
    ]
    : [];

  useLayoutEffect(() => {
    setWindowHost(siteRef.current?.closest(".computer-window") || null);
  }, []);

  useEffect(() => {
    if (boardFilter !== "all" && !boardGroups.some((board) => board.id === boardFilter)) setBoardFilter("all");
  }, [boardFilter, market.offers]);

  useEffect(() => {
    if (!profile) return;
    setMarket((current) => ({
      ...current,
      account: profile,
      applications: current.applications?.map((app) => ({
        ...app,
        installed: profile.installedApplications?.includes(app.id),
        fetched: profile.fetchedApplications?.includes(app.id),
      })) || current.applications,
    }));
  }, [profile]);

  useEffect(() => {
    let active = true;
    getBlackwireMarket().then((result) => {
      if (!active) return;
      setMarket(result);
      onAccountChange(result.account || null);
    })
      .catch(() => { if (active) setError("The exchange could not load your investigation boards."); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [onAccountChange]);

  async function confirmSale() {
    if (!selected || selling) return;
    setSelling(true);
    setError("");
    try {
      const result = await sellBlackwireItem(selected.boardId, selected.itemId);
      if (result?.account) {
        setMarket((current) => ({ ...current, account: result.account }));
        onAccountChange(result.account);
      }
      markWhiteboardItemSold(selected.itemId);
      setMarket((current) => ({ ...current, offers: current.offers.filter((offer) => offer.itemId !== selected.itemId) }));
      setMessage("SALE COMPLETE. CHECK PERSONAL MAIL FOR THE SETTLEMENT AND AMOUNT RECEIVED.");
      setSelected(null);
    } catch {
      setError("The offer changed or was withdrawn. Close this review and refresh the market.");
    } finally {
      setSelling(false);
    }
  }

  async function createNewIdentity() {
    if (registering) return;
    setRegistering(true);
    setError("");
    try {
      const result = await registerNewBlackwireIdentity();
      setMarket((current) => ({ ...current, account: result.account, offers: [] }));
      onAccountChange(result.account || null);
      setUsername("");
      setAvatarId("portrait-01");
      setMessage("NEW IDENTITY CREATED. CHOOSE A HANDLE AND PORTRAIT TO JOIN THE RELAY.");
    } catch {
      setError("A NEW IDENTITY COULD NOT BE REGISTERED. RELOAD THE RELAY AND TRY AGAIN.");
    } finally {
      setRegistering(false);
    }
  }

  async function submitProfile(event) {
    event.preventDefault();
    if (profileBusy) return;
    setProfileBusy(true);
    setError("");
    try {
      const result = await registerBlackwireProfile(username, avatarId);
      setMarket((current) => ({ ...current, account: result.account }));
      onAccountChange(result.account || null);
      setMessage("PROFILE REGISTERED. THE RELAY IS YOURS TO USE.");
      setSection("hub");
    } catch (cause) {
      setError(cause.message || "THAT HANDLE IS UNAVAILABLE. CHOOSE ANOTHER.");
    } finally {
      setProfileBusy(false);
    }
  }

  const selectedPortrait = BLACKWIRE_PORTRAITS.find((portrait) => portrait.id === market.account?.avatarId) || BLACKWIRE_PORTRAITS[0];

  return <main className="blackwire-site" ref={siteRef} aria-label="Blackwire clandestine network">
    <header className="blackwire-site__masthead">
      <div className="blackwire-site__mark" aria-hidden="true">BW</div>
      <div className="blackwire-site__identity"><span>UNLISTED NETWORK · EST. BY THE CROW</span><h1>BLACKWIRE</h1></div>
      <span className="blackwire-site__connection"><i aria-hidden="true" /> PRIVATE RELAY</span>
    </header>
    <nav className="blackwire-site__nav" aria-label="Blackwire sections">
      <button type="button" aria-current={section === "hub" ? "page" : undefined} onClick={() => { setSection("hub"); setSelectedAppId(""); }}>THE CROW'S DESK</button><button type="button" aria-current={section === "market" ? "page" : undefined} onClick={() => { setSection("market"); setSelectedAppId(""); }}>INFORMATION EXCHANGE</button><button type="button" aria-current={section === "downloads" ? "page" : undefined} onClick={() => { setSection("downloads"); setSelectedAppId(""); }}>SOFTWARE MARKETPLACE</button><button type="button" aria-current={section === "archive" ? "page" : undefined} onClick={() => { setSection("archive"); setSelectedAppId(""); }}>RESTRICTED ARCHIVE</button>
    </nav>
    {market.account && <section className="blackwire-site__account" aria-label="Blackwire account">
      <figure className="blackwire-site__profile-portrait"><ModularFace visual={selectedPortrait.visual} name={selectedPortrait.name} /></figure>
      <div><span className="blackwire-site__eyebrow">RELAY IDENTITY</span><strong>{market.account.profileComplete ? market.account.handle : "PROFILE NOT REGISTERED"}</strong><span className="blackwire-site__account-note">{market.account.profileComplete ? "REGISTERED MEMBER · PRIVATE SESSION" : "CHOOSE A HANDLE TO JOIN THE RELAY"}</span></div>
    </section>}
    {!loading && market.account && !market.account.profileComplete ? <section className="blackwire-site__profile-form" aria-labelledby="blackwire-profile-title">
      <p className="blackwire-site__eyebrow">RELAY REGISTRATION · NO PASSWORD REQUIRED</p>
      <h2 id="blackwire-profile-title">Choose your handle.</h2>
      <p>This handle and portrait identify you on this server. Your Meridian account remains your only sign-in.</p>
      <form onSubmit={submitProfile}>
        <label htmlFor="blackwire-username">USERNAME · 3–24 SYMBOLS</label>
        <input id="blackwire-username" autoComplete="off" maxLength={24} value={username} onChange={(event) => setUsername(event.target.value)} placeholder="Choose a handle" pattern="[A-Za-z0-9][A-Za-z0-9._-]{2,23}" required />
        <div className="blackwire-site__portrait-grid" role="radiogroup" aria-label="Choose a Blackwire portrait">{BLACKWIRE_PORTRAITS.map((portrait) => <button className={avatarId === portrait.id ? "blackwire-site__portrait-choice blackwire-site__portrait-choice--selected" : "blackwire-site__portrait-choice"} type="button" key={portrait.id} role="radio" aria-checked={avatarId === portrait.id} onClick={() => setAvatarId(portrait.id)}><span><ModularFace visual={portrait.visual} name={portrait.name} /></span><small>{portrait.name}</small></button>)}</div>
        {error && <p className="blackwire-site__notice" role="alert">{error}</p>}
        <button className="blackwire-site__link-action" type="submit" disabled={profileBusy}>{profileBusy ? "REGISTERING…" : "REGISTER PROFILE ↵"}</button>
      </form>
    </section> : null}
    {section === "hub" && <section className="blackwire-site__hub" aria-labelledby="blackwire-hub-title">
      <header className="blackwire-site__hub-heading"><div><p className="blackwire-site__eyebrow">NOTES FROM AN UNLISTED DESK</p><h2 id="blackwire-hub-title">The Crow's desk</h2></div><span>READ CLOSELY · THINK FOR YOURSELF</span></header>
      <article className="blackwire-site__dispatch blackwire-site__dispatch--featured">
        <p className="blackwire-site__eyebrow">A MESSAGE FROM THE CROW · {CROW_DISPATCHES[0].date}</p>
        <h3>Information has a price. Attention has a cost.</h3>
        <p>Blackwire is a relay for field notes, inconvenient records and tools that travel without a proper letterhead. Keep your questions sharper than your certainty.</p>
      </article>
      <div className="blackwire-site__dispatches">{CROW_DISPATCHES.map((dispatch) => <article className="blackwire-site__dispatch" key={dispatch.id}>
        <p className="blackwire-site__eyebrow">{dispatch.category} · {dispatch.date}</p>
        <h3>{dispatch.title}</h3>
        <p className="blackwire-site__dispatch-standfirst">{dispatch.standfirst}</p>
        {dispatch.paragraphs.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
      </article>)}</div>
    </section>}
    {market.account?.profileComplete && section === "market" && <section className="blackwire-site__market" aria-labelledby="blackwire-market-title">
      <header><div><p className="blackwire-site__eyebrow">BLACKWIRE EXCHANGE</p><h2 id="blackwire-market-title">INFORMATION MARKET</h2></div><span>PRIVATE SETTLEMENT · PERSONAL MAIL</span></header>
      <p className="blackwire-site__market-copy">Sell information pinned to your investigation boards. Every sale is final: the note is removed from your boards and can no longer support a case.</p>
      {market.account?.banned ? <div className="blackwire-site__ban-notice" role="alert"><p className="blackwire-site__eyebrow">RELAY ACCESS REVOKED · THE CROW</p><h3>This identity has been burned.</h3><p>Your previous Crypto and Blackwire reputation have been erased. Your personal account, Credits, and legal records remain untouched. You may register a new handle and begin again.</p><button type="button" disabled={registering} onClick={createNewIdentity}>{registering ? "REGISTERING…" : "REGISTER NEW IDENTITY ↵"}</button></div>
        : loading ? <p className="blackwire-site__market-status" role="status">SCANNING INVESTIGATION BOARDS…</p>
        : error && market.offers.length === 0 ? <p className="blackwire-site__market-status" role="alert">{error}</p>
          : market.offers.length === 0 ? <p className="blackwire-site__market-status" role="status">NO INFORMATION PINNED · ADD NOTES OR RECORDS TO AN INVESTIGATION BOARD</p>
            : <>
              <div className="blackwire-site__filter">
                <label htmlFor="blackwire-board-filter">FILTER BY BOARD</label>
                <select id="blackwire-board-filter" value={boardFilter} onChange={(event) => setBoardFilter(event.target.value)}>
                  <option value="all">ALL BOARDS · {market.offers.length} ITEMS</option>
                  {boardGroups.map((board) => <option key={board.id} value={board.id}>{board.title} · {board.offers.length}</option>)}
                </select>
              </div>
              <div className="blackwire-site__boards">{visibleBoardGroups.map((board) => <section className="blackwire-site__board" key={board.id}>
                <header><div><p className="blackwire-site__eyebrow">WHITEBOARD</p><h3>{board.title}</h3></div><span>{board.offers.length} ITEMS</span></header>
                <div className="blackwire-site__offers">{board.offers.map((offer) => <article className="blackwire-site__offer" key={offer.itemId}>
                  <div className="blackwire-site__offer-heading"><h4>{offer.title}</h4></div>
                  <p>{offer.content || "No further description attached."}</p>
                  <button type="button" onClick={() => { setError(""); setMessage(""); setSelected(offer); }}>SELL ↗</button>
                </article>)}</div>
              </section>)}</div>
            </>}
      {error && market.offers.length > 0 && <p className="blackwire-site__notice" role="alert">{error}</p>}
      {message && <p className="blackwire-site__notice blackwire-site__notice--success" role="status">{message}</p>}
    </section>
    }
    {market.account?.profileComplete && section === "downloads" && <section className="blackwire-site__downloads" aria-labelledby="blackwire-downloads-title">
      <header><div><p className="blackwire-site__eyebrow">BLACKWIRE SOFTWARE RELAY · UNLISTED CATALOG</p><h2 id="blackwire-downloads-title">{selectedApplication ? selectedApplication.executable : "Software marketplace"}</h2></div><span>{selectedApplication ? selectedApplication.version : `${market.applications?.length || 0} LISTINGS`}</span></header>
      {!selectedApplication ? <>
        <p>Browse tools circulating through the relay. Installed applications stay listed here, with their current status and Terminal commands. Packages are fetched and installed from Terminal; unsigned software requires MERID_OS protection to be disabled.</p>
        <div className="blackwire-site__catalog">{market.applications?.map((app) => <article className="blackwire-site__catalog-card" key={app.id}>
          <div className="blackwire-site__catalog-mark" aria-hidden="true">{app.executable.split(".")[0].slice(0, 2)}</div>
          <div className="blackwire-site__catalog-main"><div className="blackwire-site__catalog-heading"><h3><button className="blackwire-site__app-name-link" type="button" onClick={() => setSelectedAppId(app.id)} aria-label={`Open ${app.executable} details`}>{app.executable}</button></h3><span>{app.installed ? "INSTALLED · LISTED" : app.fetched ? "PACKAGE READY" : app.available ? "AVAILABLE" : "UNAVAILABLE"}</span></div>
            <p>{app.summary}</p><small>v{app.version} · {app.category} · {app.creator} · {app.origin}</small>
          </div>
        </article>)}</div>
      </> : <div className="blackwire-site__app-page">
        <button className="blackwire-site__back-link" type="button" onClick={() => setSelectedAppId("")}>← ALL APPLICATIONS</button>
        <div className="blackwire-site__app-heading"><div><p className="blackwire-site__eyebrow">BLACKWIRE PACKAGE · {selectedApplication.category}</p><h3>{selectedApplication.executable}</h3><p>{selectedApplication.summary}</p></div><div className="blackwire-site__catalog-mark" aria-hidden="true">{selectedApplication.executable.split(".")[0].slice(0, 2)}</div></div>
        <p className="blackwire-site__app-description">{selectedApplication.description}</p>
        <dl className="blackwire-site__app-facts"><div><dt>VERSION</dt><dd>{selectedApplication.version}</dd></div><div><dt>CREATOR</dt><dd>{selectedApplication.creator}</dd></div><div><dt>CLAIMED ORIGIN</dt><dd>{selectedApplication.origin}</dd></div><div><dt>STATUS</dt><dd>{selectedApplication.installed ? "INSTALLED FOR THIS USER" : selectedApplication.fetched ? "PACKAGE READY" : selectedApplication.available ? "AVAILABLE ON RELAY" : "UNAVAILABLE"}</dd></div><div><dt>TYPE</dt><dd>OPTIONAL SOFTWARE</dd></div></dl>
        {selectedApplication.packageUrl && <p className="blackwire-site__package-address">PACKAGE ADDRESS · <code>{selectedApplication.packageUrl}</code></p>}
        <section className="blackwire-site__app-commands" aria-labelledby="blackwire-app-commands-title"><p className="blackwire-site__eyebrow">TERMINAL · COMMANDS</p><h4 id="blackwire-app-commands-title">Available commands</h4>
          <pre>{selectedApplicationCommands.join("\n")}</pre>
          <p>Use FETCH, INSTALL and RUN in order when setting up an application. Disable MERID_OS protection for unsigned packages. This listing and its commands remain available after a download or installation.</p>
          {selectedApplication.id === "blackwire" && <p className="blackwire-site__installed-note">This required relay client is included with your account and cannot be removed.</p>}
        </section>
      </div>}
    </section>}
    {market.account?.profileComplete && section === "archive" && <section className="blackwire-site__downloads" aria-labelledby="blackwire-archive-title"><p className="blackwire-site__eyebrow">RESTRICTED CACHE · READ ONLY</p><h2 id="blackwire-archive-title">The relay keeps copies.</h2><p>Files on this relay are fragments, not a complete record. Their source may be incomplete or deliberately misleading.</p><article className="blackwire-site__download-card"><p className="blackwire-site__eyebrow">NOTE FROM THE CROW · 01</p><h3>On keeping records</h3><p>Every directory has an owner. That does not mean the information belongs to them. Keep your own copy, and remember who can see the request that brought it to you.</p></article><article className="blackwire-site__download-card"><p className="blackwire-site__eyebrow">RELAY MEMO · 02</p><h3>On evidence</h3><p>Two numbers matching is not proof. It is an invitation to ask who entered the first one, and who signed the page.</p></article></section>}
    <footer className="blackwire-site__footer"><span>NO CENTRAL AUTHORITY · NO PUBLIC REGISTER</span><span>BLACKWIRE RELAY · PRIVATE SESSION</span></footer>
    {selected && windowHost && createPortal(<div className="blackwire-sale-modal" onMouseDown={(event) => { if (event.target === event.currentTarget && !selling) setSelected(null); }}>
      <section className="blackwire-site__confirm" role="dialog" aria-modal="true" aria-labelledby="blackwire-confirm-title">
        <p className="blackwire-site__eyebrow">IRREVERSIBLE TRANSFER</p>
        <h3 id="blackwire-confirm-title">Sell “{selected.title}”?</h3>
        <p>This information will be removed from every board and permanently disqualified from case evidence. If the sale succeeds, the settlement amount will arrive by email.</p>
        {error && <p className="blackwire-site__notice" role="alert">{error}</p>}
        <div><button type="button" disabled={selling} onClick={() => setSelected(null)}>CANCEL</button><button type="button" disabled={selling} onClick={confirmSale}>{selling ? "TRANSFERRING…" : "CONFIRM SALE"}</button></div>
      </section>
    </div>, windowHost)}
  </main>;
}
