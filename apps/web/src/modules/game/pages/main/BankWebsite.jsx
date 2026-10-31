import { useEffect, useState } from "react";
import { getStockMarket } from "../../../world/api.js";

const PAGES = [
  { id: "home", label: "HOME" },
  { id: "accounts", label: "ACCOUNTS & SAVINGS" },
  { id: "exchange", label: "THE EXCHANGE" },
  { id: "branches", label: "BRANCHES" },
  { id: "notices", label: "PUBLIC NOTICES" },
];

function formatCredits(cents) { return (Number(cents || 0) / 100).toFixed(2); }

export function BankMasthead() {
  return <header className="bank-website__masthead">
    <div className="bank-website__brand"><span className="bank-website__seal" aria-hidden="true">U</span><span><strong>UNION MERCANTILE BANK</strong><small>THE NATIONAL BANK · ESTABLISHED 1848</small></span></div>
    <div className="bank-website__masthead-note"><b>TRUST IN EVERY TRANSACTION</b><span>Branches open Monday—Saturday</span><span>Customer office · National Relay</span></div>
  </header>;
}

export default function BankWebsite({ countryId = "", onOpenStockExchange = () => {} }) {
  const [page, setPage] = useState("home");
  const [marketPage, setMarketPage] = useState(1);
  const [market, setMarket] = useState({ listings: [], page: 1, pageSize: 25, total: 0, countryName: "" });
  const [marketError, setMarketError] = useState(false);
  const activePage = PAGES.find((item) => item.id === page) || PAGES[0];

  useEffect(() => {
    let active = true;
    getStockMarket(marketPage).then((result) => { if (active) { setMarket(result); setMarketError(false); } }).catch(() => { if (active) setMarketError(true); });
    return () => { active = false; };
  }, [countryId, marketPage]);

  useEffect(() => setMarketPage(1), [countryId]);

  return <div className="bank-website" id="institutional-content">
    <BankMasthead />
    <nav className="bank-website__nav" aria-label="Bank website sections">
      {PAGES.map((item) => <button type="button" key={item.id} aria-current={page === item.id ? "page" : undefined} onClick={() => setPage(item.id)}>{item.label}</button>)}
      <span>MERIDIAN NETWORK · PUBLIC ACCESS</span>
    </nav>
    <div className="bank-website__ticker" aria-label="Exchange bulletin"><strong>MARKET BOARD</strong><div><div className="bank-website__ticker-track">{[0, 1].map((copy) => <div className="bank-website__ticker-group" key={copy} aria-hidden={copy === 1 ? "true" : undefined}>{market.listings.slice(0, 8).map((item) => <span key={`${copy}:${item.companyId}`}>{item.companyName} <b>{formatCredits(item.priceCents)} CR</b></span>)}{marketError && <span>QUOTATIONS TEMPORARILY UNAVAILABLE</span>}</div>)}</div></div></div>

    {page === "home" ? <>
      <section className="bank-website__hero">
        <div><p className="bank-website__eyebrow">UNION MERCANTILE BANK · {activePage.label}</p><h1>A sound account.<br />A steadier tomorrow.</h1><p>For generations, Union Mercantile has kept the savings, wages and commerce of the nation moving.</p><div className="bank-website__hero-actions"><button type="button" onClick={() => setPage("accounts")}>EXPLORE SAVINGS »</button><button type="button" className="is-secondary" onClick={() => setPage("branches")}>FIND A BRANCH »</button></div></div>
        <aside className="bank-website__hero-seal"><span>UNION</span><strong>1848</strong><small>SECURE · STEADY · NATIONAL</small></aside>
      </section>
      <div className="bank-website__home-grid">
        <main className="bank-website__main">
          <section className="bank-website__section"><div className="bank-website__section-heading"><div><p className="bank-website__eyebrow">SERVICES FOR EVERY HOUSEHOLD</p><h2>Banking made dependable</h2></div><button type="button" onClick={() => setPage("accounts")}>ALL SERVICES »</button></div>
            <div className="bank-website__service-grid">
              <article><span className="bank-website__service-mark">01</span><h3>Household savings</h3><p>Put a little aside each week. Ask your branch about a passbook savings account.</p><button type="button" onClick={() => setPage("accounts")}>SAVINGS INFORMATION →</button></article>
              <article><span className="bank-website__service-mark">02</span><h3>Wages &amp; deposits</h3><p>Receive wages, deposit cheques and keep a clear record of each payment.</p><button type="button" onClick={() => setPage("accounts")}>ACCOUNT SERVICES →</button></article>
              <article><span className="bank-website__service-mark">03</span><h3>National settlements</h3><p>Send approved payments between Union Mercantile branches and registered businesses.</p><button type="button" onClick={() => setPage("exchange")}>SETTLEMENT GUIDE →</button></article>
            </div>
          </section>
          <section className="bank-website__section bank-website__market-section"><div className="bank-website__section-heading"><div><p className="bank-website__eyebrow">THE UNION EXCHANGE</p><h2>Today on the market</h2></div><button type="button" onClick={() => setPage("exchange")}>MARKET REPORT »</button></div>
            <p className="bank-website__muted">Closing quotations · Thursday, 29 October 1974 · Prices in Credits</p>
            <table className="bank-website__market-table"><thead><tr><th>LISTED COMPANY</th><th>HEAD OFFICE</th><th>LAST</th></tr></thead><tbody>{market.listings.slice(0, 8).map((item) => <tr key={item.companyId}><td>{item.companyName}</td><td>{item.cityName}</td><td>{formatCredits(item.priceCents)} cr</td></tr>)}</tbody></table>{marketError && <p className="bank-website__muted">National market quotations are temporarily unavailable.</p>}
            <button className="bank-website__exchange-link" type="button" onClick={onOpenStockExchange}>OPEN STOCK EXCHANGE APP ↗</button>
          </section>
          <section className="bank-website__ad bank-website__ad--wide"><div><small>UNION MERCANTILE · HOUSEHOLD CREDIT</small><h2>Make a home of your own.</h2><p>Ask your local branch about household improvement loans and repayment terms.</p><button type="button" onClick={() => setPage("branches")}>SPEAK TO A BRANCH CLERK »</button></div><span className="bank-website__ad-coin" aria-hidden="true">U</span></section>
        </main>
        <aside className="bank-website__sidebar">
          <section className="bank-website__sidebar-card bank-website__rates"><h2>RATES &amp; FIGURES</h2><p><span>Standard savings</span><b>3¾% <small>per annum</small></b></p><p><span>Household loan</span><b>from 6½%</b></p><p><span>Next statement day</span><b>01 NOV</b></p><small>Rates shown for public guidance. Branch terms apply.</small></section>
          <section className="bank-website__ad bank-website__ad--vertical"><small>ADVERTISEMENT · 29</small><b>THE UNION<br />POCKET LEDGER</b><span>Keep every household account in order. Available at your local branch, 2 Credits.</span><button type="button" onClick={() => setPage("branches")}>ASK AT THE COUNTER »</button></section>
          <section className="bank-website__sidebar-card"><h2>FROM THE LEDGER</h2><article><time>28 OCTOBER</time><h3>Autumn savings week begins Monday</h3><p>Branch clerks will offer passbook reviews throughout the week.</p><button type="button" onClick={() => setPage("notices")}>READ ALL NOTICES »</button></article><article><time>22 OCTOBER</time><h3>Exchange hours for the holiday</h3><p>National settlement desks will close at noon on 11 November.</p></article></section>
          <section className="bank-website__trust"><strong>YOUR MONEY, ACCOUNTED FOR.</strong><span>Every deposit receives a dated entry in your passbook.</span></section>
        </aside>
      </div>
    </> : <BankSubpage page={page} label={activePage.label} market={market} marketError={marketError} marketPage={marketPage} setMarketPage={setMarketPage} onNavigate={setPage} onOpenStockExchange={onOpenStockExchange} />}
    <footer className="bank-website__footer"><span>UNION MERCANTILE BANK · PUBLIC INFORMATION SERVICE</span><span>Member branches throughout the national network · Best viewed with Meridian Web Browser 4.1</span></footer>
  </div>;
}

function BankSubpage({ page, label, market, marketError, marketPage, setMarketPage, onNavigate, onOpenStockExchange }) {
  const content = {
    accounts: { title: "Accounts for ordinary life", intro: "Union Mercantile branches provide household savings, wage accounts and settlement services. Account opening and private account matters are handled in person at your local counter.", sections: [["Passbook savings", "A dated passbook records deposits, withdrawals and credited interest. Bring your passbook when visiting the branch."], ["Wage accounts", "Arrange regular wage deposits through your employer and review entries with a bank clerk."], ["Household credit", "Loan terms depend on the purpose, repayment period and branch review. Speak with a clerk before submitting an application."]] },
    exchange: { title: "The Union Exchange", intro: "A public summary of listed companies and national settlement notices. Quotations are published after the close of business and may change without notice.", sections: [["Daily quotations", "The market board lists the latest recorded share prices and the daily movement for participating national companies."], ["Business settlements", "Registered businesses may request transfers through their branch. Settlement instructions are reviewed during regular office hours."], ["Market conduct", "Customers should consult the published company notices before making investment decisions. The bank does not guarantee future performance."]] },
    branches: { title: "Your local Meridian branch", intro: "Meridian counters are found in the principal commercial districts of connected cities. Visit during posted hours for account questions, deposits and passbook service.", sections: [["Counter hours", "Monday through Friday, 9:00 AM to 4:30 PM. Saturday, 9:00 AM to noon. Closed on national holidays."], ["What to bring", "Bring your passbook, account reference and a letter of introduction when requesting account service."], ["Customer correspondence", "Write to the Customer Office through your local branch. Include your return address and branch name."]] },
    notices: { title: "Notices to customers", intro: "Public notices from Union Mercantile Bank and its national branch offices.", sections: [["28 October · Autumn savings week", "Passbook reviews are available at all city branches from 3 to 8 November."], ["22 October · Holiday exchange hours", "National settlement desks will close at noon on 11 November. Local deposit counters remain open until noon."], ["16 October · Passbook reissue", "Customers may request a replacement passbook at their home branch. Identification and account reference are required."]] },
  }[page];

  return <div className="bank-website__subpage">
    <div className="bank-website__subpage-layout"><main><p className="bank-website__eyebrow">UNION MERCANTILE BANK / {label}</p><h1>{content.title}</h1><p className="bank-website__subpage-intro">{content.intro}</p>
      {page === "exchange" && <><table className="bank-website__market-table"><thead><tr><th>LISTED COMPANY</th><th>HEAD OFFICE</th><th>LAST</th></tr></thead><tbody>{market.listings.map((item) => <tr key={item.companyId}><td>{item.companyName}</td><td>{item.cityName}</td><td>{formatCredits(item.priceCents)} cr</td></tr>)}</tbody></table>{marketError && <p className="bank-website__subpage-intro">The market board could not be reached. Please try again later.</p>}<nav className="bank-website__market-pages" aria-label="Market listing pages"><span>PAGE {market.page} · {market.total} COMPANIES</span><button type="button" disabled={market.page <= 1} onClick={() => setMarketPage((value) => Math.max(1, value - 1))}>← PREVIOUS</button><button type="button" disabled={market.page * market.pageSize >= market.total} onClick={() => setMarketPage((value) => value + 1)}>NEXT →</button></nav><button className="bank-website__exchange-link" type="button" onClick={onOpenStockExchange}>OPEN STOCK EXCHANGE APP ↗</button></>}
      <div className="bank-website__subpage-sections">{content.sections.map(([title, body]) => <article key={title}><h2>{title}</h2><p>{body}</p></article>)}</div>
      <button className="bank-website__back" type="button" onClick={() => onNavigate("home")}>← RETURN TO BANK HOME</button>
    </main><aside><section className="bank-website__sidebar-card bank-website__rates"><h2>UNION AT A GLANCE</h2><p><span>Founded</span><b>1848</b></p><p><span>Branches</span><b>National network</b></p><p><span>Customer desk</span><b>Open today</b></p></section><div className="bank-website__ad bank-website__ad--vertical"><small>ADVERTISEMENT</small><b>THE UNION<br />POCKET LEDGER</b><span>A tidy record for a careful household.</span><button type="button" onClick={() => onNavigate("branches")}>FIND A BRANCH »</button></div><button className="bank-website__side-link" type="button" onClick={() => onNavigate("notices")}>Customer notices and service hours →</button></aside></div>
  </div>;
}
