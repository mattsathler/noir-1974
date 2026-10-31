# M26 — Flutuação da bolsa por eventos corporativos

**Status:** Proposta  
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

- [ ] Cada empresa negociável tem uma pontuação de mercado e cotação determinadas pelo servidor.
- [ ] Um evento corporativo elegível altera somente a empresa a que está explicitamente vinculado.
- [ ] Homicídios ligados a conflito no trabalho e extorsão financeira afetam a pontuação da empresa associada ao caso.
- [ ] Homicídios sem vínculo verificável com uma empresa, casos de outras categorias e acontecimentos sem vínculo corporativo não alteram pontuação de empresa.
- [ ] O impacto do mesmo evento é aplicado uma única vez, inclusive em retry, reinício ou processamento concorrente.
- [ ] A cotação atual da bolsa reflete a pontuação atual e é a cotação usada nas compras, vendas e valorização da carteira.
- [ ] A valorização/desvalorização atualiza o valor de mercado da carteira sem alterar o saldo em Credits enquanto o jogador não vender ações.
- [ ] Personagens no mesmo servidor veem cotações iguais para as mesmas empresas; servidores distintos mantêm estados isolados.
- [ ] A API não aceita do cliente a pontuação nem a cotação como valores de escrita autoritativos.
- [ ] A interface comunica pontuação, cotação atual e variação de forma compreensível, sem expor campos internos do caso.
- [ ] Histórico do mercado permite reconstruir e inspecionar as alterações aplicadas a uma empresa.
- [ ] Atualizações de mercado não concedem fama ou Credits e não alteram geração, provas, resultado de acusação ou progressão de casos.

## Decisões necessárias antes do aceite

- Quando o mercado reage: na ocorrência do evento, quando o caso é reportado ou quando uma descoberta se torna pública.
- Como determinar direção e magnitude do impacto para cada tipo de evento; se há um escore de risco/estabilidade limitado por piso e teto.
- Como a cotação se recupera ou se estabiliza após um impacto: sem recuperação automática, recuperação temporal ou por novos eventos.
- Qual é a granularidade e retenção do histórico de alterações.
- Se outros eventos corporativos além de homicídio/extorsão entram nesta etapa ou ficam para expansões futuras.
- Como impedir que a cotação revele prematuramente conteúdo reservado de um caso. A reação deve preservar a classificação de provas e a informação interna do Crime Generator.

## Fora de escopo

- Criar ou alterar a bolsa, o app desktop, a carteira ou os endpoints de negociação cobertos pela M25.
- Recompensar casos com Credits, modificar saldo por investigação, fama ou reputação.
- Alterar provas, arquétipos, geração de casos, fórmula de acusação, prisões ou conteúdo do jornal.
- Flutuação causada por compras/vendas de jogadores, especulação, dividendos, juros, eventos políticos ou notícias sem vínculo empresarial.
- Alterar empresas de outros servidores ou instituir uma única cotação global compartilhada entre todos os servidores.
