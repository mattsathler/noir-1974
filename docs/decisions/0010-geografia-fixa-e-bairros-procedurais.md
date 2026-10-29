# ADR 0010 — Geografia fixa e bairros procedurais

- **Status:** Supersedida
- **Data:** 2026-10-20
- **Supersede:** [ADR 0012 — Computador pessoal, proxy e investigação por sites](0012-computador-proxy-e-sites-investigativos.md)

## Contexto

Noir precisa de países e cidades reconhecíveis e comuns entre servidores, além de uma forma de distribuir detetives e ocorrências sem concentrá-los em um único lugar.

## Decisão

A geografia terá duas camadas:

1. Um catálogo fixo, igual para todos os servidores, com 25 países e quatro cidades por país (100 cidades no total). O catálogo inicial está especificado no M11; a seed do servidor não altera esses nomes nem identidades.
2. Cada cidade materializada recebe 20 bairros lógicos, gerados proceduralmente a partir de um banco de palavras específico da cidade e regras determinísticas. A lista completa é criada junto à cidade; seus nomes e IDs permanecem estáveis e compartilhados dentro do servidor.

O app GATE conduz a viagem em três seleções ordenadas: país, cidade e bairro. A confirmação viaja à cidade e ao bairro escolhidos. Construções não são destinos do GATE: após a chegada, cards mostram todas as construções do bairro, e selecioná-las abre uma tela de interior.

A localização persistida do jogador identifica somente sua cidade atual. O bairro escolhido na viagem e a construção aberta são contextos transitórios da interface: não são gravados no perfil nem alteram a cidade persistida. A viagem envia cidade e bairro canônicos para validação; ações futuras em prédios enviam a construção como contexto e o backend valida que ela pertence à cidade atual, sem tratar a seleção visual como presença persistente.

Bairros existem para distribuir detetives e ocorrências e fornecer um contexto local para a geração de crimes. Não haverá mapa físico de cada bairro, nem coordenadas, distâncias, rotas ou planejamento geográfico entre bairros. Pertencer a bairros diferentes não implica distância ou necessidade de deslocamento. A criação de bairros deve ser determinística e não depender da ordem em que forem solicitados.

## Consequências

- Todos os servidores compartilham a mesma geografia de países e cidades; seus bairros e conteúdo procedural permanecem isolados por servidor.
- O catálogo inicial de 25 países e 100 cidades, assim como os bancos de palavras por cidade, são dados versionados do jogo e precisam manter compatibilidade com mundos existentes. Novos destinos entram em versões futuras do catálogo.
- A implementação deve prevenir colisões de nomes de bairros na mesma cidade e manter IDs canônicos estáveis. Uma futura alteração na quantidade de bairros deve usar uma nova versão do gerador; a versão 5 cria 20 por cidade.
- Bairros não exigem geometria, mapa próprio, coordenadas, distâncias, rotas ou relações de vizinhança.
- Todas as cidades possuem em Central os mesmos quatro prédios principais: Police Station, Bank, GATE Central e Hospital. O Hospital fica fechado e reservado para uma feature futura.
- A interface do GATE apresenta a hierarquia país → cidade → bairro; a confirmação persiste somente a cidade. Depois da chegada, os cards apresentam as construções do bairro selecionado.
- Construções não são destinos do GATE. Abrir sua tela de interior não persiste uma visita ou localização atual.
- A implementação atual contempla Lyon; o M11 expande o catálogo e habilita viagens do GATE para bairros de cidades do catálogo, inclusive entre países. Construções não são destinos do app GATE.
- O catálogo inicial de países e cidades está definido no M11. A quantidade de bairros permanece em 20 por cidade; mais bairros ou novas regras de composição exigem nova versão do gerador.
