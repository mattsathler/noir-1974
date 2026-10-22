# ADR 0010 — Geografia fixa e bairros procedurais

- **Status:** Aceita
- **Data:** 2026-10-20
- **Supersede:** nenhum

## Contexto

Noir precisa de países e cidades reconhecíveis e comuns entre servidores, além de uma forma de distribuir detetives e ocorrências sem concentrá-los em um único lugar.

## Decisão

A geografia terá duas camadas:

1. Um catálogo fixo, igual para todos os servidores, com 25 países e quatro cidades por país (100 cidades no total). O catálogo inicial está especificado no M11; a seed do servidor não altera esses nomes nem identidades.
2. Cada cidade materializada recebe 20 bairros lógicos, gerados proceduralmente a partir de um banco de palavras específico da cidade e regras determinísticas. A lista completa é criada junto à cidade; seus nomes e IDs permanecem estáveis e compartilhados dentro do servidor.

O app GATE permite viagens somente entre cidades do catálogo. Navegar entre bairros e visitar construções são ações locais, apresentadas por interfaces próprias e fora do app GATE.

Bairros existem para distribuir detetives e ocorrências e fornecer um contexto local para a geração de crimes. Não haverá mapa físico de cada bairro, nem coordenadas, distâncias, rotas ou planejamento geográfico entre bairros. Pertencer a bairros diferentes não implica distância ou necessidade de deslocamento. A criação de bairros deve ser determinística e não depender da ordem em que forem solicitados.

## Consequências

- Todos os servidores compartilham a mesma geografia de países e cidades; seus bairros e conteúdo procedural permanecem isolados por servidor.
- O catálogo inicial de 25 países e 100 cidades, assim como os bancos de palavras por cidade, são dados versionados do jogo e precisam manter compatibilidade com mundos existentes. Novos destinos entram em versões futuras do catálogo.
- A implementação deve prevenir colisões de nomes de bairros na mesma cidade e manter IDs canônicos estáveis. Uma futura alteração na quantidade de bairros deve usar uma nova versão do gerador; a versão 5 cria 20 por cidade.
- Bairros não exigem geometria, mapa próprio, coordenadas, distâncias, rotas ou relações de vizinhança.
- A implementação atual contempla Lyon; o M11 expande o catálogo às demais cidades e habilita viagens do GATE entre cidades, inclusive entre países. Bairros e construções não são destinos do app GATE.
- O catálogo inicial de países e cidades está definido no M11. A quantidade de bairros permanece em 20 por cidade; mais bairros ou novas regras de composição exigem nova versão do gerador.
