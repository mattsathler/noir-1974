# M10 — Whiteboards pessoais por caso

**Ciclo:** 10  
**Status:** Não iniciado  
**Depende de:** M09 aceito

## Objetivo

Dar ao jogador um espaço de investigação por caso para guardar e organizar os elementos que escolheu acompanhar. O whiteboard é privado por jogador e persiste no backend.

## Critérios de aceite

- [ ] O jogador pode criar um whiteboard para um caso e adicionar ou remover elementos que realmente descobriu.
- [ ] A persistência relaciona jogador, caso e IDs dos elementos escolhidos; outro jogador não pode ler ou modificar esse whiteboard.
- [ ] A lista persiste após logout, refresh e acesso em outro dispositivo.
- [ ] Elementos no whiteboard continuam sendo registros ou acontecimentos, sem rótulo de pista ou prova.
- [ ] O backend verifica que os IDs pertencem a elementos acessíveis ao jogador naquele caso; o cliente não cria fatos nem atribui classificação.
- [ ] Progresso individual não altera o caso ou o whiteboard de outros jogadores do servidor.

## Fora de escopo

Veredito, pontuação de provas, prisão de NPC e recursos colaborativos entre jogadores.
