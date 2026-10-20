# apps/api

API autoritativa de Noir 1974, escrita em Go e conectada ao PostgreSQL do Neon pela variável `DATABASE_URL`.

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

`internal/world` contém o gerador determinístico versionado. A versão 1 produz o recorte geográfico mínimo; a versão 2 acrescenta NPCs, empresas e relações e mantém a versão 1 reconstruível. `PostgresStore` persiste âncoras e relações e reconstrói atributos individuais a partir da seed. O diretório é exposto pelas rotas de busca descritas abaixo. O escopo está em [`docs/architecture/geracao-do-mundo.md`](../../docs/architecture/geracao-do-mundo.md).

O schema inicial de identidade é aplicado de forma idempotente quando a API inicia. `DATABASE_URL` é obrigatório. O código de produção fica em `internal/`, com `cmd/` restrito à composição do processo. Organização, segurança, desempenho e testes seguem as [regras de backend](../../docs/dev-rules/backend.md).

## Casos

`internal/cases` gera homicídios determinísticos por arquétipos (atualmente conflito de trabalho e ciúmes) a partir de relações da versão 2 do mundo. A definição referencia IDs do servidor, e um validador rejeita referências inválidas ou sem evidências alcançáveis. O schema separa definição, estado compartilhado do caso e progresso individual. As rotas autenticadas `GET /api/cases`, `GET /api/cases/{caseID}` e `POST /api/cases/{caseID}/progress` expõem dados públicos do caso e persistem o progresso da conta sem retornar o culpado. As tabelas são migradas na inicialização depois dos schemas de identidade e mundo. Consulte [`docs/architecture/casos.md`](../../docs/architecture/casos.md).

## Diretório do mundo

A API materializa um mundo compartilhado na inicialização. `NOIR_WORLD_SERVER_ID` e `NOIR_WORLD_SEED` podem definir seus identificadores; na ausência deles, são usados valores estáveis de desenvolvimento (`00000000-0000-4000-8000-000000000001` e `noir-1974-shared-city`). Essa seed gera 24 NPCs, 3 empresas e 12 locais com endereços em Lyon. O bootstrap também cria de forma idempotente o caso inicial e registra sua vítima como morta. As rotas autenticadas `GET /api/world/npcs?q=`, `GET /api/world/companies?q=`, `GET /api/world/locations?q=` e `GET /api/world/newspaper` alimentam o terminal e o jornal do jogo. O contrato está em [`docs/architecture/api.md`](../../docs/architecture/api.md).
