# Registros investigativos

## Geração cotidiana

`internal/records` gera atividade sob demanda para um NPC e um intervalo inclusivo. A saída não depende da ordem das consultas: cada evento é derivado da seed e versão do mundo, versão do gerador de registros, ID do NPC, data e posição do evento naquele dia. A mesma consulta reconstrói os mesmos IDs, horários, descrições e valores.

Cada dia consulta duas fontes independentes: o Union Mercantile Settlement Office e a rede GATE, operada pela Meridian Global. Dias sem atividade aparecem naturalmente; pagamentos rotineiros, remessas pessoais e acessos de viagem são misturados aos eventos vinculados a casos. A consulta não gera a atividade dos outros NPCs nem lê uma coleção global de registros.

## Registros de caso

Em casos gerados na versão 4, `case_activity_records` coloca vítima e culpado no distrito da cena antes do horário do homicídio. Em `financial_extortion`, o mesmo débito anônimo da vítima aparece como crédito na conta do culpado, usando valor e referência iguais; outras transferências entre NPCs aparecem como comparações. Em `jealousy`, passagens sincronizadas da vítima e de um possível amante apoiam a hipótese do caso. As relações internas ao caso e o papel narrativo nunca aparecem no JSON público.

O endpoint de NPC combina a rotina determinística com os eventos persistidos para aquele NPC e intervalo, aplica o filtro de fonte e ordena por data e ID. O jogador consulta o ledger financeiro pelo site Union Mercantile Bank e os deslocamentos pela subpágina GATE Access Logs em Meridian Enterprises. Cada site solicita uma tentativa de acesso ao backend e apresenta apenas a fonte correspondente; não há credenciais digitadas no cliente nem classificação interna na resposta.

Os mesmos rótulos, estilos e ordenação são usados para eventos comuns e vinculados. Não há categoria de “pista”, indicador, destaque, `caseId` ou resposta de relevância. Assim, a relação serve à validação futura no backend sem resolver a dedução para o jogador, conforme a [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md).

## Contrato de consulta

`GET /api/world/npcs/{npcID}/records?from=YYYY-MM-DD&to=YYYY-MM-DD&source=bank|gate` exige sessão autenticada, datas válidas em ordem crescente e intervalo máximo de sete dias de calendário inclusivos. O limite também é aplicado no gerador de registros. O filtro de fonte é opcional e aceita `bank` ou `gate`. A consulta está isolada pelo `NOIR_WORLD_SERVER_ID` da API; IDs e filtros do cliente não permitem trocar de servidor. NPC inexistente retorna 404.

O site Union Mercantile Bank envia o endereço do serviço como contexto transitório com `source=bank`; GATE Access Logs envia o endereço da Central GATE com `source=gate`. A API valida `accessAttemptId`, conta, endereço, serviço, cidade Proxy ativa e se o NPC consultado pertence à cidade. A jurisdição persistida é contexto de conexão, não presença física. Sem filtro, o endpoint legado inclui os dois serviços. A interface usa sempre uma página de serviço com filtro e endereço explícitos.
