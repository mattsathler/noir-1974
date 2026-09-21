# Roadmap

O desenvolvimento avança em **ciclos**. Cada ciclo tem um **milestone de sucesso** com critérios de aceite binários. O ciclo seguinte **não começa** enquanto o anterior não for aceito.

**Onde estamos:** [`estado.md`](estado.md) — milestones aceitos e porcentagem até o primeiro deploy / playtest.

Prova técnica que sustenta o restante (meta do ciclo 4):

```text
seed → gerar mundo → salvar mutations → desligar → reiniciar
  → reconstruir → idêntico + mutations aplicadas
```

## Ciclos

Ordem e regras de entrada/saída: [`ciclos.md`](ciclos.md).

| Ciclo | Milestone | Documento |
| --- | --- | --- |
| 0 | Repositório e documentação | [M00](milestones/M00-repositorio-e-docs.md) |
| 1 | Modelo de entidades e schema inicial | [M01](milestones/M01-modelo-de-entidades-e-schema.md) |
| 2 | Gerador determinístico (cidade mínima) | [M02](milestones/M02-gerador-deterministico.md) |
| 3 | Sociedade: NPCs, empresas, relações | [M03](milestones/M03-sociedade-e-relacoes.md) |
| 4 | Mutations e reconstrução | [M04](milestones/M04-mutations-e-reconstrucao.md) |
| 5 | Gerador de casos simples + validação | [M05](milestones/M05-gerador-de-casos.md) |
| 6 | REST API e cliente web mínimo | [M06](milestones/M06-api-e-cliente.md) |

Ciclos 1–5 são **fundação**. O ciclo 6 é o primeiro recorte jogável. Conteúdo, tutorial (Vesper), auth completa e simulação temporal ficam **depois** de M06 e exigirão novos milestones.

## Status

Ver [`estado.md`](estado.md). Não duplicar a tabela aqui.
