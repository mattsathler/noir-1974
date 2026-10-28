# M04 — Mutations e reconstrução

**Ciclo:** 4  
**Status:** Aceito
**Depende de:** M03 aceito

## Objetivo

Fechar a prova de arquitetura: o mundo sobrevive ao restart com a base determinística **mais** as mutations.

## Critérios de aceite

- [x] Mutation persistível (NPC `status = dead`) aplicada só naquele servidor e visível aos seus jogadores.
- [x] Após desligar e subir de novo o processo (ou equivalente de persistência), a reconstrução com a mesma seed + mutations é idêntica.
- [x] Um segundo servidor **não** vê a mutation do primeiro.
- [x] Teste automatizado cobre: gerar → mutar → “reiniciar” → reconstruir.
- [x] Estado transitório de sessão continua fora do banco; somente mutations de status são persistidas neste recorte.

## Fora de escopo

Gerador de casos, API pública, UI.
