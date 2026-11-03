export default function DeadletterApp() {
  return <main className="deadletter-app" aria-label="Dead Letter offline archive">
    <header className="deadletter-app__header">
      <p className="computer-window__eyebrow">BLACKWIRE RELAY · OFFLINE CACHE</p>
      <h1>DEAD LETTER</h1>
      <span>READ ONLY · 3 FRAGMENTS</span>
    </header>
    <section className="deadletter-app__entry">
      <p>NOTE FROM THE CROW · 01</p>
      <h2>On keeping records</h2>
      <blockquote>Every directory has an owner. That does not mean the information belongs to them. Keep your own copy, and remember who can see the request that brought it to you.</blockquote>
      <small>RELAY COPY · SOURCE UNVERIFIED</small>
    </section>
    <section className="deadletter-app__entry">
      <p>RELAY MEMO · 02</p>
      <h2>On evidence</h2>
      <blockquote>Two numbers matching is not proof. It is an invitation to ask who entered the first one, and who signed the page.</blockquote>
      <small>RELAY COPY · SOURCE UNVERIFIED</small>
    </section>
    <section className="deadletter-app__entry">
      <p>UNSIGNED NOTE · 03</p>
      <h2>On quiet routes</h2>
      <blockquote>A clean connection is not an empty trail. Leave only what you mean to leave.</blockquote>
      <small>RELAY COPY · SOURCE UNVERIFIED</small>
    </section>
    <footer>NO REPLY CHANNEL · NO ORIGINALS STORED HERE</footer>
  </main>;
}
