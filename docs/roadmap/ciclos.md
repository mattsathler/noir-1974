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
                      → M10 whiteboards pessoais + identidade visual de personagens
                        → M11 busca e apreensão
                          → M12 acusação + julgamento (fluxo criminal completo)
```

M06 é o primeiro recorte atravessável para playtest. A porcentagem geral acompanha o fluxo criminal completo até o aceite do M12, quando uma acusação válida pode levar à prisão do culpado. M10 entrega o whiteboard pessoal e a identidade visual modular antes da busca e apreensão do M11; M09 é pré-requisito da busca, cuja solicitação é feita presencialmente na Delegacia. M12 depende do M11, garantindo que whiteboard e busca estejam disponíveis para acusação.

Pular um ciclo só com ADR explicando por que o critério deixou de ser necessário.
