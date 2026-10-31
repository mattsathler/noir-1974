# ADR 0014 — Personagens e instâncias singleplayer/multiplayer

- **Status:** Proposta
- **Data:** 2026-10-31
- **Supersede:** nenhum

## Contexto

Noir já gera um mundo determinístico por servidor e compartilha mutations e casos entre jogadores que pertencem ao mesmo servidor. A direção futura é permitir que cada jogador mantenha personagens em mundos diferentes e escolha entre jogar sozinho ou compartilhar um mundo multiplayer.

É necessário distinguir a seed que reconstrói o mundo base da instância que define quem compartilha seu estado mutável. Compartilhar somente a seed não deve, por si só, compartilhar mutations, casos ou progresso.

## Proposta

- Tratar cada mundo singleplayer como uma instância privada, acessível por um único jogador, com seed própria e estado isolado.
- Tratar cada servidor multiplayer como uma instância compartilhada, com uma seed e um estado comum de mutations e casos para seus jogadores.
- Permitir que uma conta tenha personagens diferentes em instâncias diferentes; cada personagem pertence a uma única instância.
- Usar a identidade da instância — e não apenas a seed — para autorizar leituras e escritas de estado mutável. A seed e as versões dos geradores identificam como reconstruir a base determinística.
- Não assumir nesta proposta que uma instância singleplayer pode ser convertida em multiplayer. A conversão exigiria regras explícitas para compartilhar ou manter privado o estado já criado.

## Viabilidade

O modelo é compatível com os princípios atuais: `server_id` já delimita o mundo compartilhado e seus dados mutáveis, enquanto a seed e as versões dos geradores reconstroem a base. Uma instância privada pode usar o mesmo contrato de mundo com uma política de acesso de um só jogador.

Isso requer evolução de produto e de identidade: criação/listagem de personagens por instância, associação conta-personagem-instância, autorização de participação e seleção de instância no início da sessão. A arquitetura atual não deve ser interpretada como se essas telas, permissões ou ciclos de vida já estivessem implementados.

## Questões para fechar antes da aceitação

- O jogador pode entrar em servidores multiplayer existentes ou apenas criar servidores novos?
- Um personagem pode ser levado entre instâncias, ou cada instância exige personagem próprio?
- O estado de um mundo singleplayer pode ser compartilhado mais tarde? Em caso afirmativo, quais mutations, casos e progresso são copiados?
- O jogador pode ter mais de um personagem na mesma instância?
- Como convites, lista de servidores, descoberta, privacidade, moderação e saída de um servidor funcionam?
- O estado de personagem (inventário, dinheiro, nível e fama) pertence à conta, ao personagem ou à instância?

## Consequências esperadas se aceita

- A seed continua descrevendo a sociedade base; a instância é a fronteira de isolamento e compartilhamento do estado mutável.
- Singleplayer e multiplayer podem usar o mesmo gerador e os mesmos serviços de mundo, com políticas distintas de acesso.
- APIs e tabelas que hoje usam `server_id` como escopo compartilhado precisarão definir como representam instâncias privadas e personagens múltiplos.
- A transição entre instâncias não poderá misturar estado privado e compartilhado sem uma regra explícita.

## Alternativas consideradas

- **Uma seed por jogador em todos os modos:** não representa naturalmente a sociedade compartilhada de um servidor multiplayer.
- **Seed como única fronteira de compartilhamento:** rejeitada como premissa técnica, pois mundos com a mesma base determinística podem precisar de mutations e casos isolados.
