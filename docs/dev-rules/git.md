# Git e documentação

- Código em inglês; commits em português ou inglês, mas **consistentes** e no modo imperativo, focados no porquê (`explica`, `impede`, `permite`).
- Mudança de decisão de arquitetura: ADR novo + índice em `docs/decisions/README.md` + ajuste em `architecture/` se a visão mudar.
- Ponta solta não se “fecha” só no código. Ou ADR, ou continua em `pontas-soltas.md`.
- Não commitar `.env`, secrets, dumps de mundo gerado nem snapshots gigantes sem necessidade de teste.
- Um PR de feature cita o milestone do roadmap quando houver um vigente.
- PR que aceita um milestone atualiza `docs/roadmap/estado.md` via skill `atualizar-estado-roadmap` (não calcular a % à mão).
- `packages/` só quando dois apps precisarem do mesmo módulo.
