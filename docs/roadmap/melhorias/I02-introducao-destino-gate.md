# I02 — Introdução visual após viagem confirmada pelo GATE

**Status:** Proposta  
**Origem:** M17 — Animações das ações do jogo
**Escopo:** Curto

## Problema

Após uma viagem pelo GATE, falta uma introdução visual breve que situe o jogador no destino recém-alcançado.

## Resultado esperado

Uma apresentação com efeito typewriter informa o destino somente depois de a API confirmar a viagem, com um estado final acessível e sem bloquear a navegação indefinidamente.

## Escopo

- Exibir o destino confirmado após uma viagem bem-sucedida pelo GATE.
- Manter o estado visual de localização alinhado à resposta autoritativa da API.
- Permitir acesso ao texto completo sem depender da velocidade do efeito typewriter.

## Critérios de aceite

- [ ] O destino só é apresentado como localização atual depois da confirmação de sucesso da API.
- [ ] Falha ou cancelamento não apresenta o destino como alcançado.
- [ ] O efeito termina em tempo definido e não impede o uso da interface.
- [ ] Texto completo e estado da viagem são acessíveis sem depender da animação.

## Fora de escopo

Cena de viagem entre bairros, deslocamento policial, conversas com NPCs, mudanças nas regras do GATE ou criação de novos assets de destino.
