# Pontas soltas

Não tratar estes itens como decididos. Cada fechamento vira um ADR em `docs/decisions/`.

## Infra e dados

- Banco definitivo (direção atual: relacional, tendência PostgreSQL).
- Schema e índices concretos.
- Cache: o que cachear.
- Materialização: quando a entidade deixa de ser só reconstruível e ganha registro próprio.
- Modelo de relações: tabela genérica, tabelas por tipo, ou híbrido.

## Geração

- Quais atributos do NPC são seed, persistidos, materializados no primeiro contato, ou mutáveis.
- Algoritmos de famílias, casamentos, parentesco, amizades, empresas, cargos, vizinhança, histórico.
- Geração de casos: tipos de crime, suspeitos, escolha do culpado, motivos, evidências, pistas falsas, solucionabilidade, contradições.
- Validação procedural (killer ≠ vítima, entidades existem, pistas alcançáveis).
- Escala da cidade (prédios, empresas, NPCs).

## Tempo e ciclo de vida

- Mutations como estado final vs histórico de eventos.
- Exclusão/arquivamento de casos (`open → solved → archived`).
- Como crimes resolvidos afetam emprego, relações, reputação, empresas, novos casos.
- Simulação temporal (emprego, casamento, natalidade, empresas, casos automáticos).

## Segurança

- Autenticação; autorização por `world_id`; IDs não enumeráveis à vontade; rate limit; validação de mutations; anti-cheat no cliente.

## Stack ainda aberta

- Linguagem do backend.
- Ferramenta de monorepo (workspaces npm, etc.) — só fechar quando o primeiro código de app existir.
