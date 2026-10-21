# API atual

Estes são os endpoints atuais da API. Rotas de diretório, casos e progresso exigem sessão autenticada.

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET /api/me
GET /api/world
POST /api/characters
GET /api/world/npcs?q=<name>
GET /api/world/npcs/:npcID/records?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/npcs/:npcID/emails?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/companies?q=<name-or-address>
GET /api/world/locations?q=<name-or-address>
GET /api/world/transit
POST /api/world/travel
POST /api/world/visit
GET /api/world/newspaper
GET /api/cases
GET /api/cases/:caseID
POST /api/cases/:caseID/progress
```

## Servidor como autoridade

```text
CLIENT → REST API → WORLD / DATABASE → PROCEDURAL GENERATOR → CURRENT STATE
```

O frontend pede “NPC #23”. O servidor verifica o servidor do jogador, recupera seed/estado, gera atributos, consulta relações, aplica mutations e devolve só o que o cliente pode ver. Jogadores do mesmo servidor leem o mesmo estado compartilhado; progresso individual pode ser separado.

## Diretório do mundo

As rotas de busca exigem uma sessão autenticada e consultam o servidor compartilhado configurado pela API. `q` é obrigatório e aceita até 80 caracteres. A busca de pessoas retorna nome, idade, aparência, local de trabalho e estado civil; não retorna cargo nem identidade do parceiro. A busca de empresas retorna sede e funcionários com seus cargos, oferecendo uma fonte separada para descobrir funções dentro da empresa. A busca de locais retorna apenas endereços do bairro atual, incluindo os pontos fixos de serviço quando o jogador está em Central. A API recompõe atributos gerados usando a seed do mundo e lê relações e mutations persistidas.

`GET /api/world/newspaper` retorna a edição determinística do dia fictício atual. A manchete usa o caso persistido mais recente e omite o culpado; seis notas leves são escolhidas de um catálogo por servidor e data. A API cria o caso inicial uma única vez no bootstrap e registra a vítima como morta no estado compartilhado.

`GET /api/world/transit` retorna o bairro Central, seus pontos fixos de serviço e os bairros procedurais. Os nomes de edifícios de outro bairro só são exibidos após a viagem. `POST /api/world/travel` recebe `districtId` e persiste a chegada após validar o destino. `POST /api/world/visit` recebe `locationId` e aceita apenas um ponto do bairro atual; a visita preserva o bairro e registra a localização exata do jogador. `GET /api/world` e `GET /api/me` incluem `locationId`, `districtId`, o nome do bairro e, quando aplicável, `placeName`.

`GET /api/world/npcs/:npcID/records` retorna registros do NPC no intervalo inclusivo informado. `source=bank` seleciona o ledger bancário e exige que o jogador esteja no Meridian Bank da cidade atual; uma requisição remota recebe `403`. `source=gate` seleciona o histórico GATE. Sem filtro, registros bancários são omitidos fora do banco. As datas usam `YYYY-MM-DD` e aceitam no máximo sete dias de calendário inclusivos. Os resultados são ordenados por horário e ID. A consulta gera apenas os registros cotidianos daquela pessoa e intervalo e mistura registros associados a casos. O JSON público não contém `caseId` nem a classificação interna.

No desktop, banco e GATE são aplicativos distintos. Cada um apresenta o acesso pelas credenciais do protagonista como etapa narrativa e consulta apenas sua própria fonte. O endpoint continua exigindo a sessão autenticada do jogador; o passo de credenciais não é um novo mecanismo de autenticação do servidor.

O serviço bancário é operado pelo Meridian Global Settlement Office e a infraestrutura GATE pela Meridian Global. A classificação e a relação dos registros de caso ficam em `case_activity_records`; a API expõe os eventos sem revelar essa relação. A versão do gerador de registros é independente da versão da seed do mundo e ambas participam da derivação determinística.

`GET /api/world/npcs/:npcID/emails` retorna a correspondência daquela caixa no intervalo inclusivo, com limite de sete dias de calendário. Mensagens cotidianas são geradas sob demanda para o NPC consultado e misturadas às mensagens de caso persistidas. A API expõe serviço, remetente, destinatário, assunto, corpo e data; nunca expõe relação com caso ou classificação interna. O app Meridian Mail apresenta o acesso ao arquivo sob a infraestrutura corporativa da Meridian Global. Detalhes em [Correio eletrônico](correio-eletronico.md).

## Casos e progresso

`GET /api/cases` lista os casos abertos do servidor. `GET /api/cases/:caseID` retorna os dados públicos do caso, registros narrativos sem classificações internas e o progresso daquela conta. Respostas públicas não incluem `killerId`, motivo real, tipo interno das evidências, relações usadas para gerá-las ou regras de acesso.

`POST /api/cases/:caseID/progress` recebe `{ "discoveredEvidenceIds": ["…"] }`. O servidor valida que o caso existe e que cada ID de registro pertence ao caso, normaliza a lista e a salva por conta. O progresso persiste entre sessões e não altera o estado compartilhado do caso. O endpoint ainda representa o recorte mínimo de M06; ações de descoberta ficam para milestones posteriores.

Cada instância da API consulta e grava apenas o `NOIR_WORLD_SERVER_ID` configurado na composição. O cliente não pode enviar um `server_id` para trocar de mundo.

## Monorepo

Frontend em `apps/web`, backend em `apps/api`. Contratos compartilhados só quando existirem consumidores reais nos dois lados.
