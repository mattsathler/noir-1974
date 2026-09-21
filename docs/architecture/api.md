# API conceitual

A API definitiva **ainda não** foi especificada. Os endpoints abaixo são exemplos de forma, não contrato.

```http
GET /world
GET /npc/:id
GET /company/:id
GET /company/:id/employees
GET /npc/:id/relationships

GET /cases
GET /cases/:id

POST /cases/generate

POST /world/mutations
POST /cases/:id/progress
```

## Servidor como autoridade

```text
CLIENT → REST API → WORLD / DATABASE → PROCEDURAL GENERATOR → CURRENT STATE
```

O frontend pede “NPC #23”. O servidor verifica o mundo do jogador, recupera seed/estado, gera atributos, consulta relações, aplica mutations e devolve só o que o cliente pode ver.

## Monorepo

Frontend em `apps/web`, backend em `apps/api`. Contratos compartilhados só quando existirem consumidores reais nos dois lados.
