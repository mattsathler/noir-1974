import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import CompanyWebsite from "./CompanyWebsite.jsx";

afterEach(cleanup);

function makeSite(advertising = "standard") {
  return {
    company: { id: "company:atelier", name: "Ateliers de la Gare" },
    address: { full: "Ateliers de la Gare, Central, Lyon, France", city: "Lyon" },
    domain: "ateliers-de-la-gare-7c12.merid.local",
    template: "workshop",
    advertising,
    tagline: "Serving the district since 1948.",
    theme: { accent: "#176b68", paper: "#f4f2e8", ink: "#172d30", header: "ribbon", mark: "12345678", layout: "stacked-sections" },
    pages: [
      { id: "home", title: "HOME", sections: [
        { kind: "article", heading: "WELCOME", body: "A local workshop." },
        { kind: "bulletin", heading: "NOTICE", body: "New service hours." },
        { kind: "address", heading: "VISIT US", body: "Ateliers de la Gare, Central, Lyon, France" },
      ] },
      { id: "about", title: "ABOUT", sections: [{ kind: "article", heading: "OUR COMPANY", body: "A local company." }] },
      { id: "account", title: "CUSTOMER ACCOUNT", sections: [{ kind: "note", heading: "SIGN IN", body: "Unavailable." }] },
    ],
    account: { available: false, label: "CUSTOMER ACCOUNT", message: "Online account access is unavailable." },
  };
}

describe("procedural company website", () => {
  it("renders a public company page without ads and keeps account access inert", () => {
    render(<CompanyWebsite site={makeSite("none")} />);

    expect(screen.getByRole("heading", { name: "Ateliers de la Gare" })).toBeInTheDocument();
    expect(screen.getByText("Ateliers de la Gare, Central, Lyon, France")).toBeInTheDocument();
    expect(screen.queryByLabelText(/Advertisement:/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Customer account sign in is unavailable" })).toBeDisabled();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("renders crowded advertising throughout the page", () => {
    render(<CompanyWebsite site={makeSite("crowded")} />);
    expect(screen.getAllByLabelText(/Advertisement:/)).toHaveLength(11);
  });

  it("uses same-window navigation for internal pages", () => {
    const onNavigate = vi.fn();
    render(<CompanyWebsite site={makeSite("sparse")} onNavigate={onNavigate} />);
    fireEvent.click(screen.getByRole("link", { name: "ABOUT" }));
    expect(onNavigate).toHaveBeenCalledWith("about");
  });
});
