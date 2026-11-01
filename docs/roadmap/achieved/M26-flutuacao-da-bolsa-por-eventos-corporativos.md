# M26 — Flutuação da bolsa por eventos corporativos

**Status:** Implementada
**Trilha:** Independente; não altera o cálculo do fluxo criminal
**Depende de:** M25 — bolsa de valores e progressão do jogador

## Objetivo

Fazer a bolsa reagir a acontecimentos associados a empresas do mundo. Cada evento corporativo relevante deve alterar a pontuação de mercado da empresa correspondente e, por consequência, a cotação usada pela bolsa.

Esta milestone trata somente do estado de mercado das empresas, da associação entre acontecimentos e empresas e da flutuação de pontuação/cotação. A bolsa, a carteira e as operações de compra e venda são entregues pela M25.

## Direção definida

- A pontuação de mercado é específica à empresa afetada; um evento não altera indiscriminadamente as cotações de outras empresas.
- Homicídios ligados a empresas, especialmente arquétipos de extorsão financeira e conflito corporativo, devem afetar a empresa relacionada ao caso.
- A pontuação de mercado é compartilhada pelos personagens da mesma instância/servidor e isolada entre servidores, acompanhando a separação atual dos mundos e casos.
- Mudança de cotação altera o valor atual das ações na carteira, mas não movimenta Credits até o jogador negociar as ações.
- O backend é a autoridade para associar eventos, atualizar pontuação e determinar novas cotações.

## Escopo

- Definir e persistir ou reconstruir uma pontuação de mercado por empresa e servidor.
- Vincular acontecimentos corporativos a um ID estável da empresa; usar vínculos explícitos do caso, como sede da cena ou relação financeira de extorsão, em vez de inferir empresa por texto ou cidade.
- Aplicar mudanças determinísticas e idempotentes à pontuação da empresa quando um acontecimento qualificável for processado.
- Incluir pelo menos eventos de homicídio associados a conflito no trabalho e extorsão financeira. Eventos sem relação verificável com uma empresa não alteram cotações.
- Converter a pontuação de mercado em cotação para a bolsa já existente e atualizar carteira/valor de mercado apresentado ao jogador.
- Garantir que duas instâncias do mesmo evento, retries ou processamento concorrente não apliquem o mesmo impacto mais de uma vez.
- Manter histórico suficiente de alterações para que o estado atual possa ser explicado e auditado no mercado.
- Preservar isolamento por servidor: personagens do mesmo servidor observam a mesma pontuação e cotação de uma empresa; servidores diferentes podem divergir conforme seus eventos.

## Critérios de aceite

- [x] Cada empresa sem eventos usa o índice-base 1.000; empresas afetadas persistem pontuação específica por servidor e cotação calculada pelo backend.
- [x] Um evento corporativo elegível altera somente a empresa identificada pelo vínculo de emprego persistido no caso.
- [x] Homicídios ligados a conflito no trabalho e extorsão financeira afetam a pontuação da empresa associada ao caso.
- [x] Homicídios sem vínculo verificável com uma empresa, casos de outras categorias e acontecimentos sem vínculo corporativo não alteram pontuação de empresa.
- [x] A gravação do caso, seu evento de mercado e a alteração da pontuação ocorrem na mesma transação; a chave do caso garante idempotência em retry e concorrência.
- [x] A cotação da listagem, da carteira e das ordens usa a pontuação atual do servidor.
- [x] Mudanças de cotação recalculam o valor de mercado da carteira sem alterar Credits até uma negociação.
- [x] A pontuação é compartilhada dentro do servidor e isolada entre servidores.
- [x] O cliente não envia pontuação nem cotação para gravação; o backend resolve ambos.
- [x] Bolsa e site do banco apresentam índice, variação e cotação. O histórico público usa rótulos genéricos e omite IDs e campos internos do caso.
- [x] Eventos são persistidos com pontuações antes/depois e podem ser consultados pelo endpoint de histórico da empresa.
- [x] Impactos não concedem fama ou Credits e não alteram provas, acusação, prisão ou progressão de casos.

## Decisões aplicadas nesta implementação

- O mercado reage quando um caso novo entra no mundo compartilhado. A alteração é atômica à criação do caso.
- Conflito no trabalho reduz 25 pontos (2,5% do índice neutro); extorsão financeira reduz 75 pontos (7,5%). O índice fica entre 500 e 1.500.
- Após eventos negativos, o índice recupera 1 ponto por cada 24 horas do relógio do servidor, até o neutro 1.000. Empresas acima do neutro não sofrem ajuste automático.
- A recuperação é calculada ao consultar ou negociar, sem criar registros artificiais no histórico. Um novo evento parte do índice recuperado naquele momento.
- O histórico é retido sem expiração e armazena índice anterior, índice posterior, magnitude e horários de ocorrência/registro.
- Apenas `workplace_conflict` e `financial_extortion` participam desta etapa.
- A interface do mercado não apresenta tipo de arquétipo nem ID do caso; a linha pública de histórico usa “Company incident”.

## Fora de escopo

- Criar ou alterar a bolsa, o app desktop, a carteira ou os endpoints de negociação cobertos pela M25.
- Recompensar casos com Credits, modificar saldo por investigação, fama ou reputação.
- Alterar provas, arquétipos, geração de casos, fórmula de acusação, prisões ou conteúdo do jornal.
- Flutuação causada por compras/vendas de jogadores, especulação, dividendos, juros, eventos políticos ou notícias sem vínculo empresarial.
- Alterar empresas de outros servidores ou instituir uma única cotação global compartilhada entre todos os servidores.
