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

O schema inicial de identidade é aplicado de forma idempotente quando a API inicia. `DATABASE_URL` é obrigatório. O código de produção fica em `internal/`, com `cmd/` restrito à composição do processo. Organização, segurança, desempenho e testes seguem as [regras de backend](../../docs/dev-rules/backend.md).
