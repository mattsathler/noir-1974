import { useCallback, useEffect, useRef, useState } from "react";
import LoginPage from "../modules/identity/pages/login/LoginPage.jsx";
import RegisterPage from "../modules/identity/pages/register/RegisterPage.jsx";
import CharacterCreationPage from "../modules/identity/pages/character-creation/CharacterCreationPage.jsx";
import GamePage from "../modules/game/pages/main/GamePage.jsx";
import Whiteboard from "../modules/cases/pages/files/Whiteboard.jsx";
import ComputerAccessShell from "../modules/identity/components/ComputerAccessShell.jsx";
import WorkstationStartup from "../modules/identity/components/WorkstationStartup.jsx";
import { clearBootedThisLogin } from "../modules/game/pages/main/computerBoot.js";
import { currentIdentity, logoutAccount } from "../modules/identity/api.js";
import { clearWorldLocationCache } from "../modules/world/api.js";
import { playRetroSound, readRetroSoundEnabled, setRetroSoundEnabled } from "../modules/game/pages/main/retroSoundEffects.js";

function pageFromPath(pathname) {
  if (pathname === "/") return "login";
  if (pathname === "/startup") return "startup";
  if (/^\/whiteboard(?:\/[^/]+)?\/?$/.test(pathname)) return "whiteboard";
  if (pathname === "/login") return "login";
  if (pathname === "/register") return "register";
  if (pathname === "/character/new") return "character";
  if (["/game", "/archive", "/arrival"].includes(pathname)) return "game";
  if (pathname === "/loading") return "game";
  if (pathname === "/logout") return "logout";
  return "login";
}

function needsSession(page) {
  return ["game", "character", "whiteboard"].includes(page);
}

export default function App() {
  const [page, setPage] = useState(() => pageFromPath(window.location.pathname));
  const routePath = useRef(window.location.pathname);
  const [world, setWorld] = useState(null);
  const [checkingSession, setCheckingSession] = useState(() => needsSession(pageFromPath(window.location.pathname)));
  const [sessionCheckCount, setSessionCheckCount] = useState(0);
  const [startupDestination, setStartupDestination] = useState(null);

  useEffect(() => {
    setRetroSoundEnabled(readRetroSoundEnabled());
    function handlePointerDown(event) {
      if (event.button === 0 && event.isPrimary !== false && !event.target.closest?.("[data-silent-click-sound]")) playRetroSound("click");
    }
    function handleTyping(event) {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      const target = event.target;
      const acceptsText = target instanceof Element && (target.matches("input:not([type=button]):not([type=submit]):not([type=checkbox]):not([type=radio]), textarea, [contenteditable='true']") || target.isContentEditable);
      if (acceptsText && (event.key.length === 1 || ["Backspace", "Delete", "Enter", " "].includes(event.key))) playRetroSound("key");
    }
    document.addEventListener("pointerdown", handlePointerDown, true);
    document.addEventListener("keydown", handleTyping, true);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown, true);
      document.removeEventListener("keydown", handleTyping, true);
    };
  }, []);

  useEffect(() => {
    if (!checkingSession) return undefined;
    let active = true;
    const requestedPath = window.location.pathname;
    const requestedPage = pageFromPath(requestedPath);

    currentIdentity()
      .then((current) => {
        if (!active) return;
        setWorld(current.world ?? null);
        const returnTo = new URLSearchParams(window.location.search).get("returnTo");
        if (requestedPage === "whiteboard" || requestedPage === "login" && returnTo && pageFromPath(returnTo) === "whiteboard") {
          const boardPath = requestedPage === "whiteboard" ? requestedPath : returnTo;
          window.history.replaceState({}, "", boardPath);
          routePath.current = window.location.pathname;
          setPage("whiteboard");
          return;
        }
        const nextPage = current.character ? "game" : "character";
        const nextPath = current.character ? "/game" : "/character/new";
        window.history.replaceState({}, "", nextPath);
        routePath.current = window.location.pathname;
        setPage(nextPage);
      })
      .catch(() => {
        if (!active) return;
        if (requestedPage === "whiteboard") {
          window.history.replaceState({}, "", `/login?returnTo=${encodeURIComponent(requestedPath)}`);
          setPage("login");
        } else if (requestedPage !== "login") {
          window.history.replaceState({}, "", "/login");
          setPage("login");
        }
      })
      .finally(() => {
        if (active) routePath.current = window.location.pathname;
        if (active) setCheckingSession(false);
      });

    return () => { active = false; };
  }, [checkingSession, sessionCheckCount]);

  useEffect(() => {
    const handlePopState = () => {
      if (routePath.current === window.location.pathname) return;
      routePath.current = window.location.pathname;
      const nextPage = pageFromPath(window.location.pathname);
      setPage(nextPage);
      if (needsSession(nextPage)) {
        setCheckingSession(true);
        setSessionCheckCount((count) => count + 1);
      }
    };
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const navigate = useCallback((path, nextPage, nextData = null) => {
    window.history.pushState({}, "", path);
    routePath.current = window.location.pathname;
    if (nextPage === "whiteboard" && !nextData) {
      setCheckingSession(true);
      setSessionCheckCount((count) => count + 1);
    }
    if (["character", "game"].includes(nextPage)) {
      if (nextData) {
        setWorld(nextData.world ?? nextData);
      }
    }
    setPage(nextPage);
  }, []);

  useEffect(() => {
    if (page !== "logout") return undefined;
    let active = true;
    setWorld(null);
    setStartupDestination(null);
    clearBootedThisLogin();
    logoutAccount().catch(() => {}).finally(() => {
      if (active) navigate("/login", "login");
    });
    return () => { active = false; };
  }, [navigate, page]);

  if (checkingSession) {
    return <ComputerAccessShell title="Checking user session" labelledBy="session-check-title" busy>
      <section className="m-auto text-primary" aria-labelledby="session-check-title">
        <p className="text-xs tracking-wide uppercase text-success m-0 mb-3">MERID_OS · AUTHENTICATION SERVICE</p>
        <h1 id="session-check-title" className="text-xl m-0">Checking session</h1>
        <p className="text-sm text-secondary" role="status">Verifying this workstation account…</p>
      </section>
    </ComputerAccessShell>;
  }

  if (page === "startup") return <WorkstationStartup onComplete={() => {
    if (startupDestination?.character) {
      navigate("/game", "game", startupDestination);
    } else {
      navigate("/character/new", "character", startupDestination);
    }
  }} />;

  if (page === "login") {
    return (
      <LoginPage
        onRegister={() => navigate("/register", "register")}
        onAuthenticated={(authenticatedIdentity) => {
          clearWorldLocationCache();
          const returnTo = new URLSearchParams(window.location.search).get("returnTo");
          if (returnTo && pageFromPath(returnTo) === "whiteboard") return navigate(returnTo, "whiteboard", authenticatedIdentity);
          if (authenticatedIdentity.character) {
            setStartupDestination(authenticatedIdentity);
            return navigate("/startup", "startup", authenticatedIdentity);
          }
          setStartupDestination(authenticatedIdentity);
          return navigate("/startup", "startup", authenticatedIdentity);
        }}
      />
    );
  }
  if (page === "register") {
    return <RegisterPage onLogin={() => navigate("/login", "login")} onAccountCreated={() => navigate("/character/new", "character")} />;
  }
  if (page === "character") return <CharacterCreationPage onComplete={() => navigate("/game", "game")} />;
  if (page === "game") return <GamePage initialWorld={world} onLogout={() => navigate("/logout", "logout")} />;
  if (page === "whiteboard") {
    const boardPath = window.location.pathname.match(/^\/whiteboard(?:\/([^/]+))?\/?$/);
    const boardId = boardPath?.[1] ? decodeURIComponent(boardPath[1]) : undefined;
    return <main className="whiteboard-standalone"><Whiteboard initialBoardId={boardId} standalone focusBoard={(nextBoardId) => {
      window.history.replaceState({}, "", `/whiteboard/${encodeURIComponent(nextBoardId)}`);
      routePath.current = window.location.pathname;
    }} /></main>;
  }
  if (page === "logout") return <ComputerAccessShell title="Ending user session" labelledBy="session-end-title" busy>
    <section className="m-auto text-primary" aria-labelledby="session-end-title"><h1 id="session-end-title" className="text-xl m-0">Ending session</h1><p className="text-sm text-secondary">Closing this workstation account…</p></section>
  </ComputerAccessShell>;
  return <LoginPage onRegister={() => navigate("/register", "register")} onAuthenticated={(authenticatedIdentity) => navigate(authenticatedIdentity.character ? "/game" : "/character/new", authenticatedIdentity.character ? "game" : "character", authenticatedIdentity)} />;
}
