# Ciclos de desenvolvimento

## O que é um ciclo

Um ciclo é um período **fechado** de trabalho com:

1. **Entrada** — milestone anterior aceito (exceto o ciclo 0).
2. **Escopo** — só o que o milestone lista. Sem “já que estamos aqui”.
3. **Saída** — todos os critérios de aceite marcados, com evidência (teste, doc ou recorte demonstrável).
4. **Registro** — se o ciclo fechou uma decisão, há ADR; se abriu uma ponta, `pontas-soltas.md` foi atualizado. Aceitar o milestone atualiza [`estado.md`](estado.md) no mesmo PR, via skill `atualizar-estado-roadmap`.

Não há duração de calendário obrigatória. O ciclo acaba quando o milestone é aceito, não quando a semana acaba.

## Regras

- Um milestone ativo por vez.
- Critério de aceite é sim/não. “Quase” não aceita.
- Trabalho exploratório (spike) cabe no ciclo se produzir uma decisão escrita ou um experimento descartável; não vira feature paralela.
- Stack ainda aberta não bloqueia M01 (modelo pode ser documento + schema). Bloqueia implementação de API em M06 até haver ADR de stack, no máximo no início desse ciclo.
- Frontend de produção não entra antes de M06. Ferramentas internas (CLI, scripts de geração) podem existir em M02–M05.

## Sequência

```text
M00 docs + monorepo
  → M01 entidades + schema
    → M02 gerador determinístico
      → M03 relações sociais
        → M04 mutations + reboot idêntico
          → M05 casos válidos
              → M06 API + web mínimo
                → M07 registros bancários + histórico do GATE
                  → M08 correio eletrônico
                    → M09 deslocamento + pontos de serviço
                      → M10 mundo compartilhado + investigação do detetive
                        → M11 geografia, viagem e visita a locais
                            → M12 interações investigativas em locais
                                → M13 acusação + julgamento (fluxo criminal completo)
                                  → M14 otimização, cache e estabilidade da plataforma
                                    → M15 animações das ações do jogo
                                      → M16 Crime Reports
                                        → M17 identidade visual procedural e customização do personagem
                                          → M18 áudio e efeitos sonoros
                                            → M19 pedidos policiais assíncronos
                                              → M20 auditoria integral de estabilidade e consistência pré-beta
```

M06 é o primeiro recorte atravessável para playtest. A porcentagem geral acompanha o fluxo criminal completo até o aceite do M13, quando uma acusação válida pode levar à prisão do culpado. M14 é uma etapa posterior de qualidade técnica e não entra nesse cálculo. M10 reúne whiteboards, identidade visual modular e geração recorrente de casos compartilhados. M11 expande o catálogo fixo para 25 países e 100 cidades e oferece a viagem do GATE por país, cidade e bairro, seguida por cards de construções. M12 implementa as telas de interior e as ações dentro desses locais: interrogatórios, buscas e apreensões e acesso a sistemas restritos. M13 depende do M12 para garantir que whiteboard e busca estejam disponíveis para acusação.

M15 é uma etapa posterior ao fluxo criminal completo e à consolidação de estabilidade; sua primeira definição cobre a apresentação do destino confirmado pelo GATE, a cena temporária de carro para viagens entre bairros, a identidade de Bill Bouer e The Crow e uma overlay de conversas diretas com NPCs. Animações de deslocamento policial e roteiros tutoriais específicos aguardam especificação posterior. M16 adiciona Crime Reports ao computador pessoal, lista desaparecidos e publica corpos encontrados no jornal da cidade. M17 aprimora a identidade visual procedural e permite que o jogador personalize novamente seu personagem. M18 acrescenta efeitos sonoros breves às interações e transições do jogo, sincronizados com seus estados confirmados e acompanhados por controles acessíveis de volume e silenciamento; trilha contínua, dublagem e áudio espacial ficam fora do ciclo.

M19 é uma extensão do M13 solicitada após o aceite do fluxo criminal: mandados e acusações aguardam análise por alguns minutos, com resultados entregues pela caixa postal. Seu escopo não altera a porcentagem do fluxo criminal completo.

M20 realiza a auditoria integral de estabilidade e consistência visual/funcional pré-beta após a entrega das milestones M14–M19.

Pular um ciclo só com ADR explicando por que o critério deixou de ser necessário.
