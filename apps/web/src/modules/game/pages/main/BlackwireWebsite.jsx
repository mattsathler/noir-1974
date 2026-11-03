import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { getBlackwireMarket, markWhiteboardItemSold, registerNewBlackwireIdentity, sellBlackwireItem } from "../../../world/api.js";
import "./BlackwireWebsite.css";

export default function BlackwireWebsite({ onAccountChange = () => {} }) {
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

  const boardGroups = Array.from(market.offers.reduce((groups, offer) => {
    const group = groups.get(offer.boardId) || { id: offer.boardId, title: offer.boardTitle, offers: [] };
    group.offers.push(offer);
    groups.set(offer.boardId, group);
    return groups;
  }, new Map()).values());
  const visibleBoardGroups = boardFilter === "all" ? boardGroups : boardGroups.filter((board) => board.id === boardFilter);

  useLayoutEffect(() => {
    setWindowHost(siteRef.current?.closest(".computer-window") || null);
  }, []);

  useEffect(() => {
    if (boardFilter !== "all" && !boardGroups.some((board) => board.id === boardFilter)) setBoardFilter("all");
  }, [boardFilter, market.offers]);

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
      setMessage("NEW IDENTITY REGISTERED. YOUR PREVIOUS REPUTATION AND CRYPTO ARE GONE.");
    } catch {
      setError("A NEW IDENTITY COULD NOT BE REGISTERED. RELOAD THE RELAY AND TRY AGAIN.");
    } finally {
      setRegistering(false);
    }
  }

  return <main className="blackwire-site" ref={siteRef} aria-label="Blackwire clandestine network">
    <header className="blackwire-site__masthead">
      <div className="blackwire-site__mark" aria-hidden="true">BW</div>
      <div className="blackwire-site__identity"><span>UNLISTED NETWORK · EST. BY THE CROW</span><h1>BLACKWIRE</h1></div>
      <span className="blackwire-site__connection"><i aria-hidden="true" /> PRIVATE RELAY</span>
    </header>
    <nav className="blackwire-site__nav" aria-label="Blackwire sections">
      <span aria-current="page">MARKET</span><span>ARCHIVE · RESTRICTED</span><span>SOFTWARE · {market.account?.unlockedApplications?.length || 0} AVAILABLE</span>
    </nav>
    {market.account && <section className="blackwire-site__account" aria-label="Blackwire hacker account">
      <div><span className="blackwire-site__eyebrow">ALTER EGO · OPERATOR</span><strong>{market.account.handle}</strong></div>
      <div><span className="blackwire-site__eyebrow">BLACKWIRE PROGRESSION</span><strong>LEVEL {market.account.level} · {market.account.reputation}{market.account.nextLevelReputation ? ` / ${market.account.nextLevelReputation} REP` : " REP · MAX"}</strong></div>
      <div><span className="blackwire-site__eyebrow">CRYPTO BALANCE</span><strong>{(market.account.cryptoCents / 100).toFixed(2)} CRYPTO</strong></div>
      <div><span className="blackwire-site__eyebrow">AVAILABLE SOFTWARE</span><strong>{market.account.unlockedApplications.map((application) => `${application.toUpperCase()}.EXE`).join(" · ")}</strong></div>
    </section>}
    <section className="blackwire-site__welcome">
      <p className="blackwire-site__eyebrow">A MESSAGE FROM THE CROW</p>
      <h2>Information has a price.</h2>
      <p>Blackwire is a private exchange for people who know how to find what others would rather keep hidden.</p>
    </section>
    <section className="blackwire-site__market" aria-labelledby="blackwire-market-title">
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
