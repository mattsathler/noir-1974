# M29 — Sites de lore e minigames

**Status:** Aceito
**Trilha:** Independente; apresenta lore e atividades opcionais no computador sem alterar o fluxo criminal
**Depende de:** M23 — computador pessoal, Proxy e sites investigativos; M24 — Company Websites procedurais

## Objetivo

Espalhar pela web 20 sites fixos, independentes entre si e com vozes próprias — como blogs e publicações — para apresentar a história do universo do jogo, a criação do GATE, os personagens principais e Adrian Voss. O diretório nacional conecta todas as páginas e links contextuais permitem saltar entre elas, mas não existe um único arquivo central de lore. O computador pessoal também recebe o primeiro minigame jogável: Campo Minado.

As páginas de lore são opcionais e não concedem Credits, pistas de caso ou progresso investigativo. Podem apresentar relatos oficiais, registros históricos e versões conflitantes — especialmente entre Bill Bouer, The Crow, Meridian Global e outras instituições — sem transformar uma versão disputada em fato canônico sem decisão narrativa.

## Direção definida

- Sites secundários são espaços opcionais de leitura e descoberta narrativa; contam a lore e também a vida cotidiana desse universo. Cada publicação mantém identidade, ponto de vista e conteúdo próprios.
- Dentro do universo, Noir não é um lugar nem o nome do mundo; é somente o nome do jogo. Os textos diegéticos tratam as pessoas, instituições e acontecimentos como parte da vida real de seus autores.
- As publicações escrevem como blogs, colunas, jornais, boletins e páginas institucionais: alguém opina, noticia, recomenda, reclama ou explica a partir de sua posição. Evitar prosa enciclopédica sem autor ou ponto de vista.
- Não mencionar o jogo, o jogador, progresso ou mecânicas dentro das páginas. Não citar anos.
- Os assuntos podem incluir vida cotidiana, cultura, viagens, trabalho, comida, finanças, tecnologia, política e outros temas, sem precisar servir diretamente à trama principal.
- As primeiras páginas devem abordar a criação e os efeitos sociais do GATE, apresentar Bill Bouer e The Crow e ampliar a presença pública de Adrian Voss.
- O conteúdo de GATE e Adrian respeita os fatos já documentados em [`lore.md`](../../project/lore.md). Questões ainda não definidas — como a relação pessoal de Adrian com a criação do dispositivo — permanecem abertas até serem decididas na lore.
- Os sites de lore não concedem Credits, itens, reputação, pistas utilizáveis em casos, desbloqueios ou qualquer outro progresso mecânico.
- A navegação pode partir de links em websites primários e anúncios diegéticos, mas não deve usar surveys, publicidade real ou tarefas de clique como recompensa.
- O hub não exibe um índice integral das 20 publicações. A cada visita, apresenta uma seleção aleatória de matérias com deeplinks; cada site também aponta para outras publicações relacionadas e oferece retorno ao hub.
- O primeiro minigame é Campo Minado, acessível como aplicativo no computador pessoal. Ele é uma atividade opcional para passar o tempo e não concede Credits nem outras recompensas econômicas.

## Escopo

### Sites secundários de lore

- Criar vários sites secundários diegéticos, distribuídos pela web, que o jogador possa encontrar no diretório e por links contextuais em sites institucionais e corporativos existentes.
- Manter 20 sites fixos, cada um com URL, organização/publicação e conteúdo próprios.
- Evitar centralizar as histórias em um arquivo ou portal único. Cada blog/publicação conta uma parte do universo e pode apresentar uma voz institucional, jornalística ou contestatória.
- Apresentar a criação do GATE pela Meridian Global, seu lançamento e seus efeitos sobre transporte, comércio, poder corporativo e vida cotidiana, sem inventar detalhes técnicos ainda não definidos.
- Apresentar Bill Bouer, The Crow e Adrian Voss por meio de biografias públicas, entrevistas, arquivos, reportagens e versões institucionais; manter vozes e interesses distintos.
- Usar documentos e páginas com pontos de vista conflitantes quando apropriado. O material do Crow pode contestar o discurso oficial, sem ser automaticamente tratado como verdade.
- Não revelar automaticamente culpados, não fornecer pistas necessárias para casos ativos e não alterar o estado de investigação.
- Manter as páginas opcionais, acessíveis por links claros, legíveis e utilizáveis sem depender de sequência de cliques ou conclusão de tarefas.
- Dar a cada site um caminho funcional de retorno ao diretório nacional, além de deeplinks contextuais para outras publicações e serviços.
- Não oferecer Credits, Crypto, itens, descontos, fama, progresso de caso ou desbloqueios como recompensa por visitar ou ler essas páginas.

### Primeiro minigame: Campo Minado

- Disponibilizar Campo Minado como aplicativo próprio no computador pessoal, integrado ao Meridian Web Browser/MeridOS sem abrir serviço externo.
- Implementar tabuleiro, casas seguras e minas, contagem de minas adjacentes, marcar/desmarcar bandeiras, revelar casas, derrota por mina e vitória ao revelar todas as casas seguras.
- Oferecer instruções e uma configuração inicial de 9×9 casas com 10 minas, adequada ao tamanho da interface do computador.
- Permitir jogar com mouse e teclado, com foco visível, nomes acessíveis para casas, leitura do estado da partida e controles que funcionem nos viewports suportados.
- Reiniciar uma partida deve criar tabuleiro novo e estado coerente. O jogo deve impedir comandos incompatíveis após vitória ou derrota e não deve perder a partida por clique duplo acidental na primeira jogada.
- Não conceder Credits, itens ou progresso de caso por iniciar, vencer ou repetir partidas; o cronômetro registra somente a duração da partida atual.
- Manter o minigame opcional e separado dos requisitos para investigar, acusar ou concluir casos.

## Critérios de aceite

- [x] Vários sites fixos, independentes e opcionais apresentam lore sem conceder Credits, Crypto, itens, fama, pistas necessárias ou progresso mecânico.
- [x] Os 20 sites têm URL e conteúdo próprios e são acessíveis pelo diretório nacional.
- [x] Deeplinks entre sites e retornos ao diretório nacional funcionam em todos os sites de lore.
- [x] O hub apresenta matérias diferentes em visitas distintas sem listar diretamente todos os sites.
- [x] Os sites têm vozes próprias e distribuem histórias; não centralizam todo o conteúdo em um único portal.
- [x] Nenhuma página diegética chama o universo de Noir, trata as pessoas como personagens de um jogo ou cita anos.
- [x] O catálogo inclui assuntos cotidianos variados, escritos por autores e instituições com pontos de vista identificáveis.
- [x] A história da criação e dos efeitos do GATE respeita os fatos canônicos e não antecipa decisões técnicas ainda abertas.
- [x] Bill Bouer, The Crow e Adrian Voss recebem apresentação narrativa própria e coerente com seus interesses e posições no mundo.
- [x] Relatos institucionais e contestatórios podem divergir sem declarar automaticamente qual versão é verdadeira.
- [x] A lore pode ser ignorada sem bloquear navegação, investigação, acusações ou conclusão de casos.
- [x] O Campo Minado abre pelo computador pessoal e permite iniciar, jogar, vencer, perder e reiniciar uma partida funcional.
- [x] Campo Minado calcula números adjacentes, bandeiras e condição de vitória corretamente, incluindo tabuleiro e primeiro clique seguros conforme as regras implementadas.
- [x] Cada ação do Campo Minado é acessível por teclado e ponteiro, com estado anunciável, foco visível e interface utilizável nos viewports suportados.
- [x] Campo Minado é uma atividade opcional sem recompensas econômicas ou progresso de caso; o cronômetro é apenas informativo.
- [x] Nenhum site de lore ou minigame é requisito para completar casos, acusar ou prender pessoas.
- [x] A suíte de testes cobre regras do Campo Minado, navegação e acessibilidade das páginas de lore.

## Fora de escopo

- Publicidade, surveys, coleta de respostas ou compartilhamento de dados com serviços externos.
- Premiar visitas a sites com Credits ou usar páginas de lore como substituto de evidência, acusação ou resolução de casos.
- Criar neste marco vários minigames adicionais além do Campo Minado; a arquitetura e os pontos de entrada devem permitir que novos jogos sejam adicionados depois.
- Alterar as regras de bolsa e flutuação definidas pelas M25/M26 ou a geração criminal prevista nas M27/M28.
