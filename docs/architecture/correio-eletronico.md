# Correio eletrônico

## Geração e consulta

O Meridian Global Correspondence Exchange é a infraestrutura corporativa de mensagens pessoais e comerciais. A aparência de um serviço de correio privado convive com o controle da Meridian sobre a rede, os arquivos e o acesso. O aplicativo Mailroom Archive representa o uso das credenciais antigas do detetive; a API continua exigindo a sessão autenticada normal.

Mensagens cotidianas são geradas sob demanda para uma única caixa de NPC e um intervalo inclusivo. A geração usa seed, versão do gerador, versão do mundo e data; mensagens entre NPCs também usam IDs do par. IDs, remetentes, assuntos, conteúdo e horários permanecem determinísticos, sem materializar correspondência de toda a cidade. Cada conversa pessoal ou entre colegas é reconstruída a partir da mesma chave nos dois participantes; por isso a mensagem aparece como enviada em uma caixa e recebida na outra, com os mesmos campos. Colegas são selecionados entre os funcionários da mesma empresa. Mensagens da empresa são comunicados formais sobre tarefas, horários, prazos e conduta, personalizados pelo cargo, e só aparecem na caixa do funcionário destinatário. Consultas aceitam até sete dias de calendário.

Os padrões editáveis ficam em [`apps/api/internal/mail/templates.json`](../../apps/api/internal/mail/templates.json), separados nas listas `personal`, `workplace` e `colleague`. Cada item tem `subject` e `body`; para acrescentar variações basta incluir novos objetos na lista correspondente. O arquivo é embutido no binário da API.

O Crime Generator também cria mensagens de caso — por exemplo, conversa conjugal marcada por ciúme ou cobrança de extorsão. Elas são persistidas em `case_mail_messages`, relacionadas ao caso, ao destinatário e ao remetente. A consulta inclui a mesma mensagem na caixa de quem a recebeu e na caixa de quem a enviou. Mensagens comuns são geradas na leitura e misturadas cronologicamente às mensagens persistidas.

## Separação da classificação interna

`case_id`, o outro participante por ID e `internal_relation` ficam em colunas privadas da tabela e em campos de backend que não são serializados. O cliente recebe somente ID da mensagem, data, serviço, remetente, destinatário, assunto e corpo. Não há rótulo, filtro, ordenação ou destaque que indique relação com crime; mensagens cotidianas convivem com as mensagens de caso conforme a [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md).

O serviço, o acesso cobrado e as formas antigas de correspondência dentro de uma rede avançada refletem o domínio corporativo e a tensão tecnológica do universo, conforme as [Regras 0001 e 0003](../rules/README.md).

## API

`GET /api/world/npcs/:npcID/emails?from=YYYY-MM-DD&to=YYYY-MM-DD` exige sessão autenticada e consulta apenas a caixa identificada por `npcID` no servidor da sessão. `from` e `to` são inclusivos; intervalos invertidos ou superiores a sete dias de calendário são rejeitados. A resposta é ordenada cronologicamente e não contém IDs de caso nem classificação interna.
