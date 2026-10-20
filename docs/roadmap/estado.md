# Estado do produto

Fonte única do **onde estamos**. Regenerar com a skill `atualizar-estado-roadmap`
(script `scripts/recompute_estado.py`). Não editar a porcentagem à mão.

- **Alvo desta barra:** fluxo criminal completo, da geração do caso à prisão do culpado = **M11 aceito**.
- **Marco de playtest:** M06 libera o primeiro recorte atravessável no browser; ele não representa o fluxo investigativo completo.
- **Fora desta barra:** Vesper, auth completa, simulação temporal geral e mapa rico — entram em milestones posteriores.

## Agora

| Campo | Valor |
| --- | --- |
| Data de referência | 2026-10-20 |
| Milestone vigente | M07 — Registros bancários e GATE |
| Último aceito | M06 — API e cliente web |
| Primeiro playtest | M06 aceito; recorte mínimo pronto para playtest interno. |
| Fluxo do crime até a prisão | **58%** (7 de 12 milestones) |

Primeiro recorte atravessável no browser: mundo, NPC/relações e caso, com o servidor como autoridade. O fluxo investigativo completo ainda está em construção.

## Barra do fluxo criminal completo

O fluxo começa com a geração determinística do mundo e do caso, passa pelas fontes de investigação e organização dos elementos, e termina quando uma acusação válida permite prender o culpado. O M06 continua sendo um marco intermediário de playtest.

```text
M00 ████████████████████  aceito
M01 ████████████████████  aceito
M02 ████████████████████  aceito
M03 ████████████████████  aceito
M04 ████████████████████  aceito
M05 ████████████████████  aceito
M06 ████████████████████  aceito   ← primeiro recorte de playtest
M07 ░░░░░░░░░░░░░░░░░░░░  em curso
M08 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M09 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M10 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M11 ░░░░░░░░░░░░░░░░░░░░  não iniciado   ← fluxo completo até a prisão
     | 58%
```

Cada milestone do fluxo pesa **1/12 ≈ 8,3%**. Só entra na conta com status **Aceito**. “Em curso” não soma porcentagem; a porcentagem mede ciclos aceitos do percurso completo.

| Milestone | Peso | Status | Soma acumulada se aceito |
| --- | ---: | --- | ---: |
| [M00](milestones/M00-repositorio-e-docs.md) Repositório e docs | 8,3% | Aceito | 8% |
| [M01](milestones/M01-modelo-de-entidades-e-schema.md) Entidades e schema | 8,3% | Aceito | 17% |
| [M02](milestones/M02-gerador-deterministico.md) Gerador determinístico | 8,3% | Aceito | 25% |
| [M03](milestones/M03-sociedade-e-relacoes.md) Sociedade e relações | 8,3% | Aceito | 33% |
| [M04](milestones/M04-mutations-e-reconstrucao.md) Mutations e reconstrução | 8,3% | Aceito | 42% |
| [M05](milestones/M05-gerador-de-casos.md) Casos e validação | 8,3% | Aceito | 50% |
| [M06](milestones/M06-api-e-cliente.md) API e cliente web | 8,3% | Aceito | 58% |
| [M07](milestones/M07-registros-investigativos.md) Registros bancários e GATE | 8,3% | Em curso | 67% |
| [M08](milestones/M08-sistema-de-emails.md) Correio eletrônico | 8,3% | Não iniciado | 75% |
| [M09](milestones/M09-busca-e-apreensao.md) Busca e apreensão | 8,3% | Não iniciado | 83% |
| [M10](milestones/M10-whiteboards-pessoais.md) Whiteboards pessoais | 8,3% | Não iniciado | 92% |
| [M11](milestones/M11-acusacao-e-julgamento.md) Acusação e julgamento | 8,3% | Não iniciado | **100%** |

Fórmula: `milestones_aceitos_M00_a_M11 / 12`, arredondado para inteiro. O primeiro playtest continua identificado pelo aceite do M06, em campo separado.

## Como atualizar

1. Marcar o milestone no próprio arquivo (`**Status:**` + checkboxes de aceite).
2. Rodar a skill / `python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py`.
3. Se novos milestones entrarem no fluxo entre M00 e M11, atualizar `CRIME_FLOW_THROUGH`, a sequência em [`ciclos.md`](ciclos.md) e os rótulos/blurb deste script.
