# ADR 0001 — Singleplayer: um mundo por jogador

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Foi considerado um mundo global compartilhado. Isso traria concorrência, sincronização, conflitos entre jogadores, estado em tempo real e persistência muito mais complexa do que o conceito atual precisa.

## Decisão

Noir 1974 é **singleplayer**. Cada jogador tem `world_id`, seed, NPCs, empresas, casos e mutations próprios. Entidades identificam-se como `(world_id, entity_id)`. Um NPC do World A não precisa existir no World B.

## Consequências

- Mundos não sincronizam.
- Mutar o World A não afeta outros mundos.
- Não há, na versão inicial, necessidade de WebSockets para estado de mundo compartilhado.
