import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import LegacyCountryPortal from "./LegacyCountryPortal.jsx";
import { LORE_PUBLICATIONS } from "./lorePublications.js";

vi.mock("../../../world/api.js", () => ({ searchNationalCompanies: vi.fn() }));

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("National Directory reading desk", () => {
  it("shows a rotating sample of three stories, not the full publication directory", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const navigate = vi.fn();
    const { container, unmount } = render(<LegacyCountryPortal countryName="France" location="Lyon, France" programs={[]} navigate={navigate} />);
    const firstHeadlines = [...container.querySelectorAll(".legacy-portal__reading-card h3")].map((heading) => heading.textContent);

    expect(firstHeadlines).toEqual(LORE_PUBLICATIONS.slice(0, 3).map((site) => site.headline));
    expect(container.querySelectorAll(".legacy-portal__reading-card")).toHaveLength(3);
    expect(screen.queryByRole("button", { name: "THE DISTANCE LEDGER »" })).not.toBeInTheDocument();
    fireEvent.click(screen.getAllByRole("link", { name: "Read the article »" })[0]);
    expect(navigate).toHaveBeenCalledWith(expect.objectContaining({ id: LORE_PUBLICATIONS[0].id, siteSlug: LORE_PUBLICATIONS[0].slug }));
    unmount();

    vi.spyOn(Math, "random").mockReturnValue(0.999);
    const second = render(<LegacyCountryPortal countryName="France" location="Lyon, France" programs={[]} navigate={navigate} />);
    const secondHeadlines = [...second.container.querySelectorAll(".legacy-portal__reading-card h3")].map((heading) => heading.textContent);
    expect(secondHeadlines).toEqual(LORE_PUBLICATIONS.slice(-3).reverse().map((site) => site.headline));
    expect(secondHeadlines).not.toEqual(firstHeadlines);
  });
});
