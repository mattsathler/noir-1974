# M01 — Modelo de entidades e schema inicial

**Ciclo:** 1  
**Status:** Não iniciado  
**Depende de:** M00 aceito

## Objetivo

Definir o que é uma entidade, quais relações existem no recorte inicial, e o schema mínimo do banco — ainda que a engine do banco ganhe ADR neste ciclo ou no seguinte.

## Critérios de aceite

- [ ] Documento de entidades: NPC, company, location/building, case, evidence (o conjunto mínimo pode ser menor, mas o que existir está definido).
- [ ] Tipos de relação do recorte inicial nomeados (pelo menos emprego e uma relação pessoal, ex. casamento).
- [ ] Decisão escrita: tabela genérica vs tabelas por tipo (ADR).
- [ ] Schema inicial (tabelas/campos) para `worlds`, relações do recorte, `mutations`, `cases`.
- [ ] Campos obrigatórios em world e case: seed, `generator_version`.
- [ ] Modelo explicita servidor como escopo compartilhado e separa estado de caso comum do progresso individual do jogador.
- [ ] Modelo registra a expansão sob demanda e a identidade estável das regiões/entidades geradas.
- [ ] Lista explícita: o que do NPC é seed vs persistido vs mutation (pode ser provisória, mas está escrita).

## Fora de escopo

Algoritmo do gerador, API HTTP, UI.
