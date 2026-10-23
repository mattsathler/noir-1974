# M13 — Acusação e julgamento

**Ciclo:** 13  
**Status:** Em curso
**Depende de:** M12 aceito

## Objetivo

Permitir que o jogador entregue uma acusação na Delegacia escolhendo um whiteboard preparado. O whiteboard é somente um espaço de construção; a decisão fica no backend, que deduz o caso pela vítima marcada, compara acusado e provas com a verdade interna do Crime Generator e encaminha o resultado a uma instância julgadora coerente com o poder corporativo de Noir.

## Critérios de aceite

- [x] O jogador entrega a acusação na Delegacia escolhendo um whiteboard; o próprio quadro não submete acusações. O frontend envia apenas o quadro escolhido, sem IDs ou seletor de caso; o backend deduz o caso aberto pela vítima marcada.
- [x] Na Delegacia, um botão acima de POLICE TERMINAL abre o formulário e permite escolher o quadro. O envio fica bloqueado e avisa quando faltam ao menos uma prova, uma pessoa acusada ou uma vítima; o backend também valida a Delegacia, o quadro e esses requisitos.
- [x] O whiteboard permite marcar pins como PROVA, ACUSADO ou VÍTIMA; acusado e vítima são cada um um único pin de pessoa por quadro, e pode haver várias provas.
- [x] Backend valida identidade, servidor, cidade, caso, NPCs e posse/acesso ao quadro e aos pins; cálculo e classificação ficam no servidor.
- [x] Um NPC inocente nunca é preso, mesmo que sejam enviados muitos elementos.
- [x] Marcar itens que não são prova reduz a chance de prisão; provas verdadeiras elevam drasticamente a chance. A regra versionada de julgamento não permite prender um inocente.
- [x] Acusar um inocente e acusar o culpado com evidência insuficiente produzem a mesma resposta pública; emails de falha não revelam quais elementos eram válidos nem se o alvo era culpado.
- [x] O resultado da acusação chega ao jogador por email; se houver prisão, a mutation do NPC é atualizada e a notícia é publicada no jornal da cidade.
- [ ] Casos de fronteira, concorrência e repetição têm resultado determinístico e testado.

## Ideias para depois

- Mandados de busca devem ser pedidos assíncronos: o jogador recebe o resultado alguns minutos depois, sem inventário imediato.
- Acusações também devem ser pedidos assíncronos, com o resultado chegando alguns minutos depois por email.

## Fora de escopo

Recurso, apelação, julgamento conduzido por jogador ou exibição da classificação interna de qualquer elemento.
