# Melhorias

Melhorias são unidades independentes e implementáveis, com escopo curto ou médio. Não têm ordem nem dependem umas das outras; dependências técnicas de funcionalidades existentes devem ser registradas no próprio documento. Trabalho que exija várias etapas coordenadas ou passe de médio deve virar milestone.

## Como propor

Cada melhoria deve explicar o problema observável, o resultado esperado, o escopo, os critérios binários de aceite e o que fica de fora. Use o [modelo](TEMPLATE.md). Mantenha uma melhoria focada em um resultado verificável, sem agrupar assuntos só por área técnica.

## Melhorias propostas

| ID | Melhoria | Origem | Escopo |
| --- | --- | --- | --- |
| I01 | Compatibilidade da aparência procedural de NPCs | M15 — identidade visual procedural | Curto/médio |
| I02 | Introdução visual após viagem confirmada pelo GATE (substituída) | M17 — animações das ações | Curto |
| I04 | Nomes procedurais próprios para empresas | M24/M25 — websites e bolsa; gerador de empresas | Médio |

## Melhorias implementadas

| ID | Melhoria | Documento |
| --- | --- | --- |
| I03 | Minimizar janelas do computador | [I03](../achieved/I03-minimizar-janelas-computador.md) |

## Triagem das milestones atuais

| Milestone | Decisão | Motivo |
| --- | --- | --- |
| M17 — Animações das ações | Dividir entregas autônomas; I02 é substituída pela M23 | Personagens e sistema de conversa permanecem em M17; viagem física do protagonista foi removida pelo novo fluxo de Proxy. |
| M16 — Crime Reports | Manter como milestone | Cruza geração de casos, desaparecimento, descoberta e persistência de corpos, jornal e novo app; a consistência ponta a ponta excede uma melhoria. |
| M15 — Identidade visual e customização | Dividir; I01 é prioritária | Compatibilidade de peças é um defeito delimitado. Expansão de catálogo e edição persistente do personagem do jogador são escopos independentes, mas maiores, a avaliar separadamente. |
| M18 — Áudio e efeitos sonoros | Manter como milestone | Exige biblioteca de assets, integração com vários estados, preferências persistentes e acessibilidade. |
| M19 — Pedidos policiais assíncronos | Já aceito; não gerar melhoria | Trabalho concluído, conforme `achieved/`. |
| M20 — Auditoria pré-beta | Manter como milestone | É uma atividade ampla de auditoria transversal, com matriz de cobertura e resolução de achados. |
| M21 — Hospital e identificação forense | Já aceito; não gerar melhoria | Trabalho concluído, conforme `achieved/`. |

Milestones M00–M14 também estão aceitas. Melhorias que surjam de problemas nelas podem ser registradas quando houver um resultado independente e concreto; não se reabre o escopo concluído apenas para decompor retrospectivamente suas entregas.
