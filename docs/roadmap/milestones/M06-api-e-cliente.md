# M06 — REST API e cliente web mínimo

**Ciclo:** 6  
**Status:** Aceito
**Depende de:** M05 aceito

## Objetivo

Primeiro recorte atravessável no browser: o jogador (ainda que em modo de desenvolvimento) consulta mundo, NPC/empresa e um caso via REST, com o servidor como autoridade.

## Critérios de aceite

- [x] ADR de stack da API (linguagem, framework, banco).
- [x] Endpoints mínimos documentados e implementados: mundo, NPC, empresa (ou empregados), caso.
- [x] Cliente em `apps/web` consome a API e mostra pelo menos: um NPC com relações e um caso aberto.
- [x] Cliente não recebe `killer_id` (ou equivalente) antes de o estado do caso permitir.
- [x] Escrita de progresso (mínima) passa pelo servidor e sobrevive a refresh.
- [x] Isolamento: requests de um servidor não leem dados de outro, e jogadores do mesmo servidor consultam o mesmo mundo (mesmo que auth ainda seja stub de `server_id`).

## Fora de escopo

Tutorial Vesper, cadastro polido, anti-cheat completo, simulação temporal, mapa rico.
