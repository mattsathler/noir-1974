# Visão

**Noir 1974** é um jogo de investigação em formato sandbox, inspirado conceitualmente em *Shadows of Doubt*.

Jogadores conectados ao mesmo servidor compartilham um mundo persistente e resolvem **casos gerados proceduralmente** nessa sociedade comum. O mundo se expande sob demanda, sem exigir que toda a realidade seja gerada no início. O foco não é seguir uma história linear nem uma sequência fixa de diálogos: é investigar uma sociedade gerada e descobrir relações, evidências e acontecimentos.

## Diretrizes de produto

- 100% web.
- Frontend: HTML + CSS + JavaScript.
- Comunicação: REST API.
- Multiplayer: cada servidor tem um mundo compartilhado por seus jogadores.
- Servidores mantêm mundos e estados independentes.
- Casos são comuns a todos os jogadores do mesmo servidor; novas ocorrências podem surgir continuamente.
- O mundo procedural se expande sob demanda, mantendo determinismo e identidade estável dentro do servidor.
- A sociedade e os casos são procedurais, acompanhados por personagens fixos que revelam a lore e estabelecem relações com o jogador.
- O banco persiste o necessário para reconstruir o mundo e o estado do jogador.
- Geração procedural é **determinística** para a mesma seed e a mesma versão do gerador.

## Identidade do jogador

No cadastro, o jogador informa o primeiro nome. O jogo passa a se referir ao personagem pela **inicial**.

Exemplo: nome `Victor` → identidade `V`.

O jogador também poderá montar a aparência do personagem com peças visuais compatíveis. A composição será salva na conta, independente da seed do mundo. NPCs terão composições próprias reconstruídas deterministicamente para que todos no mesmo servidor vejam a mesma aparência.

O ciclo de investigação é a base repetível do jogo, não seu ponto final narrativo. A direção de dualidade, progressão hacker, casos pessoais e Online está registrada em [Direção do jogo e próximas milestones](direcao-e-proximas-milestones.md); o modelo singleplayer/multiplayer ainda é uma proposta, descrita no [ADR 0014](../decisions/0014-personagens-e-instancias-singleplayer-multiplayer.md).
