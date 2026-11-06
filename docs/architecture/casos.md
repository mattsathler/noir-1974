# Casos procedurais

Um caso não armazena uma cópia do mundo. Ele pertence ao servidor e referencia as entidades e relações já existentes. O progresso de investigação pertence individualmente a cada conta.

## Gerador modular de homicídios

Todo caso gerado continua sendo um homicídio (`case_type`). O `archetype_id` descreve a trama e a motivação, sem restringir todos os casos a conflitos de trabalho. A versão 6 mantém a hora interna do crime e os canais investigativos da v5, escolhe para ciúmes um edifício não corporativo no distrito do casal e dá a cada arquétipo uma janela noturna coerente com seu contexto, calculada no fuso da cidade. Conflitos de trabalho e extorsão continuam na sede da empresa. A chegada GATE do culpado ao bairro da cena antes do crime é garantida; em parte dos casos também há saída posterior. As versões 5 e 4 permanecem reconstruíveis, assim como a versão 3 com sua seleção e pistas anteriores.

Os arquétipos atuais são `workplace_conflict`, `jealousy` e `financial_extortion`. O primeiro usa vínculos de emprego e a sede da empresa; o segundo parte de um casamento, um colega que serve de possível amante e correspondência e passagens GATE que conectam os três; o terceiro usa uma relação profissional para estruturar uma extorsão. Em casos v5, os registros GATE do culpado provam presença no bairro da cena no horário do crime: chegada anterior é garantida e a saída posterior pode aparecer também. Extorsões geram o mesmo valor e referência de conta anônima como débito da vítima e crédito do culpado, além de transferências alheias para criar comparações. Mensagens de arrependimento ou sobre o relacionamento também conectam os NPCs; apenas uma fração pequena das mensagens oculta o remetente de quem as recebe, enquanto a caixa de enviados mantém sua identidade. A seleção escolhe um arquétipo disponível e depois um cenário dentro dele, de forma determinística por `case_seed` e versão, para que um arquétipo com mais combinações não domine os demais. Registros e mensagens de caso guardam a classificação interna separada da apresentação ao jogador.

### Casos financeiros e participantes (gerador v7)

A versão 7 mantém os arquétipos de homicídio e acrescenta os tipos `embezzlement`, `financial_extortion` e `accounting_fraud`. Cada novo caso é gerado a partir de uma empresa e de duas relações de emprego existentes: um participante principal e um cúmplice. O órgão nacional de receita aparece como destinatário narrativo nas declarações contábeis, sem criar NPC vítima ou entidade de mundo fictícia. Homicídios continuam exigindo uma vítima humana; os casos financeiros persistem `victim_entity_id` e a localização/cena como nulos.

O dossiê público de Crime Reports identifica a empresa e disponibiliza documentos narrativos sem incluir papéis, responsáveis ou vínculos internos. Os registros bancários específicos são associados aos funcionários e misturados às consultas de rotina. `case_participants` guarda `principal`, `accomplice` ou `victim`; os vínculos internos `supportsNpcIds` dos documentos indicam quais participantes cada pista pode sustentar, mas não entram na resposta pública. Homicídios persistidos antes da v7 recebem os participantes principal e vítima por migração; as versões v3–v6 continuam sendo aceitas.

O validador comum verifica seed/versão/ID, vítimas e culpados existentes e distintos, local existente, relações válidas do mundo, compatibilidade das relações com o arquétipo e evidências alcançáveis. As regras específicas de cada arquétipo ficam isoladas em seus geradores e validadores, para que novas motivações não ampliem as condições do arquétipo de trabalho.

### Assassino em série (gerador v8)

`serial_killer` é uma nova versão de caso; seeds v3–v7 continuam usando as opções e a seleção que já tinham. O gerador exige um assassino empregado, vítimas na mesma cidade e vínculos existentes que permitam ao jogador escrever para ao menos um colega ou pessoa próxima da vítima. A primeira vítima abre o caso; um alvo futuro por vez fica apenas no plano interno. O worker emite ameaças anônimas e, dois dias depois, o homicídio; depois do ataque, agenda o próximo alvo enquanto o caso continua aberto e o assassino segue vivo. O worker registra cada etapa quando ocorre, vincula o evento ao anterior e só então publica timeline, mensagens e prova.

As ameaças são frequentes e escalonam. O nome exato do empregador aparece em apenas uma das cinco mensagens iniciais; as demais dão pistas indiretas e repetem a aparência física. Nas etapas posteriores, as ameaças não repetem o nome da empresa. Colegas e companheiros da vítima podem confirmar por email os detalhes sem entregar novamente o empregador. Assim, cruzar aquela única referência com a descrição e o diretório de staff é a principal dedução. Os IDs dos alvos futuros não fazem parte da API pública antes dos eventos.

### Expansão de crimes (gerador v9)

A v9 acrescenta 15 tipos de investigação sem homicídio: `insurance_fraud`, `investment_fraud`, `money_laundering`, `financial_identity_theft`, `procurement_fraud`, `forged_will`, `cargo_theft`, `art_forgery`, `antiquities_smuggling`, `correspondence_blackmail`, `industrial_sabotage`, `arson`, `staged_disappearance`, `confidential_document_theft` e `police_corruption`. Com `embezzlement`, `financial_extortion` e `accounting_fraud`, são 18 modelos de crime patrimonial ou corporativo. Cada cenário exige uma empresa com sede e dois funcionários distintos; documentos do caso e registros financeiros relacionam o principal e o cúmplice. Os tipos novos não reutilizam os eventos posteriores de acobertamento que pertencem aos três casos financeiros anteriores.

A v9 também registra dez novas motivações de homicídio: `inheritance`, `revenge`, `silence_witness`, `conceal_fraud`, `power_struggle`, `mistaken_identity`, `blackmail`, `debt`, `protect_reputation` e `conceal_accident`. Elas usam vítimas e culpados com vínculos de emprego na mesma empresa, cena na sede e três evidências documentais vinculadas aos participantes. Com `workplace_conflict` (rivalidade profissional) e `jealousy`, completam as 12 ideias de homicídio da expansão. A API de acusação atribui força probatória aos documentos v9 sem expor seus vínculos internos.

As versões v3–v8 mantêm suas escolhas anteriores: os novos tipos só entram na seleção v9, e a v8 permanece aceita para reconstrução de casos existentes. A seleção continua equilibrada em duas etapas, primeiro pelo modelo disponível e depois pelo cenário determinístico daquele modelo.

### Ritmo narrativo e pistas (gerador v10)

A v10 preserva a seleção e a reconstrução v9, mas muda a linha do tempo e as mensagens dos homicídios. Documentos ligados ao caso recebem datas entre três e onze semanas antes do crime. Transferências financeiras e encontros GATE ligados a extorsão ou ciúme também antecedem o crime por pelo menos três semanas; registros de presença na cena permanecem no horário do homicídio.

As mensagens de caso passam a privilegiar conversas corriqueiras. Cada homicídio gera duas mensagens banais antes do crime e, quando há correspondente disponível, duas notas de colegas enviadas semanas depois, que mencionam a ausência da vítima de passagem. O catálogo traz 200 variações para cada grupo de mensagens cotidianas, pós-evento e arrependimento. Mensagens de arrependimento confessando o assassinato são determinísticas e aparecem em apenas 1 a cada 25 seeds, nunca antes de cinco semanas após o crime. A v10 também apresenta descrições documentais mais indiretas, sem expor no texto público a classificação ou o participante que cada item sustenta. Motivações simples, como silenciar uma investigação, coexistem com tramas mais elaboradas, como chantagem.

As versões v3–v9 continuam aceitas e mantêm suas datas, pistas e comportamento de mensagens; nenhuma reconstrução antiga adota o novo ritmo.

## Persistência e estado

- `cases` guarda a definição compacta: seed, versão, tipo, arquétipo, motivação, empresa afetada e os IDs opcionais de vítima e cena/local. `killer_entity_id` é legado e guarda o autor principal; decisões atuais usam participantes.
- `case_participants` guarda os NPCs e seus papéis internos (`principal`, `accomplice`, `victim`). O tipo de caso valida quais papéis e dados são exigidos.
- `case_relationships` referencia vínculos existentes de diferentes tipos do mundo.
- `case_evidence` guarda a classificação interna, apresentação, vínculos e regras de acesso em JSONB.
- `case_activity_records` relaciona registros bancários e GATE ao caso, NPC, horário e classificação interna; o JSON público do registro não inclui essa classificação.
- `case_mail_messages` relaciona mensagens ao caso, caixa do NPC, outro participante, horário e classificação interna; o JSON público contém apenas os campos de correspondência.
- `case_states` guarda o estado compartilhado do caso (`open`, `solved`, `archived`) e o evento persistido de descoberta do corpo, com horário e origem.
- `player_case_progress` guarda IDs de evidências descobertas por conta, independentemente do estado compartilhado.
- `player_case_accusations` armazena uma decisão por personagem acusado e conta; a prisão altera o estado compartilhado do NPC, enquanto o caso permanece aberto até todos os responsáveis estarem presos.

A geração com a mesma seed, versão e mundo reproduz a mesma definição. Progresso do jogador nunca é derivado da seed. `PostgresStore.Create` verifica referências de entidades e relações persistidas e grava a definição e o estado inicial em uma transação. O schema é migrado na inicialização da API depois de identidade e mundo.

A classificação das pistas e suas relações com a solução são internas ao backend, não expostas como tal ao jogador. A apresentação deve misturar pistas com acontecimentos cotidianos e deixar sua relevância como parte da dedução. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md). A geração parte das relações sociais das empresas, conforme a [Regra 0003](../rules/0003-universo-corporativo.md).

Crime Reports combina inquéritos abertos com vítimas cujo corpo ainda não foi encontrado. Esse estado é derivado no backend a partir do caso e do evento de descoberta; o perfil público comunica “missing” até o evento. Casos em cenas públicas recebem horário determinístico de descoberta automática; uma busca na residência da vítima pode registrar uma descoberta anterior. A atualização do estado e da notícia é idempotente.

Nos geradores v4 e v5, o arquivo inicial contém o briefing e os registros formais do caso; as mensagens e atividades ligadas ficam nas caixas de email, no Union Mercantile Bank e no GATE dos NPCs. Assim, ler o arquivo não substitui comparar as fontes investigativas. Casos v3 preservam a apresentação anterior.

### Interrogatórios

O telefone não recebe nem mostra um identificador de caso. O backend escolhe um único contexto aberto relacionado ao NPC, priorizando participação direta, depois relação pessoal e por último vínculo profissional; em empate, usa o caso mais recente. Para cada pergunta, o pool contém no máximo uma fala geral do arquétipo e duas falas do papel daquela pessoa. O contexto e o recorte de falas são estáveis para a combinação de NPC, caso e pergunta; os nomes dos envolvidos são inseridos na fala no servidor. Sem caso relacionado, são usadas respostas de personalidade, ocupação e rotina. O cliente recebe falas, nunca o arquétipo, papel ou classificação do NPC.

Registros de atividade e mensagens cotidianas são gerados por NPC e intervalo sob demanda. A rotina de uma pessoa depende da seed e da versão do mundo, além da versão do gerador da fonte. O serviço junta esses eventos aos registros associados a casos e ordena a resposta sem priorizar ou destacar os últimos. Para mensagens, veja a [arquitetura do correio eletrônico](correio-eletronico.md).

## Fontes investigativas, whiteboard e acusação

Registros bancários, histórico de viagens GATE, correspondências eletrônicas e objetos encontrados em busca e apreensão são fontes de acontecimentos observáveis. O Crime Generator pode associar alguns desses elementos a um crime — por exemplo, uma mensagem de ciúme ou extorsão — enquanto o sistema também apresenta acontecimentos cotidianos. Um mandado contra o culpado pode encontrar uma arma potencial (revólver, bastão, marreta ou faca), mas isso é opcional e o item não vem com atribuição forense automática. Quando aparece num board, o item conta como pista corroborativa. O jogador decide o que guardar; nenhuma fonte ou interação confirma que um elemento é prova.

Cada jogador mantém whiteboards privados no backend, com IDs dos elementos que escolheu guardar. O cliente pode manter cache local para a experiência, mas o estado persistido pelo servidor é a autoridade. Ao acusar, o cliente entrega um board com um ou mais NPCs marcados como acusados, uma vítima quando o board identifica um homicídio, e elementos marcados como prova. O backend associa os elementos ao caso e avalia cada pessoa separadamente. Para crimes financeiros, a vítima é opcional; o papel de autor ou cúmplice e o vínculo individual das evidências vêm apenas do backend.

Um NPC inocente nunca é preso. Uma acusação contra inocente e uma acusação contra o culpado com elementos insuficientes têm o mesmo resultado público. Se aceita, a prisão altera o estado compartilhado do caso e o jogador recebe o resultado por email; o jornal da cidade também é atualizado. Na v7, cada acusado selecionado é avaliado e registrado em sua própria linha. Uma pessoa aceita não é presa novamente em retries; o caso só muda para `solved` quando todos os participantes culpados estiverem presos. A pontuação é interna: registros GATE de presença na cena e transferências de extorsão verificadas valem 1; registros de encontros extraconjugais, emails de arrependimento e mensagens anônimas raras valem 0,75; apoio contextual vale 0,75 e contexto sem valor probatório vale 0. Em crimes financeiros, pistas e transferências contam apenas para os participantes associados internamente. Marcar material que não sustenta a acusação reduz a chance. Para casos v3, o peso anterior de 0,25 das mensagens anônimas e os registros de extorsão antigos são preservados. A implementação original está sequenciada em [M07–M13](../roadmap/README.md), e a expansão financeira/multirresponsável pertence à [M27](../roadmap/achieved/M27-casos-financeiros-e-multiplos-acusados.md).

Os primeiros arquétipos estabelecem o contrato. Motivações adicionais (como roubo seguido de homicídio), cadeias de pistas mais longas e pistas falsas sofisticadas permanecem em milestones posteriores. Pedidos assíncronos de mandados de busca e acusações também ficam para uma milestone futura.
