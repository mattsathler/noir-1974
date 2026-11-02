# I08 — Busca por tipos de empresas no hub

**Status:** Implementada
**Origem:** Meridian Global Company Directory
**Escopo:** Curto

## Problema

O hub permitia localizar empresas por nome, endereço ou site, mas não considerava seu tipo de atividade.

## Resultado esperado

O jogador consegue encontrar empresas do tipo desejado no hub, usando os tipos procedurais já associados às empresas.

## Escopo

- Permitir buscar empresas pelo ID, nome, categoria ou descrição do tipo de atividade no Meridian Global Company Directory.
- Preservar a busca textual existente e a jurisdição atual do hub.

## Critérios de aceite

- [x] A busca por ID, nome, categoria ou descrição do tipo retorna empresas correspondentes no hub.
- [x] A busca continua respeitando o país conectado e o limite atual de resultados.
- [x] A busca por nome, endereço e site continua funcionando.

## Fora de escopo

- Criar novos tipos de empresas ou alterar a geração procedural deles.
- Alterar o conteúdo dos websites das empresas.
