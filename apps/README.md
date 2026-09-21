# apps

Aplicações do monorepo.

| App | Papel |
| --- | --- |
| [`web`](web/README.md) | Frontend |
| [`api`](api/README.md) | Backend |

Código compartilhado (tipos, contratos de API, utilitários de seed) só entra em `packages/` quando houver um consumidor real nos dois lados. Não criar pacote “por precaução”.
