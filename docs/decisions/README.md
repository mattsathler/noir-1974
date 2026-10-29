# Decisões (ADRs)

Cada decisão de arquitetura ou de produto que afeta o desenvolvimento vira um arquivo numerado nesta pasta.

## Formato

Usar [`_template.md`](_template.md). Número com 4 dígitos. Título no kebab-case.

Status: `Proposta` | `Aceita` | `Supersedida` | `Rejeitada`.

## Índice

| ADR | Título | Status |
| --- | --- | --- |
| [0001](0001-mundo-compartilhado-por-servidor.md) | Mundo compartilhado por servidor, expansível sob demanda | Aceita |
| [0002](0002-jogo-100-web.md) | Jogo 100% web, HTML/CSS/JS, REST | Aceita |
| [0003](0003-modelo-hibrido.md) | Seed + relações + mutations | Aceita |
| [0004](0004-backend-autoritativo.md) | Backend autoritativo | Aceita |
| [0005](0005-casos-como-consequencia.md) | Casos como consequência do mundo | Aceita |
| [0006](0006-versionamento-do-gerador.md) | Versionamento do gerador | Aceita |
| [0007](0007-monorepo.md) | Monorepo frontend + backend | Aceita |
| [0008](0008-relacoes-tipadas.md) | Relações tipadas por domínio | Aceita |
| [0009](0009-stack-concreto-da-api.md) | Stack concreta da API | Aceita |
| [0010](0010-geografia-fixa-e-bairros-procedurais.md) | Geografia fixa e bairros procedurais | Supersedida |
| [0011](0011-predios-e-empresas.md) | Prédios e empresas como entidades distintas | Supersedida |
| [0012](0012-computador-proxy-e-sites-investigativos.md) | Computador pessoal, proxy e investigação por sites | Aceita |

Quando uma decisão nova **substitui** outra, marcar a antiga como `Supersedida` e apontar o ADR sucessor.
