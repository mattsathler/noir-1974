# M28 — Casos evolutivos e comunicação por email

**Status:** Concluída
**Trilha:** Independente; amplia o modelo criminal e a simulação do mundo
**Depende de:** M08 — correio eletrônico; M10 — mundo compartilhado e casos recorrentes; M27 — crimes financeiros e múltiplos acusados

## Objetivo

Refatorar a relação entre crime e caso. Um caso deixa de ser sinônimo de um único crime definido por inteiro na criação: passa a agrupar um ou mais crimes que podem acontecer em etapas enquanto o mundo continua vivendo. O caso permanece em andamento enquanto sua condição de encerramento não for alcançada.

Remover o telefone como canal de contato e interrogatório de NPCs. NPCs procedurais se comunicam diretamente com o jogador por email; NPCs fixos são uma exceção narrativa e podem iniciar conversas roteirizadas em momentos específicos.

## Direção definida

- **Crime** é um acontecimento individual com tipo, envolvidos, momento e fatos próprios.
- **Caso** é uma investigação que pode agrupar vários crimes relacionados e evoluir com novos acontecimentos.
- Um caso pode começar com um crime e receber outros depois; não é necessário revelar ou gerar toda a sequência no instante da criação.
- Os eventos do mundo podem avançar enquanto o jogador investiga outras coisas. Exemplo: Alice descobre uma fraude financeira e confronta Bianca; enquanto o jogador trabalha no board, Bianca assassina Alice.
- Um caso em andamento tem condições de encerramento explícitas. A geração de novos crimes relacionados para quando o caso é resolvido ou termina por outra condição definida.
- Um serial killer poderá cometer crimes com múltiplas vítimas em momentos diferentes, mantendo o caso aberto até ser descoberto e preso ou até outra condição de encerramento prevista.
- NPCs procedurais não têm ligação, conversa ao vivo ou interrogatório em tempo real; emails podem ser enviados, recebidos e respondidos de forma assíncrona.
- NPCs fixos não ficam disponíveis a qualquer momento. Eles iniciam conversas roteirizadas por gatilhos explícitos do jogo, por exemplo para conduzir um tutorial ou apresentar um acontecimento. Não há lista persistente para o jogador contatá-los quando quiser.
- A conversa de um NPC fixo é uma cena de diálogo própria, iniciada pelo jogo, com falas e opções de resposta; não representa uma chamada telefônica nem exige que o NPC esteja continuamente online.
- Consulta de registros, pesquisa de NPCs e acesso a websites continuam disponíveis: a remoção é do contato direto síncrono, não das fontes investigativas.

### Primeira decisão de implementação

- O telefone, a discagem, os números publicados de funcionários e a rota de entrevista em tempo real serão removidos.
- A caixa de Personal Mail passa a oferecer uma composição guiada para funcionários: o jogador informa o email publicado e escolhe entre as mesmas cinco perguntas da antiga entrevista.
- O NPC recebe uma mensagem assíncrona e a resposta chega à caixa do jogador depois de um intervalo curto. A composição nunca mostra a resposta antes da entrega.
- A resposta é registrada como email normal e mantém remetente, assunto e horário; não cria uma conversa ao vivo nem uma lista permanente de contatos.
- Emails passam a ser o único canal de contato entre jogador e NPCs; chamadas, discador e números de funcionários foram removidos.

## Escopo

### Modelo de crimes e casos

- Separar a entidade persistente `case` dos acontecimentos criminais que pertencem a ela; um caso pode referenciar vários crimes.
- Cada crime/evento mantém tipo, horário, local quando aplicável, participantes e relações causais com outros eventos.
- Permitir que um caso reúna homicídios, crimes financeiros e outros crimes relacionados conforme os modelos estabelecidos na M27.
- Suportar vítimas diferentes em crimes diferentes dentro do mesmo caso, vários acusados e cúmplices, além de crimes sem vítima individual.
- Persistir o progresso do caso incrementalmente: eventos confirmados não são recriados nem duplicados após retry, concorrência ou reinício.
- Manter a verdade, os próximos eventos planejados e as relações internas protegidos no backend; não revelar antecipadamente crimes que ainda não aconteceram.
- Manter condições de evolução e encerramento explícitas, versionadas e consistentes entre jogadores do mesmo servidor.
- Casos concluídos ou encerrados não geram novos crimes associados. Regras de encerramento devem preservar o histórico de eventos e decisões já registrados.

### Evolução do mundo

- Permitir que NPCs realizem ações e reajam a acontecimentos existentes ao longo do tempo, podendo criar novos eventos que ampliem um caso.
- O mundo pode progredir enquanto o jogador investiga outro caso, consulta fontes ou organiza o whiteboard.
- As ações devem respeitar o estado conhecido pelos NPCs, relações relevantes, local, horário e disponibilidade das pessoas envolvidas.
- Definir como NPCs descobrem informações, confrontam outras pessoas e desencadeiam consequências; não usar conhecimento secreto do jogador como conhecimento automático dos NPCs.
- A geração incremental deve produzir pistas consultáveis depois do acontecimento, mantendo coerência temporal entre emails, registros, locais e participantes.
- O primeiro recorte deve demonstrar pelo menos uma cadeia de acontecimentos que progrida em etapas depois da abertura do caso e possa terminar por descoberta/prisão dos responsáveis.
- O sistema deve aceitar futuramente um caso de serial killer com múltiplos homicídios e vítimas em tempos diferentes, sem presumir que o assassino já foi preso na abertura.
- O sistema também deve aceitar futuramente casos financeiros recorrentes, como um constante desvio de dinheiro da empresa. Salários alterados, etc.

### Comunicação com NPCs

- Remover o app/dispositivo de telefone, números de telefone em diretórios de funcionários, ligações, respostas faladas e interrogatórios de NPCs procedurais em tempo real.
- Manter o componente de conversa de NPC fixo como ferramenta narrativa acionada por momentos definidos pelo jogo, não como contato sob demanda.
- Remover do fluxo de investigação botões, estados, componentes e APIs cujo objetivo seja iniciar ou manter uma chamada síncrona com NPC.
- Usar emails assíncronos para mensagens diretas, pedidos de ajuda, denúncias, respostas e atualizações dos NPCs procedurais e mensagens não roteirizadas dos personagens fixos.
- Mensagens devem identificar participantes e horário de envio de forma coerente com o mundo; respostas não precisam acontecer imediatamente.
- Manter caixas de email, intervalo de datas, Email Archive, Personal Mail e Email Breaker segundo suas funções atuais; a M28 não substitui o hacking de caixas postais.
- Conversas ou respostas por email podem revelar contexto e orientar a investigação, mas não devem confirmar diretamente a culpa nem expor a classificação interna das provas.

### Investigação e decisão

- Apresentar um caso em andamento como uma linha temporal/lista de crimes e acontecimentos conhecidos, atualizada conforme novos eventos ficam disponíveis ao jogador.
- Permitir organizar elementos de vários crimes relacionados no mesmo whiteboard sem misturar o estado de eventos de outros casos.
- Referenciar vítimas, acusados e cúmplices por crime/evento e também no contexto agregado do caso.
- A acusação considera os crimes efetivamente ocorridos até aquele momento; eventos futuros não podem ser incluídos como evidência.
- A prisão ou encerramento de um caso interrompe os eventos futuros definidos como dependentes dos responsáveis presos, sem apagar ocorrências já realizadas.
- Eventos corporativos podem emitir os fatos vinculados à empresa para o mecanismo de mercado da M26, sem reimplementar ou alterar a lógica da bolsa.

## Critérios de aceite

- [x] Um caso pode conter vários crimes persistidos, com IDs e horários próprios, sem duplicação após retry ou reinício.
- [x] Novos crimes podem ser anexados a um caso depois de sua criação, sem gerar antecipadamente toda a sequência.
- [x] O backend define condições explícitas para manter, encerrar ou impedir a evolução de um caso.
- [x] Eventos futuros ainda não ocorridos não aparecem em Crime Reports, emails, diretórios, registros, whiteboards ou respostas da API pública.
- [x] O mundo consegue progredir e adicionar um acontecimento enquanto o jogador está ocupado com outras ferramentas/casos.
- [x] Uma cadeia de eventos demonstrativa pode evoluir em etapas a partir das ações e do conhecimento dos NPCs e produz registros investigáveis coerentes.
- [x] O modelo comporta um caso de serial killer com múltiplas vítimas em momentos diferentes, aberto até prisão/descoberta ou condição de término configurada.
- [x] As rotas e telas atuais de telefone, números de staff, chamadas e interrogatórios de NPCs procedurais em tempo real são removidas ou substituídas por email assíncrono.
- [x] NPCs procedurais usam email como único canal de comunicação direta; NPCs fixos só iniciam conversa roteirizada quando um gatilho de narrativa/tutorial ocorrer.
- [x] Não existe lista ou botão permanente para o jogador iniciar contato com NPCs fixos a qualquer momento.
- [x] Emails de casos em andamento podem chegar em momentos diferentes e não precisam aguardar uma nova sessão do jogador para existir, conforme a política temporal definida.
- [x] Mensagens e registros apresentam cronologia consistente com os eventos do caso.
- [x] Investigação, whiteboard, acusação e decisão conseguem identificar a qual crime ou conjunto de crimes cada elemento pertence.
- [x] Acusar responsáveis não utiliza eventos futuros e preserva a garantia de não prender inocentes.
- [x] Encerrar um caso impede novos eventos vinculados e preserva integralmente seu histórico.
- [x] Casos existentes de homicídio continuam reconstruíveis e consultáveis durante a migração do modelo.
- [x] Participantes, crimes, relações e provas permanecem autoritativos no backend; dados de progresso compartilhados são consistentes para jogadores do mesmo servidor.

### Implementação entregue

- Cada caso recebe um crime primário persistido ao ser criado. Casos existentes recebem esse registro por migração idempotente, preservando tipo, data, local e participantes.
- `AppendCrime` e `AppendLinkedCrime` adicionam eventos ocorridos a casos abertos, com IDs e chaves de origem estáveis, pais causais opcionais, sequência e participantes por evento. A gravação é transacional; retries idênticos são idempotentes e conflitos são rejeitados.
- Provas, registros GATE e emails de caso guardam o crime de origem. Os dados antigos são associados ao crime primário pela migração; novas provas podem ser anexadas com `AttachCrimeEvidence`.
- O backend publica a timeline apenas com eventos já ocorridos; novos eventos só ficam listáveis depois de terem prova registrada. A timeline mostra vítimas públicas por evento, sem expor acusados, cúmplices ou relações internas.
- O worker de casos recorrentes executa planos internos versionados. O primeiro arco financeiro programa, três minutos após a abertura do caso, uma adulteração de registros pelo cúmplice. O plano é privado, processado mesmo sem sessão ativa, preserva o horário planejado após reinício e é cancelado quando o caso fecha ou o responsável não estava vivo naquele horário.
- A adulteração grava numa transação o crime filho, a relação causal com o crime financeiro, participante, prova, registro GATE, email entre NPCs e uma denúncia de funcionário na caixa pessoal dos jogadores. A timeline, os registros e emails aplicam limites de tempo para não exibir mensagens futuras.
- O email de denúncia é persistido por servidor/seed para contas existentes, sem depender de login. A caixa pessoal identifica mensagens ligadas a eventos de caso.
- NPCs recebem emails do jogador na própria caixa. O canal assíncrono não bloqueia envio para NPC morto/desaparecido; o backend ignora a geração de resposta nesses estados.
- O endpoint de timeline, as provas, os registros e emails carregam IDs e tipos de crime. O whiteboard preserva essa metadata; a submissão de acusação salva os crimes representados pelas provas no momento do envio e permite associar novas vítimas ao caso sem usar eventos futuros.
- O modelo permite anexar qualquer número de crimes e vítimas a um caso aberto, incluindo homicídios em momentos diferentes; o primeiro arco procedural implementado é financeiro. Conversas dos NPCs fixos continuam sem lista de contato, acionadas pelo evento narrativo `noir:fixed-npc-contact`.

### Regras temporais e encerramento adotados

- O relógio dos eventos usa UTC do servidor. O worker consulta os planos a cada 30 segundos; a primeira ocorrência financeira fica programada para três minutos depois da criação do caso.
- Se o serviço reiniciar depois do horário programado, o evento é aplicado com o horário planejado. O evento não aparece antes de a transação de ocorrência e evidências ser confirmada.
- Somente casos `open` aceitam novos eventos. Encerrar ou resolver um caso impede anexos posteriores; planos pendentes são cancelados pelo worker.
- O agente de um plano precisa estar `alive` no horário planejado. Morte, desaparecimento ou prisão anteriores cancelam a ação; mudanças posteriores não reescrevem eventos já ocorridos.
- Cada plano tem `event_type` e `plan_version`. O primeiro recorte programa uma adulteração ligada ao crime financeiro primário; o mesmo modelo causal suporta cadeias adicionais sem materializar crimes futuros no endpoint público.

## Fora de escopo

- Criar toda a lore pessoal, campanha de serial killer, pedidos narrativos específicos ou elenco fixo completo; estes podem consumir o modelo depois.
- Implementar o mercado clandestino, níveis, fama hacker, recompensas em Credits ou novas ferramentas de hacking.
- Alterar regras de preço e flutuação da bolsa definidas pela M26.
- Adicionar chat, mensagens instantâneas ou chamadas telefônicas. Conversas roteirizadas de NPCs fixos são a exceção narrativa prevista acima.
- Simular todos os aspectos da vida de cada NPC continuamente; a simulação deve se limitar às ações/eventos necessários para evolução verificável dos casos.
