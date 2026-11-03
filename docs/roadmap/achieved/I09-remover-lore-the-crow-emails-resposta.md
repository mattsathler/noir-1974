# I09 — Remover lore do The Crow de emails de resposta

**Status:** Implementada
**Origem:** Correio recebido em resposta a ações do jogador
**Escopo:** Curto

## Problema

Emails que respondem a ações do jogador podem incluir lore sobre The Crow, embora esse assunto não faça parte da resposta solicitada.

## Resultado esperado

Emails de resposta tratam da pergunta, pedido ou decisão que os originou. Lore sobre The Crow permanece em canais próprios, como comunicados públicos.

## Escopo

- Remover referências a The Crow dos emails gerados como resposta a emails enviados pelo jogador e de respostas oficiais a pedidos investigativos.
- Manter nos emails apenas a resposta relacionada à ação que os originou.

## Critérios de aceite

- [x] Respostas a perguntas enviadas a NPCs removem referências a The Crow, Bill Bouer ou à caveira fora do assunto solicitado.
- [x] Respostas oficiais a pedidos policiais e decisões de caso permanecem focadas no pedido e não recebem lore lateral.
- [x] Comunicados públicos e outros canais independentes de resposta mantêm seu conteúdo próprio.

## Fora de escopo

- Remover The Crow de jornais, comunicados públicos, diálogos ou outros canais narrativos.
- Alterar o conteúdo central das respostas investigativas.
