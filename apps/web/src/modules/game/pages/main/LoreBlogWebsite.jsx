import InstitutionalWebsiteFrame from "./InstitutionalWebsiteFrame.jsx";
import { LORE_PUBLICATIONS } from "./lorePublications.js";

export default function LoreBlogWebsite({ blogId = "gate-journal", onNavigate }) {
  const site = LORE_PUBLICATIONS.find((publication) => publication.slug === blogId) || LORE_PUBLICATIONS[0];
  const navigate = (target) => onNavigate?.(target);
  return (
    <InstitutionalWebsiteFrame variant="lore" organizationName={site.name} onHome={() => navigate("hub")} footerContent={`${site.name} · National Relay publication`}>
      <main className={`lore-blog lore-blog--${site.theme} lore-blog--${site.layout}`}>
        <header className="lore-blog__masthead">
          <p>{site.category} · CONNECTED EDITION</p>
          <h1>{site.name}</h1>
          <span>{site.byline}</span>
        </header>
        <nav className="lore-blog__site-nav" aria-label="Publication navigation">
          <a href="#front-page">FRONT PAGE</a>
          <a href="#features">FEATURES</a>
          <a href="#advertisement">ADVERTISEMENTS</a>
          <a href="#directory" onClick={(event) => { event.preventDefault(); navigate("hub"); }}>NATIONAL DIRECTORY »</a>
        </nav>
        <div className="lore-blog__columns" id="front-page">
          <article className="lore-blog__article" id="features">
            <p className="lore-blog__eyebrow">{site.category}</p>
            <h2>{site.headline}</h2>
            {site.articles.map((article) => (
              <section className="lore-blog__story" key={article.title}>
                <h3>{article.title}</h3>
                {article.text.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
              </section>
            ))}
          </article>
          <aside className="lore-blog__rail">
            <section className="lore-blog__advert" id="advertisement">
              <small>{site.ad.label || "ADVERTISEMENT · MERIDIAN ADVERTISING EXCHANGE"}</small>
              <h2>{site.ad.title}</h2>
              <p>{site.ad.text}</p>
              <a href={`#${site.ad.target}`} onClick={(event) => { event.preventDefault(); navigate(site.ad.target); }}>
                {LORE_PUBLICATIONS.some((publication) => publication.id === site.ad.target) ? "Read related coverage »" : site.ad.target === "hub" ? "Open the national directory »" : "Visit advertiser »"}
              </a>
            </section>
            <nav className="lore-blog__related" aria-label="Related publications">
              <h2>From the reading desk</h2>
              {site.links.map((target) => {
                const related = LORE_PUBLICATIONS.find((publication) => publication.id === target);
                if (!related) return null;
                return <a key={target} href={`#${target}`} onClick={(event) => { event.preventDefault(); navigate(target); }}>{related.name} »</a>;
              })}
              <a href="#national-directory" id="directory" onClick={(event) => { event.preventDefault(); navigate("hub"); }}>Return to National Directory »</a>
              <p>Other writers see these events differently. Read their accounts and make up your own mind.</p>
            </nav>
          </aside>
        </div>
      </main>
    </InstitutionalWebsiteFrame>
  );
}
