# Pontas soltas

Não tratar estes itens como decididos. Cada fechamento vira um ADR em `docs/decisions/`.

## Infra e dados

- Estratégia operacional de migrations para aplicar/evoluir o [schema inicial](schema-inicial.sql).
- Cache: o que cachear.
- Materialização: a âncora é criada quando uma identidade precisa ser referenciada; definir limiares e gatilhos concretos com o gerador.
- Implementar o catálogo fixo de 25 países e 100 cidades e viagens internacionais pelo GATE conforme M11; novos destinos poderão ser adicionados em versões posteriores. A versão 5 define a composição de 20 bairros por cidade, inicialmente implementada para Lyon.
- Definir como distribuir detetives e ocorrências entre os 20 bairros lógicos conforme os servidores e cidades forem adicionados, preservando a identidade estável por servidor.
- Concorrência: como serializar ou resolver ações simultâneas sobre a mesma entidade, mutation ou caso.
- Distribuição dos casos compartilhados: quantidade, disponibilidade, atribuição ou reserva para evitar disputa excessiva.
- Atualização dos clientes conectados quando o mundo ou um caso compartilhado muda; transporte e frequência.
- Acusações concorrentes no mesmo caso compartilhado: ordem, idempotência, encerramento após a primeira prisão e comportamento para os demais jogadores.

## Geração

- Refinar a lista provisória de atributos do NPC (seed, persistidos, materializados ou mutáveis) durante a implementação do gerador.
- Limites e gatilhos de expansão sob demanda; quais entidades/regiões são geradas juntas para preservar coerência.
- Definir a política de colisão se um futuro banco de palavras de cidade produzir nomes duplicados ou ambíguos.
- Algoritmos de famílias, casamentos, parentesco, amizades, empresas, cargos, vizinhança, histórico.
- Geração de casos: tipos de crime, suspeitos, escolha do culpado, motivos, evidências, pistas falsas, solucionabilidade, contradições.
- Validação procedural (killer ≠ vítima, entidades existem, pistas alcançáveis).
- Escala da cidade (prédios, empresas, NPCs).
- Granularidade, intervalos consultáveis e política de retenção para registros bancários e visitas ao GATE sem simular integralmente o tempo da cidade.
- Regras de acesso às caixas de correio dos NPCs, limites de consulta, privacidade, retenção e controle corporativo sobre as mensagens.
- Fórmula de pontuação e limiar numérico/versionado para aceitar uma acusação; peso de elementos duplicados, irrelevantes ou contraditórios.

## Tempo e ciclo de vida

- Mutations como estado final vs histórico de eventos.
- Exclusão/arquivamento de casos (`open → solved → archived`).
- Como crimes resolvidos afetam emprego, relações, reputação, empresas, novos casos.
- Simulação temporal (emprego, casamento, natalidade, empresas, casos automáticos).
- Condições de autorização e consequências de busca e apreensão em residências, coerentes com a ordem corporativa e o passado do detetive na Interpol.
- Como limitar pedidos repetidos de mandados residenciais sem consultar relações ocultas entre provas e casos; avaliar custos em créditos, suborno às autoridades ou invasão ao sistema policial.

## Segurança

- Autenticação; autorização por `server_id`; IDs não enumeráveis à vontade; rate limit; validação de mutations; anti-cheat no cliente.

## Stack ainda aberta

- Linguagem do backend.
- Ferramenta de monorepo (workspaces npm, etc.) — só fechar quando o primeiro código de app existir.
