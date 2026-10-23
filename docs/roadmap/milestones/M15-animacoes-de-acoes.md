# M15 — Animações das ações do jogo

**Ciclo:** 15  
**Status:** Não iniciado  
**Depende de:** M14 aceito

## Objetivo

Dar mais presença visual às ações do jogador por meio de animações curtas e legíveis, reforçando quem está agindo, para onde está indo e o que aconteceu. A linguagem deve evocar a aventura investigativa de Carmen Sandiego sem reproduzir seus personagens, interfaces, mapas ou identidade visual, e permanecer coerente com a direção artística de Noir 1974.

## Critérios de aceite

- [ ] Toda mudança de localização do jogador — por viagem, entrada/saída de um local ou qualquer outro fluxo que altere sua localização — apresenta uma introdução em efeito typewriter com o nome do local onde o jogador está após a transição.
- [ ] A introdução só começa quando a mudança de localização é confirmada; cancelamentos e falhas não exibem o destino como localização atual. Transições consecutivas não deixam animações antigas sobrescreverem o local mais recente.
- [ ] O texto da localização permanece legível e pode ser concluído imediatamente por ação do jogador; nome e estado atual também ficam disponíveis sem depender da animação.
- [ ] Ao viajar para uma construção, a interface apresenta uma animação de deslocamento com o retrato do jogador à esquerda e o ícone da construção de destino à direita.
- [ ] A animação de viagem identifica claramente o destino e termina no estado de chegada correspondente; falhas ou cancelamentos não comunicam que a viagem foi concluída.
- [ ] Existe um personagem policial original, em pixel art e compatível com a identidade visual de Noir 1974, utilizável nas ações de busca e apreensão e de mandado de prisão.
- [ ] Ao solicitar uma busca e apreensão, a interface apresenta o policial deslocando-se até a residência do alvo, sem revelar ou persistir o endereço residencial quando as regras do fluxo não o permitem.
- [ ] Ao executar um mandado de prisão, a interface apresenta o policial deslocando-se até o alvo e comunica o resultado autoritativo da ação, sem sugerir uma prisão quando ela não ocorreu.
- [ ] As animações usam retratos, personagens, ícones e cenários coerentes entre si e com a direção artística obrigatória de pixel art; não incorporam elementos distintivos de Carmen Sandiego.
- [ ] O texto ou estado da ação permanece compreensível sem depender apenas do movimento, inclusive quando a animação é reduzida, interrompida ou não pode ser exibida.
- [ ] As animações não alteram regras, resultados, permissões ou persistência das ações; o backend continua sendo a fonte do resultado de busca, viagem e prisão.
- [ ] Cada animação tem duração e encerramento definidos, não bloqueia a interface indefinidamente e mantém controles e feedback utilizáveis em telas suportadas.

## Fora de escopo

Redesenho geral da interface, mapa navegável ou simulação de rotas e distâncias, animação contínua de personagens no mundo, alteração das regras de viagem, busca e apreensão ou prisão, e reprodução de assets ou apresentação de Carmen Sandiego.
