export const DESKTOP_WALLPAPERS = [
  { id: "meridian-night", label: "Meridian night", image: "/images/desktop-wallpapers/meridian-night.svg" },
  { id: "relay-blue", label: "Relay blue", image: "/images/desktop-wallpapers/relay-blue.svg" },
  { id: "olive-grid", label: "Olive grid", image: "/images/desktop-wallpapers/olive-grid.svg" },
  { id: "paper", label: "Warm paper", image: "/images/desktop-wallpapers/paper.svg" },
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
