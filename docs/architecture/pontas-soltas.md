# Pontas soltas

Não tratar estes itens como decididos. Cada fechamento vira um ADR em `docs/decisions/`.

## Infra e dados

- Estratégia operacional de migrations para aplicar/evoluir o [schema inicial](schema-inicial.sql).
- Cache: o que cachear.
- Materialização: a âncora é criada quando uma identidade precisa ser referenciada; definir limiares e gatilhos concretos com o gerador.
- Definir a seleção dos 50 países e o catálogo fixo de capitais e principais cidades; especificar o banco de palavras e as regras de composição dos nomes de bairros.
- Definir critérios de distribuição de detetives e ocorrências entre bairros lógicos, gatilhos de criação sob demanda e limites operacionais, preservando a identidade estável por servidor.
- Definir a cadência padrão da geração contínua de casos, o limite de casos abertos por servidor e o comportamento do agendador quando a API reinicia ou opera em múltiplas instâncias; a implementação está prevista no [M10](../roadmap/milestones/M10-mundo-compartilhado-e-investigacao.md).
- Concorrência: como serializar ou resolver ações simultâneas sobre a mesma entidade, mutation ou caso.
- Distribuição dos casos compartilhados: quantidade, disponibilidade, atribuição ou reserva para evitar disputa excessiva.
- Atualização dos clientes conectados quando o mundo ou um caso compartilhado muda; transporte e frequência.
- Acusações concorrentes no mesmo caso compartilhado: ordem, idempotência, encerramento após a primeira prisão e comportamento para os demais jogadores.

## Geração

- Refinar a lista provisória de atributos do NPC (seed, persistidos, materializados ou mutáveis) durante a implementação do gerador.
- Limites e gatilhos de expansão sob demanda; quais entidades/regiões são geradas juntas para preservar coerência.
- Evitar nomes de bairros duplicados ou ambíguos na mesma cidade e definir a política de colisão do gerador.
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

## Segurança

- Autenticação; autorização por `server_id`; IDs não enumeráveis à vontade; rate limit; validação de mutations; anti-cheat no cliente.

## Stack ainda aberta

- Linguagem do backend.
- Ferramenta de monorepo (workspaces npm, etc.) — só fechar quando o primeiro código de app existir.
