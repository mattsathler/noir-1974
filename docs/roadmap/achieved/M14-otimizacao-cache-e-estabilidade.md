# M14 — Otimização, cache e estabilidade da plataforma

**Ciclo:** 14  
**Status:** Concluída

**Depende de:** M13 aceito

## Objetivo

Reduzir o tempo de resposta e a carga desnecessária no front e no back, mantendo os dados corretos e isolados por jogador, servidor, país, cidade e contexto. Melhorar a estabilidade com tratamento consistente de falhas, limites operacionais e visibilidade suficiente para diagnosticar problemas em produção.

## Frentes priorizadas

- Reduzir round-trips e leituras excessivas nas queries dos endpoints mais usados.
- Evitar reconstruir cidades ou mundos completos quando o endpoint precisa de um escopo menor.
- Reduzir conversões repetidas entre structs de persistência, domínio e resposta da API.
- Tratar e agregar dados uma vez por requisição, removendo trabalho duplicado e padrões N+1.
- Medir latência e custo antes e depois de cada otimização, preservando as regras e os resultados do jogo.

## Otimizações implementadas e medições parciais

- **Busca de pessoas na delegacia:** deriva o país da localização do jogador, gera somente as cidades desse país, filtra NPCs persistidos pelos IDs exatos de distrito e carrega relações persistidas em lote somente para os resultados. A interface mostra a jurisdição nacional ativa.
- **Cache da busca nacional:** o conteúdo determinístico gerado fica em cache no processo por até 10 minutos, com limite de 16 escopos e chave baseada no hash da seed e nas versões do mundo/cidades. A lista de NPCs expandidos é cacheada por servidor e país e invalidada depois de `ExpandDistrictSociety`; status e relações continuam vindo do banco em cada requisição. O cache é local ao processo, não compartilhado entre réplicas. A invalidação foi validada no banco QA: após aquecer o cache e expandir `fr/lyon/district/beau-atelier`, os dois NPCs recém-criados foram encontrados na busca seguinte.
- **Busca nacional após cache:** numa bateria isolada de 100 amostras, a p50/p95 foi 1,27/3,56 s (média 1,62 s). Depois de combinar a leitura da sessão e da localização, nova bateria isolada ficou em 895 ms/2,28 s (média 1,11 s); todas as 200 chamadas responderam HTTP 200. A consulta separada de `player_locations` foi removida. A p50/p95 caíram 27%/22% em relação à bateria imediatamente anterior, e ambas ficaram abaixo da meta P0 de 1,25/2,96 s. O tempo frio permaneceu próximo de 5,8 s. O tracer também confirmou um único `select_world_entities` nas 106 requisições com cache de expansões aquecido. Detalhes e CSVs em [relatório de latências](../relatorios/M14-latencias-2026-10-24-nacional.md).
- **Reconstrução de cidade para casos:** `CityWorldForCases` cacheia a cidade determinística e os NPCs expandidos, empregos e relações persistidas. A chave inclui seed, versões global/cidade e cidade; expansões são isoladas por servidor/cidade e invalidadas após `ExpandDistrictSociety`. Status continua vindo do banco. Na medição direta de 100 chamadas, p50/p95 caiu de 1,48/1,61 s para 297/310 ms; chamadas aquecidas passaram de cinco queries para duas. O tracer viu uma consulta de NPCs expandidos e uma consulta por tipo de relação em 106 chamadas. A invalidação foi verificada após uma expansão QA, incluindo dois NPCs, dois empregos e um casamento. Detalhes e CSVs no [relatório](../relatorios/M14-latencias-2026-10-24-nacional.md).
- **Reconstrução de mundo para casos:** `WorldForCases` combina a leitura de seed/versões, cacheia a geração determinística e as expansões por servidor, e sempre consulta mutações. Em 100 chamadas aquecidas, p50/p95 ficou em 300/307 ms contra 311/398 ms antes da query fria consolidada. Um miss caiu de cinco queries para duas: configuração e um payload JSON de entidades, relações e status. `EXPLAIN ANALYZE` mediu 5–9 ms de execução PostgreSQL nas leituras anteriores, muito abaixo dos 290–600 ms do tracer; não há indicação de que índices reduzam a latência observada. Em dez misses com pool aberto, p50/p95 caiu de 1.002/2.282 ms para 836/1.295 ms, em rodadas separadas. A invalidação foi verificada após expansão QA: dois NPCs, dois empregos e uma relação pessoal apareceram sem duplicação. Linha de base, medições e CSVs estão no [relatório](../relatorios/M14-latencias-2026-10-24-nacional.md).
- **Listagem de casos:** `GET /api/cases` valida a sessão, resolve a cidade e lista os casos abertos numa única query, reduzindo a rota de três queries (linha de base inicial) para uma. Em carga concorrente de oito workers, 100 amostras HTTP 200 tiveram p50/p95 de 380/1.803 ms antes e 178/1.569 ms depois; a média caiu de 494 para 289 ms. A p95 continua acima da meta P1 de 858 ms e varia com o banco QA. Baterias seriais isoladas anteriores mediram p50/p95 de 350/655 ms. Detalhes e CSVs no [relatório](../relatorios/M14-latencias-2026-10-24-nacional.md).
- **Localização na busca:** reutiliza o `cityID` já resolvido ao carregar cenas de casos, evitando uma segunda consulta de localização.
- **Geração de emails:** carrega colegas de todos os empregadores em uma query e agrega identidade, status, empregos e relações pessoais do NPC em outra; a leitura de mensagens do caso se sobrepõe à geração determinística do email comum. A rota completa, antes/depois da reutilização da cidade, caiu de p50/p95 2,44/4,33 s para 2,12/4,24 s; `select_player_locations` caiu de três para uma execução por chamada. No benchmark direto do serviço, a p50/p95 caiu de 1.433/1.536 ms para 614/717 ms; as queries caíram de sete para três por chamada. Com oito workers, ficou em 532 ms/2,15 s. A rota completa ainda não atingiu a meta P3; a cauda e o perfil com múltiplos empregadores ficam como acompanhamento da M15.
- **Telemetria:** tracer `pgx` registra contagem e histograma por rota, operação/tabela e status, sem guardar SQL ou argumentos. A query `select_world_entities` caiu de 2,51 s em uma chamada antes do filtro exato para média de 413 ms em 11 chamadas depois. A média HTTP dessas 11 chamadas foi 4,03 s; essa amostra não confirma a meta de latência.
- **Carga HTTP concorrente:** as rodadas tiveram 100 amostras por rota e oito workers, todas HTTP 200. Para casos, a comparação pareada pré/pós-query única foi p50/p95 de 380/1.803 ms contra 178/1.569 ms; outra repetição pós-mudança com pool padrão mediu 162/1.086 ms. Com oito conexões no pool, p95 piorou para 5.235 ms e aquisição/query médias subiram para 244/306 ms; a configuração experimental foi descartada. A busca nacional ainda não tem comparação concorrente pré-otimização. A p95 de casos segue acima da meta P1. CSVs e método estão no [relatório](../relatorios/M14-latencias-2026-10-24-nacional.md).
- **Métricas de cache e aquisição do pool:** `/metrics` expõe contadores `noir_cache_requests_total` por cache e resultado (`hit`/`miss`), além de separar `noir_db_pool_acquire_duration_seconds` da duração SQL. No perfil padrão, sete de 105 aquisições passaram de 500 ms, enquanto todas as queries ficaram abaixo de 250 ms. As métricas são agregadas por processo e têm cardinalidade limitada; erros HTTP mantêm `request_id` nos logs e na resposta, sem transformar esse identificador em label Prometheus.
- **Deadline da API:** operações HTTP da API recebem contexto com prazo de 30 s, propagado às consultas; deadline excedido cancela o trabalho e retorna HTTP 504 com código `REQUEST_TIMEOUT` e o request ID da resposta. Teste unitário cobre cancelamento e envelope sem expor o erro interno.
- **Erros recuperáveis no front:** `APIErrorNotice` mostra a mensagem pública, código e request ID, sem renderizar `details`. Jornal e consultas do GATE oferecem retry para GETs; falha de viagem permite reabrir o seletor sem repetir automaticamente o POST. O carregamento inicial do jogo também oferece nova tentativa. Testes cobrem filtro de detalhes e retry.

## Acompanhamento transferido para M15

- `WorldForCases` em cache frio: dez misses após a query consolidada ficaram em p50/p95 de 836/1.295 ms; repetir a comparação fria pré/pós de forma intercalada e sob carga concorrente.
- `GET /api/cases`: a query caiu de duas para uma e a p50 concorrente caiu 53%, mas p95 permaneceu acima de 858 ms. Ampliar amostras sob carga controlada. Elevar o pool de quatro para oito piorou o p95 para 5.235 ms no QA e não será adotado.
- Emails: a rota HTTP completa ainda tem p95 de 4,24 s contra meta P3 de 2,97 s. Repetir com carga isolada e perfil com múltiplos empregadores; o benchmark atual tinha um vínculo de emprego.
- A busca nacional tem baseline concorrente pós-otimização, sem amostra concorrente prévia equivalente; coletar comparação quando houver ambiente QA estável.

Esses itens permanecem visíveis como acompanhamento, mas não bloqueiam a conclusão da M14: as otimizações estruturais foram entregues e medidas, as metas seriais prioritárias foram verificadas e os desvios de cauda foram documentados com dados reproduzíveis. O banco QA apresentou variabilidade e não oferece evidência suficiente para perseguir ajustes de pool ou índices nesta milestone.

## Linha de base e prioridades

A linha de base HTTP inicial está em [relatório de latências](../relatorios/M14-latencias-2026-10-23.md), com [amostras brutas](../relatorios/M14-latencias-2026-10-23.csv) e [roteiro de coleta](../../../apps/api/scripts/measure_m14_latency.py). Foram medidos seis endpoints autenticados, com 40 amostras por endpoint, cinco aquecimentos e chamadas seriais no banco QA. A amostra de busca é municipal e anterior ao cache nacional; não mede nem valida o desempenho da implementação atual.

| Prioridade | Endpoint | Linha de base p50 / p95 | Última medição | Meta inicial após otimização |
|---|---|---:|---:|---:|
| P0 | `GET /api/world/npcs?q=a` | 1,78 s / 4,23 s (bateria com rotas alternadas) | 895 ms / 2,28 s (cache aquecido; isolada) | ≤ 1,25 s / ≤ 2,96 s |
| P1 | `GET /api/cases` | 550 ms / 1,37 s (bateria isolada, antes da mudança) | 350 ms / 655 ms (bateria isolada) | ≤ 372 ms / ≤ 858 ms |
| P2 | `GET /api/me`, `/api/world`, `/api/whiteboards`, `/api/world/newspaper` | p50 348–531 ms; p95 612–952 ms | — | Evitar regressões; investigar picos acima de 1 s |
| P3 | `GET /api/world/npcs/{npcID}/emails` | 2,12 s / 4,24 s | — | ≤ 1,48 s / ≤ 2,97 s |

As metas P0/P1 de redução de aproximadamente 30% usam os resultados atuais medidos em 24/10 sob carga serial; não são garantias para produção. A bateria P0 isolada não é diretamente comparável às baterias que alternaram seis rotas. A primeira coleta, de 23/10, usava escopo municipal e seus alvos já não se aplicam. A linha de base e as medições estão no [relatório de 24/10](../relatorios/M14-latencias-2026-10-24-nacional.md), com [CSV nacional](../relatorios/M14-latencias-2026-10-24-nacional.csv), [segunda bateria](../relatorios/M14-latencias-2026-10-24-pos-otimizacao-busca.csv) e [bateria isolada após cache](../relatorios/M14-latencias-2026-10-24-pos-cache-expansoes.csv). A busca de NPCs também chegou a 7,74 s e a lista de casos a 5,00 s na coleta anterior, indicando cauda longa que precisa ser observada separadamente.

## Plano de ataque

1. [x] Comparar a busca nacional aquecida e a lista de casos após as otimizações com 100 amostras; separar queries lentas, pool e reconstrução determinística.
2. [x] Comparar busca e listagem de casos com a mesma seed, conta QA e roteiro; as metas seriais iniciais foram verificadas.
3. [x] Medir a rota HTTP e o serviço de geração de emails antes/depois. A rota completa não atingiu a meta P3; o teste com perfil de múltiplos empregadores foi transferido para M15, pois a fixture disponível tinha um empregador.
4. [x] Executar carga HTTP concorrente nos endpoints prioritários; `/api/cases` e `/api/world/npcs?q=a` tiveram 100 amostras cada com oito workers, incluindo autenticação e serialização. Primeira referência concorrente pós-otimização; não há rodada concorrente pré-otimização comparável.

A API expõe `/metrics` em formato Prometheus: `noir_http_*` mede latência HTTP, `noir_db_queries_total` e `noir_db_query_duration_seconds_*` medem contagem e duração por rota/operação SQL, `noir_db_pool_*` mede espera de conexão, e `noir_cache_requests_total` mede hit/miss. O recorder vive em memória e reinicia junto com o processo; Prometheus deve coletá-lo para manter histórico. Métricas têm agregação por rota/cache, enquanto erros mantêm o request ID em logs e respostas; não se usa request ID como label devido à cardinalidade. A telemetria atual ainda não separa geração em memória e serialização HTTP.

## Critérios de aceite

- [x] Linha de base serial inicial documentada para seis endpoints, com metas iniciais p50/p95 definidas para os dois alvos prioritários.
- [x] Há carga concorrente controlada e 100 amostras por endpoint prioritário, medidas por HTTP real com autenticação, autorização, banco e serialização, em oito workers. `/api/cases` tem comparação concorrente pré/pós; `/api/world/npcs?q=a` tem referência pós-otimização e ainda precisa de linha prévia equivalente.
- [x] Requisições GET simultâneas no front são deduplicadas por URL, credenciais, cabeçalhos e contexto; cache local de whiteboards usa chave versionada, expiração de 24 horas e invalidação ao descartar um rascunho. Coberto por testes do cliente API e do cache local.
- [x] Leituras reutilizáveis têm cache de processo limitado, TTL de 10 minutos e invalidação após expansão. O Postgres continua sendo a fonte compartilhada e canônica para entidades persistidas, estado do jogador e mutações; projeções determinísticas são os únicos dados reutilizados sem nova leitura. Não há cache cross-replica nesta arquitetura, e dados pessoais/sensíveis não entram nos caches de mundo.
- [x] Chaves e invalidação isolam expansões por servidor/país/cidade; projeções-base usam hash de seed e versões de mundo/cidade. Estado mutável e pessoal permanece em leituras do banco. Testes cobrem isolamento entre servidores/países e revisão de invalidação.
- [x] A resposta GET do whiteboard verifica se houve mutação local enquanto a requisição estava em voo; nesse caso retorna o cache atualizado e não grava a resposta antiga. Teste cobre a corrida entre resposta remota antiga e rascunho atualizado.
- [x] Métricas de cache expõem hits/misses com labels limitados e teste de exportação. Rotas, operações SQL e pool são agregados; logs de erros e respostas estruturadas incluem request ID para investigação, sem labels de alta cardinalidade.
- [x] Duração e quantidade de queries SQL podem ser correlacionadas por rota; as queries/reconstruções prioritárias foram medidas e otimizadas com baterias comparáveis, reduzindo round-trips e trabalho repetido sem alterar os dados retornados. A rota de emails não alcançou a meta p95 e segue explicitamente para M15.
- [x] A API aplica prazo de 30 s e propaga cancelamento às operações dependentes de banco; deadline excedido retorna HTTP 504 estruturado com código e request ID consistentes. Erros de handlers também passam pelo envelope JSON com código, status e request ID.
- [x] O front apresenta falhas recuperáveis e permite nova tentativa apropriada nos fluxos principais; mensagens exibem código/request ID sem credenciais, SQL ou `details`. Jornal, GATE e carregamento inicial oferecem retry; viagem permite reabrir o seletor sem repetir a mutação.
- [x] Métricas agregadas acompanham latência HTTP, SQL, aquisição do pool e hit/miss de cache; erros incluem request ID na resposta e no log. O ID não é label das séries Prometheus, evitando cardinalidade ilimitada. A telemetria correlaciona falhas por request ID e desempenho agregado por rota/cache.
- [x] Testes cobrem isolamento de cache entre servidores/países, invalidação, expiração local, resposta antiga de whiteboard, deduplicação, timeout e erros recuperáveis. Os benchmarks dos fluxos prioritários foram executados e os desvios das metas concorrentes/P3 estão documentados como acompanhamento da M15, sem alterações especulativas no pool.

## Fora de escopo

Alterações nas regras de jogo ou na verdade dos casos, cache de conteúdo sensível sem política de autorização explícita, e otimizações especulativas sem medição que demonstre impacto.
