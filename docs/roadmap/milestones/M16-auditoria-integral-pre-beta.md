# M16 — Auditoria integral de estabilidade e consistência pré-beta

**Ciclo:** 16
**Status:** Não iniciado
**Depende de:** M14 aceito e M15 aceito

## Objetivo

Executar uma auditoria ampla e sistemática do produto antes do lançamento beta, cobrindo front-end, back-end e os contratos entre eles. Encontrar e corrigir bugs, falhas de estabilidade e inconsistências visuais ou funcionais nos fluxos completos, incluindo estados raros, erros recuperáveis, diferentes tamanhos de tela e dados representativos. A auditoria deve produzir evidências reproduzíveis e um registro priorizado de achados; nenhum problema bloqueador ou crítico pode permanecer aberto no aceite.

## Escopo da auditoria

- **Front-end e consistência visual:** navegação e hierarquia, alinhamento, espaçamento, tipografia, cores e tokens, ícones, imagens, clipping, overflow, scroll, estados vazios/carregando/erro/sucesso, modais, foco e teclado, responsividade e legibilidade. Conferir telas e componentes em resoluções suportadas, zoom e conteúdo curto/longo, incluindo traduções e nomes extensos quando aplicável.
- **Fluxos de ponta a ponta:** criação/entrada no mundo, investigação, registros, correio, viagem, visita a locais, interrogatório, busca e apreensão, organização de evidências, acusação e prisão. Verificar transições, atualização após mutações, recarga, navegação para trás/para frente e repetição de ações.
- **Back-end e API:** validação de entrada, autorização e isolamento por jogador/servidor/cidade, contratos e códigos HTTP, consistência transacional, concorrência, idempotência quando necessária, erros e timeouts, limites, consultas/reconstruções, integridade e migrações relevantes.
- **Integração e resiliência:** latência ou indisponibilidade de rede/banco, respostas lentas, duplicadas, fora de ordem ou malformadas, retry e cancelamento, cache/invalidação, sessão expirada, falhas parciais e recuperação sem perda ou duplicação de estado.
- **Compatibilidade e operação:** navegadores e ambientes oficialmente suportados, logs e métricas úteis, request IDs e diagnóstico sem exposição de dados sensíveis; confirmar que os caminhos críticos podem ser observados durante o beta.

## Critérios de aceite

- [ ] Existe uma matriz de auditoria versionada que mapeia cada fluxo crítico, tela/estado, endpoint e combinação relevante de viewport/ambiente para os cenários executados; lacunas e itens não aplicáveis têm justificativa.
- [ ] Os fluxos ponta a ponta críticos foram percorridos em sucesso, erro recuperável, estado vazio/carregando e repetição/retomada, com resultados e evidências reproduzíveis registrados.
- [ ] As telas prioritárias foram inspecionadas em todos os viewports e navegadores suportados; inconsistências visuais têm evidência antes/depois e resolução registrada.
- [ ] Os contratos front/back foram confrontados com implementação e documentação; validação, autorização, isolamento, erros, timeouts e comportamento após mutação foram verificados nos endpoints críticos.
- [ ] Concorrência, requests duplicadas ou fora de ordem, falhas de rede/banco, cache e invalidação foram exercitados nos caminhos em que podem afetar integridade ou experiência.
- [ ] Todo achado tem severidade, passos para reproduzir, ambiente, evidência, área responsável e estado. Achados críticos e bloqueadores foram corrigidos e revalidados; os demais têm correção aceita ou plano explícito com responsável e prazo antes do beta.
- [ ] Não há bugs bloqueadores/críticos conhecidos nem regressões visuais críticas abertas. A lista residual de problemas e riscos foi revisada e aceita pelos responsáveis pelo lançamento beta.
- [ ] As correções foram cobertas por verificações automatizadas ou uma justificativa documentada quando a automação não for adequada; a suíte de regressão e os checks de integração relevantes passam no ambiente de referência.
- [ ] Existe um relatório final de auditoria com cobertura, achados, correções, pendências aceitas, evidências, ambiente e decisão explícita de prontidão para beta.

## Fora de escopo

Adicionar funcionalidades ou mudar regras de jogo sem relação com um achado; redesenhar a direção artística; suportar navegadores, dispositivos ou ambientes fora da matriz oficial. Mudanças de escopo descobertas durante a auditoria devem virar trabalho separado, salvo correção necessária para remover bloqueador ou risco crítico.
