# Roadmap

O desenvolvimento avança em **ciclos**. Cada ciclo tem um **milestone de sucesso** com critérios de aceite binários. O ciclo seguinte **não começa** enquanto o anterior não for aceito.

**Onde estamos:** [`estado.md`](estado.md) — milestones aceitos e porcentagem do fluxo completo do crime, da geração à prisão do culpado. O M06 indica separadamente quando o primeiro recorte está pronto para playtest.

Milestones concluídas ficam em [`achieved/`](achieved/); milestones em andamento ou futuras ficam em [`milestones/`](milestones/).

Melhorias independentes, de escopo curto ou médio, ficam em [`melhorias/`](melhorias/). Uma melhoria não depende da conclusão de outra melhoria. Se o trabalho exigir várias etapas coordenadas ou ultrapassar escopo médio, deve ser descrito como milestone.

Milestones independentes de várias etapas também ficam em [`milestones/`](milestones/), mas em uma trilha própria: não bloqueiam o ciclo principal nem entram no cálculo do fluxo criminal. Consulte a tabela abaixo.

Prova técnica que sustenta o restante (meta do ciclo 4):

```text
seed → gerar mundo → salvar mutations → desligar → reiniciar
  → reconstruir → idêntico + mutations aplicadas
```

## Ciclos

Ordem e regras de entrada/saída: [`ciclos.md`](ciclos.md).

| Ciclo | Milestone | Documento |
| --- | --- | --- |
| 0 | Repositório e documentação | [M00](achieved/M00-repositorio-e-docs.md) |
| 1 | Modelo de entidades e schema inicial | [M01](achieved/M01-modelo-de-entidades-e-schema.md) |
| 2 | Gerador determinístico (cidade mínima) | [M02](achieved/M02-gerador-deterministico.md) |
| 3 | Sociedade: NPCs, empresas, relações | [M03](achieved/M03-sociedade-e-relacoes.md) |
| 4 | Mutations e reconstrução | [M04](achieved/M04-mutations-e-reconstrucao.md) |
| 5 | Gerador de casos simples + validação | [M05](achieved/M05-gerador-de-casos.md) |
| 6 | REST API e cliente web mínimo | [M06](achieved/M06-api-e-cliente.md) |
| 7 | Registros bancários e histórico do GATE | [M07](achieved/M07-registros-investigativos.md) |
| 8 | Sistema de correio eletrônico | [M08](achieved/M08-sistema-de-emails.md) |
| 9 | Deslocamento e pontos de serviço | [M09](achieved/M09-deslocamento-e-pontos-de-servico.md) |
| 10 | Mundo compartilhado e investigação do detetive | [M10](achieved/M10-mundo-compartilhado-e-investigacao.md) |
| 11 | Geografia, viagem e visita a locais | [M11](achieved/M11-geografia-internacional-e-viagem.md) |
| 12 | Interações investigativas em locais | [M12](achieved/M12-busca-e-apreensao.md) |
| 13 | Acusação e julgamento | [M13](achieved/M13-acusacao-e-julgamento.md) |
| 14 | Otimização, cache e estabilidade | [M14](achieved/M14-otimizacao-cache-e-estabilidade.md) |
| 15 | Identidade visual procedural e customização do personagem | [M15](achieved/M15-identidade-visual-e-customizacao.md) |
| 16 | Crime Reports | [M16](achieved/M16-crime-reports.md) |
| 17 | Animações das ações do jogo | [M17](milestones/M17-animacoes-de-acoes.md) |
| 18 | Áudio e efeitos sonoros | [M18](milestones/M18-audio-e-efeitos-sonoros.md) |
| 19 | Pedidos policiais assíncronos | [M19](achieved/M19-pedidos-assincronos.md) |
| 20 | Auditoria integral de estabilidade e consistência pré-beta | [M20](milestones/M20-auditoria-integral-pre-beta.md) |
| 21 | Hospital e identificação forense | [M21](achieved/M21-hospital-e-identificacao-forense.md) |

## Trilha independente

| Milestone | Tema | Documento |
| --- | --- | --- |
| 22 | Refatoração do bypass e navegação de arquivos com Ghostkey | [M22](milestones/M22-refatoracao-ghostkey.md) |

Ciclos 1–5 são **fundação**. O ciclo 6 é o primeiro recorte jogável para playtest, não a conclusão do jogo. M07–M13 constroem o fluxo investigativo completo: registros, correio, deslocamento, organização de elementos, expansão geográfica, visita e investigação de locais, acusação e prisão do culpado. O M10 também entrega identidades visuais modulares para personagens e geração recorrente de casos compartilhados. M14 vem depois do fluxo criminal completo e consolida desempenho, cache e estabilidade da plataforma; ele não altera a porcentagem de conclusão desse fluxo. M15 amplia o catálogo de aparência procedural e permite editar a aparência do personagem do jogador. M16 adiciona Crime Reports ao computador pessoal como início investigativo, lista desaparecidos e publica corpos encontrados no jornal. M17 usa uma cena temporária de carro em pixel art para viagens entre bairros e cria uma overlay de conversa direta com opções de resposta; também estabelece Bill Bouer e The Crow. Animações de deslocamento policial e roteiros tutoriais específicos ficam para especificações posteriores. M18 adiciona áudio às interações. M19 implementa a análise assíncrona de mandados e acusações, como extensão do fluxo de M13. M20 realiza a auditoria integral de estabilidade e consistência pré-beta. M21 abre o Hospital e acrescenta prontuários, códigos estáveis de impressão digital e vestígios forenses comparáveis no fluxo de investigação. Tutorial (Vesper), auth completa, simulação temporal geral e outros sistemas continuam exigindo milestones próprios.

## Status

Ver [`estado.md`](estado.md). Não duplicar a tabela aqui.
