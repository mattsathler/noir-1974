# ADR 0005 — Casos como consequência do mundo

- **Status:** Aceita
- **Data:** 2026-09-21
- **Supersede:** nenhum

## Contexto

Casos gerados como “dois NPCs aleatórios + crime aleatório” produzem pistas artificiais e uma sociedade que não se conecta à investigação.

## Decisão

Gerar estrutura social primeiro. Casos referenciam entidades e relações já existentes. Pistas devem poder nascer dessas relações. O caso tem seed, versão e estado próprios e **não** copia o mundo.

## Consequências

- Gerador de casos depende de relações consultáveis.
- Validação de solucionabilidade é obrigatória antes de um caso ser considerado válido (detalhe ainda aberto).
