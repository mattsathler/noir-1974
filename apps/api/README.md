# apps/api

API autoritativa de Noir 1974, escrita em Go e conectada ao PostgreSQL do Neon pela variável `DATABASE_URL`.

Na raiz do projeto, `npm run dev:api` inicia a API em modo de desenvolvimento. O processo observa `apps/api` e recompila/reinicia a API quando arquivos Go, JSON, SQL, YAML ou TOML mudam. `npm run dev` inicia esse watcher junto com o Vite.

## Identidade

O módulo `internal/identity` mantém contas de jogador, sessões e personagens. A conta guarda usuário, e-mail e hash de senha; cada conta pode ter um personagem, criado no primeiro acesso com nome e uma das seis opções de retrato.

- `POST /api/auth/register` cria a conta e inicia uma sessão.
- `POST /api/auth/login` autentica por e-mail e senha.
- `POST /api/auth/logout` encerra a sessão.
- `GET /api/me` retorna a conta e o personagem da sessão.
- `GET /api/world` retorna a localização atual e o horário do servidor para a sessão autenticada. O horário é UTC em ISO 8601; o cliente o exibe no fuso local do mundo.
- `POST /api/characters` cria o personagem, uma única vez por conta.
- `GET /api/health` verifica a API e a conexão com o banco.

As sessões usam tokens aleatórios em cookie `HttpOnly`, `SameSite=Lax` e `Secure` quando `APP_ENV=production`; somente o hash do token é armazenado. Senhas são protegidas com bcrypt. Confirmação de e-mail e recuperação de senha ficam fora deste primeiro fluxo.

## Geração do mundo

`internal/world` contém o gerador determinístico versionado. A versão 1 produz o recorte geográfico mínimo; a versão 2 acrescenta NPCs, empresas e relações; a versão 3 mantém essa sociedade e usa catálogos de nomes por cidade. Todas as versões anteriores continuam reconstruíveis. `PostgresStore` persiste âncoras e relações e reconstrói atributos individuais a partir da seed. O diretório é exposto pelas rotas de busca descritas abaixo. O escopo está em [`docs/architecture/geracao-do-mundo.md`](../../docs/architecture/geracao-do-mundo.md).

O schema inicial de identidade é aplicado de forma idempotente quando a API inicia. `DATABASE_URL` é obrigatório. O código de produção fica em `internal/`, com `cmd/` restrito à composição do processo. Organização, segurança, desempenho e testes seguem as [regras de backend](../../docs/dev-rules/backend.md).

## Casos

`internal/cases` gera homicídios determinísticos por arquétipos (conflito de trabalho, ciúmes e extorsão financeira) a partir de relações das versões 2 e 3 do mundo. O gerador de casos está na versão 3; o gerador de registros possui versão própria. Casos de extorsão vinculam uma remessa financeira da vítima e uma passagem GATE do culpado à definição interna do caso. As rotas autenticadas `GET /api/cases`, `GET /api/cases/{caseID}` e `POST /api/cases/{caseID}/progress` continuam omitindo o culpado e classificações internas.

`GET /api/world/npcs/{npcID}/records?from=YYYY-MM-DD&to=YYYY-MM-DD` consulta uma janela inclusiva de até 366 dias. O serviço calcula atividade cotidiana apenas para aquele NPC e intervalo e mistura os registros ligados a casos antes de ordenar tudo por horário e ID. A API não sinaliza quais eventos se relacionam a um crime. Consulte [`docs/architecture/registros-investigativos.md`](../../docs/architecture/registros-investigativos.md) e [`docs/architecture/api.md`](../../docs/architecture/api.md).

## Diretório do mundo

A API materializa um mundo compartilhado na inicialização. `NOIR_WORLD_SERVER_ID` e `NOIR_WORLD_SEED` podem definir seus identificadores; na ausência deles, são usados valores estáveis de desenvolvimento (`00000000-0000-4000-8000-000000000001` e `noir-1974-shared-city`). Essa seed gera 24 NPCs, 3 empresas e 12 locais com endereços em Lyon. Mundos novos usam a versão 3 do gerador e o catálogo de Lyon com 300 prenomes e 300 sobrenomes em `internal/world/names/france/lyon.go`; mundos existentes mantêm a versão gravada e seus nomes determinísticos. O bootstrap também cria de forma idempotente o caso inicial e registra sua vítima como morta. As rotas autenticadas `GET /api/world/npcs?q=`, `GET /api/world/companies?q=`, `GET /api/world/locations?q=` e `GET /api/world/newspaper` alimentam o terminal e o jornal do jogo. O contrato está em [`docs/architecture/api.md`](../../docs/architecture/api.md).
