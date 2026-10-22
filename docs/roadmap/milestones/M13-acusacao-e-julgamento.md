# M13 — Acusação e julgamento

**Ciclo:** 13  
**Status:** Não iniciado  
**Depende de:** M12 aceito

## Objetivo

Permitir acusar um NPC usando elementos que o jogador guardou no whiteboard. O backend cruza alvo e elementos selecionados com a verdade interna do Crime Generator e encaminha o resultado a uma instância julgadora coerente com o poder corporativo de Noir.

## Critérios de aceite

- [ ] A acusação informa um NPC e IDs de elementos existentes no whiteboard do jogador para aquele caso.
- [ ] Backend valida identidade, servidor, caso, alvo e posse/acesso aos elementos; cálculo e classificação ficam no servidor.
- [ ] Um NPC inocente nunca é preso, mesmo que sejam enviados muitos elementos.
- [ ] O culpado só é preso quando a pontuação de elementos válidos alcança o limiar definido e versionado para o julgamento.
- [ ] Acusar um inocente e acusar o culpado com evidência insuficiente produzem o mesmo resultado público; respostas não revelam quais elementos eram válidos nem se o alvo era culpado.
- [ ] Se a acusação for aceita, o NPC é preso no estado compartilhado do caso e o jogador recebe o feedback de sucesso.
- [ ] Casos de fronteira, concorrência e repetição têm resultado determinístico e testado.

## Fora de escopo

Recurso, apelação, julgamento conduzido por jogador ou exibição da classificação interna de qualquer elemento.
