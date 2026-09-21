# ADR 0003 — Modelo híbrido: seed, relações e mutations

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Gerar tudo e persistir tudo não escala conceitualmente. Tratar a seed como banco também falha: relações precisam ser consultáveis (quem trabalha na empresa X?) sem materializar o mundo inteiro.

## Decisão

- Seed + `generator_version` geram o mundo base de forma determinística.
- Relações entre entidades são persistidas/indexadas para consulta.
- Mutations registram o que mudou em relação à base.
- Estado transitório de sessão não é persistido por padrão.

## Consequências

- O banco não substitui o gerador; o gerador não substitui o banco.
- Schema concreto ainda é ponta solta; a *forma* do modelo não é.
