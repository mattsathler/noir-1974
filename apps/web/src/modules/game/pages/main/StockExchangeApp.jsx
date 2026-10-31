import { useCallback, useEffect, useMemo, useState } from "react";
import { getStockMarket, getStockPortfolio, placeStockOrder } from "../../../world/api.js";
import { BankMasthead } from "./BankWebsite.jsx";

const NOOP = () => {};
function credits(cents = 0) { return `${(Number(cents) / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} CR`; }
function stockSymbol(name = "") {
  const words = name.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toUpperCase().match(/[A-Z0-9]+/g) || [];
  const suffix = words.find((word) => /^\d+$/.test(word));
  const letters = words.filter((word) => !/^\d+$/.test(word));
  return `${letters.slice(0, 3).map((word) => word[0]).join("") || "CO"}${suffix || ""}`;
}
function readableError(error) {
  const messages = {
    INSUFFICIENT_CREDITS: "INSUFFICIENT CREDITS FOR THIS ORDER.",
    INSUFFICIENT_SHARES: "YOUR PORTFOLIO DOES NOT HAVE ENOUGH SHARES.",
    DUPLICATE_STOCK_ORDER: "THIS ORDER WAS ALREADY RECEIVED. PORTFOLIO REFRESHED.",
    COMPANY_NOT_IN_ACTIVE_MARKET: "CONNECT TO THE COMPANY'S COUNTRY THROUGH ROGUE DNS.",
  };
  return messages[error?.code] || error?.message || "THE ORDER COULD NOT BE COMPLETED.";
}

export default function StockExchangeApp({ countryId = "", onProgressionChange = NOOP }) {
  const [page, setPage] = useState(1);
  const [tab, setTab] = useState("market");
  const [market, setMarket] = useState({ listings: [], page: 1, pageSize: 25, total: 0, countryName: "" });
  const [progression, setProgression] = useState(null);
  const [positions, setPositions] = useState([]);
  const [quantity, setQuantity] = useState("1");
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [orderDraft, setOrderDraft] = useState(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const refreshPortfolio = useCallback(async () => {
    const result = await getStockPortfolio();
    setProgression(result.progression);
    setPositions(result.positions || []);
    onProgressionChange(result.progression);
  }, [onProgressionChange]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    Promise.all([getStockMarket(page, search), getStockPortfolio()]).then(([marketResult, portfolioResult]) => {
      if (!active) return;
      setMarket(marketResult);
      setProgression(portfolioResult.progression);
      setPositions(portfolioResult.positions || []);
      onProgressionChange(portfolioResult.progression);
    }).catch((cause) => { if (active) setError(readableError(cause)); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [page, countryId, search, onProgressionChange]);

  const positionByCompany = useMemo(() => new Map(positions.map((position) => [position.companyId, position])), [positions]);
  const portfolioValue = useMemo(() => positions.reduce((sum, position) => sum + position.marketValueCents, 0), [positions]);
  const pageCount = Math.max(1, Math.ceil(market.total / market.pageSize));
  const draftShares = Number(quantity);
  const draftPosition = orderDraft ? positionByCompany.get(orderDraft.listing.companyId) : null;
  const draftValid = Number.isSafeInteger(draftShares) && draftShares > 0 && draftShares <= (orderDraft?.side === "sell" ? draftPosition?.shares || 0 : 10000);

  useEffect(() => setPage(1), [countryId]);

  function openOrder(listing, side) {
    setQuantity("1");
    setOrderDraft({ listing, side });
    setError("");
  }

  async function submitOrder() {
    if (!orderDraft) return;
    const shares = Number(quantity);
    if (!Number.isSafeInteger(shares) || shares < 1 || shares > 10000) {
      setError("ENTER A WHOLE SHARE QUANTITY BETWEEN 1 AND 10,000.");
      return;
    }
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const clientOrderId = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const { listing, side } = orderDraft;
      const result = await placeStockOrder({ clientOrderId, companyId: listing.companyId, side, shares });
      const prior = positionByCompany.get(listing.companyId);
      const updatedPositions = side === "buy"
        ? (() => {
            const nextShares = (prior?.shares || 0) + shares;
            const nextAverage = prior ? Math.round(((prior.shares * prior.averageCostCents) + (shares * listing.priceCents)) / nextShares) : listing.priceCents;
            const position = { ...listing, shares: nextShares, averageCostCents: nextAverage, marketValueCents: nextShares * listing.priceCents };
            return prior ? positions.map((item) => item.companyId === listing.companyId ? position : item) : [...positions, position];
          })()
        : positions.flatMap((item) => {
            if (item.companyId !== listing.companyId) return [item];
            const remaining = item.shares - shares;
            return remaining > 0 ? [{ ...item, shares: remaining, marketValueCents: remaining * item.priceCents }] : [];
          });
      const updatedProgression = { ...(progression || {}), creditsCents: result.creditsCents };
      setProgression(updatedProgression);
      setPositions(updatedPositions);
      onProgressionChange(updatedProgression);
      setNotice(`${side.toUpperCase()} ORDER FILLED · ${shares} SHARE${shares === 1 ? "" : "S"} · ${credits(result.order.totalCents)}`);
      setOrderDraft(null);
    } catch (cause) {
      setError(readableError(cause));
      if (cause?.code === "DUPLICATE_STOCK_ORDER") await refreshPortfolio().catch(() => {});
    } finally {
      setBusy(false);
    }
  }

  return <main className="stock-exchange" aria-label="Stock exchange application">
    <BankMasthead />
    <header className="stock-exchange__header"><div><span className="stock-exchange__eyebrow">UNION MERCANTILE · PUBLIC EXCHANGE</span><h1>STOCK EXCHANGE</h1><p>NATIONAL LISTINGS · {market.countryName || countryId.toUpperCase() || "PROXY COUNTRY"}</p></div><span className="stock-exchange__status"><i /> MARKET OPEN</span></header>
    <section className="stock-exchange__account-strip" aria-label="Player progression">
      <div><span>AVAILABLE CREDITS</span><strong>{progression ? credits(progression.creditsCents) : "—"}</strong></div>
      <div><span>PORTFOLIO VALUE</span><strong>{credits(portfolioValue)}</strong></div>
      <div><span>POSITIONS</span><strong>{positions.length}</strong></div>
    </section>
    <nav className="stock-exchange__tabs" aria-label="Stock exchange sections"><button type="button" aria-current={tab === "market" ? "page" : undefined} onClick={() => setTab("market")}>NATIONAL MARKET</button><button type="button" aria-current={tab === "portfolio" ? "page" : undefined} onClick={() => setTab("portfolio")}>MY PORTFOLIO ({positions.length})</button></nav>
    <section className="stock-exchange__orders"><span>Whole shares · no brokerage fee</span>{tab === "market" && <form className="stock-exchange__search-form" onSubmit={(event) => { event.preventDefault(); setSearch(searchInput.trim()); setPage(1); }}><label htmlFor="stock-company-search">FIND A COMPANY</label><input id="stock-company-search" className="stock-exchange__search" type="search" placeholder="Company name" value={searchInput} onChange={(event) => setSearchInput(event.target.value)} disabled={busy} /><button type="submit" disabled={busy}>SEARCH ↵</button></form>}</section>
    {error && <p className="stock-exchange__message stock-exchange__message--error" role="alert">{error}</p>}
    {notice && <p className="stock-exchange__message stock-exchange__message--success" role="status">{notice}</p>}
    {loading ? <p className="stock-exchange__loading" role="status">READING NATIONAL EXCHANGE…</p> : tab === "market" ? <>
      <div className="stock-exchange__table-wrap"><table className="stock-exchange__table stock-exchange__table--market"><thead><tr><th>COMPANY</th><th>HEAD OFFICE</th><th>PRICE</th><th>OWNED</th><th>ORDER</th></tr></thead><tbody>
        {market.listings.map((listing) => { const position = positionByCompany.get(listing.companyId); return <tr key={listing.companyId}><td><strong className="stock-exchange__symbol">{stockSymbol(listing.companyName)}</strong><small className="stock-exchange__company-name">{listing.companyName}</small></td><td>{listing.cityName}</td><td>{credits(listing.priceCents)}</td><td>{position?.shares || 0}</td><td className="stock-exchange__actions"><button type="button" disabled={busy} onClick={() => openOrder(listing, "buy")}>BUY ↗</button><button type="button" disabled={busy || !position} onClick={() => openOrder(listing, "sell")}>SELL ↘</button></td></tr>; })}
      </tbody></table>{market.listings.length === 0 && <p className="stock-exchange__empty">NO COMPANIES ARE LISTED IN THIS COUNTRY.</p>}</div>
      <footer className="stock-exchange__pagination"><span>{market.total} NATIONAL LISTINGS · PAGE {market.page} OF {pageCount}</span><div><button type="button" disabled={busy || page <= 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>← PREVIOUS</button><button type="button" disabled={busy || page >= pageCount} onClick={() => setPage((value) => Math.min(pageCount, value + 1))}>NEXT →</button></div></footer>
    </> : <div className="stock-exchange__table-wrap"><table className="stock-exchange__table stock-exchange__table--portfolio"><thead><tr><th>COMPANY</th><th>SHARES</th><th>AVG. COST</th><th>MARKET PRICE</th><th>POSITION VALUE</th><th>ORDER</th></tr></thead><tbody>
      {positions.map((position) => <tr key={position.companyId}><td><strong>{stockSymbol(position.companyName)}</strong><small>{position.companyName} · {position.cityName} · {position.countryId.toUpperCase()}</small></td><td>{position.shares}</td><td>{credits(position.averageCostCents)}</td><td>{credits(position.priceCents)}</td><td>{credits(position.marketValueCents)}</td><td className="stock-exchange__actions">{position.countryId === market.countryId ? <><button type="button" disabled={busy} onClick={() => openOrder(position, "buy")}>BUY ↗</button><button type="button" disabled={busy} onClick={() => openOrder(position, "sell")}>SELL ↘</button></> : <small>CHANGE PROXY TO TRADE</small>}</td></tr>)}
      </tbody></table>{positions.length === 0 && <p className="stock-exchange__empty">NO SHARES HELD. SELECT A COMPANY FROM THE NATIONAL MARKET.</p>}</div>}
    <footer className="stock-exchange__footnote">Quotes are set by the national exchange. Prices remain fixed until a market bulletin is issued. Trades require a Proxy in the listing's country.</footer>
    {orderDraft && <div className="stock-exchange__modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) setOrderDraft(null); }}><section className="stock-exchange__order-modal" role="dialog" aria-modal="true" aria-labelledby="stock-order-title"><header><div><span>UNION MERCANTILE · THE UNION EXCHANGE</span><h2 id="stock-order-title">{orderDraft.side === "buy" ? "Buy shares" : "Sell shares"}</h2></div><button type="button" aria-label="Close order" disabled={busy} onClick={() => setOrderDraft(null)}>×</button></header><p className="stock-exchange__modal-company"><strong>{stockSymbol(orderDraft.listing.companyName)}</strong><span>{orderDraft.listing.companyName}</span></p><label className="stock-exchange__modal-quantity" htmlFor="stock-share-quantity">SHARE QUANTITY<input id="stock-share-quantity" type="number" min="1" max={orderDraft.side === "sell" ? draftPosition?.shares || 0 : 10000} step="1" value={quantity} onChange={(event) => setQuantity(event.target.value)} disabled={busy} /></label><dl><div><dt>Current share price</dt><dd>{credits(orderDraft.listing.priceCents)}</dd></div><div><dt>{orderDraft.side === "buy" ? "Estimated debit" : "Estimated proceeds"}</dt><dd>{credits((Number(quantity) || 0) * orderDraft.listing.priceCents)}</dd></div>{orderDraft.side === "sell" && <div><dt>Shares held</dt><dd>{draftPosition?.shares || 0}</dd></div>}{orderDraft.side === "buy" && progression && <div><dt>Available after order</dt><dd>{credits(progression.creditsCents - (Number(quantity) || 0) * orderDraft.listing.priceCents)}</dd></div>}</dl>{error && <p className="stock-exchange__message stock-exchange__message--error" role="alert">{error}</p>}<footer><button className="stock-exchange__cancel" type="button" disabled={busy} onClick={() => setOrderDraft(null)}>CANCEL</button><button type="button" disabled={busy || !draftValid || (orderDraft.side === "buy" && (Number(quantity) || 0) * orderDraft.listing.priceCents > (progression?.creditsCents ?? 0))} onClick={submitOrder}>{busy ? "PROCESSING…" : `CONFIRM ${orderDraft.side.toUpperCase()}`}</button></footer></section></div>}
  </main>;
}
