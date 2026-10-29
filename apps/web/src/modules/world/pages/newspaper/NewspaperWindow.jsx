import { useCallback, useEffect, useRef, useState } from "react";
import { getNewspaperIssue } from "../../api.js";
import APIErrorNotice from "../../../../shared/APIErrorNotice.jsx";

export default function NewspaperWindow({ cityId, issueDay, onClose, embedded = false }) {
  const [issue, setIssue] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const dialogRef = useRef(null);
  const closeRef = useRef(null);

  const loadIssue = useCallback(async () => {
    setLoading(true);
    setError("");
    setIssue(null);
    try {
      setIssue(await getNewspaperIssue(cityId, issueDay));
    } catch (requestError) {
      setError(requestError);
    } finally {
      setLoading(false);
    }
  }, [cityId, issueDay]);

  useEffect(() => {
    loadIssue();
  }, [loadIssue]);

  useEffect(() => {
    if (embedded) return undefined;
    const previousFocus = document.activeElement;
    closeRef.current?.focus();
    function handleKeyDown(event) {
      if (event.key === "Escape") {
        onClose?.();
        return;
      }
      if (event.key !== "Tab" || !dialogRef.current) return;
      const focusable = dialogRef.current.querySelectorAll("button:not(:disabled), a[href], input:not(:disabled), [tabindex]:not([tabindex='-1'])");
      if (!focusable.length) return;
      if (event.shiftKey && document.activeElement === focusable[0]) {
        event.preventDefault();
        focusable[focusable.length - 1].focus();
      } else if (!event.shiftKey && document.activeElement === focusable[focusable.length - 1]) {
        event.preventDefault();
        focusable[0].focus();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      if (typeof previousFocus?.focus === "function") previousFocus.focus();
    };
  }, [embedded, onClose]);

  const sheet = (
      <article className={`newspaper-sheet${embedded ? " newspaper-sheet--app" : ""}`}>
        <header className="newspaper-utility">
          <span>{issue?.city ? `${issue.city.toUpperCase()} CITY EDITION` : "CITY EDITION"}</span>
          <span>PRICE: 20 CREDITS</span>
          {!embedded && <button ref={closeRef} type="button" className="newspaper-close" aria-label="Close newspaper" onClick={onClose}>×</button>}
        </header>

        {loading && <div className="newspaper-state" role="status" aria-live="polite">The presses are running…</div>}
        {error && <APIErrorNotice error={error} fallback="The newspaper could not be retrieved." onRetry={loadIssue} retrying={loading} className="newspaper-state newspaper-state--error" />}

        {issue && !loading && (
          <>
            <header className="newspaper-masthead">
              <p className="newspaper-masthead__place">{issue.city} · {issue.country}</p>
              <h1>{issue.publication}</h1>
              <div className="newspaper-masthead__rule" />
              <div className="newspaper-masthead__meta"><span>THE EVENING EDITION</span><time>{issue.issueDate}</time></div>
            </header>

            <section className="newspaper-lead" aria-labelledby="newspaper-lead-title">
              <p className="newspaper-section-label">{issue.lead.section} · LATEST NEWS</p>
              <h2 id="newspaper-lead-title">{issue.lead.headline}</h2>
              <p className="newspaper-lead__summary">{issue.lead.summary}</p>
            </section>

            <section className="newspaper-columns" aria-label="Other city news">
              {issue.articles.map((article, index) => (
                <article className="newspaper-article" key={article.id}>
                  <p className="newspaper-section-label">{article.section}</p>
                  <h3>{article.headline}</h3>
                  <p>{article.summary}</p>
                  <span className="newspaper-article__number">{String(index + 1).padStart(2, "0")}</span>
                </article>
              ))}
            </section>
            <footer className="newspaper-footer">PRINTED IN {issue.city.toUpperCase()} · ALL THE NEWS THAT FITS THE PAGE</footer>
          </>
        )}
      </article>
  );

  if (embedded) return <div className="newspaper-app">{sheet}</div>;
  return <div className="modal-overlay newspaper-overlay" ref={dialogRef} role="dialog" aria-modal="true" aria-label="Daily newspaper" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.(); }}>{sheet}</div>;
}
