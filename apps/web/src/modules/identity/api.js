async function request(path, options = {}) {
  const { includeWorldFallback = false, ...requestOptions } = options;
  const response = await fetch(path, {
    credentials: "same-origin",
    ...requestOptions,
    headers: {
      ...(requestOptions.body ? { "Content-Type": "application/json" } : {}),
      ...requestOptions.headers,
    },
  });

  if (response.status === 204) return null;

  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json")
    ? await response.json()
    : { message: await response.text() };

  if (!response.ok) {
    throw new Error(payload.message?.trim() || "The request could not be completed.");
  }

  if (includeWorldFallback && !payload.world?.serverTime) {
    payload.world = {
      city: "Lyon",
      country: "France",
      timeZone: "Europe/Paris",
      locationId: "lyon",
      serverTime: new Date().toISOString(),
    };
  }

  return payload;
}

export function registerAccount({ username, email, password }) {
  return request("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, email, password }),
  });
}

export function loginAccount({ email, password }) {
  return request("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export function createCharacter({ displayName, visual }) {
  return request("/api/characters", {
    method: "POST",
    body: JSON.stringify({ displayName, visual }),
  });
}

export function getCharacterVisuals() {
  return request("/api/character-visuals");
}

export function logoutAccount() {
  return request("/api/auth/logout", { method: "POST" });
}

export function currentIdentity() {
  return request("/api/me", { includeWorldFallback: true });
}
