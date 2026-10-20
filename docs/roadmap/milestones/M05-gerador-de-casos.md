# M05 — Gerador de casos simples e validação

**Ciclo:** 5  
**Status:** Não iniciado  
**Depende de:** M04 aceito

## Objetivo

Um tipo simples de caso (homicídio entre conhecidos) nasce das relações do mundo, guarda seed/versão/estado compacto, e só é aceito se passar na validação.

## Critérios de aceite

- [ ] Caso compartilhado referencia vítima, culpado e local por IDs do servidor — não duplica o grafo.
- [ ] Culpado e vítima são entidades existentes e distintas.
- [ ] Há pelo menos uma relação pré-existente aproveitada (ex. colegas).
- [ ] Validador rejeita caso insolúvel segundo regras mínimas documentadas (entidades existem; killer ≠ victim; pelo menos uma pista alcançável definida).
- [ ] Estado do caso (aberto / progresso mínimo) persiste separado da definition.
- [ ] O estado compartilhado do caso é distinguido do progresso individual dos jogadores.
- [ ] Mesma `case_seed` + versão reconstrói a definition; o estado do jogador não vem da seed.

## Fora de escopo

Múltiplos tipos de crime, pistas falsas sofisticadas, UI de investigação, Vesper.
