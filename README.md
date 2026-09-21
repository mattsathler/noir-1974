# Noir 1974

Jogo de investigação sandbox, 100% web. O jogador resolve casos gerados proceduralmente em um mundo próprio — uma sociedade coerente, não uma sequência linear de diálogos.

Cada jogador tem a sua instância: seed, NPCs, empresas, relações, mutações e casos não são compartilhados.

## Monorepo

| Caminho | Responsabilidade |
| --- | --- |
| `apps/web` | Cliente web (HTML, CSS, JavaScript) |
| `apps/api` | Backend autoritativo (REST API + gerador + persistência) |
| `docs/` | Projeto, arquitetura, decisões, regras de desenvolvimento e roadmap |

O cliente **não** é fonte de verdade do estado. Toda reconstrução do mundo passa pelo servidor.

## Arquitetura (resumo)

> A seed cria a realidade base; as relações dão estrutura à sociedade; as mutations registram o que mudou; e os casos utilizam essa sociedade para criar investigações coerentes.

Detalhe em [`docs/architecture`](docs/architecture/README.md). Índice de toda a documentação em [`docs/README.md`](docs/README.md).

## Começar por aqui

1. Visão do jogo: [`docs/project/visao.md`](docs/project/visao.md)
2. Princípios: [`docs/architecture/principios.md`](docs/architecture/principios.md)
3. Regras de desenvolvimento: [`docs/dev-rules/README.md`](docs/dev-rules/README.md)
4. Roadmap e ciclos: [`docs/roadmap/README.md`](docs/roadmap/README.md)
5. Estado atual (milestones aceitos e % até o playtest): [`docs/roadmap/estado.md`](docs/roadmap/estado.md)
