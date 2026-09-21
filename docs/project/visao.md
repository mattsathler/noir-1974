# Visão

**Noir 1974** é um jogo de investigação em formato sandbox, inspirado conceitualmente em *Shadows of Doubt*.

O jogador resolve **casos gerados proceduralmente** dentro de um mundo próprio. O foco não é seguir uma história linear nem uma sequência fixa de diálogos: é investigar uma sociedade gerada e descobrir relações, evidências e acontecimentos.

## Diretrizes de produto

- 100% web.
- Frontend: HTML + CSS + JavaScript.
- Comunicação: REST API.
- Singleplayer: cada jogador tem a própria instância/mundo.
- Mundos **não** são sincronizados entre jogadores.
- NPCs e casos são procedurais, não um elenco escrito à mão.
- O banco persiste o necessário para reconstruir o mundo e o estado do jogador.
- Geração procedural é **determinística** para a mesma seed e a mesma versão do gerador.

## Identidade do jogador

No cadastro, o jogador informa o primeiro nome. O jogo passa a se referir ao personagem pela **inicial**.

Exemplo: nome `Victor` → identidade `V`.
