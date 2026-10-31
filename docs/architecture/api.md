# API atual

Estes são os endpoints atuais da API. Rotas de diretório, casos e progresso exigem sessão autenticada.

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET /api/me
GET /api/character-visuals
GET /api/world
GET /api/world/proxy/nodes
POST /api/world/proxy/connect
POST /api/characters
GET /api/world/npcs?q=<name>
GET /api/world/npcs/:npcID/records?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/npcs/:npcID/emails?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/companies?q=<name-or-address>
GET /api/world/companies/national?q=<name-or-address>
GET /api/market?page=<page>
GET /api/player/stock-portfolio
POST /api/player/stock-orders
GET /api/world/locations?q=<name-or-address>
GET /api/world/transit
GET /api/world/gate/destinations
POST /api/world/gate/travel
GET /api/world/newspaper
GET /api/crime-reports
GET /api/hospital/patients?q=<name>&accessAttemptId=<id>
GET /api/cases
GET /api/cases/:caseID
POST /api/cases/:caseID/progress
GET /api/cases/interrogation/:npcID
POST /api/police/warrants
GET /api/player/emails
POST /api/facilities/access
```

## Servidor como autoridade

```text
CLIENT → REST API → WORLD / DATABASE → PROCEDURAL GENERATOR → CURRENT STATE
```

O frontend pede “NPC #23”. O servidor verifica o servidor do jogador, recupera seed/estado, gera atributos, consulta relações, aplica mutations e devolve só o que o cliente pode ver. Jogadores do mesmo servidor leem o mesmo estado compartilhado; progresso individual pode ser separado.

`GET /api/character-visuals` retorna os IDs aceitos nas categorias `accessories`, `beard`, `eyes`, `face`, `hair` e `skin`. `POST /api/characters` recebe `{ "displayName": "Ada", "visual": { "hairId": "hair_short_brown", "skinId": "skin_white", "eyesId": "eyes_green", "faceId": "face_scar_short", "beardId": "beard_moustache_brown", "accessoryIds": ["accessories_glasses_round_tortoise"] } }`. Face, barba e um acessório são opcionais. `PUT /api/characters/appearance` aceita `{ "visual": { ... } }` e atualiza somente o personagem da conta autenticada; a validação e a gravação são atômicas. `GET /api/me` retorna a aparência persistida. Identidades no formato anterior são adaptadas ao serem lidas. NPCs das versões 4 a 8 retornam IDs em `npc.visual`; a versão 8 acrescenta perfis regionais sintéticos e detalhes faciais. O frontend compõe imagens PNG e SVG; o backend não envia arquivos de arte.

`GET /api/market?page=1` lista 25 empresas nacionais da jurisdição Proxy atual e calcula cotações fixas de 50,00 a 1.000,00 Credits no backend. `GET /api/player/stock-portfolio` retorna fama, saldo em centavos e posições do personagem atual. `POST /api/player/stock-orders` recebe `{ "clientOrderId": "...", "companyId": "...", "side": "buy|sell", "shares": 1 }`; a empresa, país e preço são resolvidos pelo servidor a partir do Proxy ativo. Saldo, ordem e posição são atualizados numa transação, com bloqueio por conta e idempotência por `clientOrderId`. O saldo inicial é 1.000,00 Credits; fama começa em zero e não muda até que sua regra de progressão seja definida. As posições e saldo são isolados pelo servidor e pela conta que possui o personagem.

## Jurisdição, Proxy e buscas

As rotas de busca exigem sessão autenticada e consultam o servidor compartilhado configurado pela API. O estado persistido em `player_locations.location_id` representa a jurisdição Proxy ativa; não representa presença física. `GET /api/world/proxy/nodes` lista cidades disponíveis. `POST /api/world/proxy/connect` recebe `{ "cityId": "fr/paris" }`, valida o ID canônico, prepara os dados da cidade quando necessário e só então atualiza a jurisdição da conta. Uma falha de preparação ou persistência não confirma a troca.

`GET /api/world/npcs?q=...` é Person Search nacional dentro do país da jurisdição ativa, limitado a dez pessoas. Retorna perfis resumidos e informações permitidas; o backend resolve fatos de caso pela cidade do NPC e não expõe classificações internas. A busca é separada dos diretórios de funcionários.

`GET /api/world/companies/national?q=...` pesquisa o National Directory nas cidades do país ativo e retorna no máximo dez registros leves de empresa, endereço, domínio e descrição. Não carrega funcionários nem páginas completas. `GET /api/world/companies/{companyID}/website` resolve a empresa dentro do país Proxy ativo e gera sob demanda um perfil público determinístico com a seed/versão do mundo, versão do perfil de website e ID da empresa. O perfil inclui um dos modelos de composição (`shop-window`, `editorial`, `corporate`, `workshop`, `catalogue`, `directory`, `classifieds`, `bulletin`, `ledger` ou `brochure`), uma família de cabeçalho e a densidade de anúncios (`none`, `sparse`, `standard` ou `crowded`). `GET /api/world/companies?locationId=...` lista empresas associadas ao endereço exato. Acrescentar `companyId` e `accessAttemptId` solicita somente o diretório daquela empresa e exige uma tentativa de acesso válida. As consultas de funcionários permanecem sob demanda; o endpoint de website nunca retorna equipe nem autentica o login diegético.

`GET /api/world/locations?q=...` e `GET /api/world/transit?cityId=...` servem como catálogos de endereços e distritos da cidade. Esses dados identificam instituições e locais, mas a navegação investigativa do jogador não exige visitar bairros ou prédios. Os endpoints GATE de destinos e viagem continuam disponíveis no backend; o desktop da M23 usa Proxy para trocar a jurisdição do detetive. NPCs continuam viajando pelo GATE, e seus registros permanecem fontes investigativas.

`GET /api/world/newspaper` retorna a edição da cidade atual e publica apenas corpos já encontrados. Notícias de descoberta usam o horário e local da descoberta, separados do horário do crime. A edição atualiza o cache quando há mudanças de estado; notas locais continuam determinísticas por servidor, cidade e data. `GET /api/crime-reports?from=YYYY-MM-DD&to=YYYY-MM-DD` resolve a cidade pela localização persistida da conta e retorna inquéritos ainda abertos no período e pessoas atualmente desaparecidas, sem expor culpado ou classificações internas. O intervalo usa dias locais da cidade e, se omitido, cobre os últimos 30 dias. Nomes e retratos são reconstruídos do mundo determinístico. O backend agenda a descoberta automática em cenas públicas e registra buscas residenciais que localizam corpos; vítimas ainda não encontradas aparecem internamente como desaparecidas.

`GET /api/world/npcs/:npcID/records` retorna registros do NPC no intervalo inclusivo informado. O alvo precisa pertencer à cidade Proxy ativa. `source=bank` seleciona o ledger bancário e `source=gate` seleciona o histórico GATE; ambos recebem o endereço canônico do serviço e `accessAttemptId`, que o backend valida antes da consulta. O filtro de fonte é aplicado na geração cotidiana e na consulta SQL, que usa índice por servidor, NPC, fonte e data; não carrega nem serializa eventos do outro serviço. Sem filtro, ambos os serviços são incluídos. As datas usam `YYYY-MM-DD` e aceitam no máximo sete dias de calendário inclusivos. Os resultados são ordenados por horário e ID. O JSON público não contém `caseId` nem a classificação interna.

O Union Mercantile Bank e os GATE Access Logs são sites institucionais. Os registros são páginas desses sites e consultam somente suas próprias fontes. O acesso é validado pela tentativa server-side, não por credenciais digitadas no cliente.

O serviço bancário é operado pelo Union Mercantile Settlement Office e a infraestrutura GATE pela Meridian Global. A classificação e a relação dos registros de caso ficam em `case_activity_records`; a API expõe os eventos sem revelar essa relação. A versão do gerador de registros é independente da versão da seed do mundo e ambas participam da derivação determinística.

`GET /api/hospital/patients?q=...&accessAttemptId=...` consulta pacientes da cidade Proxy ativa. A rota exige uma tentativa concedida para o hospital daquela cidade. Relatórios post-mortem são apresentados como uma página interna do site do hospital; descobertas retornadas pela busca podem ser registradas no whiteboard da conta.

### Interações da M12

`GET /api/cases/interrogation/:npcID` exige sessão autenticada e retorna respostas narrativas escolhidas pela pergunta, relação do NPC com a vítima e arquétipos abertos na cidade. As falas não são descobertas, não expõem classificação interna e não confirmam diretamente a culpa.

`POST /api/police/warrants` recebe `{ "targetNpcId": "…", "stationLocationId": "…" }`. A conta vem da sessão; cidade atual, Delegacia e alvo são validados no servidor. Toda solicitação válida é autorizada, sem consultar descobertas nem classificações internas dos itens. A busca pode incluir NPCs falecidos e não persiste nem retorna endereço residencial. O relatório e os anexos são enviados à caixa privada retornada por `GET /api/player/emails`; os anexos também são registrados como descobertas antes de serem pinados em um whiteboard. Buscas ao culpado podem incluir, de forma opcional, um objeto que poderia servir como arma do crime; sua presença não atribui automaticamente o objeto ao homicídio. Limites ou custos contra pedidos repetidos continuam como regra futura.

`POST /api/facilities/access` recebe `{ "locationId": "...", "companyId": "...", "facility": "company" }`. A sessão e a cidade vêm do estado persistido da conta. Para sites empresariais, o servidor resolve a cidade canônica do endereço e exige igualdade exata com a jurisdição Proxy, além de validar a relação empresa/endereço. Serviços institucionais precisam corresponder ao endereço canônico daquele serviço na cidade conectada. Cada tentativa recebe um resultado e um registro no backend. A interface abre um site restrito somente após receber `granted`; tentativas negadas ou inválidas não liberam dados.

As rotas restritas recebem `accessAttemptId` e validam no servidor o servidor de jogo, a conta, a cidade, o endereço, a empresa, o sistema acessado, o status concedido e a validade temporal. A autorização dura 30 minutos e não pode ser reutilizada em outra conta, cidade, empresa ou sistema.

O portal inicial oferece atalhos para National Hospital, Police Department, Meridian Enterprises e Meridian Email. O National Directory é um site dentro do portal nacional. Meridian Enterprises é o site institucional da corporação e GATE Access Logs é uma subpágina. O jornal continua como app independente, com sua estética editorial. Ao criar um caso novo, a API envia uma notificação não investigativa às caixas pessoais de jogadores cuja jurisdição ativa corresponde àquela cidade.

`GET /api/world/npcs/:npcID/emails` consulta o arquivo policial de correspondência do NPC no intervalo inclusivo, com limite de sete dias de calendário. O alvo precisa pertencer à cidade Proxy ativa, e a rota recebe o endereço da Delegacia e `accessAttemptId` para validar o acesso. Mensagens cotidianas são geradas sob demanda para o NPC consultado e misturadas às mensagens de caso persistidas. A API nunca expõe relação com caso ou classificação interna. As notificações privadas da análise de caso são apresentadas como enviadas por Bill Bouer. Meridian Email é a caixa pessoal do jogador e não substitui o arquivo policial de NPCs. Detalhes em [Correio eletrônico](correio-eletronico.md).

## Casos e progresso

`GET /api/cases` lista os casos abertos da cidade atual do jogador. A geração recorrente distribui novos casos entre as cidades materializadas e usa apenas NPCs, empresas e locais da cidade selecionada; cada caso persiste seu `city_id`. `GET /api/cases/:caseID` retorna os dados públicos do caso, registros narrativos sem classificações internas e o progresso daquela conta. Respostas públicas não incluem `killerId`, motivo real, tipo interno das evidências, relações usadas para gerá-las ou regras de acesso.

`POST /api/cases/:caseID/progress` recebe `{ "discoveredEvidenceIds": ["…"] }`. O servidor valida que o caso existe e que cada ID de registro pertence ao caso, normaliza a lista e a salva por conta. O progresso persiste entre sessões e não altera o estado compartilhado do caso. O endpoint ainda representa o recorte mínimo de M06; ações de descoberta ficam para milestones posteriores.

Cada instância da API consulta e grava apenas o `NOIR_WORLD_SERVER_ID` configurado na composição. O cliente não pode enviar um `server_id` para trocar de mundo.

## Monorepo

Frontend em `apps/web`, backend em `apps/api`. Contratos compartilhados só quando existirem consumidores reais nos dois lados.
