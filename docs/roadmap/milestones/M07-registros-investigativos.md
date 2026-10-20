# M07 — Registros bancários e histórico do GATE

**Ciclo:** 7  
**Status:** Em andamento  
**Depende de:** M06 aceito

## Objetivo

Permitir consultar registros procedurais associados a NPCs: movimentações bancárias e locais visitados pela rede GATE. Os registros servem à vida comum e também podem conter fatos criados pelo Crime Generator para um caso.

## Critérios de aceite

- [ ] Consultas retornam registros de um NPC e intervalo definidos, sem gerar ou carregar todos os registros da cidade.
- [ ] Movimentações e visitas GATE são determinísticas e continuam estáveis para a mesma seed e versão do gerador.
- [ ] Crime Generator pode relacionar transferências e deslocamentos a um crime, preservando essa classificação apenas no backend.
- [ ] Ao menos um caso gerado contém uma transferência ligada a extorsão e um deslocamento investigável, misturados a registros cotidianos.
- [ ] Registros cotidianos sem relação com crimes aparecem junto aos demais.
- [ ] API, ordenação, rótulos e apresentação não indicam quais registros têm relação com um caso, conforme a [Regra 0002](../../rules/0002-pistas-sem-sinalizacao.md).
- [ ] Consultas e acessos respeitam o domínio corporativo sobre bancos e infraestrutura GATE, conforme a [Regra 0003](../../rules/0003-universo-corporativo.md).

## Fora de escopo

Busca domiciliar, armazenamento em whiteboard, acusação, julgamento e simulação temporal geral da cidade.
