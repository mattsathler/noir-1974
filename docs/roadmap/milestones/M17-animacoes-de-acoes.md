# M17 — Animações das ações do jogo

**Ciclo:** 17  
**Status:** Em andamento
**Depende de:** M14 aceito

> A introdução do destino após confirmação da viagem foi separada para [I02](../melhorias/I02-introducao-destino-gate.md), originada nesta milestone antes da troca de numeração.

## Objetivo

Dar mais presença visual às viagens entre bairros, estabelecer os personagens centrais da narrativa policial e criar um sistema de conversa direta entre NPCs e jogador. A direção visual permanece coerente com a identidade de Noir 1974, em pixel art.

## Critérios de aceite

- [ ] Viagens entre bairros apresentam uma cena de deslocamento com um carro em pixel art sobre um fundo de rua. A cena identifica o bairro de destino e termina na chegada confirmada.
- [ ] Enquanto os PNGs finais do carro e do cenário não estiverem prontos, a cena usa substitutos visuais temporários. A estrutura permite trocar esses substitutos pelos assets finais sem alterar as regras da viagem.
- [ ] Bill Bouer é estabelecido como o personagem policial protagonista, e The Crow como seu antagonista. Os dois devem manter identidade visual própria e coerente com a direção em pixel art de Noir 1974.
- [ ] Conversas diretas com NPCs abrem uma overlay que cobre a tela, escurece o fundo e posiciona a face do NPC atual no centro.
- [ ] A fala do NPC aparece em uma box quadrada no canto inferior da overlay, com efeito typewriter.
- [ ] As opções de resposta aparecem em botões quadrados abaixo da fala e seguem a linguagem visual das opções da conversa telefônica.
- [ ] Escolher uma resposta avança a conversa e apresenta a próxima fala e suas opções sem fechar a overlay ou perder o NPC atual.
- [ ] O sistema aceita Bill Bouer e The Crow como interlocutores e comporta futuras conversas tutoriais em formato de chat. O conteúdo e a sequência específicos do tutorial podem ser escritos como etapa narrativa própria.
- [ ] A conversa permanece legível e operável por teclado e leitor de tela; o typewriter não é o único meio de acessar o texto completo e as opções.
- [ ] As cenas apresentam estados de partida, viagem e chegada sem sugerir uma chegada quando a viagem falhar ou for cancelada.
- [ ] Texto e estado da viagem continuam compreensíveis sem depender apenas do movimento, inclusive quando animações não puderem ser exibidas.
- [ ] As animações não alteram regras, resultados, permissões ou persistência: o backend continua sendo a fonte autoritativa do resultado da viagem.
- [ ] Cada animação tem duração e encerramento definidos, não bloqueia a interface indefinidamente e mantém feedback utilizável nas telas suportadas.

## A definir em etapa posterior

- [ ] Animações de deslocamento de personagens durante buscas e apreensões, mandados de prisão ou outras ações policiais. Não implementar deslocamentos do policial nesta etapa; a direção e o funcionamento serão detalhados em um próximo prompt.
- [ ] Roteiro, ordem e gatilhos das conversas tutoriais específicas de Bill Bouer e The Crow; esta etapa entrega o sistema visual e interativo para suportá-las.

## Fora de escopo

Redesenho geral da interface, mapa navegável ou simulação de rotas e distâncias, animação contínua de personagens no mundo, deslocamentos animados do policial durante buscas ou prisões até que sejam especificados, alteração das regras de viagem, busca e apreensão ou prisão, e reprodução de personagens, interfaces ou identidade visual de qualquer outra obra.
