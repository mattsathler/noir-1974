# Casos procedurais

Um caso **não** armazena uma cópia do mundo.

```text
CASE DEFINITION + CASE STATE → CURRENT CASE
```

## Case definition

Determinada por `case_seed`, `generator_version`, entidades envolvidas, relações, localização, vítima, suspeitos, culpado, método, motivo, evidências.

## Case state

Persistido conforme o progresso do jogador (status, vítima identificada, assassino identificado, motivo descoberto, arma encontrada, etc.).

O registro persistido tende a ser compacto: ids, seeds, versão, status, progresso, timestamps. O restante reconstrói a partir do mundo.

## Crime como consequência

Evitar: gerar dois NPCs e um crime aleatório.

Preferir: mundo → relações → histórico → caso. Assim as pistas (colegas, mesmo local, rotina, histórico profissional) nascem da estrutura social.

## Persistência pretendida (conceitual)

```text
case_id
world_id
case_seed
generator_version
status
victim_id
killer_id
location_id
progress / state
created_at
updated_at
```
