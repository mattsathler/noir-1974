# M32 — Alter ego hacker e progressão ilegal

**Status:** Aceita
**Trilha:** Progressão ilegal, identidade e economia paralela  
**Depende de:** M30 — Blackwire e mercado de informações; progressão legal e Credits existentes

## Objetivo

Separar completamente a progressão legal do personagem da progressão criminosa de seu alter ego hacker. O jogador pode atuar como detetive e cooperar com o Estado ou desenvolver uma carreira clandestina na Blackwire, sem misturar saldos, níveis ou desbloqueios entre as duas identidades.

## Direção definida

- A progressão legal continua vinculada ao personagem e usa fama e Credits do jogo.
- A progressão ilegal pertence ao alter ego hacker e à sua conta Blackwire.
- A conta Blackwire é a origem da progressão clandestina: nível ou reputação hacker, Crypto e desbloqueios de aplicativos e áreas restritas.
- Vendas de post-its e demais atividades do mercado Blackwire concedem Crypto, nunca Credits.
- Crypto e Credits são saldos distintos. A venda de informações não credita a economia legal diretamente.
- Crypto poderá ser lavada e convertida para Credits por processos demorados, sujeitos a regras e riscos definidos posteriormente.
- Aplicativos criminosos, acesso a áreas restritas e outros recursos ilegais devem ser desbloqueados pela progressão Blackwire, sem depender da progressão legal.
- A separação será estrutural, incluindo persistência e escopo de identidade, e não apenas uma distinção visual entre dois saldos.
- A identidade hacker poderá sofrer sanções independentes do personagem. Milestones posteriores podem fazer a exposição excessiva resultar em perda da conta Blackwire e do progresso criminoso, exigindo começar com outra conta.
- A punição da conta hacker não apaga automaticamente a identidade, a progressão ou os Credits legais do personagem.
- Cada personagem tem uma conta Blackwire por servidor; a conta é provisionada no primeiro acesso e recebe um handle gerado pelo servidor.
- Ghostkey e Mailbreaker são ferramentas iniciais. Bankrupt é desbloqueado ao alcançar o nível 2 da Blackwire.
- Reputação aumenta com vendas concluídas, conforme a natureza dos dados vendidos. O nível é derivado da reputação; rotina concede pouco, informação de caso e dados corporativos concedem mais.
- A conta Blackwire guarda Crypto, reputação e nível em uma tabela própria vinculada ao mesmo escopo personagem/servidor. Apagar ou reiniciar essa conta não altera a economia legal.
- A lavagem de Crypto e novos downloads além do Bankrupt permanecem para etapas posteriores; até lá não existe conversão de Crypto em Credits.
- O primeiro acesso ao desktop cria a conta do alter ego e mostra handle e saldo Crypto no widget de identidade; o site Blackwire também mostra nível, reputação e ferramentas autorizadas.
- Vendas anteriores registradas pela Blackwire recriam reputação; somente o valor ainda disponível, até o total dessas vendas, passa de Credits para Crypto. O total movido é descontado do saldo legal na mesma migração, sem permitir saldo negativo.

## Critérios de aceite

- [x] O personagem possui estado e progressão legal separados da conta Blackwire.
- [x] A conta Blackwire possui Crypto e progressão ilegal próprias.
- [x] Vendas na Blackwire creditam somente Crypto.
- [x] Crypto e Credits não podem ser usados como se fossem o mesmo saldo.
- [x] Recursos e aplicativos ilegais verificam a progressão Blackwire.
- [x] Não há conversão automática de Crypto em Credits; a lavagem dependerá de um fluxo explícito e demorado em uma etapa futura.
- [x] A perda futura de uma conta Blackwire pode remover a progressão ilegal sem apagar o personagem legal.

## Decisões registradas

- Há uma conta hacker por personagem/servidor; ela é criada automaticamente no primeiro acesso à Blackwire e recebe um handle determinístico. Recuperação de conta não se aplica nesta etapa.
- Vendas de informação são a primeira atividade que concede Crypto e reputação. Uma venda por item continua garantida pela unicidade do registro de venda.
- Reputação mede a progressão. Os limiares iniciais são 50, 150, 400 e 900 pontos para os níveis 2 a 5.
- Ghostkey e Mailbreaker são ferramentas iniciais; Bankrupt exige nível 2 tanto na interface quanto na API.
- Uma conta reiniciada começa com saldo, reputação e nível zerados; o handle pode ser gerado novamente. A identidade legal e Credits são preservados.
- Vendas antigas reconstituem a reputação. A parcela ainda disponível em Credits, limitada ao total de vendas antigas, migra para Crypto uma única vez; o total movido é deduzido de Credits.
- Não há conversão de Crypto para Credits nesta milestone. A lavagem exige uma milestone futura para definir demora, canal, taxa e risco.

## Fora de escopo

- Unificar Credits e Crypto ou permitir conversão direta entre eles.
- Implementar a lavagem de Crypto antes de definir sua operação e seus riscos.
- Alterar a progressão legal, os casos ou as regras de acusação.
- Definir nesta milestone os mecanismos de exposição e investigação de hackers previstos para M33.
