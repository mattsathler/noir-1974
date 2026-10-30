import { afterEach, describe, expect, it } from "vitest";
import { DEFAULT_DESKTOP_PREFERENCES, readDesktopPreferences, saveDesktopPreferences } from "./desktopPreferences.js";

const STORAGE_KEY = "noir:desktop:preferences:v1";

afterEach(() => localStorage.removeItem(STORAGE_KEY));

describe("desktop preferences", () => {
  it("uses defaults when no saved preferences exist", () => {
    expect(readDesktopPreferences()).toEqual(DEFAULT_DESKTOP_PREFERENCES);
  });

  it("persists wallpaper and theme and falls back from invalid values", () => {
    expect(saveDesktopPreferences({ wallpaper: "relay-blue", theme: "light" })).toEqual({ wallpaper: "relay-blue", theme: "light" });
    expect(readDesktopPreferences()).toEqual({ wallpaper: "relay-blue", theme: "light" });

    localStorage.setItem(STORAGE_KEY, JSON.stringify({ wallpaper: "missing", theme: "sepia" }));
    expect(readDesktopPreferences()).toEqual(DEFAULT_DESKTOP_PREFERENCES);
  });
});
