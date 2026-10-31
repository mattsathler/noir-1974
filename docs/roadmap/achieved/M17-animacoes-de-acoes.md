# M17 — Animações das ações do jogo

**Ciclo:** 17  
**Status:** Concluída
**Depende de:** M14 aceito

> A cena de viagem entre bairros e a introdução visual de destino foram substituídas pelo fluxo de Proxy e sites da [M23](M23-computador-os-proxy-e-sites.md). Permanecem nesta milestone os personagens centrais e o sistema de conversas.

> A apresentação dos personagens também deve estabelecer o clima político do universo. O conflito entre Bill Bouer e The Crow aparece de forma recorrente nos canais cotidianos do jogador: a imprensa estatal atribui a Bouer relatos da investigação e enquadra os atos do Crow como crimes; o Crow justifica invasões e roubos como combate ao sistema. O jornal também alerta contra Ghostkey e outras ferramentas ilícitas, deixando visível a contradição de um regime autoritário que criminaliza o acesso à informação.

## Objetivo

Estabelecer os personagens centrais da narrativa policial e criar um sistema de conversa direta entre NPCs e jogador. A direção visual permanece coerente com a identidade de Noir 1974, em pixel art.

## Critérios de aceite

- [x] Bill Bouer é estabelecido como o personagem policial protagonista, e The Crow como seu antagonista. Os dois têm retratos distintos em pixel art e vozes próprias nas conversas diretas; a caveira do Crow aparece como sua assinatura visual.
- [x] Cada edição diária do jornal inclui um relato atribuído a Bill Bouer sobre a investigação para capturar The Crow e um aviso estatal sobre Ghostkey ou outras ferramentas ilícitas. As versões variam ao longo dos dias e apresentam a justificativa política do Crow junto ao enquadramento oficial, sem o reduzir a herói ou vilão.
- [x] O jornal, as conversas diretas e os emails oficiais assinados por Bill reforçam a disputa entre Bouer e The Crow e a tensão entre segurança estatal, vigilância e acesso à informação. O jornal e as mensagens variam; as conversas apresentam vozes próprias.
- [x] Relatos telefônicos de NPCs e correspondência cotidiana incorporam essa disputa no conteúdo dos depoimentos e mensagens, com variação de voz e sem repetir um único texto em toda parte.
- [x] Conversas diretas com NPCs abrem uma overlay que cobre a tela, escurece o fundo e posiciona a face do NPC atual no centro.
- [x] A fala do NPC aparece em uma box quadrada na parte inferior da overlay, com efeito typewriter.
- [x] As opções de resposta aparecem em botões quadrados abaixo da fala e seguem a linguagem visual das opções da conversa telefônica.
- [x] Escolher uma resposta avança a conversa e apresenta a próxima fala e suas opções sem fechar a overlay ou perder o NPC atual.
- [x] O sistema aceita Bill Bouer e The Crow como interlocutores e comporta futuras conversas tutoriais em formato de chat. O roteiro tutorial continua definido como etapa narrativa.
- [x] A conversa permanece legível e operável por teclado e leitor de tela; o texto completo fica disponível desde o início para tecnologias assistivas e o typewriter respeita preferência por movimento reduzido.
- [x] O typewriter das conversas e as esperas temporizadas das ligações têm duração e encerramento definidos. O texto completo continua acessível, e o typewriter respeita a preferência por movimento reduzido.
- [x] O jogador tem uma central de notificações overlay com snackbar e histórico. Novos emails recebidos durante a sessão e conexões telefônicas bem-sucedidas geram notificações; o histórico permanece acessível durante a sessão.

## A definir em etapa posterior

- [ ] Roteiro, ordem e gatilhos das conversas tutoriais específicas de Bill Bouer e The Crow; esta etapa entrega o sistema visual e interativo para suportá-las.

## Fora de escopo

Redesenho geral da interface, mapa navegável ou simulação de rotas e distâncias, animação contínua de personagens no mundo, deslocamentos animados do policial durante buscas ou prisões até que sejam especificados, alterações às regras de investigação ou prisão, e reprodução de personagens, interfaces ou identidade visual de qualquer outra obra. O sistema operacional, Proxy e navegação por sites pertencem à M23.
