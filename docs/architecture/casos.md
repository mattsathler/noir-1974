# Casos procedurais

Um caso não armazena uma cópia do mundo. Ele pertence ao servidor e referencia as entidades e relações já existentes. O progresso de investigação pertence individualmente a cada conta.

## Gerador modular de homicídios

Todo caso gerado continua sendo um homicídio (`case_type`). O `archetype_id` descreve a trama e a motivação, sem restringir todos os casos a conflitos de trabalho. A versão 2 do gerador reúne estratégias de arquétipo que enumeram participantes e locais a partir do mundo e depois cria relações e pistas próprias da trama.

Os primeiros arquétipos são `workplace_conflict` e `jealousy`. O primeiro usa vínculos de emprego e a sede da empresa; o segundo parte de um vínculo de casamento e de um local do distrito do casal. A seleção escolhe um arquétipo disponível e depois um cenário dentro dele, ambos de forma determinística por `case_seed` e versão, para que um arquétipo com mais combinações não domine os demais. Cada cenário gera três evidências com tipo interno e apresentação/conteúdo narrativo separados. Os vínculos da evidência apontam para fatos existentes no mundo e as regras de acesso indicam de onde ela pode ser descoberta.

O validador comum verifica seed/versão/ID, vítimas e culpados existentes e distintos, local existente, relações válidas do mundo, compatibilidade das relações com o arquétipo e evidências alcançáveis. As regras específicas de cada arquétipo ficam isoladas em seus geradores e validadores, para que novas motivações não ampliem as condições do arquétipo de trabalho.

## Persistência e estado

- `cases` guarda a definição compacta: seed, versão, tipo, arquétipo, motivação e IDs de vítima, culpado e local.
- `case_relationships` referencia vínculos existentes de diferentes tipos do mundo.
- `case_evidence` guarda a classificação interna, apresentação, vínculos e regras de acesso em JSONB.
- `case_states` guarda o estado compartilhado do caso (`open`, `solved`, `archived`).
- `player_case_progress` guarda IDs de evidências descobertas por conta, independentemente do estado compartilhado.

A geração com a mesma seed, versão e mundo reproduz a mesma definição. Progresso do jogador nunca é derivado da seed. `PostgresStore.Create` verifica referências de entidades e relações persistidas e grava a definição e o estado inicial em uma transação. O schema é migrado na inicialização da API depois de identidade e mundo.

A classificação das pistas e suas relações com a solução são internas ao backend, não expostas como tal ao jogador. A apresentação deve misturar pistas com acontecimentos cotidianos e deixar sua relevância como parte da dedução. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md). A geração parte das relações sociais das empresas, conforme a [Regra 0003](../rules/0003-universo-corporativo.md).

## Fontes investigativas, whiteboard e acusação

Registros bancários, histórico de viagens GATE, correspondências eletrônicas e objetos encontrados em busca e apreensão são fontes de acontecimentos observáveis. O Crime Generator pode associar alguns desses elementos a um crime — por exemplo, uma mensagem de ciúme ou extorsão — enquanto o sistema também apresenta acontecimentos cotidianos. O jogador decide o que guardar; nenhuma fonte ou interação confirma que um elemento é prova.

Cada jogador mantém um whiteboard privado por caso, salvo no backend, com IDs dos elementos que escolheu guardar. O cliente pode manter cache local para a experiência, mas o estado persistido pelo servidor é a autoridade. Ao acusar, o cliente envia o NPC e os IDs escolhidos; o backend verifica acesso e calcula a suficiência sem expor quais IDs eram válidos.

Um NPC inocente nunca é preso. Uma acusação contra inocente e uma acusação contra o culpado com elementos insuficientes têm o mesmo resultado público. Se aceita, a prisão altera o estado compartilhado do caso e o jogador recebe feedback de sucesso. Fórmula e limiar de julgamento, autorização de buscas e concorrência entre acusações permanecem em aberto em [pontas soltas](pontas-soltas.md). A implementação está sequenciada em [M07–M11](../roadmap/README.md).

Os primeiros arquétipos estabelecem o contrato. Motivações adicionais (como roubo seguido de homicídio e extorsão para silenciar uma testemunha), cadeias de pistas mais longas, pistas falsas sofisticadas, API pública, UI de investigação e Vesper ainda são trabalho futuro.
