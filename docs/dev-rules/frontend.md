# Frontend

- `apps/web` renderiza e envia intenções (`quero o NPC`, `avanço o caso`). Não gera mundo, não escolhe culpado, não aplica mutation local como verdade.
- Estado local é cache de UI. Conflito com o servidor: o servidor ganha.
- Sem lógica de “resolver o caso no cliente” que o backend não possa recusar.
- HTML/CSS/JS; não introduzir framework sem ADR.

## Organização por módulos e páginas

- `apps/web/src/app/` contém inicialização da aplicação, composição global e roteamento.
- `apps/web/src/modules/` contém os módulos do jogo. Cada módulo representa uma área funcional (por exemplo, `prologue`, `identity`, `cases`, `world`).
- Cada página tem sua própria pasta dentro do módulo: `modules/<modulo>/pages/<pagina>/`.
- Componentes, estado e lógica exclusivos de uma página ficam junto dela. Elementos reutilizáveis entre módulos pertencem ao `apps/@noir-ui` ou a uma área compartilhada explicitamente justificada.
- Testes unitários e de integração ficam junto ao código testado, na mesma pasta da página ou do módulo. Testes e2e ficam em `apps/web/e2e/` e descrevem fluxos completos do usuário.
- Não criar pastas genéricas como `pages/`, `components/` ou `utils/` na raiz de `src` para código que pertence a um módulo.

Estrutura de referência:

```text
apps/web/src/
├── app/                         # bootstrap, composição e rotas
└── modules/
    ├── prologue/
    │   └── pages/
    │       └── introduction/    # página, componentes e testes locais
    ├── identity/
    │   └── pages/
    │       └── login/
    └── cases/
        └── pages/
            └── case-board/
```

## Cobertura e estratégia de testes

- Código de produção deve manter 100% de cobertura de linhas, funções, branches e statements. Cobertura é um gate obrigatório de CI e não pode ser reduzida para fazer uma alteração passar.
- Código novo é entregue com testes no mesmo change. Cobertura excluída só pode incluir artefatos gerados ou código não executado no produto, com justificativa explícita e revisão; não se exclui código apenas para atingir o limite.
- Cada teste deve ter escopo e comportamento esperado definidos antes de sua implementação. Testes precisam ser determinísticos, isolados de relógio, rede e estado compartilhado, e usar fixtures pequenas e explícitas. Dependências externas são substituídas por contratos controlados, exceto nos testes que as verificam deliberadamente.
- A suíte deve cobrir níveis complementares: unitários e de regressão para regras e componentes; integração para contratos entre módulos e API; e2e para jornadas críticas do jogador; mutação para avaliar se os testes detectam alterações comportamentais; e análise de vulnerabilidade para dependências e superfícies expostas.
- Testes de regressão reproduzem o defeito antes de validar a correção. Testes de mutação devem ter limite de mutantes sobreviventes documentado e tratado como gate. E2E deve cobrir jornadas selecionadas, sem substituir cobertura unitária ou de integração.
- Análise de vulnerabilidade roda no fluxo de CI e em atualizações de dependências. Exceções a achados precisam de avaliação, responsável e prazo; vulnerabilidade não é considerada coberta pela métrica de linhas.
- CI executa os gates de cobertura, tipos de teste aplicáveis e análise de segurança antes de aceitar mudanças. Um gate ausente ou desabilitado não conta como aprovação.
