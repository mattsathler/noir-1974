# Pontas soltas

Não tratar estes itens como decididos. Cada fechamento vira um ADR em `docs/decisions/`.

## Infra e dados

- Estratégia operacional de migrations para aplicar/evoluir o [schema inicial](schema-inicial.sql).
- Cache: o que cachear.
- Materialização: a âncora é criada quando uma identidade precisa ser referenciada; definir limiares e gatilhos concretos com o gerador.
- Particionamento do mundo expansível: o `region_id` canônico e a identidade estável estão definidos; falta definir geografia, vizinhança e limites de expansão.
- Concorrência: como serializar ou resolver ações simultâneas sobre a mesma entidade, mutation ou caso.
- Distribuição dos casos compartilhados: quantidade, disponibilidade, atribuição ou reserva para evitar disputa excessiva.
- Atualização dos clientes conectados quando o mundo ou um caso compartilhado muda; transporte e frequência.

## Geração

- Refinar a lista provisória de atributos do NPC (seed, persistidos, materializados ou mutáveis) durante a implementação do gerador.
- Limites e gatilhos de expansão sob demanda; quais entidades/regiões são geradas juntas para preservar coerência.
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

- Autenticação; autorização por `server_id`; IDs não enumeráveis à vontade; rate limit; validação de mutations; anti-cheat no cliente.

## Stack ainda aberta

- Linguagem do backend.
- Ferramenta de monorepo (workspaces npm, etc.) — só fechar quando o primeiro código de app existir.
