# Arquitetura desejada

A arquitetura pode ser resumida nesta frase:

> **A seed cria a realidade base; as relações dão estrutura à sociedade; as mutations registram o que mudou; e os casos utilizam essa sociedade para criar investigações coerentes.**

## Documentos

| Documento | Conteúdo |
| --- | --- |
| [Princípios](principios.md) | Os dez princípios consolidados |
| [Modelo híbrido](modelo-hibrido.md) | Seed, relações, mutations, o que persistir |
| [Geração do mundo](geracao-do-mundo.md) | Camadas, grafo, entidades |
| [Casos](casos.md) | Definição vs estado; crime como consequência |
| [API conceitual](api.md) | REST, servidor autoritativo |
| [Pontas soltas](pontas-soltas.md) | O que ainda não está fechado |

## Diagrama

```text
                         JOGADOR
                            │
                            ▼
                       WEB CLIENT
                            │
                         REST API
                            │
              ┌─────────────┴─────────────┐
              │                           │
              ▼                           ▼
         WORLD SERVICE               CASE SERVICE
              │                           │
       ┌──────┴──────┐             ┌──────┴──────┐
       ▼             ▼             ▼             ▼
   GENERATOR    RELATIONSHIPS   CASE SEED    CASE STATE
       └──────┬──────┘             └──────┬──────┘
              ▼                           ▼
          BASE WORLD                 CASE DEFINITION
              └─────────────┬─────────────┘
                            ▼
                       MUTATIONS
                            │
                            ▼
                      CURRENT WORLD
```

## Reconstrução

```text
                 WORLD SEED
                     │
                     ▼
            PROCEDURAL GENERATOR
                     │
                     ▼
                 BASE WORLD
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
   RELATIONSHIPS            MUTATIONS
          └──────────┬──────────┘
                     ▼
                CURRENT WORLD
```
