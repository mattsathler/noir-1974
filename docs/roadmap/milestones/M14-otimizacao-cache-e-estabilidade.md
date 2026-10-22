# M14 — Otimização, cache e estabilidade da plataforma

**Ciclo:** 14  
**Status:** Não iniciado  
**Depende de:** M13 aceito

## Objetivo

Reduzir o tempo de resposta e a carga desnecessária no front e no back, mantendo os dados corretos e isolados por jogador, servidor, cidade e contexto. Melhorar a estabilidade com tratamento consistente de falhas, limites operacionais e visibilidade suficiente para diagnosticar problemas em produção.

## Critérios de aceite

- [ ] Os fluxos e endpoints de maior uso têm uma linha de base de latência e carga documentada; metas de latência p50/p95 e carga são definidas a partir dessa medição antes das otimizações.
- [ ] Requisições repetidas e simultâneas no front são deduplicadas quando compartilham o mesmo contexto; cache local tem chaves explícitas, política de expiração e invalidação após mutações relevantes.
- [ ] Leituras reutilizáveis no back têm cache compartilhado com TTL e invalidação documentados; dados personalizados ou sensíveis não vazam entre jogadores ou servidores.
- [ ] Cache respeita os escopos corretos de servidor, cidade, jogador, consulta e dia do mundo. Dados do mundo compartilhado e dados pessoais permanecem separados.
- [ ] Respostas desatualizadas não substituem silenciosamente dados autoritativos após uma mutação; falhas de cache degradam para a fonte de dados quando isso for seguro.
- [ ] Consultas e reconstruções caras dos fluxos prioritários são medidas e otimizadas com evidência, reduzindo round-trips, trabalho duplicado ou leitura excessiva sem alterar seus resultados funcionais.
- [ ] A API aplica prazos e cancelamento apropriados a operações dependentes de rede ou banco, e retorna erros estruturados com código, status HTTP e request ID consistentes.
- [ ] O front apresenta erros recuperáveis de forma clara e permite nova tentativa quando apropriado; respostas e mensagens públicas não expõem credenciais, SQL ou detalhes internos.
- [ ] Métricas e logs permitem acompanhar latência, erros e acertos/falhas de cache por fluxo, com request ID para correlacionar uma falha entre front e back.
- [ ] Testes automatizados cobrem isolamento, expiração e invalidação de cache, deduplicação de requests e os principais caminhos de timeout e erro; benchmarks dos fluxos prioritários confirmam as metas definidas.

## Fora de escopo

Alterações nas regras de jogo ou na verdade dos casos, cache de conteúdo sensível sem política de autorização explícita, e otimizações especulativas sem medição que demonstre impacto.
