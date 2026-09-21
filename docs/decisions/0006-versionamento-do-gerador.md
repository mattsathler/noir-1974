# ADR 0006 — Versionamento do gerador

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Se o algoritmo mudar, a mesma seed gera outro mundo. Jogadores antigos perderiam a realidade já criada.

## Decisão

Todo mundo e todo caso guardam `generator_version`. Mundos/casos novos podem usar versão nova. Antigos reconstruem com a versão original. Mudança no gerador **não** altera retroativamente uma realidade já criada.

## Consequências

- Código do gerador precisa ser despachável por versão (módulos versionados ou equivalente).
- Testes de regressão: seed conhecida + versão N = snapshot estável.
