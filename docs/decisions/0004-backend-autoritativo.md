# ADR 0004 — Backend autoritativo

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Se o cliente gerar o mundo ou aplicar mutations por conta própria, o estado deixa de ser confiável (cheat, divergência, versões diferentes do gerador).

## Decisão

O servidor é a autoridade. O cliente solicita entidades e progresso; o backend verifica o mundo do jogador, reconstrói, aplica mutations e devolve o estado permitido.

## Consequências

- Gerador vive no backend (ou em biblioteca usada só pelo backend, até haver motivo real para compartilhar).
- Mutations e progresso de caso só entram via API validada.
- Segurança (auth, `world_id`, rate limit) é requisito, ainda não detalhado.
