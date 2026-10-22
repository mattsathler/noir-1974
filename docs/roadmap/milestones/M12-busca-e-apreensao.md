# M12 — Busca e apreensão

**Ciclo:** 12  
**Status:** Não iniciado  
**Depende de:** M11 aceito

## Objetivo

Permitir que o jogador solicite uma autorização de busca na Delegacia da cidade e investigue uma residência associada a um NPC. A solicitação exige presença física na Delegacia; depois de autorizada, a busca pode revelar objetos relacionados a um caso ou à vida comum.

## Critérios de aceite

- [ ] API valida o alvo, o local, a sessão do jogador e as condições de acesso antes de executar uma busca.
- [ ] O jogador precisa estar na Delegacia da cidade para protocolar a solicitação de busca e apreensão; o backend valida a localização atual e registra a autorização concedida ou negada.
- [ ] Uma autorização concedida identifica o local residencial e permite ao jogador viajar até lá para realizar a busca.
- [ ] O inventário de cada residência é procedural, estável e consultado por local; a cidade inteira não é materializada para responder à ação.
- [ ] Crime Generator pode incluir objetos ligados ao crime no inventário do local correto.
- [ ] Objetos relevantes e cotidianos são apresentados sem distinção visual, textual ou comportamental.
- [ ] A busca persiste os elementos efetivamente encontrados e disponíveis para o jogador, sem revelar sua classificação interna.
- [ ] O fluxo e seus acessos respeitam as [Regras 0001, 0002 e 0003](../../rules/README.md).

## Fora de escopo

Arbitragem da acusação. A organização dos elementos em whiteboards é coberta pelo M10.
