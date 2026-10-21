# M08 — Sistema de correio eletrônico

**Ciclo:** 8  
**Status:** Aceito
**Depende de:** M07 aceito

## Objetivo

Permitir consultar correspondências eletrônicas associadas a NPCs. O Crime Generator poderá criar mensagens relacionadas a motivações e relações — como ciúme e extorsão — junto de mensagens cotidianas sem relação com crimes.

## Critérios de aceite

- [x] NPCs possuem correspondência procedural coerente com seus vínculos, ocupações e atividades no mundo.
- [x] O Crime Generator pode criar mensagens relacionadas a um caso, incluindo conversas de ciúme e mensagens de extorsão, sem marcá-las como pista ou prova na apresentação.
- [x] Correspondências cotidianas e irrelevantes aparecem junto às mensagens relacionadas a crimes.
- [x] Consultas são limitadas às caixas de entrada, participantes e intervalos permitidos; o sistema não gera nem carrega todas as mensagens do mundo para uma consulta.
- [x] Backend mantém a relação interna entre mensagem e crime; cliente, API e interface não expõem essa classificação, conforme a [Regra 0002](../../rules/0002-pistas-sem-sinalizacao.md).
- [x] Acesso e infraestrutura de mensagens refletem o controle corporativo e a tensão entre tecnologia avançada e costumes antigos, conforme as [Regras 0001 e 0003](../../rules/README.md).
- [x] O desktop do detetive oferece um aplicativo para pesquisar caixas de correio e consultar mensagens por intervalo.

## Fora de escopo

Busca e apreensão, whiteboards, acusação, julgamento e comunicação entre jogadores.
