# Backend

- O gerador e a persistência vivem na API (`apps/api`), não no browser.
- Endpoints de escrita validam: servidor do jogador, existência da entidade, se a mutation é permitida e conflitos com alterações concorrentes.
- Não expor ao cliente campos de autoridade (ex.: `killer_id`) só porque o caso já os tem no banco. O que o jogador ainda não descobriu não viaja na resposta.
- A classificação interna de pistas e sua relação com crimes/acontecimentos ficam no backend; não expor metadados ou sinais que permitam ao cliente distinguir pistas de eventos comuns. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md).
- Consultas de relação são queries (ou índice equivalente), não “gerar todos os NPCs e filtrar”.
- Linguagem e framework ainda abertos; quando existirem, esta regra ganha um ADR de stack.

## Organização por módulos

- `apps/api/cmd/` contém somente pontos de entrada e composição de processos. Regras de domínio e handlers não ficam em `main.go`.
- Código de produção fica em `apps/api/internal/`, agrupado pelo módulo de domínio que possui a regra: por exemplo, `world`, `cases`, `identity` e `servers`.
- Cada módulo mantém juntos os casos de uso, modelos, validações e portas de persistência que mudam pela mesma razão. Adaptadores HTTP e de banco ficam próximos do módulo que os utiliza; não criar camadas vazias ou abstrações sem consumidor real.
- Dependências entre módulos passam por contratos explícitos. Um módulo não importa detalhes internos de outro; extrações para `internal/platform/` só ocorrem quando há necessidade compartilhada comprovada.
- Testes unitários e de integração ficam no pacote testado, usando o sufixo `_test.go`. Testes de API ponta a ponta ficam em `apps/api/test/e2e/` e cobrem contratos observáveis pelo cliente.
- `cmd/` deve permanecer fino: carregar configuração, construir dependências, iniciar e encerrar o servidor.

Estrutura de referência:

```text
apps/api/
├── cmd/server/                 # composição e ciclo de vida do processo
├── internal/
│   ├── cases/                  # domínio, casos de uso, persistência e testes
│   ├── world/
│   ├── identity/
│   └── platform/               # infraestrutura realmente compartilhada
└── test/e2e/                   # jornadas e contratos HTTP completos
```

## Cobertura e estratégia de testes

- Código de produção deve manter 100% de cobertura de statements, funções e branches. Cobertura é um gate obrigatório de CI e não pode ser reduzida para fazer uma alteração passar.
- A métrica deve refletir o que a linguagem consegue medir. Em Go, `go test -coverprofile` mede statements, mas não fornece cobertura de branches; a cobertura de branches deve usar ferramenta adicional ou medição equivalente. Não declarar 100% de branches sem evidência dessa métrica.
- Código novo é entregue com testes no mesmo change. Cobertura excluída só pode incluir artefatos gerados ou código não executado no produto, com justificativa explícita e revisão; não se exclui código apenas para atingir o limite.
- Testes são planejados a partir dos comportamentos e falhas esperados, com dados explícitos e resultados verificáveis. Devem ser determinísticos e isolados de relógio, rede e estado compartilhado. Testes que usam PostgreSQL usam banco/branch efêmero, schema isolado e limpeza garantida; nunca usam dados de produção.
- A suíte combina unitários para regras de domínio e casos de borda; integração para banco, filas e adaptadores; e2e para contratos e jornadas HTTP; regressão para defeitos corrigidos; mutação para validar a força da suíte; e análise de vulnerabilidade para dependências e superfícies expostas.
- Testes de regressão reproduzem o defeito antes de validar a correção. Testes de mutação têm limite de mutantes sobreviventes documentado e tratado como gate. Testes e2e não substituem os níveis unitário e de integração.
- CI executa os gates de cobertura, testes aplicáveis, análise de vulnerabilidade e verificações estáticas antes de aceitar mudanças. Achados de segurança exigem correção ou exceção registrada com responsável e prazo. Um gate ausente ou desabilitado não conta como aprovação.

## Desempenho

- Cada endpoint e processo de geração define, antes de ser aceito, a carga representativa, escala de dados, métricas e orçamento de desempenho relevantes. Metas de latência, throughput e memória são critérios explícitos do milestone ou da operação; não se escolhem limites sem carga de referência.
- Benchmarks usam dados e seeds reproduzíveis, distinguem aquecimento de medição e registram ambiente, versão, tamanho dos dados, latência (incluindo percentis quando aplicável), throughput, alocações e memória. Cargas de benchmark não compartilham estado mutável.
- Mudanças em caminhos críticos são comparadas a uma baseline sob a mesma carga. Regressões acima do orçamento definido bloqueiam a mudança até serem explicadas e corrigidas ou aceitas formalmente.
- Otimização é orientada por profiling e evidência. Consultas devem respeitar índices e evitar N+1, filtragem em memória de conjuntos que podem ser consultados no banco e carregamento de dados que o endpoint não usa.
- Trabalho de rede, banco e processamento limitado por requisição recebe timeout ou cancelamento via `context.Context`; limites de concorrência e tamanho de entrada devem ser explícitos para operações que possam crescer com a carga.
- Benchmarks medem tendência e regressão de desempenho; não substituem testes funcionais nem devem tornar a suíte unitária instável por depender de relógio compartilhado ou condições da máquina.
