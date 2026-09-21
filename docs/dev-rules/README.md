# Regras de desenvolvimento

Regras para quem implementa neste repositório. Arquitetura descreve o *quê*; isto descreve o *como*.

| Documento | Escopo |
| --- | --- |
| [Determinismo](determinismo.md) | Seeds, versões, testes de reconstrução |
| [Persistência](persistencia.md) | O que vai ao banco, o que não vai |
| [Identificação](identificacao.md) | IDs, `world_id`, referências |
| [Backend](backend.md) | Autoridade, API, gerador |
| [Frontend](frontend.md) | Cliente, o que não pode decidir |
| [Git e docs](git.md) | Commits, ADRs, documentação |

## Precedência

1. ADR aceito em `docs/decisions/`
2. Arquitetura em `docs/architecture/`
3. Estas regras
4. Convenção local do arquivo

Se o código precisar violar um ADR, a mudança é um **novo ADR**, não um comentário no PR.
