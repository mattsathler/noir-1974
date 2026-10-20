# M03 — Sociedade: NPCs, empresas e relações

**Ciclo:** 3  
**Status:** Aceito
**Depende de:** M02 aceito

## Objetivo

O mundo mínimo deixa de ser só lugares: há NPCs, empresas e relações consultáveis sem gerar a população inteira a cada pergunta.

## Critérios de aceite

- [x] NPCs gerados deterministicamente (nome e atributos do recorte M01).
- [x] Empresas geradas e ligadas a locais, se o modelo M01 exigir.
- [x] Relação de emprego persistida/indexada: “funcionários da empresa X” sem varrer todos os NPCs.
- [x] Pelo menos uma relação pessoal persistida (ex. casamento) consultável nos dois sentidos.
- [x] Consultar um NPC devolve atributos gerados + relações, alinhado ao modelo.
- [x] Dois mundos com seeds diferentes não compartilham as mesmas pessoas.

## Fora de escopo

Casos, mutations de morte/prisão, cliente web.
