# M39 — Relógio acelerado do jogo

**Status:** Proposta  
**Trilha:** Simulação do mundo  
**Depende de:** M28 — Casos evolutivos e comunicação por email; M35 — Login e jogo no desktop

## Objetivo

Dar ao mundo um relógio próprio, mais rápido que o tempo real e com ritmo familiar a jogos offline. O avanço moderado do calendário permitirá que casos, mensagens e acontecimentos evoluam sem exigir que o jogador espere dias reais.

## Decisões de escopo

- A velocidade inicial é **2×**: um minuto real com o jogo em execução avança dois minutos no calendário do jogo. Um dia de jogo dura 12 horas reais.
- O relógio usa uma referência de início persistida e avança enquanto a sessão do jogo está ativa. Ao fechar ou suspender o jogo, o tempo simulado pausa; reabrir não aplica o tempo real transcorrido fora da sessão.
- O servidor é a autoridade do horário simulado. O cliente exibe esse horário e não pode definir nem acelerar o relógio por conta própria.
- O relógio é da instância/universo, não do personagem. Todos os jogadores de uma instância compartilhada observam o mesmo horário; uma partida singleplayer terá sua própria instância.
- A implementação deve separar horário simulado de timestamps técnicos em tempo real. Sessões, auditoria, expiração de cache e telemetria continuam usando tempo real; agendas, prazos e acontecimentos do mundo passam a usar o relógio do jogo quando explicitamente migrados.
- Dados já persistidos em tempo real não devem ser reinterpretados silenciosamente como horário de jogo. Cada sistema migrado define uma conversão ou mantém sua semântica atual até ser incluído.

## Critérios de aceite

- [ ] A API fornece horário simulado consistente para uma instância, calculado a 2× enquanto a sessão está ativa.
- [ ] Requisições simultâneas e reinícios do servidor não causam saltos, regressões nem avanço duplicado do relógio.
- [ ] O horário fica pausado quando não há sessão de jogo ativa e retoma do ponto persistido ao reabrir.
- [ ] Todos os jogadores da mesma instância recebem o mesmo horário; instâncias distintas mantêm relógios isolados.
- [ ] A interface do desktop mostra o horário simulado e atualiza sem depender de uma nova ação do jogador.
- [ ] Pelo menos um sistema temporal do mundo usa o novo relógio de ponta a ponta, com comportamento coberto por critérios determinísticos.
- [ ] Timestamps operacionais e demais sistemas ainda não migrados continuam preservando o comportamento atual.

## Fora de escopo

- Acelerar o relógio quando o jogo está fechado ou em segundo plano.
- Adicionar controles de pausa, velocidade variável ou avanço instantâneo.
- Migrar todos os prazos, agendamentos e timestamps existentes em uma única etapa; cada sistema deve ser migrado explicitamente.
- Alterar o modelo de identidade, armazenamento local ou distribuição offline. A compatibilidade com singleplayer offline deve orientar o desenho, mas pertence a milestones próprios.
