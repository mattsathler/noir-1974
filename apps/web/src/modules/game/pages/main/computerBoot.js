export const COMPUTER_BOOT_LINES = [
  "PhoenixBIOS Meridian Systems 4.1 Release",
  "Copyright 1974-1984 Meridian Enterprises",
  "All Rights Reserved",
  "BIOS version 4.1.74",
  "System ID = M_OS-00250850",
  "Build Time: 10/18/74 09:09:52",
  "CPU: Meridian 16-bit Processor 1066 MHz",
  "640K System RAM Passed",
  "254M Extended RAM Passed",
  "512K Cache SRAM Passed",
  "System BIOS shadowed",
  "Video BIOS shadowed",
  "Detecting primary IDE controller… OK",
  "Checking local file system… OK",
  "Initializing Meridian network stack…",
  "Private relay connection established",
  "Loading Meridian OS workstation services…",
  "System startup complete.",
];

const LOGIN_BOOT_STORAGE_KEY = "noir:computer:booted-this-login";

export function hasBootedThisLogin() {
  try { return sessionStorage.getItem(LOGIN_BOOT_STORAGE_KEY) === "true"; }
  catch { return false; }
}

export function markBootedThisLogin() {
  try { sessionStorage.setItem(LOGIN_BOOT_STORAGE_KEY, "true"); }
  catch { /* Boot remains available when browser storage is disabled. */ }
}

export function clearBootedThisLogin() {
  try { sessionStorage.removeItem(LOGIN_BOOT_STORAGE_KEY); }
  catch { /* The next session will still be able to boot. */ }
}
