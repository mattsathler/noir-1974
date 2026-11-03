export default function ComputerAccessShell({ title, children, labelledBy, busy = false, variant = "default" }) {
  return (
    <main className={`computer-access-shell computer-access-shell--${variant}`} aria-label="Meridian computer access" aria-busy={busy}>
      <section className="computer-access-shell__screen" aria-label="Meridian 1974 workstation">
        <div className="computer-access-shell__desktop">
          <section className="computer-access-shell__window" aria-labelledby={labelledBy}>
            <header className="computer-access-shell__titlebar">
              <span><i className="computer-status-light" aria-hidden="true" /> {title}</span>
            </header>
            <div className="computer-access-shell__content">{children}</div>
          </section>
        </div>
        <div className="computer-access-shell__scanlines" aria-hidden="true" />
      </section>
    </main>
  );
}
