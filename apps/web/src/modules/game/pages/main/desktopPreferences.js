export const DESKTOP_WALLPAPERS = [
  { id: "meridian-night", label: "Meridian night", image: "/images/desktop-wallpapers/meridian-night.svg", accent: "#d2b955", mode: "dark" },
  { id: "relay-blue", label: "Relay blue", image: "/images/desktop-wallpapers/relay-blue.svg", accent: "#b8d7c9", mode: "dark" },
  { id: "olive-grid", label: "Olive grid", image: "/images/desktop-wallpapers/olive-grid.svg", accent: "#e3d58b", mode: "dark" },
  { id: "paper", label: "Warm paper", image: "/images/desktop-wallpapers/paper.svg", accent: "#08796e", mode: "light" },
  { id: "amber-circuit", label: "Amber circuit", image: "/images/desktop-wallpapers/amber-circuit.svg", accent: "#edaa62", mode: "dark" },
  { id: "atlantic-fog", label: "Atlantic fog", image: "/images/desktop-wallpapers/atlantic-fog.svg", accent: "#a8d7cf", mode: "dark" },
  { id: "copper-dusk", label: "Copper dusk", image: "/images/desktop-wallpapers/copper-dusk.svg", accent: "#efa286", mode: "dark" },
  { id: "desert-signal", label: "Desert signal", image: "/images/desktop-wallpapers/desert-signal.svg", accent: "#ffe0a3", mode: "dark" },
  { id: "forest-relay", label: "Forest relay", image: "/images/desktop-wallpapers/forest-relay.svg", accent: "#c2d68e", mode: "dark" },
  { id: "violet-frequency", label: "Violet frequency", image: "/images/desktop-wallpapers/violet-frequency.svg", accent: "#c2b5ff", mode: "dark" },
  { id: "redline", label: "Redline", image: "/images/desktop-wallpapers/redline.svg", accent: "#ff9a83", mode: "dark" },
  { id: "glacier", label: "Glacier", image: "/images/desktop-wallpapers/glacier.svg", accent: "#c1f0e8", mode: "dark" },
  { id: "midnight-orbit", label: "Midnight orbit", image: "/images/desktop-wallpapers/midnight-orbit.svg", accent: "#aabbff", mode: "dark" },
  { id: "golden-hour", label: "Golden hour", image: "/images/desktop-wallpapers/golden-hour.svg", accent: "#ffe4a0", mode: "dark" },
  { id: "monochrome", label: "Monochrome", image: "/images/desktop-wallpapers/monochrome.svg", accent: "#e1e6e2", mode: "dark" },
  { id: "jade-current", label: "Jade current", image: "/images/desktop-wallpapers/jade-current.svg", accent: "#a4efca", mode: "dark" },
  { id: "rose-terminal", label: "Rose terminal", image: "/images/desktop-wallpapers/rose-terminal.svg", accent: "#ffd0cb", mode: "dark" },
  { id: "silver-grid", label: "Silver grid", image: "/images/desktop-wallpapers/silver-grid.svg", accent: "#e1ebeb", mode: "dark" },
  { id: "sahara-lines", label: "Sahara lines", image: "/images/desktop-wallpapers/sahara-lines.svg", accent: "#ffe0a6", mode: "dark" },
  { id: "polar-night", label: "Polar night", image: "/images/desktop-wallpapers/polar-night.svg", accent: "#a6eaff", mode: "dark" },
];

export const DEFAULT_DESKTOP_PREFERENCES = { wallpaper: "meridian-night" };
const STORAGE_KEY = "noir:desktop:preferences:v1";

export function readDesktopPreferences() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "null");
    return { wallpaper: DESKTOP_WALLPAPERS.some(({ id }) => id === saved?.wallpaper) ? saved.wallpaper : DEFAULT_DESKTOP_PREFERENCES.wallpaper };
  } catch {
    return DEFAULT_DESKTOP_PREFERENCES;
  }
}

export function saveDesktopPreferences(preferences) {
  const next = { wallpaper: DESKTOP_WALLPAPERS.some(({ id }) => id === preferences.wallpaper) ? preferences.wallpaper : DEFAULT_DESKTOP_PREFERENCES.wallpaper };
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify(next)); } catch { /* Keep preferences usable for this session if storage is unavailable. */ }
  return next;
}
