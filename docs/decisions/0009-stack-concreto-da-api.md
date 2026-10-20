# ADR 0009 — Stack concreta da API

- **Status:** Aceita
- **Data:** 2026-10-20
- **Supersede:** ADR 0002 apenas quanto à implementação concreta do backend

## Contexto

O ADR 0002 definiu REST e o cliente web, mas deixou a linguagem, o framework e o banco da API em aberto. M06 precisa de um contrato reproduzível para consultas do mundo, casos e progresso.

## Decisão

- Backend em Go, usando `net/http` da biblioteca padrão e rotas com `http.ServeMux`.
- PostgreSQL como banco relacional, acessado com `pgx`/`pgxpool`.
- O ambiente atual usa Neon para hospedar PostgreSQL; a aplicação depende de `DATABASE_URL`, sem acoplar o domínio ao provedor.
- Cliente em React e Vite, consumindo endpoints REST JSON.
- Sessões autenticadas por cookie `HttpOnly`; cada instância da API usa o `NOIR_WORLD_SERVER_ID` configurado e não aceita que o cliente escolha outro servidor.

## Consequências

- A composição do servidor fica em `cmd/server`; regras ficam em `internal/`.
- Toda leitura e escrita de mundo, caso e progresso recebe o ID do servidor da composição da API.
- O cliente não calcula ou altera estado autoritativo.
- Trocar Neon não exige reescrever o domínio, desde que o novo serviço ofereça PostgreSQL compatível.
