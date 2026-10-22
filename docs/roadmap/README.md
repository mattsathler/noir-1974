# Roadmap

O desenvolvimento avança em **ciclos**. Cada ciclo tem um **milestone de sucesso** com critérios de aceite binários. O ciclo seguinte **não começa** enquanto o anterior não for aceito.

**Onde estamos:** [`estado.md`](estado.md) — milestones aceitos e porcentagem do fluxo completo do crime, da geração à prisão do culpado. O M06 indica separadamente quando o primeiro recorte está pronto para playtest.

Prova técnica que sustenta o restante (meta do ciclo 4):

```text
seed → gerar mundo → salvar mutations → desligar → reiniciar
  → reconstruir → idêntico + mutations aplicadas
```

## Ciclos

Ordem e regras de entrada/saída: [`ciclos.md`](ciclos.md).

| Ciclo | Milestone | Documento |
| --- | --- | --- |
| 0 | Repositório e documentação | [M00](milestones/M00-repositorio-e-docs.md) |
| 1 | Modelo de entidades e schema inicial | [M01](milestones/M01-modelo-de-entidades-e-schema.md) |
| 2 | Gerador determinístico (cidade mínima) | [M02](milestones/M02-gerador-deterministico.md) |
| 3 | Sociedade: NPCs, empresas, relações | [M03](milestones/M03-sociedade-e-relacoes.md) |
| 4 | Mutations e reconstrução | [M04](milestones/M04-mutations-e-reconstrucao.md) |
| 5 | Gerador de casos simples + validação | [M05](milestones/M05-gerador-de-casos.md) |
| 6 | REST API e cliente web mínimo | [M06](milestones/M06-api-e-cliente.md) |
| 7 | Registros bancários e histórico do GATE | [M07](milestones/M07-registros-investigativos.md) |
| 8 | Sistema de correio eletrônico | [M08](milestones/M08-sistema-de-emails.md) |
| 9 | Deslocamento e pontos de serviço | [M09](milestones/M09-deslocamento-e-pontos-de-servico.md) |
| 10 | Mundo compartilhado e investigação do detetive | [M10](milestones/M10-mundo-compartilhado-e-investigacao.md) |
| 11 | Geografia, viagem e visita a locais | [M11](milestones/M11-geografia-internacional-e-viagem.md) |
| 12 | Interações investigativas em locais | [M12](milestones/M12-busca-e-apreensao.md) |
| 13 | Acusação e julgamento | [M13](milestones/M13-acusacao-e-julgamento.md) |
| 14 | Otimização, cache e estabilidade | [M14](milestones/M14-otimizacao-cache-e-estabilidade.md) |

Ciclos 1–5 são **fundação**. O ciclo 6 é o primeiro recorte jogável para playtest, não a conclusão do jogo. M07–M13 constroem o fluxo investigativo completo: registros, correio, deslocamento, organização de elementos, expansão geográfica, visita e investigação de locais, acusação e prisão do culpado. O M10 também entrega identidades visuais modulares para personagens e geração recorrente de casos compartilhados. M14 vem depois do fluxo criminal completo e consolida desempenho, cache e estabilidade da plataforma; ele não altera a porcentagem de conclusão desse fluxo. Tutorial (Vesper), auth completa, simulação temporal geral e outros sistemas continuam exigindo milestones próprios.

## Status

Ver [`estado.md`](estado.md). Não duplicar a tabela aqui.
