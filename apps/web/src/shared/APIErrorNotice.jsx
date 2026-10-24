export default function APIErrorNotice({ error, fallback = "The request could not be completed.", onRetry, retrying = false, retryLabel = "TRY AGAIN", className = "api-error-notice" }) {
  const message = typeof error === "string" ? error : error?.message || fallback;
  const reference = error?.requestId;
  const code = error?.code;

  return (
    <div className={className} role="alert">
      <p className="api-error-notice__message">{message}</p>
      {(reference || code) && <small className="api-error-notice__reference">REFERENCE {reference || "UNAVAILABLE"}{code ? ` · ${code}` : ""}</small>}
      {onRetry && <button className="api-error-notice__retry" type="button" disabled={retrying} onClick={onRetry}>{retrying ? "RETRYING…" : retryLabel}</button>}
    </div>
  );
}
