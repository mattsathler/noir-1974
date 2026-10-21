# Estado do produto

Fonte única do **onde estamos**. Regenerar com a skill `atualizar-estado-roadmap`
(script `scripts/recompute_estado.py`). Não editar a porcentagem à mão.

- **Alvo desta barra:** fluxo criminal completo, da geração do caso à prisão do culpado = **M12 aceito**.
- **Marco de playtest:** M06 libera o primeiro recorte atravessável no browser; ele não representa o fluxo investigativo completo.
- **Fora desta barra:** Vesper, auth completa, simulação temporal geral e mapa rico — entram em milestones posteriores.

## Agora

| Campo | Valor |
| --- | --- |
| Data de referência | 2026-10-21 |
| Milestone vigente | M10 — Whiteboards e identidade visual |
| Último aceito | M09 — Deslocamento e pontos de serviço |
| Primeiro playtest | M06 aceito; recorte mínimo pronto para playtest interno. |
| Fluxo do crime até a prisão | **77%** (10 de 13 milestones) |

O jogador já pode consultar registros e correspondências e se deslocar entre cidades e pontos de serviço. Whiteboards pessoais e identidade visual modular estão em implementação.

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
M07 ████████████████████  aceito
M08 ████████████████████  aceito
M09 ████████████████████  aceito
M10 ░░░░░░░░░░░░░░░░░░░░  em curso
M11 ░░░░░░░░░░░░░░░░░░░░  não iniciado
M12 ░░░░░░░░░░░░░░░░░░░░  não iniciado   ← fluxo completo até a prisão
     | 77%
```

Cada milestone do fluxo pesa **1/13 ≈ 7,7%**. Só entra na conta com status **Aceito**. “Em curso” não soma porcentagem; a porcentagem mede ciclos aceitos do percurso completo.

| Milestone | Peso | Status | Soma acumulada se aceito |
| --- | ---: | --- | ---: |
| [M00](milestones/M00-repositorio-e-docs.md) Repositório e docs | 7,7% | Aceito | 8% |
| [M01](milestones/M01-modelo-de-entidades-e-schema.md) Entidades e schema | 7,7% | Aceito | 15% |
| [M02](milestones/M02-gerador-deterministico.md) Gerador determinístico | 7,7% | Aceito | 23% |
| [M03](milestones/M03-sociedade-e-relacoes.md) Sociedade e relações | 7,7% | Aceito | 31% |
| [M04](milestones/M04-mutations-e-reconstrucao.md) Mutations e reconstrução | 7,7% | Aceito | 38% |
| [M05](milestones/M05-gerador-de-casos.md) Casos e validação | 7,7% | Aceito | 46% |
| [M06](milestones/M06-api-e-cliente.md) API e cliente web | 7,7% | Aceito | 54% |
| [M07](milestones/M07-registros-investigativos.md) Registros bancários e GATE | 7,7% | Aceito | 62% |
| [M08](milestones/M08-sistema-de-emails.md) Correio eletrônico | 7,7% | Aceito | 69% |
| [M09](milestones/M09-deslocamento-e-pontos-de-servico.md) Deslocamento e pontos de serviço | 7,7% | Aceito | 77% |
| [M10](milestones/M10-whiteboards-e-identidade-visual.md) Whiteboards e identidade visual | 7,7% | Em curso | 85% |
| [M11](milestones/M11-busca-e-apreensao.md) Busca e apreensão | 7,7% | Não iniciado | 92% |
| [M12](milestones/M12-acusacao-e-julgamento.md) Acusação e julgamento | 7,7% | Não iniciado | **100%** |

Fórmula: `milestones_aceitos_M00_a_M12 / 13`, arredondado para inteiro. O primeiro playtest continua identificado pelo aceite do M06, em campo separado.

## Como atualizar

1. Marcar o milestone no próprio arquivo (`**Status:**` + checkboxes de aceite).
2. Rodar a skill / `python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py`.
3. Se novos milestones entrarem no fluxo, atualizar `CRIME_FLOW_THROUGH`, a sequência em [`ciclos.md`](ciclos.md) e os rótulos/blurb deste script.
