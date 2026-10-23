const contactsByNumber = new Map();

export function phoneNumberForNPC(npcId) {
  const identity = String(npcId || "unknown-npc");
  let hash = 2166136261;
  for (let index = 0; index < identity.length; index += 1) {
    hash = Math.imul(hash ^ identity.charCodeAt(index), 16777619) >>> 0;
  }
  const digits = String(1000000 + (hash % 9000000));
  return `${digits.slice(0, 3)}-${digits.slice(3)}`;
}

export function registerPhoneContacts(companies) {
  for (const { company, employees = [] } of companies) {
    for (const employee of employees) {
      const number = phoneNumberForNPC(employee.id);
      contactsByNumber.set(number.replace(/\D/gu, ""), {
        id: employee.id,
        name: employee.name,
        role: employee.role || "",
        status: employee.status || "alive",
        visual: employee.visual,
        company: company.name,
        number,
      });
    }
  }
}

export function findPhoneContact(number) {
  return contactsByNumber.get(String(number).replace(/\D/gu, "")) ?? null;
}
