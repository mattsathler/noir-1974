# M12 — Interações investigativas em locais

**Ciclo:** 12  
**Status:** Não iniciado  
**Depende de:** M11 aceito

## Objetivo

Construir as telas de interior dos locais e transformá-los em espaços de investigação. A partir do card de uma construção no M11, o jogador entra em sua tela de interior e poderá interrogar NPCs, solicitar na Delegacia uma autorização de busca e apreensão, investigar residências e burlar controles de acesso a sistemas restritos. A seleção do bairro e da construção continua transitória; cada ação envia o contexto necessário e o backend valida cidade, alvo e permissões. Buscas podem revelar artefatos e vestígios pináveis; interrogatórios usam falas fixas que dão contexto e direção ao jogador, mas não viram elementos do whiteboard. Achados, falas e resultados misturam fatos cotidianos e elementos ligados a crimes sem sinalizar ao jogador o que é pista.

## Critérios de aceite

- [ ] API valida o alvo, o local, a sessão do jogador e as condições de acesso antes de executar uma busca.
- [ ] Cada construção disponível possui uma tela de interior compatível com a direção visual do jogo; entrar e sair dela não persiste bairro ou prédio atual. O Hospital permanece fechado e não pode ser aberto nesta milestone.
- [ ] O interior recebe a construção selecionada no M11 e oferece os NPCs, serviços ou sistemas pertinentes àquele local, sem carregar conteúdo de outra cidade.
- [ ] Ações enviam o local pretendido como contexto transitório; o backend valida sessão, cidade persistida, cidade do local e permissões. Não depende de uma visita salva nem de um campo de prédio atual no perfil do jogador.
- [ ] NPCs presentes em locais podem ser interrogados por uma interface própria, com opções de perguntas em frases fixas; diálogos respeitam o universo corporativo de Noir e nunca identificam uma fala como pista ou prova.
- [ ] O interrogatório oferece no máximo estas cinco perguntas genéricas, aplicáveis a qualquer NPC e local: “Como tem sido sua rotina ultimamente?”, “Com quem você costuma passar o tempo?”, “Aconteceu algo fora do comum por aqui recentemente?”, “Como vão as coisas no trabalho?” e “Há algo que esteja preocupando você?”.
- [ ] Respostas de interrogatório são texto contextual para orientar a investigação e não são pináveis, não entram no whiteboard e não contam como evidência na acusação.
- [ ] Respostas podem corroborar, contradizer ou não acrescentar informação útil; incluem fatos cotidianos e detalhes relacionados ao caso, sem confirmar diretamente a culpa de um NPC nem expor a classificação interna do conteúdo.
- [ ] Respostas usam modelos de fala com variações coerentes com cotidiano, relações e ocupação do NPC, podendo incluir detalhes relacionados a um caso sem pressupor que o NPC conheça a vítima ou saiba qual caso o jogador investiga; os detalhes sugerem caminhos sem entregar uma conclusão pronta.
- [ ] A solicitação de busca e apreensão identifica a Delegacia da cidade atual como local de protocolo; o backend valida a cidade persistida e registra a autorização concedida ou negada, sem persistir que o jogador está dentro da Delegacia.
- [ ] Uma autorização concedida identifica o local residencial e permite ao jogador viajar até lá para realizar a busca.
- [ ] O inventário de cada residência é procedural, estável e consultado por local; a cidade inteira não é materializada para responder à ação.
- [ ] Crime Generator pode incluir objetos ligados ao crime no inventário do local correto.
- [ ] Objetos relevantes e cotidianos são apresentados sem distinção visual, textual ou comportamental.
- [ ] A busca pode revelar artefatos físicos, documentos e vestígios do local, incluindo achados cotidianos e relacionados ao caso.
- [ ] Achados investigativos persistidos pela busca podem ser pinados no whiteboard como elementos do caso; sua apresentação registra descrição observável e contexto da descoberta, sem revelar classificação ou validade interna.
- [ ] Artefatos e vestígios podem fornecer detalhes verificáveis — por exemplo, inscrição, data, condição ou local de origem — que o jogador pode comparar com emails, registros, relatos e outros achados.
- [ ] A busca persiste os elementos efetivamente encontrados e disponíveis para o jogador, sem revelar sua classificação interna.
- [ ] Sistemas e áreas restritas exigem credenciais; ao solicitar invasão, o jogador passa pelo fluxo de hacking antes de receber ou ter negado acesso, conforme as regras existentes de hacking. O prédio selecionado permanece contexto da ação e não é gravado como localização do jogador.
- [ ] O backend autoriza e registra tentativas de acesso restrito; o frontend não concede acesso nem determina o resultado da invasão.
- [ ] Interrogatórios, buscas e invasões podem produzir resultados cotidianos ou relacionados a casos, sem revelar a classificação interna de nenhum resultado.
- [ ] O fluxo e seus acessos respeitam as [Regras 0001, 0002 e 0003](../../rules/README.md).

## Fora de escopo

Arbitragem da acusação. A organização dos elementos em whiteboards é coberta pelo M10; M12 apenas disponibiliza novos elementos pináveis encontrados em buscas. Análise forense especializada, comparação automática de artefatos e julgamento de credibilidade de depoimentos. Viagem entre cidades e visita inicial a locais são cobertas pelo M11.
