# M10 — Mundo compartilhado e investigação do detetive

**Ciclo:** 10  
**Status:** Em andamento  
**Depende de:** M09 aceito

## Objetivo

Dar ao jogador um espaço pessoal de investigação dentro do computador, onde possa reunir e organizar informações que encontrou em emails, registros bancários, logs do GATE e outras fontes; oferecer identidade visual modular para NPCs e personagens de jogadores; e manter novas ocorrências surgindo automaticamente no mundo compartilhado enquanto a API estiver ativa.

## Critérios de aceite

### Whiteboards pessoais

- [x] O jogador pode criar várias páginas de investigação, alternar entre elas e renomear ou excluir uma página com confirmação.
- [x] O jogador pode fixar informações descobertas em um board existente; se ainda não houver boards, um é criado automaticamente.
- [x] O jogador pode mover os registros fixados e conectar ou desconectar dois deles com uma linha visual.
- [x] Boards, registros fixados, posições e conexões são persistidos no backend, isolados por conta e servidor.
- [x] O backend só aceita registros que a conta tenha descoberto; o cliente não cria fatos nem atribui classificação investigativa.
- [x] O estado do board é atualizado na interface sem aguardar a resposta de cada gravação, com indicador de salvamento e cache local.
- [x] O whiteboard pode permanecer aberto junto a outras janelas do computador; janelas podem ser trazidas para frente e o estado aberto é restaurado por até cinco minutos.
- [x] Informações no board são apresentadas como registros organizados pelo jogador, sem revelar classificação interna como pista ou prova.

### Identidade visual modular

- [ ] Existe um conjunto modular de peças visuais combináveis, com alinhamento e escala consistentes para compor personagens em pixel art.
- [ ] O catálogo oferece variações de rostos, cabelos, tons de pele, características visuais e acessórios, sem restringir combinações por gênero ou origem.
- [ ] Cada NPC recebe uma composição visual determinística a partir da seed do servidor, da versão do gerador e de sua chave estável; todos os jogadores do servidor veem a mesma aparência para aquele NPC.
- [ ] A nova aparência procedural de NPCs usa uma nova versão do gerador mundial; versões anteriores continuam disponíveis para reconstruir mundos existentes sem alterar seus personagens.
- [ ] O jogador substitui a escolha entre seis retratos fixos por um editor modular com prévia e persiste as peças escolhidas na própria conta, independente da seed do mundo.
- [ ] A API valida as peças do jogador contra o catálogo visual permitido e retorna dados que o cliente consegue compor sem duplicar regras de compatibilidade.
- [ ] NPCs e personagem do jogador seguem literalmente a [Regra 0001](../../rules/0001-direcao-artistica.md), incluindo pixel art, paleta, proporções e influência de Scotland Yard.
- [ ] A seed e a versão de mundos existentes não são sobrescritas; a nova versão é usada para mundos novos, e a reconstrução de um mundo existente exige processo explícito.
- [ ] Testes cobrem determinismo visual de NPCs, consistência entre jogadores, validação e persistência da aparência do personagem, além da reconstrução de versões antigas do gerador.

### Geração recorrente de casos

- [x] Um agendador no backend tenta gerar casos em intervalos configuráveis enquanto o serviço está ativo, sem depender de requisições dos jogadores.
- [x] O agendador para junto com a API e recupera o cursor persistido após reinício; repetir uma criação usa o mesmo seed e a criação do caso é idempotente.
- [x] Um advisory lock transacional por servidor coordena múltiplas instâncias da API para evitar disparos simultâneos.
- [x] Cada caso criado é compartilhado por todos os jogadores do servidor, sem alterar o progresso privado deles.
- [x] A cadência, o limite de casos abertos e o comportamento ao atingir o limite são configuráveis; se não houver candidatos, o mundo ganha novos NPCs antes da próxima seleção.
- [x] A expansão populacional seleciona o bairro existente menos povoado, com desempate determinístico pelo identificador.
- [ ] Gerar novos bairros sob demanda a partir do banco de palavras, além de expandir a população dos bairros existentes.
- [ ] Testes cobrem disparo recorrente, concorrência entre instâncias, recuperação após reinício, limites e visibilidade compartilhada dos casos.

## Fora de escopo

Veredito, pontuação de provas e prisão de NPC; animações corporais, roupas com atributos de jogo, editor livre de pixels, compartilhamento de aparências e mudanças visuais de NPC causadas por mutations; simulação temporal completa dos NPCs.
