# Direção do jogo e próximas milestones

Este documento registra a direção de produto acordada após o fechamento do ciclo principal de investigação. O fluxo de recolher provas, acusar e prender continua sendo uma base jogável, mas não é o destino final da experiência: ele deve servir para contar a história do mundo e aproximar o jogador dos casos e das pessoas envolvidas.

## Direção aceita

### Dualidade do jogador

Hacking é ilegal em qualquer circunstância, inclusive quando usado para ajudar alguém. O jogador pode invadir sistemas para investigar casos ou para obter bens e vantagens para si. O jogo deve permitir que ele aja por altruísmo, interesse próprio ou uma mistura dos dois, e fazer essas escolhas terem consequências.

Bill Bouer e The Crow representam posições opostas diante da lei, do Estado e das grandes corporações. Bill trabalha dentro das instituições, mesmo quando elas são autoritárias e corporativistas. The Crow é um foragido que se opõe ao Estado e às corporações, mas não é automaticamente inimigo do jogador; ele também usa hacking e rouba de empresas. Essas perspectivas não devem se reduzir a uma divisão simples entre herói e vilão.

### Casos como narrativa

Resolver um caso é uma parte essencial do jogo, mas o ciclo de investigação é uma estrutura de jogo repetível, não um ponto final narrativo. As próximas histórias devem dar contexto às vítimas, suspeitos, instituições e consequências das escolhas do jogador.

O correio eletrônico será uma fonte direta de casos e relações. Além de spam, NPCs poderão procurar o jogador para pedir ajuda em desaparecimentos negligenciados pelas autoridades, denunciar desvios dentro de empresas ou relatar outros problemas. Essas mensagens devem criar vínculos pessoais com as histórias, complementando pistas públicas como o jornal.

O modelo de casos também deve sair do homicídio obrigatório com uma única vítima: crimes financeiros e corporativos poderão envolver empresas ou instituições afetadas sem vítima humana, e uma trama poderá ter mais de um acusado, incluindo cúmplices.

Depois dessa expansão, um caso poderá reunir vários crimes que surgem em etapas enquanto o mundo progride, em vez de nascer completo. O telefone e as conversas em tempo real serão removidos; emails assíncronos serão o único canal de comunicação direta entre o jogador e NPCs. Isso permite que relações e ações dos NPCs produzam novos acontecimentos — inclusive outras vítimas — enquanto o jogador investiga.

NPCs fixos do universo devem ganhar histórias, diálogos, emails e relações que se revelem progressivamente conforme os casos avançam. O que o jogador descobre e decide fazer deve influenciar como essas pessoas se relacionam com ele.

### Progressão hacker

O `black_market.exe` será a porta de entrada para baixar e usar ferramentas ilegais. A progressão deverá conectar nível, dinheiro e fama online do jogador ao acesso a esses aplicativos.

Ghostkey e Mailbreaker poderão começar desinstalados. Um tutorial curto deverá explicar como obter e usar ferramentas pelo mercado clandestino, introduzindo seus riscos e consequências.

A primeira etapa da progressão adicionará fama em pontos e saldo em Credits por personagem/servidor, sem ranking. O jogador poderá investir em ações de empresas nacionais pelo desktop. Descobrir ou solucionar casos não pagará Credits nesta etapa. Uma etapa seguinte fará a cotação de cada empresa reagir a acontecimentos associados a ela, principalmente homicídios ligados a extorsões ou conflitos corporativos; a pontuação de mercado será isolada por servidor e não pagará Credits diretamente.

## Online — proposta em avaliação

O modelo desejado é permitir que cada jogador crie personagens vinculados a mundos/servidores diferentes e escolha entre:

- **Singleplayer:** mundo e estado privados do personagem.
- **Multiplayer:** personagem participando de um mundo cujo estado é compartilhado com os demais jogadores daquele servidor.

A arquitetura atual já associa um mundo determinístico a cada servidor por meio de uma seed, e isola as mutations e os casos por `server_id`. Portanto, mundos públicos compartilhados e mundos privados de jogador são compatíveis com a base conceitual existente. Um mundo singleplayer pode ser tratado como uma instância privada, equivalente a um servidor de um único jogador.

Uma seed, sozinha, não define o compartilhamento. Ela reconstrói a sociedade base; a identidade da instância determina o escopo das mutations, dos casos e do progresso. Assim, dois mundos podem partir da mesma seed e ainda manter estados mutáveis isolados. Compartilhar um mundo multiplayer exige compartilhar tanto a base determinística quanto a instância de estado.

O modelo de personagem por mundo/servidor ainda precisa de decisão formal. Antes de implementar, definir se um personagem pode existir em mais de uma instância, como funciona a entrada em servidores existentes, e se um mundo singleplayer pode ser compartilhado depois sem perder ou expor seu estado privado.

## Direção para as próximas milestones

As próximas milestones devem desenvolver a lore do universo e apresentá-la ao jogador por sistemas que já fazem parte da rotina — principalmente emails, diálogos, casos, jornal, websites e relações com NPCs. Explorar deve ser recompensante: fuçar páginas, sistemas e conteúdo opcional pode revelar lore, contexto investigativo e recompensas da economia do jogo. Essas recompensas devem refletir o mundo e respeitar o saldo isolado por personagem/servidor, sem transformar a investigação numa sequência de acusações idênticas.

A M29 inicia essa direção com atividades opcionais nos websites e Campo Minado no computador pessoal. Seu escopo base usa Credits fictícios e surveys diegéticos sem coleta externa; pagamentos em dinheiro real, anúncios de terceiros ou surveys comerciais exigem decisão e escopo separados antes de qualquer implementação.

O Online permanece uma proposta arquitetural até que o modelo de personagens, instâncias e compartilhamento de estado seja fechado em um ADR aceito.
