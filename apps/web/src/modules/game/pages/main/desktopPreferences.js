export const DESKTOP_WALLPAPERS = [
  { id: "meridian-night", label: "Meridian night", image: "/images/desktop-wallpapers/meridian-night.svg" },
  { id: "relay-blue", label: "Relay blue", image: "/images/desktop-wallpapers/relay-blue.svg" },
  { id: "olive-grid", label: "Olive grid", image: "/images/desktop-wallpapers/olive-grid.svg" },
  { id: "paper", label: "Warm paper", image: "/images/desktop-wallpapers/paper.svg" },
  { id: "amber-circuit", label: "Amber circuit", image: "/images/desktop-wallpapers/amber-circuit.svg" },
  { id: "atlantic-fog", label: "Atlantic fog", image: "/images/desktop-wallpapers/atlantic-fog.svg" },
  { id: "copper-dusk", label: "Copper dusk", image: "/images/desktop-wallpapers/copper-dusk.svg" },
  { id: "desert-signal", label: "Desert signal", image: "/images/desktop-wallpapers/desert-signal.svg" },
  { id: "forest-relay", label: "Forest relay", image: "/images/desktop-wallpapers/forest-relay.svg" },
  { id: "violet-frequency", label: "Violet frequency", image: "/images/desktop-wallpapers/violet-frequency.svg" },
  { id: "redline", label: "Redline", image: "/images/desktop-wallpapers/redline.svg" },
  { id: "glacier", label: "Glacier", image: "/images/desktop-wallpapers/glacier.svg" },
  { id: "midnight-orbit", label: "Midnight orbit", image: "/images/desktop-wallpapers/midnight-orbit.svg" },
  { id: "golden-hour", label: "Golden hour", image: "/images/desktop-wallpapers/golden-hour.svg" },
  { id: "monochrome", label: "Monochrome", image: "/images/desktop-wallpapers/monochrome.svg" },
  { id: "jade-current", label: "Jade current", image: "/images/desktop-wallpapers/jade-current.svg" },
  { id: "rose-terminal", label: "Rose terminal", image: "/images/desktop-wallpapers/rose-terminal.svg" },
  { id: "silver-grid", label: "Silver grid", image: "/images/desktop-wallpapers/silver-grid.svg" },
  { id: "sahara-lines", label: "Sahara lines", image: "/images/desktop-wallpapers/sahara-lines.svg" },
  { id: "polar-night", label: "Polar night", image: "/images/desktop-wallpapers/polar-night.svg" },
];

export const DEFAULT_DESKTOP_PREFERENCES = { wallpaper: "meridian-night", theme: "dark" };
const STORAGE_KEY = "noir:desktop:preferences:v1";

export function readDesktopPreferences() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "null");
    return {
      wallpaper: DESKTOP_WALLPAPERS.some(({ id }) => id === saved?.wallpaper) ? saved.wallpaper : DEFAULT_DESKTOP_PREFERENCES.wallpaper,
      theme: saved?.theme === "light" ? "light" : "dark",
    };
  } catch {
    return DEFAULT_DESKTOP_PREFERENCES;
  }
}

export function saveDesktopPreferences(preferences) {
  const next = {
    wallpaper: DESKTOP_WALLPAPERS.some(({ id }) => id === preferences.wallpaper) ? preferences.wallpaper : DEFAULT_DESKTOP_PREFERENCES.wallpaper,
    theme: preferences.theme === "light" ? "light" : "dark",
  };
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify(next)); } catch { /* Keep preferences usable for this session if storage is unavailable. */ }
  return next;
}
