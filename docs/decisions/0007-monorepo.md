# ADR 0007 — Monorepo frontend + backend

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Cliente e API evoluem juntos (contratos, entidades, progresso de caso). Repositórios separados atrasariam o ciclo único de produto neste estágio.

## Decisão

Um único repositório:

- `apps/web` — frontend
- `apps/api` — backend
- `docs/` — documentação de produto, arquitetura, decisões, regras e roadmap

Ferramenta de workspaces e linguagem da API ficam em aberto até o primeiro código de aplicação.

## Consequências

- Um PR pode alterar contrato e cliente juntos.
- Código compartilhado só em `packages/` quando houver dois consumidores reais.
