import { useCallback, useEffect, useRef, useState } from "react";
import IntroductionPage, { LocationIntroductionPage } from "../modules/prologue/pages/introduction/IntroductionPage.jsx";
import LoginPage from "../modules/identity/pages/login/LoginPage.jsx";
import RegisterPage from "../modules/identity/pages/register/RegisterPage.jsx";
import CharacterCreationPage from "../modules/identity/pages/character-creation/CharacterCreationPage.jsx";
import ArchivePage from "../modules/identity/pages/archive/ArchivePage.jsx";
import GamePage from "../modules/game/pages/main/GamePage.jsx";
import { currentIdentity, logoutAccount } from "../modules/identity/api.js";

const COMPLETION_TRANSITION_DELAY = 2500;

function pageFromPath(pathname) {
  if (pathname === "/login") return "login";
  if (pathname === "/register") return "register";
  if (pathname === "/character/new") return "character";
  if (pathname === "/archive") return "archive";
  if (pathname === "/arrival") return "arrival";
  if (pathname === "/game") return "game";
  if (pathname === "/logout") return "logout";
  return "intro";
}

export default function App() {
  const [page, setPage] = useState(() => pageFromPath(window.location.pathname));
  const [world, setWorld] = useState(null);
  const [identity, setIdentity] = useState(null);
  const [checkingSession, setCheckingSession] = useState(() => ["intro", "login"].includes(pageFromPath(window.location.pathname)));
  const [sessionCheckCount, setSessionCheckCount] = useState(0);
  const completionTimer = useRef(null);

  useEffect(() => {
    if (!checkingSession) return undefined;
    let active = true;

    currentIdentity()
      .then((current) => {
        if (!active) return;
        setIdentity(current);
        setWorld(current.world ?? null);
        const nextPage = current.character ? "archive" : "character";
        const nextPath = current.character ? "/archive" : "/character/new";
        window.history.replaceState({}, "", nextPath);
        setPage(nextPage);
      })
      .catch(() => {})
      .finally(() => {
        if (active) setCheckingSession(false);
      });

    return () => { active = false; };
  }, [checkingSession, sessionCheckCount]);

  useEffect(() => {
    const handlePopState = () => {
      const nextPage = pageFromPath(window.location.pathname);
      setPage(nextPage);
      if (["intro", "login"].includes(nextPage)) {
        setCheckingSession(true);
        setSessionCheckCount((count) => count + 1);
      }
    };
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const navigate = useCallback((path, nextPage, nextData = null) => {
    window.clearTimeout(completionTimer.current);
    completionTimer.current = null;
    window.history.pushState({}, "", path);
    if (["intro", "login"].includes(nextPage)) {
      setCheckingSession(true);
      setSessionCheckCount((count) => count + 1);
    }
    if (nextPage === "archive") setIdentity(nextData);
    if ((nextPage === "arrival" || nextPage === "game") && nextData) setWorld(nextData);
    setPage(nextPage);
  }, []);

  useEffect(() => {
    if (page !== "logout") return undefined;
    let active = true;
    setIdentity(null);
    setWorld(null);
    logoutAccount().catch(() => {}).finally(() => {
      if (active) navigate("/login", "login");
    });
    return () => { active = false; };
  }, [navigate, page]);

  const navigateAfterCompletion = useCallback((path, nextPage, nextData = null) => {
    window.clearTimeout(completionTimer.current);
    completionTimer.current = window.setTimeout(() => {
      completionTimer.current = null;
      navigate(path, nextPage, nextData);
    }, COMPLETION_TRANSITION_DELAY);
  }, [navigate]);

  useEffect(() => () => window.clearTimeout(completionTimer.current), []);

  if (checkingSession) {
    return <main className="min-h-screen bg-deep" aria-label="Checking session" aria-busy="true" />;
  }

  if (page === "login") {
    return (
      <LoginPage
        onRegister={() => navigate("/register", "register")}
        onAuthenticated={(authenticatedIdentity) => navigate(
          authenticatedIdentity.character ? "/archive" : "/character/new",
          authenticatedIdentity.character ? "archive" : "character",
          authenticatedIdentity.character ? authenticatedIdentity : null,
        )}
      />
    );
  }
  if (page === "register") {
    return <RegisterPage onLogin={() => navigate("/login", "login")} onAccountCreated={() => navigateAfterCompletion("/character/new", "character")} />;
  }
  if (page === "character") return <CharacterCreationPage onComplete={() => navigateAfterCompletion("/archive", "archive")} />;
  if (page === "archive") return <ArchivePage initialIdentity={identity} onLoaded={(loadedWorld) => navigateAfterCompletion("/arrival", "arrival", loadedWorld)} />;
  if (page === "arrival") return <LocationIntroductionPage world={world} onComplete={(loadedWorld) => navigateAfterCompletion("/game", "game", loadedWorld)} />;
  if (page === "game") return <GamePage initialWorld={world} onLogout={() => navigate("/logout", "logout")} />;
  if (page === "logout") return <main className="min-h-screen bg-deep" aria-label="Ending session" aria-busy="true" />;
  return <IntroductionPage onComplete={() => navigateAfterCompletion("/login", "login")} />;
}
