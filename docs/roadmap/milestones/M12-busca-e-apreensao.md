# M12 — Interações investigativas em locais

**Ciclo:** 12  
**Status:** Em curso
**Depende de:** M11 aceito

## Objetivo

Transformar os espaços de trabalho das empresas e os serviços públicos em pontos de investigação. O jogador seleciona uma empresa diretamente no Neighborhood; não há etapa intermediária de visita ao prédio. O computador e as interações pertencem à empresa escolhida, enquanto bancos, Delegacia e Central da GATE são destinos de serviço próprios em Central. Para interrogar um NPC, o jogador invade o computador da empresa e consulta o Staff Directory. Cada NPC tem um telefone gerado proceduralmente no cliente, sem persistência no backend. O telefone é uma janela independente; a presença física e o horário de funcionamento da empresa não condicionam a ligação. As respostas variam por pergunta, relação com a vítima e arquétipo dos casos abertos na cidade, com conteúdo organizado em arquivos JSON e sem confirmar diretamente o culpado. Emails anônimos podem trazer ameaças e extorsões anteriores à morte; registros bancários e GATE oferecem dados mais verificáveis. O jornal omite o local da morte, que pode ser descoberto na ficha da vítima pelo Person Search. Também poderá solicitar na Delegacia uma autorização de busca e apreensão, investigar residências e burlar controles de acesso a sistemas restritos. A seleção do bairro, endereço e empresa continua transitória; cada ação envia o contexto necessário e o backend valida cidade, alvo e permissões. Buscas podem revelar artefatos e vestígios pináveis; respostas do interrogatório são texto contextual e não viram elementos do whiteboard. Achados, falas e resultados misturam fatos cotidianos e elementos ligados a crimes sem sinalizar ao jogador o que é pista.

## Critérios de aceite

### Concluído

- [x] O telefone é uma janela independente do computador, com visor de entrada colável, keypad e controles de chamada; os números do Staff Directory são gerados proceduralmente por NPC e não são persistidos no backend.
- [x] O jogador pode ligar para contatos obtidos no Staff Directory sem depender de horário comercial ou da presença física do NPC. O tempo até atender varia; números fora do diretório e NPCs mortos não completam a ligação.
- [x] O interrogatório apresenta as cinco perguntas “Como tem sido sua rotina ultimamente?”, “Com quem você costuma passar o tempo?”, “Aconteceu algo fora do comum por aqui recentemente?”, “Como vão as coisas no trabalho?” e “Há algo que esteja preocupando você?”, com respostas em typewriter e botão de desligar. A altura do modal permanece fixa durante a ligação e a grade decorativa fica no rodapé.
- [x] As respostas são selecionadas por pergunta e por relação do NPC com a vítima. NPCs sem relação direta recebem respostas gerais; colegas, pessoas ligadas à vítima e o próprio responsável têm variações próprias, incluindo respostas incompletas e recusas.
- [x] O conteúdo do interrogatório está organizado em `apps/api/internal/cases/interrogation/`: um arquivo geral e um arquivo por arquétipo de caso, com ao menos 100 respostas compartilhadas em cada arquétipo e variações adicionais por papel.
- [x] As falas podem sugerir linhas de investigação, mas não são evidências, não são pináveis e não expõem a classificação interna nem confirmam diretamente a culpa de um NPC.
- [x] Os casos geram emails anônimos anteriores à morte, com ameaças ou exigências de pagamento conforme o arquétipo. Falas telefônicas foram calibradas para serem menos assertivas que dados de emails, banco e GATE.
- [x] O jornal noticia a morte sem informar o local. O Person Search mostra o local reportado da cena na ficha da vítima; o arquivo público do caso não revela esse local.
- [x] As listas de pessoas nos arquivos de email, banco e GATE exibem retratos para ajudar a distinguir NPCs homônimos.

### Em aberto

- [ ] API valida o alvo, o local, a sessão do jogador e as condições de acesso antes de executar uma busca.
- [ ] Cada empresa disponível possui um espaço de trabalho compatível com a direção visual do jogo; entrar e sair dele não persiste bairro, prédio ou empresa atual. O Hospital permanece fechado e não pode ser aberto nesta milestone.
- [ ] O Neighborhood lista empresas ligadas ao endereço exato, sem deduzir ocupantes pelo bairro e sem limitar a cinco resultados quando a consulta for por localização.
- [ ] Selecionar uma empresa abre diretamente seu espaço de trabalho; não existe lobby, elevador ou tela intermediária de prédio.
- [ ] As opções de interação com a empresa aparecem no centro da tela, com os aplicativos apresentados no padrão visual do desktop do computador.
- [ ] O espaço de trabalho recebe a empresa selecionada no Neighborhood e oferece os NPCs, serviços ou sistemas pertinentes àquela empresa, sem carregar conteúdo de outra cidade.
- [ ] O computador pessoal não oferece banco, registros GATE, emails ou busca de pessoas; registros bancários só podem ser consultados no Banco, registros GATE na Central GATE e emails/busca de pessoas na Delegacia da cidade atual.
- [ ] O acesso a sistemas locais abre uma janela de terminal; o jogador digita comandos para estabelecer a sessão restrita e solicitar dados, com saída textual simulando a negociação e a resposta do sistema.
- [ ] Ações enviam o local pretendido como contexto transitório; o backend valida sessão, cidade persistida, cidade do local e permissões. Não depende de uma visita salva nem de um campo de prédio atual no perfil do jogador.
- [x] O jogador acessa o computador da empresa pelo fluxo de hacking existente e consulta o Staff Directory para obter o telefone de um funcionário daquela empresa; o diretório não oferece funcionários de outras empresas.
- [ ] Respostas futuras podem refletir com maior granularidade a rotina, personalidade e ocupação individual do NPC, além das relações e do contexto de caso já cobertos.
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
