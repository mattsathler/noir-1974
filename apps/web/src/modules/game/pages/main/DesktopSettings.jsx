import { DESKTOP_WALLPAPERS } from "./desktopPreferences.js";

export default function DesktopSettings({ preferences, onChange, soundEffectsEnabled = true, onSoundEffectsChange }) {
  return <main className="desktop-settings" aria-labelledby="desktop-settings-title">
    <header className="desktop-settings__heading">
      <p>MERID_OS · CONTROL PANEL</p>
      <h1 id="desktop-settings-title">Desktop settings</h1>
      <span>Changes are saved automatically on this computer.</span>
    </header>

    <div className="desktop-settings__panel" id="desktop-settings-panel-desktop">
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

      <section className="desktop-settings__section desktop-settings__sound" aria-labelledby="desktop-sound-title">
        <div className="desktop-settings__section-heading">
          <h2 id="desktop-sound-title">System sounds</h2>
          <span>8-BIT INTERFACE EFFECTS</span>
        </div>
        <label className="desktop-settings__sound-toggle">
          <input type="checkbox" checked={soundEffectsEnabled} onChange={(event) => onSoundEffectsChange?.(event.target.checked)} />
          <span><strong>Enable sound effects</strong><small>Keyboard, mouse and notification sounds.</small></span>
        </label>
      </section>
    </div>
  </main>;
}
