# M04 — Mutations e reconstrução

**Ciclo:** 4  
**Status:** Não iniciado  
**Depende de:** M03 aceito

## Objetivo

Fechar a prova de arquitetura: o mundo sobrevive ao restart com a base determinística **mais** as mutations.

## Critérios de aceite

- [ ] Mutation persistível (ex. NPC `status = dead`) aplicada só naquele `world_id`.
- [ ] Após desligar e subir de novo o processo (ou equivalente de persistência), a reconstrução com a mesma seed + mutations é idêntica.
- [ ] Um segundo mundo com outra seed **não** vê a mutation do primeiro.
- [ ] Teste automatizado cobre: gerar → mutar → “reiniciar” → reconstruir.
- [ ] Estado transitório de sessão continua fora do banco.

## Fora de escopo

Gerador de casos, API pública, UI.
