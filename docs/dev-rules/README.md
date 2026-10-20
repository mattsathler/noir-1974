# Regras de desenvolvimento

Regras para quem implementa neste repositório. Arquitetura descreve o *quê*; isto descreve o *como*.

| Documento | Escopo |
| --- | --- |
| [Determinismo](determinismo.md) | Seeds, versões, testes de reconstrução |
| [Persistência](persistencia.md) | O que vai ao banco, o que não vai |
| [Identificação](identificacao.md) | IDs, `server_id`, referências |
| [Backend](backend.md) | Autoridade, API, gerador |
| [Frontend](frontend.md) | Cliente, o que não pode decidir |
| [Git e docs](git.md) | Commits, ADRs, documentação |

## Precedência

1. Regras obrigatórias em `docs/rules/` — nenhum artefato ou decisão pode contrariá-las.
2. ADR aceito em `docs/decisions/`, desde que compatível com as regras.
3. Arquitetura em `docs/architecture/`, desde que compatível com regras e ADRs.
4. Estas regras de desenvolvimento.
5. Convenção local do arquivo.

Se o código precisar violar um ADR, a mudança é um **novo ADR**, não um comentário no PR.
