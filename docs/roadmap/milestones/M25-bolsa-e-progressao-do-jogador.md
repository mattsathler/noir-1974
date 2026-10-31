# M25 — Bolsa de valores e progressão do jogador

**Status:** Implementada
**Trilha:** Independente; não altera o cálculo do fluxo criminal  
**Depende de:** identidade e sessão atuais; catálogo nacional de empresas; desktop MeridOS

## Objetivo

Estabelecer uma base persistente de progressão econômica e fama por personagem/servidor e oferecer uma bolsa de valores acessível pelo desktop. O jogador poderá consultar empresas do país conectado, negociar ações e administrar sua carteira.

## Decisões definidas

- Cada personagem em um servidor tem sua própria pontuação de fama e saldo em Credits. Os valores não acompanham a conta para outros personagens ou servidores.
- Fama é uma pontuação global do personagem, não limitada à cidade ou ao país atual. Esta milestone não cria ranking, leaderboard nem comparação pública entre jogadores.
- Descobrir, investigar ou solucionar casos não concede Credits nesta milestone. O saldo inicial e as transações da bolsa são as fontes econômicas consideradas pelo escopo inicial.
- A bolsa lista empresas nacionais: ao mudar o país Proxy, o jogador consulta o mercado daquele país.
- A negociação de ações e a carteira pertencem ao personagem/servidor da sessão. Não há transferência de ações ou Credits entre jogadores nesta milestone.
- O acesso à bolsa é feito por um aplicativo próprio no desktop pessoal.
- A primeira versão usa cotações iniciais definidas pelo backend e estáveis. A flutuação em resposta a acontecimentos corporativos pertence à M26.
- Novos personagens começam com **1.000,00 Credits** e fama **0**. A fama fica persistida e visível; nenhuma ação aumenta a pontuação até a regra de progressão ser definida.
- As cotações iniciais são determinísticas por país e empresa, variam entre 50,00 e 1.000,00 Credits e permanecem estáveis nesta milestone.
- Ordens usam ações inteiras, preço em centavos e não cobram corretagem.
- A carteira e o saldo são armazenados pelo servidor e pela conta do personagem atual; o servidor permanece parte da chave de isolamento.

## Escopo

### Perfil de progressão

- Persistir fama e saldo em Credits no escopo do personagem/servidor e devolvê-los pela identidade autenticada.
- Mostrar os dois valores no perfil ou em uma superfície persistente do desktop e dentro do app de bolsa.
- Apresentar fama como pontuação numérica; não criar ranking, faixas de nível, recompensas automáticas ou vantagens de acesso nesta etapa.
- Creditar e debitar Credits apenas por operações explicitamente definidas nesta milestone. Nenhuma rota de descoberta, investigação, acusação ou prisão deve pagar Credits.

### Bolsa de valores

- Criar o app **STOCK EXCHANGE** no desktop pessoal.
- Listar as empresas do país conectado, sem limitar o resultado ao Proxy de uma única cidade. A listagem deve ser paginada e não carregar diretórios de funcionários ou websites completos.
- Mostrar por empresa as informações mínimas necessárias à decisão: nome, cidade, cotação atual e identificador estável.
- Permitir compra e venda de ações inteiras usando Credits do personagem/servidor atual.
- Rejeitar compras sem saldo suficiente, quantidades inválidas e vendas acima da quantidade possuída.
- Executar cada ordem no backend de forma autoritativa e atômica: débito/crédito de Credits e atualização da carteira não podem divergir em falha ou concorrência.
- Oferecer uma carteira que mostre as empresas com ações do jogador, quantidade possuída, custo médio de aquisição, cotação atual e valor de mercado da posição.
- Atualizar saldo, carteira e valores exibidos após cada negociação confirmada. Recarregar o app ou trocar de cidade não pode duplicar nem perder transações.
- Garantir que o preço usado na ordem seja o mesmo preço confirmado pelo backend, nunca um valor enviado e aceito sem validação pelo cliente.

## Critérios de aceite

- [x] Cada personagem/servidor tem saldo em Credits e pontuação de fama persistentes e isolados dos demais personagens e servidores.
- [x] O saldo inicial é 1.000 Credits, e nenhum fluxo de descoberta, investigação ou solução de caso cria Credits.
- [x] Fama aparece como pontuação global do personagem, sem ranking público nem comparação entre jogadores.
- [x] O desktop pessoal oferece o app **STOCK EXCHANGE**.
- [x] A bolsa lista empresas do país Proxy com paginação; trocar o país muda o mercado sem misturar listagens.
- [x] Preços e cotações usados nas ordens são determinados pelo backend.
- [x] Comprar ações debita o saldo e registra a quantidade na carteira em uma única operação atômica.
- [x] Vender ações credita o saldo e reduz a carteira em uma única operação atômica.
- [x] Saldo insuficiente, quantidade inválida, venda acima da posição e empresa indisponível são recusados sem alterar saldo ou carteira.
- [x] A carteira mostra quantidade, custo médio, cotação e valor atual por posição e total.
- [x] IDs de ordem impedem duplicação e as transações bloqueiam saldo e posição para impedir valores negativos em concorrência.
- [x] Fama, Credits e carteira persistem após logout/login e permanecem isolados por personagem/servidor.
- [x] As APIs validam sessão e escopo; o frontend não é autoridade para saldo, fama, preço, ordem ou propriedade das ações.

## Decisões necessárias antes do aceite

- Quais ações aumentam fama e quantos pontos cada ação concede. Até isso ser definido, a pontuação começa em zero e permanece estável.
- Se haverá reposição de Credits além das vendas de ações.

## Fora de escopo

- Recompensas em Credits por descobrir, investigar ou solucionar casos.
- Ranking, leaderboard, níveis com faixas ou benefícios mecânicos ligados à fama.
- Dividendos, venda a descoberto, empréstimos, transferências entre jogadores ou ordens condicionais.
- Simulação de mercado influenciada por notícias, casos, crimes, decisões do jogador ou eventos corporativos.
- Compra de bens, imóveis ou aplicativos pelo mercado clandestino; `black_market.exe` permanece em direção futura.
- Decidir se fama, saldo e ações migram quando um personagem muda de servidor ou quando uma instância singleplayer é compartilhada. O modelo atual desta milestone mantém tudo separado por personagem/servidor.
