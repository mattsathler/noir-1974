# Estado do produto

Fonte única do **onde estamos**. Regenerar com a skill `atualizar-estado-roadmap`
(script `scripts/recompute_estado.py`). Não editar a porcentagem à mão.

- **Alvo desta barra:** primeiro deploy + playtest interno = **M06 aceito**.
- **Fora desta barra:** Vesper, auth completa, simulação temporal, mapa rico — entram em milestones posteriores.

## Agora

| Campo | Valor |
| --- | --- |
| Data de referência | 2026-10-20 |
| Milestone vigente | M02 — Gerador determinístico |
| Último aceito | M01 — Entidades e schema |
| Deploy / playtest | Não. Ainda não há mundo gerado nem recorte de investigação jogável. |
| Caminho até o 1º playtest | **29%** (2 de 7 milestones) |

Modelo de entidades e schema inicial estão escritos. Ainda não há gerador nem mundo materializado.

## Barra até o primeiro deploy e playtest

Playtest = recorte M06 no browser: consultar mundo, um NPC com relações e um caso, com o servidor como autoridade.

```text
M00 ████████████████████  aceito
M01 ████████████████████  aceito
M02 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M03 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M04 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M05 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M06 ░░░░░░░░░░░░░░░░░░░░  não iniciado   ← deploy / playtest
     | 29%
```

Cada milestone do caminho pesa **1/7 ≈ 14,3%**. Só entra na conta com status **Aceito**. “Em curso” não soma porcentagem (evita 40% de trabalho pela metade).

| Milestone | Peso | Status | Soma acumulada se aceito |
| --- | ---: | --- | ---: |
| [M00](milestones/M00-repositorio-e-docs.md) Repositório e docs | 14,3% | Aceito | 14% |
| [M01](milestones/M01-modelo-de-entidades-e-schema.md) Entidades e schema | 14,3% | Aceito | 29% |
| [M02](milestones/M02-gerador-deterministico.md) Gerador determinístico | 14,3% | Não iniciado | 43% |
| [M03](milestones/M03-sociedade-e-relacoes.md) Sociedade e relações | 14,3% | Não iniciado | 57% |
| [M04](milestones/M04-mutations-e-reconstrucao.md) Mutations e reconstrução | 14,3% | Não iniciado | 71% |
| [M05](milestones/M05-gerador-de-casos.md) Casos e validação | 14,3% | Não iniciado | 86% |
| [M06](milestones/M06-api-e-cliente.md) API e cliente web | 14,3% | Não iniciado | **100%** |

Fórmula: `aceitos_no_caminho / 7`, arredondado para inteiro.

## Como atualizar

1. Marcar o milestone no próprio arquivo (`**Status:**` + checkboxes de aceite).
2. Rodar a skill / `python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py`.
3. Se o primeiro playtest ganhar milestones extras **antes** de M06, ajustar `PLAYTEST_THROUGH` no script e [`ciclos.md`](ciclos.md).
