import { useState } from "react";
import { DESKTOP_WALLPAPERS } from "./desktopPreferences.js";

export default function DesktopSettings({ preferences, onChange, soundEffectsEnabled = true, onSoundEffectsChange, soundEffectsVolume = 1, onSoundEffectsVolumeChange }) {
  const [activeTab, setActiveTab] = useState("display");
  const volumePercent = Math.round(soundEffectsVolume * 100);
  function handleTabKeyDown(event) {
    const currentTab = event.currentTarget.id.endsWith("sound") ? "sound" : "display";
    const nextTab = event.key === "ArrowRight" ? currentTab === "display" ? "sound" : "display"
      : event.key === "ArrowLeft" ? currentTab === "sound" ? "display" : "sound"
        : event.key === "Home" ? "display"
          : event.key === "End" ? "sound"
            : "";
    if (!nextTab) return;
    event.preventDefault();
    setActiveTab(nextTab);
    document.getElementById(`desktop-settings-tab-${nextTab}`)?.focus();
  }

  return <main className="desktop-settings" aria-labelledby="desktop-settings-title">
    <header className="desktop-settings__heading">
      <p>MERID_OS · CONTROL PANEL</p>
      <h1 id="desktop-settings-title">Desktop settings</h1>
      <span>Changes are saved automatically on this computer.</span>
    </header>

    <div className="desktop-settings__tabs" role="tablist" aria-label="Settings categories">
      <button type="button" id="desktop-settings-tab-display" role="tab" aria-selected={activeTab === "display"} aria-controls="desktop-settings-panel-display" tabIndex={activeTab === "display" ? 0 : -1} onKeyDown={handleTabKeyDown} onClick={() => setActiveTab("display")}>Desktop</button>
      <button type="button" id="desktop-settings-tab-sound" role="tab" aria-selected={activeTab === "sound"} aria-controls="desktop-settings-panel-sound" tabIndex={activeTab === "sound" ? 0 : -1} onKeyDown={handleTabKeyDown} onClick={() => setActiveTab("sound")}>Sound</button>
    </div>

    <section className="desktop-settings__panel" id="desktop-settings-panel-display" role="tabpanel" aria-labelledby="desktop-settings-tab-display" hidden={activeTab !== "display"}>
      <section className="desktop-settings__section" aria-labelledby="desktop-wallpaper-title">
        <div className="desktop-settings__section-heading">
          <h2 id="desktop-wallpaper-title">Desktop themes</h2>
          <span>WALLPAPER · ICON COLORS · CONTRAST</span>
        </div>
        <div className="desktop-settings__wallpapers">
          {DESKTOP_WALLPAPERS.map((wallpaper) => <button
            type="button"
            key={wallpaper.id}
            className={`desktop-settings__wallpaper${preferences.wallpaper === wallpaper.id ? " desktop-settings__wallpaper--selected" : ""}`}
            aria-pressed={preferences.wallpaper === wallpaper.id}
            onClick={() => onChange({ wallpaper: wallpaper.id })}
          >
            <span className="desktop-settings__wallpaper-preview" style={{ backgroundImage: `url("${wallpaper.image}")`, "--wallpaper-accent": wallpaper.accent }} aria-hidden="true"><i>▣</i></span>
            <span className="desktop-settings__wallpaper-label"><strong>{wallpaper.label}</strong><small>ICON COLOR · {wallpaper.accent.toUpperCase()}</small></span>
          </button>)}
        </div>
      </section>
    </section>

    <section className="desktop-settings__panel desktop-settings__panel--sound" id="desktop-settings-panel-sound" role="tabpanel" aria-labelledby="desktop-settings-tab-sound" hidden={activeTab !== "sound"}>
      <section className="desktop-settings__section" aria-labelledby="desktop-sound-title">
        <div className="desktop-settings__section-heading">
          <h2 id="desktop-sound-title">System sound effects</h2>
          <span>8-BIT INTERFACE EFFECTS</span>
        </div>
        <label className="desktop-settings__sound-toggle">
          <input type="checkbox" checked={soundEffectsEnabled} onChange={(event) => onSoundEffectsChange?.(event.target.checked)} />
          <span><strong>Enable sound effects</strong><small>Keyboard, mouse and notification sounds.</small></span>
        </label>
        <label className={`desktop-settings__volume${soundEffectsEnabled ? "" : " desktop-settings__volume--disabled"}`} htmlFor="desktop-sound-volume">
          <span><strong>Volume</strong><output htmlFor="desktop-sound-volume">{volumePercent}%</output></span>
          <input id="desktop-sound-volume" type="range" min="0" max="100" step="1" value={volumePercent} disabled={!soundEffectsEnabled} aria-valuetext={`${volumePercent}%`} onChange={(event) => onSoundEffectsVolumeChange?.(Number(event.currentTarget.value) / 100)} />
          <small>Applies to all system sound effects.</small>
        </label>
      </section>
    </section>
  </main>;
}
