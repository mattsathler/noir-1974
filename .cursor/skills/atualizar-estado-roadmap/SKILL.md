---
name: atualizar-estado-roadmap
description: >-
  Recalcula docs/roadmap/estado.md a partir do Status dos milestones
  (M00–M13): aceitos, vigente, progresso do fluxo completo do crime até a prisão.
  Use when accepting or reopening a milestone, updating product status,
  playtest progress, estado.md, or when the user asks to atualizar o estado
  do roadmap / porcentagem de desenvolvimento.
---

# Atualizar estado do roadmap

Não editar a barra nem a porcentagem de `docs/roadmap/estado.md` à mão. A fonte de verdade do status é o campo `**Status:**` em cada arquivo de `docs/roadmap/milestones/`.

## Quando

- Milestone marcado como Aceito / Em curso / Não iniciado
- Pedido para atualizar o estado do produto, a % até o playtest, ou `estado.md`
- PR que fecha um ciclo do roadmap

## Passos

1. Confirmar o `**Status:**` de cada `docs/roadmap/milestones/M*.md` (valores: `Aceito`, `Em curso`, `Não iniciado`). Aceite só com os checkboxes de critério preenchidos.
2. Na raiz do repo, executar:

```bash
python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py
```

3. Conferir o stdout (milestones aceitos, porcentagem e estado de prontidão para playtest) e o diff de `docs/roadmap/estado.md`.
4. No mesmo PR: milestone + `estado.md`. Não commitar só a porcentagem sem o Status que a justifica.

## Regras de cálculo (já no script)

- Alvo da porcentagem: fluxo completo do crime, da geração do caso à prisão do culpado: **M00..M13** (`CRIME_FLOW_THROUGH = 13`).
- Peso igual: milestones aceitos / 14, arredondado para inteiro.
- Só **Aceito** soma. **Em curso** aparece na barra como vazio.
- Milestone vigente = primeiro que ainda não está Aceito, em ordem.
- M06 aceito ⇒ primeiro recorte de playtest está pronto; isso não conclui o fluxo criminal.
- M13 aceito ⇒ fluxo completo, incluindo acusação e prisão, está concluído.

## Extra no fluxo criminal

Se milestones forem adicionados, removidos ou reordenados entre M00 e M13, atualizar `CRIME_FLOW_THROUGH`, os rótulos e as descrições no script e `docs/roadmap/ciclos.md`. Depois rerodar o script.
