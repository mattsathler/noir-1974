export const LORE_PUBLICATIONS = [
  {
    id: "lore-gate-journal", slug: "gate-journal", name: "Meridian Transit Journal", category: "TRANSPORT · MERIDIAN GLOBAL", theme: "olive",
    headline: "When distance became a service", byline: "From the Meridian Global public information desk",
    articles: [
      { title: "A network built for passage", text: ["We developed the General Access Transit Exchange so people could travel between equipped locations in moments. Its fare is average by design: remarkable reach should remain within ordinary means.", "Since GATE opened, our network has become essential to governments, businesses and households. We see that dependence as proof of a service well made."] },
      { title: "What a connection changes", text: ["A nearby relay can bring work, markets and family within reach. Where there is no terminal, distance still decides who can arrive and on what terms.", "GATE is physical transport. A message sent across a computer network reaches its recipient; it does not carry its sender through a station."] },
    ], ad: { title: "Meridian Transit Atlas", text: "Printed route guides for offices, travelers and station desks.", target: "meridian-enterprises" }, links: ["lore-passage-gazette", "lore-distance-ledger", "lore-terminal-gazette"],
  },
  {
    id: "lore-city-desk", slug: "city-desk", name: "The City Desk", category: "CIVIC LIFE · PUBLIC ORDER", theme: "blue",
    headline: "Bill Bouer works inside the system", byline: "By Elise Marchand · City Desk correspondent",
    articles: [
      { title: "A badge and its institution", text: ["Bill Bouer is a police officer working for law and state. I keep returning to a question: who defines public order when large corporations hold so much power?", "A badge describes an office; it does not settle what a citizen should think of the order that office serves."] },
      { title: "The argument around authority", text: ["Bouer and The Crow stand on opposing sides of a conflict over authority, security and access to information. Neither man can be understood by repeating only the account offered by his allies.", "When I read a public statement, I ask what the institution behind it hopes I will notice—and what it hopes I will leave unread."] },
    ], ad: { title: "Leclerc Typewriter Company", text: "Quiet keys, dependable ribbons, and service at your neighborhood counter.", target: "lore-old-customs" }, links: ["lore-civic-watch", "lore-unfiled-page", "lore-courier-features"],
  },
  {
    id: "lore-unfiled-page", slug: "unfiled-page", name: "The Unfiled Page", category: "UNSIGNED COLUMN · SOURCE UNVERIFIED", theme: "rust",
    headline: "The law has an owner, too", byline: "Unsigned letter · source claims to be The Crow",
    articles: [
      { title: "A note from outside", text: ["They call me a fugitive because I refuse to ask permission from the state or the corporations that own half the things people need. I have broken into their systems and stolen their money. I will not pretend that makes me clean.", "Bill Bouer works for the law and the state. I work against them. You still have to decide what to do when the law protects corporate power."] },
      { title: "A question of trust", text: ["You cannot verify who sent this page. Treat my words with the same suspicion you give an official statement.", "Opposition is not innocence. Authority is not proof. Keep your own judgment, and do not let either of us borrow it."] },
    ], ad: { label: "FUNDING NOTICE · NO PAID ADVERTISEMENTS", title: "No advertiser listed", text: "This publication declines to identify its sponsors. Source and funding remain unverified.", target: "lore-corporate-review" }, links: ["lore-hacker-margin", "lore-city-desk", "lore-citizen-correspondent"],
  },
  {
    id: "lore-distance-ledger", slug: "distance-ledger", name: "The Distance Ledger", category: "BUSINESS · INFRASTRUCTURE & POWER", theme: "gold",
    headline: "The man who owns the distance", byline: "By Lucien Moreau · Senior market correspondent",
    articles: [
      { title: "Adrian Voss and Meridian Global", text: ["Adrian Voss is Meridian Global's chief executive and the most influential, powerful and wealthy person in the world. His company controls the GATE infrastructure on which governments, firms and individuals depend.", "He is reserved, elegant and implacable, and rarely appears in public. I would call that restraint a luxury most executives cannot afford."] },
      { title: "A commodity called distance", text: ["GATE turned distance and geographic access into goods and sources of influence. Meridian Global accumulated extraordinary wealth after the network changed transportation.", "A quotation attributed to Voss reads: “Borders were never meant to divide the world. They were meant to charge for crossing it.” His personal role in developing GATE remains undocumented, and I have found no reason to fill that silence with rumor."] },
    ], ad: { title: "Union Mercantile Bank", text: "Savings, settlements and published exchange information for account holders.", target: "bank-website" }, links: ["lore-meridian-dividend", "lore-gate-journal", "lore-corporate-review"],
  },
  {
    id: "lore-passage-gazette", slug: "passage-gazette", name: "The Passage Gazette", category: "TRAVEL · CITY CONNECTIONS", theme: "blue",
    headline: "A world measured in reachable places", byline: "By Camille Duret · The Passage Gazette travel desk",
    articles: [
      { title: "The station on your street", text: ["I used to judge a trip by the roads between two places. Now I check the relay map first: the stations tell me whether a far city is close enough for an afternoon visit.", "The fare is manageable for many households, but every journey still begins and ends at a physical station. A town without one remains a long way out."] },
      { title: "Who benefits from a route?", text: ["A new connection can bring a market, office or relative within reach. It can also make the districts left off the network feel further away than before.", "I write about the people who wait for a route as much as the travelers who already have one."] },
    ], ad: { title: "Maison des Voyageurs", text: "Station-side rooms, luggage tags and printed city guides.", target: "lore-relay-almanac" }, links: ["lore-gate-journal", "lore-beyond-map", "lore-relay-almanac"],
  },
  {
    id: "lore-relay-almanac", slug: "relay-almanac", name: "National Relay Almanac", category: "REFERENCE · NETWORK & SOCIETY", theme: "olive",
    headline: "The connected world, explained", byline: "By Alain Morel · National Relay Almanac reference desk",
    articles: [
      { title: "A network of services", text: ["Our office receives letters asking whether a company directory, a hospital page and private mail belong to one system. They do not. The relay connects services; each institution keeps its own records and rules.", "A shared connection makes a service easier to reach. It does not make every file public or every operator the same organization."] },
      { title: "Connection is not travel", text: ["A computer can consult information associated with another city without its user going there. We still hear people describe a remote connection as a visit; it is not.", "GATE stations move people. Networks carry messages and records. The distinction matters whenever someone claims to have been present."] },
    ], ad: { title: "Meridian Web Browser 4.1", text: "A familiar index for public services and national listings.", target: "hub" }, links: ["lore-household-circuit", "lore-public-relay", "lore-passage-gazette"],
  },
  {
    id: "lore-beyond-map", slug: "beyond-the-map", name: "Beyond the Map", category: "PERSONAL ESSAYS · TRAVEL", theme: "rust",
    headline: "The old meaning of far away", byline: "By a traveler who prefers the window seat",
    articles: [
      { title: "A journey without a road", text: ["At the station I watched a porter mark trunks with a pencil while passengers discussed a journey that would take only moments. That contrast still pleases me: the machine is astonishing; departure is familiar.", "The nearest terminal and its fare matter more than any grand map. A place without a station is still far away."] },
      { title: "Distance after arrival", text: ["I have crossed half the map and found the same breakfast arguments waiting for me. A swift arrival does not make people less themselves.", "These are only my travel notes. Ask another passenger and you will get a different account of the same city."] },
    ], ad: { title: "Renard Luggage & Leather", text: "Hand-sewn trunks made for the modern traveler and the old station floor.", target: "lore-old-customs" }, links: ["lore-passage-gazette", "lore-old-customs", "lore-terminal-gazette"],
  },
  {
    id: "lore-household-circuit", slug: "household-circuit", name: "The Household Circuit", category: "HOME · COMPUTING & CORRESPONDENCE", theme: "blue",
    headline: "The computer enters the front room", byline: "By Jeanne Arnaud · Household Circuit home desk",
    articles: [
      { title: "A desk connected to the world", text: ["My husband wanted the computer in the front room; I wanted the sewing basket there. We compromised by moving the desk beside the window and keeping the table clear for supper.", "Connected services are useful for work and correspondence, but no directory tells you which neighbor is bringing a pie on Sunday."] },
      { title: "A message still takes care", text: ["I tell readers to save a copy before sending an important letter. Machines can lose a draft as easily as a dog can carry off a note.", "Mail arrives quickly. Trust still takes the time it takes, and a hurried reply can cause more trouble than a slow one."] },
    ], ad: { title: "Meridian Office Machines", text: "Keyboards, paper trays and service plans for the connected household.", target: "lore-old-customs" }, links: ["lore-relay-almanac", "lore-old-customs", "lore-work-register"],
  },
  {
    id: "lore-office-citizen", slug: "office-and-citizen", name: "Office & Citizen", category: "PUBLIC LIFE · CORPORATE ORDER", theme: "gold",
    headline: "Who answers when a company governs?", byline: "By Marceline Perrin · Office & Citizen editorial",
    articles: [
      { title: "Power beyond the town hall", text: ["A company runs the relay my office uses, employs my sister and owns the building where I buy bread. Calling that merely commerce seems too small a word.", "Government still claims authority, but companies own much of the infrastructure that makes daily life possible. I want both sides to explain their obligations plainly."] },
      { title: "The cost of dependence", text: ["A useful service can still concentrate power in the hands of its operator. Convenience does not excuse us from asking who sets the rules.", "My readers disagree about where public duty ends and corporate authority begins. I would rather keep the argument open than pretend it has been settled."] },
    ], ad: { title: "Civic Correspondence Office", text: "Printed forms, registered letters and advice on filing a public inquiry.", target: "police-website" }, links: ["lore-civic-watch", "lore-meridian-dividend", "lore-unfiled-page"],
  },
  {
    id: "lore-civic-watch", slug: "civic-watch", name: "Civic Watch Bulletin", category: "LAW · STATE & PUBLIC SAFETY", theme: "blue",
    headline: "The promise and price of order", byline: "Meridian Civic Watch · Public safety office",
    articles: [
      { title: "Authority in a corporate world", text: ["Our officers act in the name of law and state. We know citizens also see the corporations whose services shape the streets, workplaces and networks they depend on.", "Public safety requires cooperation; public confidence requires clear accounts of what the service does. We welcome questions submitted through the proper office."] },
      { title: "Bouer and the public record", text: ["Officer Bill Bouer is one of the people charged with maintaining public order. His work is described in official notices and in the press.", "An official report records the service's account. It should not be confused with a witness's recollection or a newspaper's interpretation."] },
    ], ad: { title: "Meridian Police Service", text: "Public information, community guidance and published records.", target: "police-website" }, links: ["lore-city-desk", "lore-office-citizen", "lore-courier-features"],
  },
  {
    id: "lore-hacker-margin", slug: "hackers-margin", name: "The Hacker's Margin", category: "TECHNOLOGY · ETHICS & ACCESS", theme: "rust",
    headline: "A key can open more than a door", byline: "By an unnamed systems clerk · Reader's essay",
    articles: [
      { title: "Access is an action", text: ["I maintain a small office network. When somebody enters it without permission, I have to explain to my colleagues why private letters and payroll files were suddenly at risk.", "A person may claim a good purpose. The intrusion still reaches people who never agreed to take part in that argument."] },
      { title: "No clean hands", text: ["The Crow opposes corporate and state power, and he has also hacked systems and stolen company money. I can understand the anger without pretending the theft is harmless.", "Before sharing a secret, ask who benefits from its release and who will have to live with the consequences."] },
    ], ad: { title: "The Unfiled Page", text: "An anonymous column about power, access and the cost of resistance.", target: "lore-unfiled-page" }, links: ["lore-unfiled-page", "lore-household-circuit", "lore-corporate-review"],
  },
  {
    id: "lore-meridian-dividend", slug: "meridian-dividend", name: "The Meridian Dividend", category: "FINANCE · CORPORATE INFLUENCE", theme: "olive",
    headline: "A company becomes infrastructure", byline: "By Étienne Laurent · The Meridian Dividend market letter",
    articles: [
      { title: "Growth after GATE", text: ["Meridian Global became the world's largest company after GATE opened. Its network now carries governments, businesses and individual travelers.", "That is a remarkable position for any firm. I advise readers to admire the reach and still ask how much dependence one company should hold."] },
      { title: "The value of a route", text: ["A route has economic and political value when it makes a distant market reachable. The distribution of stations can shape opportunity as much as travel time.", "Investors study company notices. I also watch what happens to the districts that remain outside the network."] },
    ], ad: { title: "Union Exchange Listings", text: "Published company information and national market services.", target: "bank-website" }, links: ["lore-distance-ledger", "lore-passage-gazette", "lore-office-citizen"],
  },
  {
    id: "lore-old-customs", slug: "old-customs-review", name: "Old Customs Review", category: "CULTURE · TECHNOLOGY & TRADITION", theme: "rust",
    headline: "The future arrived in an old coat", byline: "By Margot Legrand · Old Customs Review",
    articles: [
      { title: "Advanced machines, familiar rooms", text: ["I saw a GATE cabin last week: a machine with astonishing reach, dressed in the noisy, antique form of an old carriage. That contrast is more interesting than another polished brass advertisement.", "Our clothes, manners and household objects change at their own pace. A new machine does not tell a family how to arrange its parlor."] },
      { title: "Change does not travel evenly", text: ["I still mend cuffs by hand and keep a paper notebook beside the computer. My niece says the notebook is old-fashioned; she borrows it whenever the network is down.", "A fast connection and familiar habits fit in the same home. That is not a contradiction I am eager to resolve."] },
    ], ad: { title: "Leclerc Typewriter Company", text: "Traditional workmanship for modern offices. Ribbon replacement while you wait.", target: "lore-household-circuit" }, links: ["lore-beyond-map", "lore-household-circuit", "lore-terminal-gazette"],
  },
  {
    id: "lore-work-register", slug: "work-register", name: "The Work Register", category: "LABOR · THE CORPORATE WORLD", theme: "blue",
    headline: "Employment under the company seal", byline: "By Renée Martin · Shop-floor correspondent",
    articles: [
      { title: "The employer as a daily institution", text: ["My sister works for a company that provides her housing; my brother refuses to work anywhere with a company dining room. Both say their employer has too much say in their day.", "A job can mean a wage, a schedule and access to services. It can also make a person's livelihood depend on one office's decisions."] },
      { title: "The small rules of a working day", text: ["Readers write to me about meal breaks, travel allowances and the cost of replacing worn uniforms. Those details rarely reach the company brochure, but they decide whether a wage stretches to Sunday.", "I print the notices employers send us and the letters workers send back. The two accounts do not always agree."] },
    ], ad: { title: "National Office Placement", text: "Employment notices and clerical openings from connected cities.", target: "meridian-enterprises" }, links: ["lore-office-citizen", "lore-household-circuit", "lore-courier-features"],
  },
  {
    id: "lore-common-table", slug: "common-table", name: "The Common Table", category: "HOUSEHOLD ECONOMY · CREDITS", theme: "gold",
    headline: "One currency, many kinds of life", byline: "By Odette Fabre · The Common Table household column",
    articles: [
      { title: "A week's shopping", text: ["Credits spend the same across cities, but my market basket does not cost the same everywhere. I keep a little notebook of bread, lamp oil and apples, and the totals surprise me each month.", "A common currency makes the sums easy to compare. It does not make every household's income or choices equal."] },
      { title: "The figures behind a household", text: ["My husband calls the household account dull. I call it the reason we can afford a day trip without borrowing from next week's wages.", "A bank statement shows what arrived and what left. It cannot tell you whether the new coat was a necessity, a gift or a foolish delight."] },
    ], ad: { title: "Union Mercantile Bank", text: "Savings and settlements for households across the national network.", target: "bank-website" }, links: ["lore-work-register", "lore-meridian-dividend", "lore-distance-ledger"],
  },
  {
    id: "lore-courier-features", slug: "courier-features", name: "The Evening Courier · Long Reads", category: "PRESS · PEOPLE & POWER", theme: "blue",
    headline: "The people behind the headlines", byline: "By the Evening Courier feature desk",
    articles: [
      { title: "A police officer in the public eye", text: ["Our readers know Bill Bouer as a police officer, but most meet him only through notices, headlines and second-hand accounts. We asked several Lyon residents what his name means to them.", "Their answers ranged from reassurance to suspicion. The difference says as much about the institution as it does about the man."] },
      { title: "The fugitive in the margin", text: ["The Crow opposes the state and large corporations. He also hacks their systems and steals company money. A headline that includes only one of those facts is selling a simple story.", "We prefer a longer read. Public order and corporate power deserve scrutiny in the same column."] },
    ], ad: { title: "The Evening Courier", text: "Morning edition, evening supplement and a weekly review of public affairs.", target: "newspaper" }, links: ["lore-city-desk", "lore-unfiled-page", "lore-civic-watch"],
  },
  {
    id: "lore-public-relay", slug: "public-relay", name: "The Balcony Garden", category: "GARDENING · CITY HOMES", theme: "olive",
    headline: "A little green above the street", byline: "By Paul Vasseur · Notes from a fifth-floor balcony",
    articles: [
      { title: "Tomatoes in a narrow space", text: ["My balcony gets sun only after lunch, so I moved the tomato pots against the warm wall. They are smaller than the market's, but the first ripe one tasted like a victory.", "A reader in the north district writes that their wind is too strong for tall stems. Try low boxes and tie the canes before the first storm, not after."] },
      { title: "The neighbor's cuttings", text: ["Mme. Girard gave me a cutting of her geranium and refused payment. I returned the favor with a jar of bean seeds and a note about the soil mix.", "A city garden grows through small exchanges. Send your questions by post; I answer the letters that include a return address."] },
    ], ad: { title: "Jardin du Rhône Seed House", text: "Heritage seeds, clay pots and balcony soil delivered by local cart.", target: "lore-beyond-map" }, links: ["lore-household-circuit", "lore-common-table", "lore-old-customs"],
  },
  {
    id: "lore-terminal-gazette", slug: "terminal-gazette", name: "The Terminal Gazette", category: "GATE · STATION LIFE", theme: "rust",
    headline: "At the edge of arrival", byline: "By Henri Vidal · Bas Passage station reporter",
    articles: [
      { title: "The morning queue", text: ["At opening, I counted two porters, a family with a wicker basket and one man who had forgotten his ticket. The GATE cabin hummed behind them like an old boiler.", "The trip may take moments; the line still moves at the pace of the person searching every pocket for a receipt."] },
      { title: "A station is a neighborhood", text: ["The café beside the entrance knows who travels on market days. The bookstall keeps a stack of local papers for passengers who arrive early.", "Corporate notices talk about terminals as points on a map. To those of us who work here, a station is also a workplace, a shelter and a meeting place."] },
    ], ad: { title: "Maison des Voyageurs", text: "Maps, station refreshments and dependable luggage repairs.", target: "lore-passage-gazette" }, links: ["lore-gate-journal", "lore-old-customs", "lore-beyond-map"],
  },
  {
    id: "lore-citizen-correspondent", slug: "citizen-correspondent", name: "The Citizen Correspondent", category: "LETTERS · NEIGHBORHOOD LIFE", theme: "gold",
    headline: "Notes from the people next door", byline: "Edited by Lucie Bernard · Readers' letters",
    articles: [
      { title: "The missing blue umbrella", text: ["A reader from Rue Victor Hugo found a blue umbrella left beside the bakery counter. The owner may claim it by describing the small tear in the lining.", "We print neighborhood notices without charge. Please keep them kind, specific and short enough to fit beside the bread prices."] },
      { title: "A Saturday table", text: ["Three families on our block have started sharing supper on the first clear evening of the week. One brings soup, one brings a tart, and one brings the folding chairs.", "Send us your small neighborhood customs. They tell more about a city than another grand speech from a company office."] },
    ], ad: { title: "Le Courrier Correspondence School", text: "Practical lessons in clear letters, careful filing and public argument.", target: "lore-courier-features" }, links: ["lore-unfiled-page", "lore-courier-features", "lore-hacker-margin"],
  },
  {
    id: "lore-corporate-review", slug: "corporate-review", name: "Corporate Review & Inquiry", category: "BUSINESS · ACCOUNTABILITY", theme: "blue",
    headline: "When private systems become public concerns", byline: "By the Corporate Review & Inquiry reporting desk",
    articles: [
      { title: "The reach of a large company", text: ["Major corporations shape infrastructure, work and social life. Their decisions affect people who have no direct seat at the company table.", "Meridian Global illustrates the scale of that influence: its GATE network supports governments, other companies and individuals."] },
      { title: "A question of evidence", text: ["A leak may reveal information, but publication does not prove every allegation that accompanies it. Records should be checked against their source and context.", "Accountability depends on what can be substantiated, not only on which side claims to represent the public."] },
    ], ad: { title: "The Meridian Dividend", text: "Independent commentary on corporate infrastructure and markets.", target: "lore-meridian-dividend" }, links: ["lore-office-citizen", "lore-hacker-margin", "lore-distance-ledger"],
  },
].map((site, index) => ({ ...site, layout: site.layout || ["broadsheet", "column", "bulletin", "ledger", "notebook"][index % 5] }));

export function chooseLorePublications(sites = LORE_PUBLICATIONS, count = 3) {
  const remaining = [...sites];
  const selected = [];
  while (remaining.length && selected.length < count) {
    const index = Math.floor(Math.random() * remaining.length);
    selected.push(remaining.splice(index, 1)[0]);
  }
  return selected;
}

export function loreProgramForSite(siteId) {
  const site = LORE_PUBLICATIONS.find((publication) => publication.id === siteId);
  return site ? { id: site.id, label: site.name.toUpperCase(), siteSlug: site.slug, blogId: site.slug, state: "ONLINE", available: true } : null;
}
