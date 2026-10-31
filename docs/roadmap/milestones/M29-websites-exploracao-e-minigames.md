# M29 — Websites, exploração e minigames

**Status:** Proposta  
**Trilha:** Independente; amplia exploração e atividades no computador sem alterar o fluxo criminal  
**Depende de:** M23 — computador pessoal, Proxy e sites investigativos; M24 — Company Websites procedurais; M25 — saldo em Credits por personagem/servidor  

## Objetivo

Fazer da exploração da web uma atividade recompensante e integrada à investigação. O jogador que percorre páginas, examina anúncios, encontra conteúdo escondido e participa de atividades opcionais deve descobrir lore, pistas e recompensas do mundo. O computador pessoal também recebe o primeiro minigame jogável: Campo Minado.

As recompensas devem despertar curiosidade e dar contexto ao universo sem transformar navegação em cliques repetitivos, revelar automaticamente culpados ou substituir as ferramentas de investigação. O conteúdo pode mostrar versões conflitantes dos fatos — inclusive as disputas entre Bill Bouer, The Crow, instituições e empresas — sem declarar uma única interpretação como verdade antes das evidências.

## Direção definida

- Websites são espaços para investigar e descobrir histórias, não apenas páginas de consulta.
- Anúncios podem esconder referências, mensagens, pistas de lore ou oportunidades de recompensa. O jogador precisa explorar e interpretar o conteúdo; anúncios não devem ser pop-ups obrigatórios.
- Surveys são atividades diegéticas, opcionais e sobre produtos, instituições e acontecimentos fictícios do universo. Não coletam respostas nem dados pessoais fora do jogo.
- A moeda desta entrega é **Credits**, já existente no perfil por personagem/servidor. Ganhos são registrados no escopo econômico atual e não podem ser criados exclusivamente pelo cliente.
- O primeiro minigame é Campo Minado, acessível como aplicativo no computador pessoal. O resultado do minigame pode conceder recompensa limitada, mas não pode virar fonte repetível de saldo sem limite.
- Recompensas em dinheiro real, pagamentos externos, anúncios de terceiros e surveys comerciais não fazem parte do contrato implícito desta milestone. Se essa ideia for mantida, precisa de decisão própria sobre pagamentos, elegibilidade, consentimento, privacidade, fraude e requisitos legais antes de entrar em escopo implementável.

## Escopo

### Exploração recompensante de websites

- Criar atividades descobríveis em sites existentes, começando por módulos de anúncios e conteúdo promocional dos Company Websites e páginas públicas selecionadas.
- Incluir oportunidades de três tipos: descoberta de lore, informação útil para investigação e atividade opcional com recompensa em Credits.
- Criar surveys ficcionais curtos, com contexto diegético, perguntas sobre o mundo do jogo e conclusão explícita; respostas não alteram perfil psicológico nem são enviadas a serviços externos.
- Permitir que uma pista seja contextualizada sem marcar automaticamente um NPC como culpado. Quando adequada, a informação investigativa deve poder ser registrada/revisitada na interface de investigação existente.
- Distribuir achados em páginas e empresas distintas, relacionando-os ao mundo, à empresa, ao país ou à narrativa; a exploração não deve depender de sequência de cliques sem significado.
- Apresentar anúncios como parte visual e textual dos websites, mantendo links funcionais claros, rótulos acessíveis e diferença legível entre publicidade diegética e controles de navegação.
- Reutilizar o sistema procedural/versionado dos websites quando o conteúdo for derivado da empresa, com resultado estável para a mesma identidade e versão do mundo.

### Créditos e integridade de recompensas

- Definir tipos e valores explícitos de recompensa para atividades incluídas, com limite ou condição de conclusão verificável.
- Registrar recompensas de forma autoritativa e idempotente no backend, isoladas por personagem/servidor; retries, recarga e cliques concorrentes não podem creditar duas vezes a mesma conclusão.
- Exibir o ganho e o motivo de forma clara ao jogador, atualizando saldo e histórico sem exigir reiniciar o computador.
- Definir quais achados são repetíveis, únicos ou renováveis, evitando que recarregar a página, limpar estado local ou refazer o survey gere saldo ilimitado.
- Manter atividades sem recompensa econômica quando sua função principal for revelar lore ou oferecer um desafio opcional.

### Primeiro minigame: Campo Minado

- Disponibilizar Campo Minado como aplicativo próprio no computador pessoal, integrado ao NOIR Browser/MeridOS sem abrir serviço externo.
- Implementar tabuleiro, casas seguras e minas, contagem de minas adjacentes, marcar/desmarcar bandeiras, revelar casas, derrota por mina e vitória ao revelar todas as casas seguras.
- Oferecer instruções e pelo menos uma configuração de dificuldade adequada ao tamanho da interface do computador.
- Permitir jogar com mouse e teclado, com foco visível, nomes acessíveis para casas, leitura do estado da partida e controles que funcionem nos viewports suportados.
- Reiniciar uma partida deve criar tabuleiro novo e estado coerente. O jogo deve impedir comandos incompatíveis após vitória ou derrota e não deve perder a partida por clique duplo acidental na primeira jogada.
- Integrar uma recompensa de Credits limitada à conclusão elegível definida pela milestone; repetir partidas não pode gerar recompensa ilimitada.
- Manter o minigame opcional e separado dos requisitos para investigar, acusar ou concluir casos.

## Critérios de aceite

- [ ] Websites têm descobertas opcionais que recompensam exploração com lore, contexto investigativo ou atividade com recompensa; não basta abrir uma página sem explorar seu conteúdo.
- [ ] Conteúdo promocional e anúncios são diegéticos, navegáveis quando apresentados como links, acessíveis e não bloqueiam o uso normal do website.
- [ ] Surveys são ficcionais, voluntários, explicam sua finalidade dentro do mundo e não enviam dados pessoais ou respostas para fora do jogo.
- [ ] Uma descoberta narrativa pode apresentar versões conflitantes de Bill Bouer, The Crow ou instituições sem declarar culpa como fato investigativo não comprovado.
- [ ] Recompensas econômicas usam Credits do personagem/servidor atual, têm motivo e valor comunicados e atualizam o saldo exibido.
- [ ] O backend valida e registra cada recompensa de modo atômico/idempotente; retries, concorrência, reload, replay de requisição ou manipulação do cliente não duplicam saldo.
- [ ] Atividades únicas ou limitadas continuam concluídas após recarga e não podem ser repetidas para farmar Credits; regras de renovação, se existirem, são explícitas.
- [ ] O Campo Minado abre pelo computador pessoal e permite iniciar, jogar, vencer, perder e reiniciar uma partida funcional.
- [ ] Campo Minado calcula números adjacentes, bandeiras e condição de vitória corretamente, incluindo tabuleiro e primeiro clique seguros conforme as regras implementadas.
- [ ] Cada ação do Campo Minado é acessível por teclado e ponteiro, com estado anunciável, foco visível e interface utilizável nos viewports suportados.
- [ ] A recompensa do Campo Minado é validada no backend e limitada por uma regra explícita; não há repetição ilimitada de ganho.
- [ ] Nenhuma atividade de websites ou minigame é requisito para completar casos, acusar ou prender pessoas.
- [ ] A suíte de testes cobre regras do Campo Minado, navegação/atividades web, persistência e idempotência de recompensas e isolamento entre personagens/servidores.

## Decisões necessárias antes do aceite

- Valor e limites das recompensas em Credits por descoberta, survey e minigame; quais ações são únicas, repetíveis ou renováveis.
- Quais empresas, instituições e sites recebem as primeiras atividades e como a distribuição procedural evita duplicação de conteúdo.
- Se pistas descobertas em sites entram no whiteboard automaticamente ou ficam guardadas numa caixa de descobertas para o jogador registrar.
- Quais temas e formatos de survey fictício cabem na ambientação autoritária e como deixar claro que responder é opcional.
- Níveis de dificuldade, dimensões e configuração de minas do Campo Minado; se recordes e estatísticas ficam locais ou persistem por personagem/servidor.
- Se recompensas reais externas continuam sendo um objetivo de produto. Qualquer entrega desse tipo exige escopo e autorização separados, integração de pagamentos explicitamente definida e revisão de privacidade, fraude e conformidade; M29 usa Credits fictícios até essa decisão.

## Fora de escopo

- Pagamento, saque, conversão de Credits em dinheiro real ou qualquer promessa de ganho financeiro fora do jogo.
- Anúncios de terceiros, rastreamento publicitário, coleta de respostas reais, pesquisa de mercado ou compartilhamento de dados com patrocinadores.
- Transferir Credits entre jogadores, personagens ou servidores.
- Usar cliques, pesquisas ou minigames como substituto de evidência validada, de acusação ou de resolução de casos.
- Criar neste marco vários minigames adicionais além do Campo Minado; a arquitetura e os pontos de entrada devem permitir que novos jogos sejam adicionados depois.
- Alterar as regras de bolsa e flutuação definidas pelas M25/M26 ou a geração criminal prevista nas M27/M28.
