# Documentação — Noir 1974

Esta pasta é a fonte de verdade escrita do projeto. Código implementa o que está aqui; quando a implementação forçar uma mudança, registra-se a decisão em `decisions/` e atualiza-se a arquitetura.

| Pasta | Para quê |
| --- | --- |
| [`project/`](project/visao.md) | O que o jogo é: visão, loop, universo narrativo e [lore](project/lore.md) |
| [`architecture/`](architecture/README.md) | Arquitetura desejada (modelo híbrido, geração, API conceitual) |
| [`decisions/`](decisions/README.md) | Decisões tomadas no desenvolvimento (ADRs) |
| [`dev-rules/`](dev-rules/README.md) | Regras para quem escreve código neste repositório |
| [`rules/`](rules/README.md) | Regras obrigatórias de identidade e conteúdo do jogo |
| [`roadmap/`](roadmap/README.md) | Milestones, ciclos e [estado atual / % até o playtest](roadmap/estado.md) |

## Como usar

- **Entender o jogo** → `project/`
- **Saber o que construir** → `architecture/` + milestone atual em `roadmap/`
- **Saber o que já foi fechado** → `decisions/`
- **Saber como contribuir** → `dev-rules/`
- **Auditar decisões e criações do jogo** → `rules/`
- **Saber o que vem depois** → `roadmap/`
- **Saber a % até o primeiro playtest** → [`roadmap/estado.md`](roadmap/estado.md)

Decisões ainda abertas ficam em [`architecture/pontas-soltas.md`](architecture/pontas-soltas.md). Não implementar como se estivessem fechadas.
