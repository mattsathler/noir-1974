# API atual

Estes são os endpoints atuais da API. Rotas de diretório, casos e progresso exigem sessão autenticada.

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET /api/me
GET /api/character-visuals
GET /api/world
POST /api/characters
GET /api/world/npcs?q=<name>
GET /api/world/npcs/:npcID/records?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/npcs/:npcID/emails?from=YYYY-MM-DD&to=YYYY-MM-DD
GET /api/world/companies?q=<name-or-address>
GET /api/world/locations?q=<name-or-address>
GET /api/world/transit
GET /api/world/gate/destinations
POST /api/world/gate/travel
GET /api/world/newspaper
GET /api/crime-reports
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

`GET /api/character-visuals` retorna os IDs aceitos nas categorias `accessories`, `beard`, `eyes`, `face`, `hair` e `skin`. `POST /api/characters` recebe `{ "displayName": "Ada", "visual": { "hairId": "hair_short_brown", "skinId": "skin_white", "eyesId": "eyes_green", "faceId": "face_scar_short", "beardId": "beard_moustache_brown", "accessoryIds": ["accessories_glasses_round_tortoise"] } }`. Face, barba e um acessório são opcionais. A API valida cada ID e salva a aparência na conta; `GET /api/me` retorna os mesmos campos. Identidades persistidas no formato anterior são adaptadas ao serem lidas. NPCs das versões 4 e 5 retornam IDs em `npc.visual`; os da versão 5 também usam a geração visual calibrada por idade. O frontend compõe as imagens PNG; o backend não envia arquivos de arte.`

## Diretório do mundo

As rotas de busca exigem uma sessão autenticada e consultam o servidor compartilhado configurado pela API. `q` é obrigatório e aceita até 80 caracteres. A busca de pessoas retorna nome, idade, aparência, IDs visuais determinísticos, local de trabalho e estado civil; não retorna cargo nem identidade do parceiro. A busca de empresas retorna sede e funcionários com seus cargos, oferecendo uma fonte separada para descobrir funções dentro da empresa. A busca de locais retorna apenas endereços do bairro atual, incluindo os pontos fixos de serviço quando o jogador está em Central. A API recompõe atributos gerados usando a seed do mundo e lê relações e mutations persistidas.

`GET /api/world/newspaper` retorna a edição da cidade atual e publica apenas corpos já encontrados. Notícias de descoberta usam o horário e local da descoberta, separados do horário do crime. A edição atualiza o cache quando há mudanças de estado; notas locais continuam determinísticas por servidor, cidade e data. `GET /api/crime-reports?from=YYYY-MM-DD&to=YYYY-MM-DD` resolve a cidade pela localização persistida da conta e retorna inquéritos ainda abertos no período e pessoas atualmente desaparecidas, sem expor culpado ou classificações internas. O intervalo usa dias locais da cidade e, se omitido, cobre os últimos 30 dias. Nomes e retratos são reconstruídos do mundo determinístico. O backend agenda a descoberta automática em cenas públicas e registra buscas residenciais que localizam corpos; vítimas ainda não encontradas aparecem internamente como desaparecidas.

As buscas textuais de pessoas, empresas e locais retornam no máximo cinco resultados, em ordem alfabética. A busca de pessoas monta os perfis completos somente para os cinco primeiros resultados; a busca de empresas consulta os funcionários somente para as cinco empresas retornadas. `GET /api/world/companies?locationId=...` lista todas as empresas relacionadas ao endereço exato por `company_locations`, sem incluir os funcionários nem procurar empresas apenas pelo bairro; a interface usa essas relações para listar empresas diretamente no Neighborhood. O computador da empresa recebe seu `companyId` como contexto; o staff directory consulta `GET /api/world/companies?locationId=...&companyId=...` depois do acesso e retorna somente os funcionários daquela empresa. A consulta de registros bancários/GATE também retorna no máximo cinco registros por chamada.

`GET /api/world/gate/destinations` devolve o catálogo versionado de cidades, agrupadas por país na interface. `GET /api/world/transit?cityId=...` lista os bairros da cidade selecionada. O GATE apresenta essas opções em sequência: país, cidade e bairro. `POST /api/world/gate/travel` recebe `cityId` e `districtId`, valida a relação entre cidade e bairro e materializa a cidade de destino se necessário. Após a viagem, a interface lista empresas associadas aos endereços daquele bairro usando `GET /api/world/companies?locationId=...`; prédios não são destinos visitáveis. Somente `cityId` é persistido como localização do jogador. Em Central, todas as cidades apresentam Police Station, Bank, GATE Central e Hospital como serviços próprios; o Hospital é exibido fechado para uma feature futura. Selecionar uma empresa ou serviço abre sua representação no frontend, mantendo o contexto apenas no estado transitório; não há endpoint que persista bairro ou destino visitado. As consultas de empresa por endereço validam a sessão e a cidade persistida conforme a seção de interações investigativas da M12. `GET /api/world` e `GET /api/me` identificam a cidade atual, país e fuso horário; não expõem bairro selecionado ou destino visitado como estado persistente.

`GET /api/world/npcs/:npcID/records` retorna registros do NPC no intervalo inclusivo informado. O alvo precisa pertencer à cidade persistida do jogador. `source=bank` seleciona o ledger bancário e exige `locationId` correspondente ao Meridian Bank da cidade; `source=gate` seleciona o histórico GATE e exige o endereço da Central GATE daquela cidade. O prédio atual não é persistido. O filtro de fonte é aplicado na geração cotidiana e na consulta SQL, que usa índice por servidor, NPC, fonte e data; não carrega nem serializa eventos do outro serviço. Sem filtro, ambos os serviços são incluídos. As datas usam `YYYY-MM-DD` e aceitam no máximo sete dias de calendário inclusivos. Os resultados são ordenados por horário e ID. O JSON público não contém `caseId` nem a classificação interna.

No desktop, banco e GATE são aplicativos distintos. Cada um apresenta o acesso pelas credenciais do protagonista como etapa narrativa e consulta apenas sua própria fonte. O endpoint continua exigindo a sessão autenticada do jogador; o passo de credenciais não é um novo mecanismo de autenticação do servidor.

O serviço bancário é operado pelo Meridian Global Settlement Office e a infraestrutura GATE pela Meridian Global. A classificação e a relação dos registros de caso ficam em `case_activity_records`; a API expõe os eventos sem revelar essa relação. A versão do gerador de registros é independente da versão da seed do mundo e ambas participam da derivação determinística.

### Interações da M12

`GET /api/cases/interrogation/:npcID` exige sessão autenticada e retorna respostas narrativas escolhidas pela pergunta, relação do NPC com a vítima e arquétipos abertos na cidade. As falas não são descobertas, não expõem classificação interna e não confirmam diretamente a culpa.

`POST /api/police/warrants` recebe `{ "targetNpcId": "…", "stationLocationId": "…" }`. A conta vem da sessão; cidade atual, Delegacia e alvo são validados no servidor. Toda solicitação válida é autorizada, sem consultar descobertas nem classificações internas dos itens. A busca pode incluir NPCs falecidos e não persiste nem retorna endereço residencial. O relatório e os anexos são enviados à caixa privada retornada por `GET /api/player/emails`; os anexos também são registrados como descobertas antes de serem pinados em um whiteboard. Buscas ao culpado podem incluir, de forma opcional, um objeto que poderia servir como arma do crime; sua presença não atribui automaticamente o objeto ao homicídio. Limites ou custos contra pedidos repetidos continuam como regra futura.

`POST /api/facilities/access` recebe o endereço local, tipo de sistema e comando digitado no terminal. A sessão e a cidade vêm do estado persistido da conta; a API valida o endereço exato e, para uma empresa, a relação entre empresa e endereço. Cada tentativa recebe um resultado e um registro no backend. O cliente só libera os aplicativos locais após receber `granted`; tentativas negadas ou inválidas não liberam a sessão. As rotas que consultam funcionários, correspondência ou registros financeiros e GATE também recebem `accessAttemptId` e verificam conta, cidade, endereço, serviço e empresa. A autorização vale por 30 minutos e não pode ser reutilizada em outro computador local.

As consultas de empresa usam `GET /api/world/companies?locationId=...` para listar todas as empresas ligadas ao endereço exato, sem limite de cinco, e `GET /api/world/companies?locationId=...&companyId=...&accessAttemptId=...` após acesso autorizado para obter apenas os funcionários da empresa selecionada. O backend exige sessão e valida que o endereço, empresa e eventual `cityId` correspondem à cidade persistida do jogador. O contexto de bairro, endereço e empresa permanece nas requisições; não é persistido como localização do jogador. Ao criar um caso novo, a API envia uma notificação não investigativa às caixas pessoais de jogadores localizados naquela cidade.

`GET /api/world/npcs/:npcID/emails` exige o endereço da Delegacia da cidade atual e retorna a correspondência daquela caixa no intervalo inclusivo, com limite de sete dias de calendário. O alvo precisa pertencer à cidade persistida. Mensagens cotidianas são geradas sob demanda para o NPC consultado e misturadas às mensagens de caso persistidas. A API expõe serviço, remetente, destinatário, assunto, corpo e data; nunca expõe relação com caso ou classificação interna. As notificações privadas da análise de caso são apresentadas como enviadas por Bill Bouer. O app Meridian Mail apresenta o acesso ao arquivo sob a infraestrutura corporativa da Meridian Global. Detalhes em [Correio eletrônico](correio-eletronico.md).

## Casos e progresso

`GET /api/cases` lista os casos abertos da cidade atual do jogador. A geração recorrente distribui novos casos entre as cidades materializadas e usa apenas NPCs, empresas e locais da cidade selecionada; cada caso persiste seu `city_id`. `GET /api/cases/:caseID` retorna os dados públicos do caso, registros narrativos sem classificações internas e o progresso daquela conta. Respostas públicas não incluem `killerId`, motivo real, tipo interno das evidências, relações usadas para gerá-las ou regras de acesso.

`POST /api/cases/:caseID/progress` recebe `{ "discoveredEvidenceIds": ["…"] }`. O servidor valida que o caso existe e que cada ID de registro pertence ao caso, normaliza a lista e a salva por conta. O progresso persiste entre sessões e não altera o estado compartilhado do caso. O endpoint ainda representa o recorte mínimo de M06; ações de descoberta ficam para milestones posteriores.

Cada instância da API consulta e grava apenas o `NOIR_WORLD_SERVER_ID` configurado na composição. O cliente não pode enviar um `server_id` para trocar de mundo.

## Monorepo

Frontend em `apps/web`, backend em `apps/api`. Contratos compartilhados só quando existirem consumidores reais nos dois lados.
