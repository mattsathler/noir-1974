# M06 — REST API e cliente web mínimo

**Ciclo:** 6  
**Status:** Não iniciado  
**Depende de:** M05 aceito

## Objetivo

Primeiro recorte atravessável no browser: o jogador (ainda que em modo de desenvolvimento) consulta mundo, NPC/empresa e um caso via REST, com o servidor como autoridade.

## Critérios de aceite

- [ ] ADR de stack da API (linguagem, framework, banco).
- [ ] Endpoints mínimos documentados e implementados: mundo, NPC, empresa (ou empregados), caso.
- [ ] Cliente em `apps/web` consome a API e mostra pelo menos: um NPC com relações e um caso aberto.
- [ ] Cliente não recebe `killer_id` (ou equivalente) antes de o estado do caso permitir.
- [ ] Escrita de progresso (mínima) passa pelo servidor e sobrevive a refresh.
- [ ] Isolation: requests de um mundo não leem dados de outro (mesmo que auth ainda seja stub de `world_id`).

## Fora de escopo

Tutorial Vesper, cadastro polido, anti-cheat completo, simulação temporal, mapa rico.
