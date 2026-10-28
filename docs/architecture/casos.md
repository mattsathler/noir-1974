# Casos procedurais

Um caso não armazena uma cópia do mundo. Ele pertence ao servidor e referencia as entidades e relações já existentes. O progresso de investigação pertence individualmente a cada conta.

## Gerador modular de homicídios

Todo caso gerado continua sendo um homicídio (`case_type`). O `archetype_id` descreve a trama e a motivação, sem restringir todos os casos a conflitos de trabalho. A versão 6 mantém a hora interna do crime e os canais investigativos da v5, escolhe para ciúmes um edifício não corporativo no distrito do casal e dá a cada arquétipo uma janela noturna coerente com seu contexto, calculada no fuso da cidade. Conflitos de trabalho e extorsão continuam na sede da empresa. A chegada GATE do culpado ao bairro da cena antes do crime é garantida; em parte dos casos também há saída posterior. As versões 5 e 4 permanecem reconstruíveis, assim como a versão 3 com sua seleção e pistas anteriores.

Os arquétipos atuais são `workplace_conflict`, `jealousy` e `financial_extortion`. O primeiro usa vínculos de emprego e a sede da empresa; o segundo parte de um casamento, um colega que serve de possível amante e correspondência e passagens GATE que conectam os três; o terceiro usa uma relação profissional para estruturar uma extorsão. Em casos v5, os registros GATE do culpado provam presença no bairro da cena no horário do crime: chegada anterior é garantida e a saída posterior pode aparecer também. Extorsões geram o mesmo valor e referência de conta anônima como débito da vítima e crédito do culpado, além de transferências alheias para criar comparações. Mensagens de arrependimento ou sobre o relacionamento também conectam os NPCs; apenas uma fração pequena das mensagens oculta o remetente de quem as recebe, enquanto a caixa de enviados mantém sua identidade. A seleção escolhe um arquétipo disponível e depois um cenário dentro dele, de forma determinística por `case_seed` e versão, para que um arquétipo com mais combinações não domine os demais. Registros e mensagens de caso guardam a classificação interna separada da apresentação ao jogador.

O validador comum verifica seed/versão/ID, vítimas e culpados existentes e distintos, local existente, relações válidas do mundo, compatibilidade das relações com o arquétipo e evidências alcançáveis. As regras específicas de cada arquétipo ficam isoladas em seus geradores e validadores, para que novas motivações não ampliem as condições do arquétipo de trabalho.

## Persistência e estado

- `cases` guarda a definição compacta: seed, versão, tipo, arquétipo, motivação e IDs de vítima, culpado e local.
- `case_relationships` referencia vínculos existentes de diferentes tipos do mundo.
- `case_evidence` guarda a classificação interna, apresentação, vínculos e regras de acesso em JSONB.
- `case_activity_records` relaciona registros bancários e GATE ao caso, NPC, horário e classificação interna; o JSON público do registro não inclui essa classificação.
- `case_mail_messages` relaciona mensagens ao caso, caixa do NPC, outro participante, horário e classificação interna; o JSON público contém apenas os campos de correspondência.
- `case_states` guarda o estado compartilhado do caso (`open`, `solved`, `archived`) e o evento persistido de descoberta do corpo, com horário e origem.
- `player_case_progress` guarda IDs de evidências descobertas por conta, independentemente do estado compartilhado.

A geração com a mesma seed, versão e mundo reproduz a mesma definição. Progresso do jogador nunca é derivado da seed. `PostgresStore.Create` verifica referências de entidades e relações persistidas e grava a definição e o estado inicial em uma transação. O schema é migrado na inicialização da API depois de identidade e mundo.

A classificação das pistas e suas relações com a solução são internas ao backend, não expostas como tal ao jogador. A apresentação deve misturar pistas com acontecimentos cotidianos e deixar sua relevância como parte da dedução. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md). A geração parte das relações sociais das empresas, conforme a [Regra 0003](../rules/0003-universo-corporativo.md).

Crime Reports combina inquéritos abertos com vítimas cujo corpo ainda não foi encontrado. Esse estado é derivado no backend a partir do caso e do evento de descoberta; o perfil público comunica “missing” até o evento. Casos em cenas públicas recebem horário determinístico de descoberta automática; uma busca na residência da vítima pode registrar uma descoberta anterior. A atualização do estado e da notícia é idempotente.

Nos geradores v4 e v5, o arquivo inicial contém o briefing e os registros formais do caso; as mensagens e atividades ligadas ficam nas caixas de email, no Meridian Bank e no GATE dos NPCs. Assim, ler o arquivo não substitui comparar as fontes investigativas. Casos v3 preservam a apresentação anterior.

### Interrogatórios

O telefone não recebe nem mostra um identificador de caso. O backend escolhe um único contexto aberto relacionado ao NPC, priorizando participação direta, depois relação pessoal e por último vínculo profissional; em empate, usa o caso mais recente. Para cada pergunta, o pool contém no máximo uma fala geral do arquétipo e duas falas do papel daquela pessoa. O contexto e o recorte de falas são estáveis para a combinação de NPC, caso e pergunta; os nomes dos envolvidos são inseridos na fala no servidor. Sem caso relacionado, são usadas respostas de personalidade, ocupação e rotina. O cliente recebe falas, nunca o arquétipo, papel ou classificação do NPC.

Registros de atividade e mensagens cotidianas são gerados por NPC e intervalo sob demanda. A rotina de uma pessoa depende da seed e da versão do mundo, além da versão do gerador da fonte. O serviço junta esses eventos aos registros associados a casos e ordena a resposta sem priorizar ou destacar os últimos. Para mensagens, veja a [arquitetura do correio eletrônico](correio-eletronico.md).

## Fontes investigativas, whiteboard e acusação

Registros bancários, histórico de viagens GATE, correspondências eletrônicas e objetos encontrados em busca e apreensão são fontes de acontecimentos observáveis. O Crime Generator pode associar alguns desses elementos a um crime — por exemplo, uma mensagem de ciúme ou extorsão — enquanto o sistema também apresenta acontecimentos cotidianos. Um mandado contra o culpado pode encontrar uma arma potencial (revólver, bastão, marreta ou faca), mas isso é opcional e o item não vem com atribuição forense automática. Quando aparece num board, o item conta como pista corroborativa. O jogador decide o que guardar; nenhuma fonte ou interação confirma que um elemento é prova.

Cada jogador mantém whiteboards privados no backend, com IDs dos elementos que escolheu guardar. O cliente pode manter cache local para a experiência, mas o estado persistido pelo servidor é a autoridade. Ao acusar, o cliente entrega um board; o backend infere o homicídio pela vítima marcada e valida o acusado, a vítima e as provas sem expor quais IDs eram válidos.

Um NPC inocente nunca é preso. Uma acusação contra inocente e uma acusação contra o culpado com elementos insuficientes têm o mesmo resultado público. Se aceita, a prisão altera o estado compartilhado do caso e o jogador recebe o resultado por email; o jornal da cidade também é atualizado. A pontuação é interna: registros GATE de presença na cena e transferências de extorsão verificadas valem 1; registros de encontros extraconjugais, emails de arrependimento e mensagens anônimas raras valem 0,75; apoio contextual vale 0,75 e contexto sem valor probatório vale 0. Marcar material que não sustenta a acusação reduz a chance. Para casos v3, o peso anterior de 0,25 das mensagens anônimas e os registros de extorsão antigos são preservados. A implementação está sequenciada em [M07–M13](../roadmap/README.md).

Os primeiros arquétipos estabelecem o contrato. Motivações adicionais (como roubo seguido de homicídio), cadeias de pistas mais longas e pistas falsas sofisticadas permanecem em milestones posteriores. Pedidos assíncronos de mandados de busca e acusações também ficam para uma milestone futura.
