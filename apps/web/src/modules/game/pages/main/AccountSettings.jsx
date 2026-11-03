import ModularFace from "../../../identity/components/ModularFace.jsx";

export default function AccountSettings({ accountUsername = "SYSTEM USER", character, characterPreviewVisual, location = "MERIDIAN CITY", progression = null, onEditCharacter }) {
  const characterName = character?.displayName || "NO USER PROFILE SELECTED";
  const visual = characterPreviewVisual || character?.visual;

  return <main className="desktop-settings" aria-labelledby="account-settings-title">
    <header className="desktop-settings__heading">
      <p>MERID_OS · USER PROFILE</p>
      <h1 id="account-settings-title">Account settings</h1>
      <span>Manage the active account and its user profile.</span>
    </header>
    <div className="desktop-settings__panel desktop-settings__panel--account">
      <section className="desktop-settings__account" aria-labelledby="account-settings-identity">
        <div className="desktop-settings__account-profile">
          <div className="desktop-settings__character-portrait">
            {visual
              ? <ModularFace visual={visual} name={characterName} />
              : <span aria-hidden="true">{characterName.slice(0, 1) || "?"}</span>}
          </div>
          <div className="desktop-settings__account-identity">
            <p>MERIDIAN SYSTEM USER</p>
            <h2 id="account-settings-identity">{accountUsername || "SYSTEM USER"}</h2>
            <span>USER PROFILE · {characterName}</span>
            <span>ONLINE FAME · {progression?.fame ?? "—"}</span>
          </div>
        </div>
        <section className="desktop-settings__system" aria-labelledby="account-workstation-title">
          <button type="button" className="desktop-settings__change-character" disabled={!character} onClick={onEditCharacter}>CHANGE APPEARANCE</button>
          <div className="desktop-settings__section-heading">
            <h2 id="account-workstation-title">Computer information</h2>
            <span>THIS COMPUTER</span>
          </div>
          <dl className="desktop-settings__system-details">
            <div><dt>Manufacturer</dt><dd>Atlas Business Machines</dd></div>
            <div><dt>System</dt><dd>Meridian OS Workstation 4.1</dd></div>
            <div><dt>Computer name</dt><dd>MERIDIAN-74-042</dd></div>
            <div><dt>Registered organization</dt><dd>{location || "Meridian City"} Municipal Records Office</dd></div>
            <div><dt>Processor</dt><dd>Atlas A-8 · 8 MHz</dd></div>
            <div><dt>Installed memory</dt><dd>640 KB RAM</dd></div>
            <div><dt>Network</dt><dd>GATE local relay · Connected</dd></div>
            <div><dt>Support</dt><dd>Meridian Service Desk · Extension 214</dd></div>
          </dl>
        </section>
      </section>
    </div>
  </main>;
}
