# ADR 0008 — Relações tipadas por domínio

- **Status:** Aceita
- **Data:** 2026-10-20
- **Supersede:** nenhum

## Contexto

O mundo precisa consultar relações como emprego e vínculos pessoais nos dois sentidos, sem gerar toda a população. Uma única tabela genérica com tipos e endpoints polimórficos facilita adicionar arestas, mas não consegue impor chaves estrangeiras convencionais para NPCs, empresas e locais e deixa o significado de cada relação implícito.

## Decisão

Usar uma tabela por tipo de relação (`employment_relationships`, `personal_relationships`, `company_locations`) com campos explícitos para cada extremo e índices para as direções de consulta necessárias. Todas as tabelas carregam `server_id` e referenciam a âncora `world_entities(server_id, entity_id)`, impedindo relações entre servidores.

O tipo de cada extremo (por exemplo, funcionário é NPC e empregador é empresa) também é validado pela camada de domínio. A tabela de identidade é comum; os fatos de relação não são armazenados numa tabela genérica.

## Consequências

- Consultas e índices refletem o domínio e seus caminhos conhecidos.
- Chaves estrangeiras garantem existência e isolamento de servidor; validação de tipo fica no domínio.
- Um novo tipo de relação exige schema e índice próprios quando necessário, em vez de apenas inserir um novo valor genérico.
- A forma relacional inicial está documentada em [schema-inicial.sql](../architecture/schema-inicial.sql); estratégia de migrations e materialização operacional ficam para os ciclos de implementação.
