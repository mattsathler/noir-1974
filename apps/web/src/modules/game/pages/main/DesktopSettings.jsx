import { useState } from "react";
import ModularFace from "../../../identity/components/ModularFace.jsx";
import { DESKTOP_WALLPAPERS } from "./desktopPreferences.js";

export default function DesktopSettings({ preferences, onChange, accountUsername = "SYSTEM USER", character, characterPreviewVisual, location = "MERIDIAN CITY", onEditCharacter, onLogout }) {
  const [activeTab, setActiveTab] = useState("desktop");
  const username = accountUsername || "SYSTEM USER";
  const characterName = character?.displayName || "NO USER PROFILE SELECTED";
  function handleTabKeyDown(event) {
    const tabIDs = ["desktop", "account"];
    const currentIndex = tabIDs.indexOf(activeTab);
    const nextIndex = event.key === "ArrowRight" || event.key === "Home"
      ? (event.key === "Home" ? 0 : (currentIndex + 1) % tabIDs.length)
      : event.key === "ArrowLeft" || event.key === "End"
        ? (event.key === "End" ? tabIDs.length - 1 : (currentIndex - 1 + tabIDs.length) % tabIDs.length)
        : -1;
    if (nextIndex < 0) return;
    event.preventDefault();
    const nextTab = tabIDs[nextIndex];
    setActiveTab(nextTab);
    document.getElementById(`desktop-settings-tab-${nextTab}`)?.focus();
  }
  return <main className="desktop-settings" aria-labelledby="desktop-settings-title">
    <header className="desktop-settings__heading">
      <p>MERID_OS · CONTROL PANEL</p>
      <h1 id="desktop-settings-title">Desktop settings</h1>
      <span>Changes are saved automatically on this computer.</span>
    </header>

    <nav className="desktop-settings__tabs" role="tablist" aria-label="Settings categories">
      <button id="desktop-settings-tab-desktop" type="button" role="tab" tabIndex={activeTab === "desktop" ? 0 : -1} aria-selected={activeTab === "desktop"} aria-controls="desktop-settings-panel-desktop" onKeyDown={handleTabKeyDown} onClick={() => setActiveTab("desktop")}>DESKTOP</button>
      <button id="desktop-settings-tab-account" type="button" role="tab" tabIndex={activeTab === "account" ? 0 : -1} aria-selected={activeTab === "account"} aria-controls="desktop-settings-panel-account" onKeyDown={handleTabKeyDown} onClick={() => setActiveTab("account")}>MY ACCOUNT</button>
    </nav>

    {activeTab === "desktop" ? <div className="desktop-settings__panel" id="desktop-settings-panel-desktop" role="tabpanel" aria-labelledby="desktop-settings-tab-desktop" tabIndex={0}>
      <section className="desktop-settings__section" aria-labelledby="desktop-wallpaper-title">
        <div className="desktop-settings__section-heading">
          <h2 id="desktop-wallpaper-title">Wallpaper</h2>
          <span>DESKTOP BACKGROUND</span>
        </div>
        <div className="desktop-settings__wallpapers">
          {DESKTOP_WALLPAPERS.map((wallpaper) => <button
            type="button"
            key={wallpaper.id}
            className={`desktop-settings__wallpaper${preferences.wallpaper === wallpaper.id ? " desktop-settings__wallpaper--selected" : ""}`}
            aria-pressed={preferences.wallpaper === wallpaper.id}
            onClick={() => onChange({ wallpaper: wallpaper.id })}
          >
            <span className="desktop-settings__wallpaper-preview" style={{ backgroundImage: `url("${wallpaper.image}")` }} aria-hidden="true" />
            <span>{wallpaper.label}</span>
          </button>)}
        </div>
      </section>

      <section className="desktop-settings__section" aria-labelledby="desktop-theme-title">
        <div className="desktop-settings__section-heading">
          <h2 id="desktop-theme-title">Color theme</h2>
          <span>WINDOWS AND DESKTOP</span>
        </div>
        <div className="desktop-settings__themes" role="group" aria-label="Color theme">
          {[{ id: "dark", label: "Dark", sample: "A quiet, low-light desktop." }, { id: "light", label: "Light", sample: "A brighter desktop and controls." }].map((theme) => <button
            type="button"
            key={theme.id}
            className={`desktop-settings__theme desktop-settings__theme--${theme.id}${preferences.theme === theme.id ? " desktop-settings__theme--selected" : ""}`}
            aria-pressed={preferences.theme === theme.id}
            onClick={() => onChange({ theme: theme.id })}
          >
            <span className="desktop-settings__theme-swatch" aria-hidden="true"><i /><i /><i /></span>
            <span><strong>{theme.label}</strong><small>{theme.sample}</small></span>
          </button>)}
        </div>
      </section>

    </div> : <div className="desktop-settings__panel desktop-settings__panel--account" id="desktop-settings-panel-account" role="tabpanel" aria-labelledby="desktop-settings-tab-account" tabIndex={0}>
      <section className="desktop-settings__account" aria-labelledby="desktop-account-title">
        <div className="desktop-settings__account-profile">
          <div className="desktop-settings__character-portrait">
            {characterPreviewVisual || character?.visual
                ? <ModularFace visual={characterPreviewVisual || character.visual} name={characterName || "System user"} />
                : <span aria-hidden="true">{characterName?.slice(0, 1) || "?"}</span>}
          </div>
          <div className="desktop-settings__account-identity">
            <p>MERIDIAN SYSTEM USER</p>
            <h2 id="desktop-account-title">{username}</h2>
            <span>USER PROFILE · {characterName}</span>
          </div>
        </div>
        <section className="desktop-settings__system" aria-labelledby="desktop-system-title">
          <button type="button" className="desktop-settings__change-character" disabled={!character} onClick={onEditCharacter}>CHANGE APPEARANCE</button>
          <div className="desktop-settings__section-heading">
            <h2 id="desktop-system-title">Computer information</h2>
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
        <footer className="desktop-settings__account-actions">
          <button type="button" className="desktop-settings__logout" onClick={onLogout}>LOG OUT ↗</button>
        </footer>
      </section>
    </div>}
  </main>;
}
