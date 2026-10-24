export function phoneNumberForNPC(npcId) {
  const identity = String(npcId || "unknown-npc");
  let hash = 2166136261;
  for (let index = 0; index < identity.length; index += 1) {
    hash = Math.imul(hash ^ identity.charCodeAt(index), 16777619) >>> 0;
  }
  const digits = String(1000000 + (hash % 9000000));
  return `${digits.slice(0, 3)}-${digits.slice(3)}`;
}
