import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import LoreBlogWebsite from "./LoreBlogWebsite.jsx";
import { LORE_PUBLICATIONS, chooseLorePublications } from "./lorePublications.js";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("lore publication network", () => {
  it("contains twenty fixed, individually addressed publications", () => {
    expect(LORE_PUBLICATIONS).toHaveLength(20);
    expect(new Set(LORE_PUBLICATIONS.map((site) => site.id)).size).toBe(20);
    expect(new Set(LORE_PUBLICATIONS.map((site) => site.slug)).size).toBe(20);
  });

  it("connects every publication to the rest of the reading network", () => {
    const publications = new Map(LORE_PUBLICATIONS.map((site) => [site.id, site]));
    for (const origin of LORE_PUBLICATIONS) {
      const visited = new Set();
      const pending = [origin.id];
      while (pending.length) {
        const currentId = pending.pop();
        if (visited.has(currentId)) continue;
        visited.add(currentId);
        const current = publications.get(currentId);
        current.links.filter((target) => publications.has(target)).forEach((target) => pending.push(target));
        if (publications.has(current.ad.target)) pending.push(current.ad.target);
      }
      expect(visited.size).toBe(LORE_PUBLICATIONS.length);
    }
  });

  it("chooses a new small, duplicate-free reading sample instead of exposing the whole catalogue", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const firstVisit = chooseLorePublications(LORE_PUBLICATIONS, 3);
    vi.spyOn(Math, "random").mockReturnValue(0.999);
    const secondVisit = chooseLorePublications(LORE_PUBLICATIONS, 3);

    expect(firstVisit).toHaveLength(3);
    expect(new Set(firstVisit.map((site) => site.id)).size).toBe(3);
    expect(secondVisit).toHaveLength(3);
    expect(new Set(secondVisit.map((site) => site.id)).size).toBe(3);
    expect(firstVisit.map((site) => site.id)).not.toEqual(secondVisit.map((site) => site.id));
  });

  it.each(LORE_PUBLICATIONS)("$name reads like an in-world publication and avoids game framing or years", (site) => {
    const { container } = render(<LoreBlogWebsite blogId={site.slug} onNavigate={vi.fn()} />);
    const copy = container.textContent;

    expect(screen.getByRole("heading", { name: site.name, level: 1 })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: site.headline, level: 2 })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "NATIONAL DIRECTORY »" })).toBeInTheDocument();
    expect(copy).not.toMatch(/\bNoir\b|\bplayer\b|\bgame\b|\b(?:19|20)\d{2}\b/iu);
  });

  it.each(LORE_PUBLICATIONS)("$name has working deeplinks to its related publications", (site) => {
    const onNavigate = vi.fn();
    render(<LoreBlogWebsite blogId={site.slug} onNavigate={onNavigate} />);

    site.links.forEach((target) => {
      const destination = LORE_PUBLICATIONS.find((publication) => publication.id === target);
      expect(destination).toBeTruthy();
      fireEvent.click(screen.getByRole("link", { name: destination.name + " »" }));
      expect(onNavigate).toHaveBeenLastCalledWith(target);
    });

    fireEvent.click(screen.getByRole("link", { name: "NATIONAL DIRECTORY »" }));
    expect(onNavigate).toHaveBeenLastCalledWith("hub");
  });
});
