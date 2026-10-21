# Arquitetura desejada

A arquitetura pode ser resumida nesta frase:

> **A seed cria a realidade base compartilhada por servidor; o mundo se expande sob demanda, as relações dão estrutura à sociedade, as mutations registram o que mudou e os casos utilizam essa sociedade para criar investigações comuns aos jogadores do servidor.**

## Documentos

| Documento | Conteúdo |
| --- | --- |
| [Princípios](principios.md) | Os dez princípios consolidados |
| [Modelo híbrido](modelo-hibrido.md) | Seed, relações, mutations, o que persistir |
| [Entidades](entidades.md) | Modelo do recorte inicial e origem dos dados |
| [Schema inicial](schema-inicial.sql) | Tabelas e índices PostgreSQL de referência para M01 |
| [Geração do mundo](geracao-do-mundo.md) | Camadas, grafo, entidades |
| [Como funciona o mundo](como-funciona-o-mundo.md) | Explicação simples da seed, geração, persistência e consultas |
| [Casos](casos.md) | Definição vs estado; crime como consequência |
| [Correio eletrônico](correio-eletronico.md) | Geração de mensagens, caixas consultáveis e separação de classificação interna |
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
