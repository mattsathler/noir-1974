# M14 — Linha de base nacional e latências aquecidas

**Coleta:** 24/10/2026, horário local  
**Ambiente:** API local atualizada, conectada ao banco QA  
**Carga:** uma conta temporária de QA; chamadas seriais por conexão HTTP persistente

## Método

- Uma chamada de `/api/world/npcs?q=a` foi feita primeiro para registrar a busca nacional com cache de geração frio.
- Depois foram feitas cinco rodadas de aquecimento e 100 rodadas medidas. As seis rotas foram alternadas em ordem pseudoaleatória fixa em cada rodada.
- Todas as rotas tiveram 100 respostas HTTP 200. A medição não foi concorrente.
- O CSV contém a chamada fria, aquecimentos e amostras medidas: [amostras brutas](M14-latencias-2026-10-24-nacional.csv).
- Percentis usam *nearest rank*. O tempo é medido pelo cliente e inclui processamento HTTP, acesso ao banco e serialização.

## Resultados

| Rota | Amostras 200 | Mín. | p50 | p95 | Média | Máx. |
|---|---:|---:|---:|---:|---:|---:|
| `GET /api/me` | 100 | 309 ms | 347 ms | 548 ms | 424 ms | 2.166 ms |
| `GET /api/world` | 100 | 306 ms | 347 ms | 1.232 ms | 449 ms | 2.411 ms |
| `GET /api/cases` | 100 | 465 ms | 535 ms | 1.325 ms | 614 ms | 2.268 ms |
| `GET /api/whiteboards` | 100 | 308 ms | 347 ms | 1.070 ms | 453 ms | 2.607 ms |
| `GET /api/world/npcs?q=a` | 100 | 1.722 ms | 1.935 ms | 4.435 ms | 2.302 ms | 6.712 ms |
| `GET /api/world/newspaper` | 100 | 461 ms | 533 ms | 1.536 ms | 721 ms | 4.556 ms |

A primeira busca nacional, antes dos aquecimentos, respondeu em **6.306 ms**. É uma observação isolada de cache frio, não um percentil. As medições das 100 amostras ocorreram com o conteúdo gerado aquecido.

Esta é a primeira linha de base medida para a busca nacional. A coleta municipal de 23/10 não é comparável diretamente. Para orientar o próximo ciclo de otimização, a meta inicial da busca aquecida fica em p50 ≤ **1,36 s** e p95 ≤ **3,10 s**, correspondendo a cerca de 30% de redução sobre esta linha de base, na mesma carga serial. A primeira chamada fria deve continuar sendo acompanhada separadamente.

## Queries SQL observadas

O recorder em memória contou 106 requisições da busca nacional no processo: uma chamada fria, cinco aquecimentos e 100 amostras. As principais leituras foram repetidas em todas as chamadas:

| Operação | Chamadas | Duração média por query |
|---|---:|---:|
| `select_world_entities` | 106 | 286 ms |
| `select_personal_relationships` | 106 | 258 ms |
| `select_worlds` | 106 | 213 ms |
| `select_employment_relationships` | 106 | 209 ms |
| `select_world_city_generators` | 106 | 195 ms |
| `select_world_mutations` | 106 | 192 ms |
| `insert_player_discovered_items` | 106 | 313 ms |

Também foram observadas consultas de sessão, localização e cenas do caso. As durações acumuladas das queries do endpoint somaram aproximadamente **2,28 s por chamada**; os tempos se aproximam da média HTTP de 2,30 s. A geração em cache evita reconstruir as cidades nas chamadas aquecidas, mas não elimina essas leituras repetidas do banco.

## Leitura e limites

- A busca nacional aquecida ficou em p50 de 1,94 s e p95 de 4,43 s. A meta inicial de p50 já é inferior ao resultado atual; a cauda longa permanece acima da meta de p95.
- `/api/cases` teve p50 de 535 ms e p95 de 1.325 ms. A p50 ficou próxima da coleta anterior; a p95 subiu, reforçando a variabilidade do ambiente QA.
- Os picos nas rotas `/api/world`, `/api/cases`, `/api/whiteboards` e `/api/world/newspaper` passaram de 1 s. O ambiente serial não permite atribuir esses picos a concorrência.
- O recorder mede duração de queries, não espera por conexão, geração ou serialização separadamente. Os números SQL são médias por operação no período do processo, não latência por amostra HTTP.
- A bateria de linha de base não incluiu a rota de emails nem carga concorrente controlada; essas medições complementares aparecem nas seções seguintes.

## Comparação após reduzir as queries da busca nacional

Após combinar seed/versões em uma consulta e executar em paralelo a leitura de status e das relações, a mesma bateria foi repetida com a mesma conta QA, ordem pseudoaleatória, cinco aquecimentos e 100 amostras por rota. O CSV desta segunda rodada está em [amostras após a otimização](M14-latencias-2026-10-24-pos-otimizacao-busca.csv).

| Rota | Antes p50 / p95 | Depois p50 / p95 |
|---|---:|---:|
| `GET /api/me` | 347 / 548 ms | 349 / 676 ms |
| `GET /api/world` | 347 / 1.232 ms | 348 / 612 ms |
| `GET /api/cases` | 535 / 1.325 ms | 532 / 1.226 ms |
| `GET /api/whiteboards` | 347 / 1.070 ms | 349 / 952 ms |
| `GET /api/world/npcs?q=a` | 1.935 / 4.435 s | 1.784 / 4.234 s |
| `GET /api/world/newspaper` | 533 / 1.536 ms | 531 / 904 ms |

A busca nacional melhorou cerca de **8% na p50** e **5% na p95**; a primeira chamada fria passou de 6.306 ms para 5.607 ms. O recorder mostra nove operações SQL por requisição da busca após a mudança, contra dez antes: `select_world_city_generators` deixou de ser uma query separada. Status, vínculos de trabalho e relações pessoais são consultados em paralelo.

O tempo médio de `select_world_entities`, entretanto, subiu de **286 ms** para **603 ms** entre as duas rodadas. As somas do recorder são médias do processo, e a diferença indica variabilidade do banco QA que limita a atribuição do ganho apenas à mudança de código. Essa consulta agora é o maior alvo visível da busca nacional; avaliar o plano e quantas entidades ela lê antes de outra alteração.

As chamadas continuam seriais, e cada bateria alterna seis rotas; não são um benchmark isolado da busca nem uma medição concorrente. O resultado aponta ganho modesto e não atinge a meta P0 de p50 ≤ 1,36 s / p95 ≤ 3,10 s.

## Busca de emails

Foi medida a rota autenticada `GET /api/world/npcs/{npcID}/emails` com cinco aquecimentos e 100 chamadas para o mesmo NPC de Lyon, mesmo intervalo de sete dias e mesma sessão QA. As 100 chamadas responderam HTTP 200. O [CSV de emails](M14-latencias-2026-10-24-emails.csv) registra duração/status.

| Amostras 200 | Mín. | p50 | p95 | Média | Máx. |
|---:|---:|---:|---:|---:|---:|
| 100 | 2.220 ms | 2.440 ms | 4.328 ms | 2.770 ms | 5.393 ms |

O tracer registrou 105 requisições, incluindo aquecimentos. Cada chamada fez três leituras de `player_locations` (315 no total) e duas de `employment_relationships` (210 no total), além de validar o acesso local e montar a caixa postal. A segunda consulta de emprego é compartilhada para os colegas; falta uma medição anterior equivalente para quantificar o efeito do lote. As três leituras de localização são redundantes dentro desta rota e são o próximo alvo de redução.

Após reutilizar o contexto da cidade para validar a jurisdição e o acesso ao terminal, a rota foi repetida nas mesmas condições. O [CSV após a otimização](M14-latencias-2026-10-24-emails-apos-localizacao.csv) contém 100 respostas HTTP 200.

| Medição | Antes | Depois |
|---|---:|---:|
| p50 | 2.440 ms | 2.116 ms |
| p95 | 4.328 ms | 4.238 ms |
| Média | 2.770 ms | 2.513 ms |
| `select_player_locations` | 3 por requisição | 1 por requisição |

A p50 caiu cerca de **13%**, e a média cerca de **9%**. O p95 caiu 2%; o máximo subiu de 5.393 ms para 7.005 ms, então a cauda ainda precisa de investigação. `select_employment_relationships` permaneceu em duas execuções por chamada; esta bateria usa um NPC com um vínculo de emprego e não isola o ganho do lote de colegas.

### Serviço de emails sem leitura duplicada da configuração

Para medir o trabalho do serviço sem iniciar a API, `mail.Service.NPCEmails` foi executado diretamente para o mesmo NPC empregado de Lyon e o mesmo intervalo de sete dias, com uma chamada fria, cinco aquecimentos e 100 amostras. As duas rodadas usaram o mesmo código de perfil e colegas; a referência emula o fluxo anterior que lia `WorldConfig` antes de `GetNPC`. CSVs: [fluxo anterior](M14-npc-emails-service-base.csv) e [configuração combinada](M14-npc-emails-service-apos-config-combinada.csv).

| Serviço direto | Fluxo anterior | Configuração combinada |
|---|---:|---:|
| p50 | 1.433 ms | 1.229 ms |
| p95 | 1.536 ms | 1.331 ms |
| Média | 1.363 ms | 1.218 ms |
| `select_worlds` por chamada | 1 | 0 |
| `select_employment_relationships` por chamada | 2 | 2 |

A p50 e a p95 caíram aproximadamente **14%** e **13%**. `GetNPCWithWorldConfig` fornece seed e versão global na mesma leitura usada para carregar o NPC; a versão municipal continua sendo usada para gerar sua identidade. NPCs sem perfil persistido continuam usando a leitura de configuração necessária ao fallback do mundo gerado. Depois, status, emprego e relações pessoais foram consultados em paralelo, e a leitura de status foi incorporada à query de identidade do NPC com um `LATERAL`, mantendo a regra de selecionar a mutação mais recente.

Em seguida, empregos e relações pessoais do perfil foram agregados em JSON na query do NPC, e a query de caso passou a ocorrer enquanto o serviço gera a correspondência determinística. No último ciclo, o serviço ficou em p50/p95 de **614/717 ms** em série e **532/2.150 ms** com oito workers. Frente ao fluxo anterior (1.433/1.536 ms), a coleta serial mais recente caiu 57%/53%; as queries caíram de sete para três por chamada. Os CSVs registram a [última rodada serial](M14-npc-emails-service-relacoes-agregadas.csv) e a [última rodada concorrente](M14-npc-emails-service-concorrencia-8-relacoes-agregadas.csv). Uma verificação curta após reorganizar o SQL para manter rótulos de telemetria mostrou as mesmas três queries por chamada e p50 de 614 ms. A comparação mede o serviço, sem autenticação, autorização local nem serialização HTTP; não substitui a bateria da rota completa. Sob oito workers, a p95 do serviço permanece acima de 2 s, alvo para a próxima investigação. O lote de colegas continua em uma query por chamada para este NPC com um empregador, então seu ganho específico permanece sem isolamento.

## Cache das expansões de NPCs

Com a busca nacional isolada, foram feitas uma chamada fria, cinco aquecimentos e 100 amostras medidas. Todas as 100 amostras responderam HTTP 200. O [CSV após o cache de expansões](M14-latencias-2026-10-24-pos-cache-expansoes.csv) contém os tempos por chamada.

| Fase | Amostras 200 | Mín. | p50 | p95 | Média | Máx. |
|---|---:|---:|---:|---:|---:|---:|
| Fria (uma observação) | 1 | — | — | — | 5.723 ms | 5.723 ms |
| Aquecida | 100 | 995 ms | 1.272 ms | 3.555 ms | 1.616 ms | 5.048 ms |

Na bateria anterior, feita enquanto seis rotas eram alternadas, a busca teve p50/p95 de 1,784/4,234 s. A nova coleta foi isolada e não é uma comparação perfeitamente controlada; indica queda de 29% na p50 e 16% na p95, com p50 perto da meta de 1,25 s e p95 ainda acima de 2,96 s. A chamada fria de 5,723 s permanece semelhante à medição fria anterior de 5,607 s.

As métricas da API registraram 106 buscas no total (fria, aquecimentos e amostras) e somente **uma** execução de `select_world_entities`; as 105 buscas seguintes acertaram o cache da lista de expansões. Essa execução durou 349 ms. As demais oito operações recorrentes continuaram em todas as 106 chamadas. A invalidação após `ExpandDistrictSociety` foi validada com uma expansão real e os NPCs novos apareceram na busca, conforme detalhado abaixo.

### Contexto de sessão e localização na busca

Foi feita outra bateria isolada, agora comparando a rota antes e depois de reaproveitar a consulta combinada de sessão/localização usada por `/api/cases`. Os [dados antes](M14-latencias-2026-10-24-busca-base-contexto.csv) e [depois](M14-latencias-2026-10-24-busca-pos-contexto.csv) incluem 100 amostras HTTP 200 cada, cinco aquecimentos e uma chamada fria.

| Medição aquecida | Antes | Depois |
|---|---:|---:|
| p50 | 1.225 ms | 895 ms |
| p95 | 2.931 ms | 2.278 ms |
| Média | 1.416 ms | 1.111 ms |
| Queries por chamada | 8 (após aquecer o cache) | 7 (após aquecer o cache) |
| `select_player_locations` | 1 por chamada | removida |

A p50 caiu **27%** e a p95 **22%**. As chamadas frias foram 5.720 ms e 5.832 ms, respectivamente; o custo de reconstruir o conteúdo determinístico segue praticamente igual. A meta P0 serial de p50 ≤ 1,25 s / p95 ≤ 2,96 s foi atingida nesta bateria.

### Validação operacional da invalidação

Com uma instância de `PostgresStore`, a validação aqueceu o cache da França, expandiu o distrito `fr/lyon/district/beau-atelier` pelo método `ExpandDistrictSociety` e pesquisou os dois NPCs recém-criados pelos nomes gerados. Os dois foram encontrados. A expansão também criou vínculos associados no banco QA; essa alteração é descartável. Isso confirma que a revisão de cache invalida as expansões após o commit e que a busca subsequente recarrega as entidades.

## Listagem de casos e contexto de sessão/localização

`GET /api/cases` foi medido isoladamente com uma chamada fria, cinco aquecimentos e 100 amostras antes e depois da otimização. As duas rodadas usaram a mesma conta de QA e todas as 100 amostras de cada uma responderam HTTP 200. Os CSVs registram a [linha de base](M14-latencias-2026-10-24-cases-base.csv) e a [medição após a mudança](M14-latencias-2026-10-24-cases-pos-contexto.csv).

| Medição aquecida | Antes | Depois |
|---|---:|---:|
| p50 | 550 ms | 350 ms |
| p95 | 1.373 ms | 655 ms |
| Média | 654 ms | 455 ms |
| Queries por chamada | 3 | 2 |

A mudança combina a validação da sessão e a leitura da localização em uma consulta, mantendo a consulta de casos. O tracer confirma duas operações por requisição após a mudança. A p50 caiu 36%, a p95 52% e a média 30%. O tempo frio variou de 1,01 s para 2,31 s, então não é comparável entre rodadas; a API reiniciou entre as medições e a amostra fria também inclui os efeitos da conexão inicial.

Na linha de base, as médias SQL foram 219 ms para sessão, 235 ms para localização e 196 ms para casos. Depois da mudança, a consulta combinada de sessão/localização teve média de 226 ms e a consulta de casos 242 ms. A latência de cada query variou no banco QA; o ganho observado no HTTP é consistente com a remoção de um round-trip, mas esta coleta não separa a variação externa ao processo.

## Referência HTTP sob concorrência controlada

Em 24/10/2026 foi executada a primeira bateria concorrente dos endpoints prioritários após as otimizações. Uma conta e um personagem temporários de QA foram usados nas 200 amostras; cada rota teve cinco aquecimentos e 100 medidas, distribuídas por oito workers com conexões HTTP independentes. O início dos workers foi sincronizado. O cronômetro do cliente inclui autenticação, autorização, consultas, montagem do JSON e transferência da resposta. O [CSV inicial](M14-http-concorrencia-2026-10-24.csv) contém cada status, duração e erro.

| Endpoint | Amostras HTTP 200 | Concorrência | p50 | p95 | Média | Máx. |
|---|---:|---:|---:|---:|---:|---:|
| `GET /api/cases` | 100/100 | 8 | 320 ms | 1.256 ms | 388 ms | 1.602 ms |
| `GET /api/world/npcs?q=a` | 100/100 | 8 | 866 ms | 1.090 ms | 897 ms | 1.714 ms |

O recorder da API contou 105 requisições por rota (cinco aquecimentos e 100 amostras medidas). `/api/cases` executou duas queries por chamada, `select_player_sessions` e `select_cases`. A busca teve uma leitura de `select_world_entities` nas 105 chamadas, confirmando o cache aquecido de expansões; as demais leituras, incluindo mutations e inserção de descobertas, permaneceram por requisição. As métricas agregadas em memória são fornecidas pelo endpoint `/metrics`.

Esta primeira coleta foi uma referência da implementação atual, antes de consolidar a listagem de casos com a validação da sessão. Uma repetição pré-otimização e a coleta após a mudança usaram a mesma conta, carga, número de amostras e processo da API; os [dados pré-otimização](M14-cases-concorrencia-pre-otimizacao-2026-10-24.csv) e [dados pós-otimização](M14-cases-concorrencia-pos-otimizacao-2026-10-24.csv) mostram:

| `GET /api/cases`, oito workers | Pré-otimização | Pós-otimização | Variação |
|---|---:|---:|---:|
| p50 | 380 ms | 178 ms | -53% |
| p95 | 1.803 ms | 1.569 ms | -13% |
| Média | 494 ms | 289 ms | -42% |
| Queries por chamada | 2 | 1 | -50% |

As 100 amostras medidas de cada rodada foram HTTP 200. O recorder confirmou 105 execuções da query consolidada após a mudança (cinco aquecimentos e 100 amostras), sem a antiga query separada `select_cases`. A p95 continua acima da meta P1 de 858 ms. O banco QA mostrou variação alta: a p95 pré-otimização desta repetição foi 1.803 ms, contra 1.256 ms na primeira referência concorrente, e a p95 de busca NPC variou entre 1.090 e 2.433 ms nas rodadas. A redução da p50/média e de um round-trip é consistente; a melhora da p95 ainda não é suficiente nem estável. A rodada não incluiu a rota de emails, cujo custo e autorização local exigem uma combinação representativa de NPC e acesso.

### Pool de conexões e cauda de `/api/cases`

O tracer agora publica `noir_db_pool_acquire_duration_seconds` e `noir_db_pool_acquisitions_total`, separados das métricas de execução SQL. Em uma rodada com o pool padrão de quatro conexões, 100 amostras concorrentes de casos ficaram em p50/p95 de **162/1.086 ms** (média 229 ms); o [CSV da rodada](M14-cases-concorrencia-pos-metrica-pool-2026-10-24.csv) contém os tempos HTTP. O recorder incluiu os cinco aquecimentos: 105 aquisições tiveram média de **67 ms**, e 105 queries tiveram média de **161 ms**; todas as queries ficaram abaixo de 250 ms, enquanto sete aquisições passaram de 500 ms. Isso indica que aquisição do pool contribui para parte da cauda, mas não explica sozinho toda a latência.

Foi feita uma tentativa exploratória elevando o pool para oito conexões, com a mesma conta e carga. O resultado piorou: p50/p95 de casos foi **164/5.235 ms**, e média de aquisição e query agregadas subiram para **244/306 ms**. O [CSV dessa rodada](M14-cases-pool8-concorrencia-2026-10-24.csv) registra as amostras. O aumento de conexões foi descartado: nesta configuração e banco QA, mais concorrência coincidiu com maior latência tanto na aquisição quanto na execução SQL. O limite padrão de quatro conexões foi mantido.

## Reconstrução de `CityWorldForCases`

O método foi medido diretamente com o roteiro [measure_city_world.go](../../../apps/api/scripts/measure_city_world.go), cinco aquecimentos e 100 chamadas medidas para Lyon. O tracer ficou conectado ao pool PostgreSQL e registrou as queries do método. A medição não inclui autenticação ou serialização de um endpoint HTTP real. Os CSVs guardam a [linha de base](M14-city-world-base.csv), a [redução de relações lidas](M14-city-world-relacoes-otimizadas.csv) e o [cache aquecido](M14-city-world-cache.csv).

| Implementação | p50 | p95 | Média | Queries por chamada aquecida |
|---|---:|---:|---:|---:|
| Sem cache, lendo relações da cidade inteira | 1.484 ms | 1.615 ms | 1.501 ms | 5 |
| Relações consultadas só para NPCs expandidos | 1.136 ms | 1.319 ms | 1.173 ms | 5 |
| Base e expansões cacheadas; status continua no banco | 297 ms | 310 ms | 299 ms | 2 |

O cache aquecido reduziu p50 em **80%** e p95 em **81%** contra a linha de base. Em 106 chamadas, `select_world_entities`, `select_employment_relationships` e `select_personal_relationships` executaram uma vez cada; `select_worlds` e `select_world_mutations` continuaram em todas as chamadas. Assim, os NPCs, empregos e relações pessoais determinísticos ou já materializados são reutilizados, enquanto versões do mundo e status continuam sendo lidos do banco. A primeira chamada ainda levou cerca de 2,04 s, perto das anteriores (2,52 s sem cache e 2,07 s após filtrar vínculos).

O cache da base usa hash da seed, versões global e da cidade e ID da cidade. As expansões e seus vínculos ficam separados por servidor/cidade e são invalidados após o commit de `ExpandDistrictSociety`. A validação real aqueceu o cache, expandiu `fr/lyon/district/beau-atelier` e confirmou os dois NPCs, dois empregos e um casamento novos; a chamada aquecida seguinte não duplicou os dados. As expansões e vínculos ficam no banco QA e são descartáveis. Os caches são locais ao processo e expiram após dez minutos.

## Reconstrução de `WorldForCases`

O método foi medido diretamente com o mesmo roteiro, cinco aquecimentos e 100 chamadas medidas. A linha de base foi coletada antes de combinar as leituras de mundo e versões de cidade; a rodada após cache usou o mesmo servidor QA e a mesma carga serial. Os CSVs registram a [linha de base](M14-world-for-cases-base.csv) e a [medição após cache](M14-world-for-cases-cache.csv).

| Implementação | p50 | p95 | Média | Queries por chamada aquecida |
|---|---:|---:|---:|---:|
| Sem cache de geração e expansões | 1.425 ms | 1.675 ms | 1.497 ms | 7 |
| Base determinística e expansões cacheadas; mutações continuam no banco | 311 ms | 398 ms | 351 ms | 2 |

A p50 caiu **78%** e a p95 **76%**. A chamada aquecida lê a configuração combinada do mundo/cidades e as mutações; NPCs e vínculos determinísticos são reaproveitados. Na medição após a consolidação da query fria, `select_world_entities` ocorreu uma vez em 106 chamadas e `select_world_mutations`, nas 105 chamadas com cache já quente. A primeira observação da coleta antiga levou 3,18 s, mas incluía pool de conexões frio.

O status mais recente dos NPCs continua sendo lido do banco em todos os caminhos: na query do payload em um miss e na query de mutações em cache hit. A chave da base usa hash da seed, versão global e versões das cidades; expansões são isoladas por servidor, protegidas por revisão e invalidadas após o commit. O cache é local ao processo e expira em dez minutos.

### Chamada fria com conexões já abertas

A primeira leitura anterior (3,18 s) misturava cache frio com abertura de conexões: o roteiro só fazia `Ping` antes de disparar leituras paralelas. Para separar esses custos, o roteiro agora pode abrir e devolver conexões ao pool sem chamar `WorldForCases`, e repetir misses com instâncias novas do store, preservando o pool aberto. Dez misses no QA deram p50 de **1.002 ms**, p95 de **2.282 ms**, média de **1.146 ms** e máximo de **2.282 ms**. O [CSV de misses com pool aquecido](M14-world-for-cases-cold-cache-pool-aquecido.csv) traz os tempos individuais.

O `EXPLAIN (ANALYZE, BUFFERS)` das leituras anteriores mostrou execução PostgreSQL de 5–9 ms por query para o volume QA. O tracer observava 290–600 ms, indicando que round-trip e latência do pool dominavam; não há evidência para adicionar índices nessas leituras. Por isso, entidade, vínculos e status foram agregados em JSON numa única query de miss. Os dez misses com pool aquecido caíram de cinco leituras para duas (configuração e payload); p50/p95 foi **836/1.295 ms**, média de 862 ms. Na rodada anterior comparável, os dez misses tiveram p50/p95 de 1.002/2.282 ms e média de 1.146 ms. Isso representa queda observada de 17%/43% em p50/p95, mas as rodadas não foram intercaladas; a redução de cinco para duas queries é a evidência estrutural, e a variação do QA impede atribuir todo o ganho aos round-trips removidos. O [CSV após consolidação](M14-world-for-cases-cold-query-unica.csv) contém os misses.

Após a mudança, uma bateria de 100 chamadas aquecidas ficou em p50/p95 de **300/307 ms**, sem regressão frente aos 311/398 ms anteriores. O [CSV com chamadas aquecidas após consolidação](M14-world-for-cases-cache-query-unica.csv) registra todas as amostras. A invalidação e a ausência de duplicação também foram verificadas após uma expansão QA: 166→168 NPCs, 166→168 empregos e 83→84 relações pessoais; a chamada subsequente manteve as contagens. Os dados adicionados são descartáveis.

### Carga concorrente

Com o cache aquecido, 100 chamadas diretas em oito workers ficaram em p50/p95 de **305/603 ms**, média de 355 ms e máximo de 1.377 ms. A p50 permaneceu próxima da serial; a p95 dobrou em relação à bateria serial de 297/300 ms. O [CSV concorrente](M14-world-for-cases-concorrencia-8.csv) registra as amostras. É um perfil do método do store, sem autenticação ou serialização HTTP.
