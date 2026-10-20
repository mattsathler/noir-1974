# ADR 0010 — Geografia fixa e bairros procedurais

- **Status:** Aceita
- **Data:** 2026-10-20
- **Supersede:** nenhum

## Contexto

Noir precisa de países e cidades reconhecíveis e comuns entre servidores, além de uma forma de distribuir detetives e ocorrências sem concentrá-los em um único lugar.

## Decisão

A geografia terá duas camadas:

1. Um catálogo fixo, igual para todos os servidores, com 50 países selecionados de diferentes partes do mundo e, para cada país, sua capital e principais cidades. A seed do servidor não altera esses nomes nem identidades.
2. Bairros lógicos gerados proceduralmente sob demanda dentro das cidades. O gerador compõe seus nomes a partir de um banco de palavras e regras determinísticas; cada bairro criado mantém identidade estável e compartilhada dentro do servidor.

Bairros existem para distribuir detetives e ocorrências e fornecer um contexto local para a geração de crimes. Não haverá mapa físico de cada bairro, nem coordenadas, distâncias, rotas ou planejamento geográfico entre bairros. Pertencer a bairros diferentes não implica distância ou necessidade de deslocamento. A criação de bairros deve ser determinística e não depender da ordem em que forem solicitados.

## Consequências

- Todos os servidores compartilham a mesma geografia de países e cidades; seus bairros e conteúdo procedural permanecem isolados por servidor.
- O catálogo de 50 países/cidades e o banco de palavras são dados versionados do jogo e precisam manter compatibilidade com mundos existentes.
- A implementação deve prevenir colisões de nomes de bairros na mesma cidade e manter IDs canônicos estáveis.
- Bairros não exigem geometria, mapa próprio, coordenadas, distâncias, rotas ou relações de vizinhança.
- A implementação atual, limitada a Lyon e catálogos pequenos, ainda precisa evoluir para cumprir este modelo.
- A seleção dos países, a lista de cidades, as regras lexicais, os critérios de distribuição e os limites operacionais de expansão permanecem detalhados em aberto.
