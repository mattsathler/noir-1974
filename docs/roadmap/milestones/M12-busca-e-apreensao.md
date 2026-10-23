# M12 — Interações investigativas em locais

**Ciclo:** 12  
**Status:** Aceito
**Depende de:** M11 aceito

## Objetivo

Transformar os espaços de trabalho das empresas e os serviços públicos em pontos de investigação. O jogador seleciona uma empresa diretamente no Neighborhood; não há etapa intermediária de visita ao prédio. O computador e as interações pertencem à empresa escolhida, enquanto bancos, Delegacia e Central da GATE são destinos de serviço próprios em Central. Para interrogar um NPC, o jogador invade o computador da empresa e consulta o Staff Directory. Cada NPC tem um telefone gerado proceduralmente no cliente, sem persistência no backend. O telefone é uma janela independente; a presença física e o horário de funcionamento da empresa não condicionam a ligação. As respostas variam por pergunta, relação com a vítima e arquétipo dos casos abertos na cidade, com conteúdo organizado em arquivos JSON e sem confirmar diretamente o culpado. Emails anônimos podem trazer ameaças e extorsões anteriores à morte; registros bancários e GATE oferecem dados mais verificáveis. O jornal omite o local da morte, que pode ser descoberto na ficha da vítima pelo Person Search. Na Delegacia, o jogador solicita uma busca residencial pelo nome do NPC; a execução não depende de sua localização ou estado vital, e o endereço residencial não é apresentado. O relatório chega ao email pessoal do computador do jogador, com anexos individuais que podem ser fixados em qualquer whiteboard. Inventários determinísticos misturam objetos cotidianos e achados contextuais sem sinalizar quais têm relação com o crime. A seleção do bairro, endereço e empresa continua transitória; cada ação envia o contexto necessário e o backend valida cidade, alvo e permissões. Respostas do interrogatório são texto contextual e não viram elementos do whiteboard.

## Critérios de aceite

### Concluído

- [x] O telefone é uma janela independente do computador, com visor de entrada colável, keypad e controles de chamada; os números do Staff Directory são gerados proceduralmente por NPC e não são persistidos no backend.
- [x] O jogador pode ligar para contatos obtidos no Staff Directory sem depender de horário comercial ou da presença física do NPC. O tempo até atender varia; números fora do diretório e NPCs mortos não completam a ligação.
- [x] O interrogatório apresenta as cinco perguntas “Como tem sido sua rotina ultimamente?”, “Com quem você costuma passar o tempo?”, “Aconteceu algo fora do comum por aqui recentemente?”, “Como vão as coisas no trabalho?” e “Há algo que esteja preocupando você?”, com respostas em typewriter e botão de desligar. A altura do modal permanece fixa durante a ligação e a grade decorativa fica no rodapé.
- [x] As respostas são selecionadas por pergunta e refletem relação com a vítima, personalidade gerada, ocupação e rotina de trabalho do NPC; colegas, pessoas ligadas à vítima e o próprio responsável têm variações próprias, incluindo respostas incompletas e recusas.
- [x] O conteúdo do interrogatório está organizado em `apps/api/internal/cases/interrogation/`: um arquivo geral e um arquivo por arquétipo de caso, com ao menos 100 respostas compartilhadas em cada arquétipo e variações adicionais por papel.
- [x] As falas podem sugerir linhas de investigação, mas não são evidências, não são pináveis e não expõem a classificação interna nem confirmam diretamente a culpa de um NPC.
- [x] Os casos geram emails anônimos anteriores à morte, com ameaças ou exigências de pagamento conforme o arquétipo. Falas telefônicas foram calibradas para serem menos assertivas que dados de emails, banco e GATE.
- [x] O jornal noticia a morte sem informar o local. O Person Search mostra o local reportado da cena na ficha da vítima; o arquivo público do caso não revela esse local.
- [x] As listas de pessoas nos arquivos de email, banco e GATE exibem retratos para ajudar a distinguir NPCs homônimos.
- [x] A Delegacia oferece um formulário de busca e apreensão acima do acesso ao computador; o jogador localiza um NPC por nome e pode solicitar a busca mesmo quando o cadastro indica falecimento.
- [x] A API valida a sessão, a Delegacia da cidade persistida e o NPC pertencente à cidade, registra a autorização e executa a busca sem persistir uma localização residencial ou mostrar seu endereço.
- [x] O inventário residencial gera oito achados estáveis por jogador, cidade e NPC a partir de catálogos JSON modulares com mais de cem itens cotidianos e contextuais; vítimas, responsáveis, colegas e associados ligados a casos recebem um achado contextual variável por papel e arquétipo, com detalhes de registros do próprio caso quando disponíveis.
- [x] O email pessoal está disponível no computador do jogador. O relatório da busca é persistido na caixa de entrada e contém todos os achados, misturando itens cotidianos e contextuais.
- [x] Cada anexo é um card no estilo visual dos pins e possui ação própria para escolher um whiteboard de destino; os itens são persistidos como descobertas antes de serem fixados.
- [x] A API autentica a sessão e valida que o alvo pertence à cidade persistida e que a solicitação veio da Delegacia daquela cidade. Toda solicitação válida é autorizada; a decisão não consulta relações ocultas entre itens e casos.
- [x] Cada empresa disponível abre um espaço de trabalho; entrar e sair dele não persiste bairro, prédio ou empresa atual. O Hospital permanece fechado e não pode ser aberto nesta milestone.
- [x] O Neighborhood lista empresas ligadas ao endereço exato, sem deduzir ocupantes pelo bairro e sem limitar a cinco resultados quando a consulta for por localização.
- [x] Selecionar uma empresa abre diretamente seu espaço de trabalho; não existe lobby, elevador ou tela intermediária de prédio.
- [x] As opções de interação com a empresa aparecem no desktop do computador, com os aplicativos apresentados no padrão visual do desktop.
- [x] O espaço de trabalho recebe a empresa selecionada no Neighborhood e oferece os funcionários daquela empresa, sem carregar funcionários de outra cidade.
- [x] O computador pessoal oferece a caixa de email privada do jogador para relatórios da investigação; busca de pessoas e arquivo de correspondência de NPCs continuam restritos à Delegacia, registros bancários ao Banco e registros GATE à Central GATE.
- [x] O acesso a sistemas locais abre uma janela de terminal; o jogador digita um comando para estabelecer a sessão restrita e recebe saída textual simulando a negociação com o sistema. Depois do acesso, os dados são solicitados pelos aplicativos disponíveis naquele computador.
- [x] O jogador acessa o computador da empresa pelo fluxo de hacking existente e consulta o Staff Directory para obter o telefone de um funcionário daquela empresa; o diretório não oferece funcionários de outras empresas.
- [x] A Delegacia concede qualquer solicitação válida depois de validar a sessão, a cidade, o alvo e o local; a autorização não depende de evidências descobertas nem da classificação interna dos itens.
- [x] A busca persiste os elementos efetivamente encontrados e disponíveis para o jogador, sem revelar sua classificação interna.
- [x] Sistemas locais exigem o comando de acesso no terminal; o jogador só recebe ou tem negado o acesso após a resposta da API. O prédio selecionado permanece contexto transitório e não é gravado como localização do jogador.
- [x] O backend valida a cidade atual, o endereço e a empresa ou serviço solicitado, decide o resultado e persiste cada tentativa de acesso; o frontend só abre os aplicativos após autorização.
- [x] As rotas de dados de sistemas restritos exigem uma autorização recente da mesma conta, cidade, empresa ou serviço; chamadas diretas sem a sessão de acesso são negadas.
- [x] A caixa pessoal recebe uma notificação quando um novo caso é aberto na cidade do jogador, além dos relatórios de busca e apreensão.
- [x] Interrogatórios, buscas residenciais e dados obtidos por acesso local podem misturar conteúdo cotidiano e relacionado a casos sem expor a classificação interna dos resultados.
- [x] Ações locais enviam o endereço ou serviço pretendido como contexto transitório; o backend valida sessão, cidade persistida, alvo e local. Não depende de visita salva nem de campo de prédio atual no perfil do jogador.

### Itens fechados nesta revisão

- [x] Respostas refletem personalidade gerada, ocupação e rotina individual do NPC, além das relações e do contexto de caso.
- [x] O gerador de achados associa itens contextuais a relações de trabalho e vínculos pessoais com pessoas envolvidas no caso, além dos papéis de vítima e responsável.
- [x] Achados contextuais incluem detalhes verificáveis — inscrições, datas, horários, valores, referências, condição ou origem — e reutilizam dados de registros financeiros/GATE relacionados ao alvo quando disponíveis, para comparação com outros elementos.
- [x] O fluxo e seus acessos respeitam as [Regras 0001, 0002 e 0003](../../rules/README.md), conforme auditoria abaixo.

### Auditoria das regras (2026-10-23)

- **0001 — Direção artística:** telefone, janelas e anexos usam os componentes e tokens visuais do terminal; retratos e efeitos preservam renderização pixelada. Os novos itens são conteúdo textual nos cards existentes, sem introduzir uma linguagem visual paralela.
- **0002 — Pistas sem sinalização:** falas e itens são apresentados como conteúdo comum. Relações, papéis e ligação interna a registros são usados apenas no backend para selecionar conteúdo; respostas e metadados dos anexos não expõem classificação de pista, papel no caso ou confirmação de culpa.
- **0003 — Universo corporativo:** as novas falas e itens citam rotas GATE, Créditos, escalas, registros, balcões, formulários, companhia e rotinas de escritório, mantendo os costumes e suportes materiais coerentes com o cenário.

## Fora de escopo

Arbitragem da acusação. A organização dos elementos em whiteboards é coberta pelo M10; M12 apenas disponibiliza novos elementos pináveis encontrados em buscas. Análise forense especializada, comparação automática de artefatos e julgamento de credibilidade de depoimentos. Limites ou custos contra pedidos repetidos de mandados ficam para uma regra futura. Viagem entre cidades e visita inicial a locais são cobertas pelo M11.
