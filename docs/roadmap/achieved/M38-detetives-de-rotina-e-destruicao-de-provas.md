# M38 — Detetives de rotina e destruição de provas

**Status:** Concluída
**Trilha:** Investigação cotidiana  
**Depende de:** M12 — Interações investigativas em locais; M13 — Acusação e julgamento; M16 — Crime Reports; M27 — Casos financeiros e múltiplos acusados; M28 — Casos evolutivos e comunicação por email; M30 — Blackwire: mercado de informações; M35 — Login e jogo no desktop

## Objetivo

Fazer a polícia investigar casos em paralelo com o jogador. Detetives procedurais da cidade recolhem provas ao longo de uma semana e podem encerrar um caso se o material reunido permitir prender todos os responsáveis. O jogador também pode destruir permanentemente provas ligadas a um caso, afetando a investigação de todos. A venda de uma anotação pela Blackwire continua sendo uma ação pessoal e não destrói a fonte original.

## Decisões de escopo

- Cada caso aberto recebe um detetive procedural do departamento da cidade. O agente é estável para aquele caso, mas não é inserido no diretório público de cidadãos, não participa de relações sociais e não pode ser interrogado.
- O prazo é de sete dias corridos desde a criação do caso. O detetive tenta reunir até três itens diferentes durante esse prazo, em intervalos determinísticos derivados do caso. Não há barra, pontuação ou aviso de progresso policial para o jogador.
- A polícia avalia apenas os itens que seu detetive recolheu. A lista vem das fontes investigativas persistidas e normalmente acessíveis no jogo: documentos de caso, registros bancários e GATE e mensagens vinculadas ao caso. Anexos privados da caixa de entrada de um jogador não são copiados automaticamente para o dossiê policial.
- No fim do prazo, a avaliação usa as regras atuais de relevância, suporte a participantes e força probatória. A polícia só encerra o caso se a evidência não destruída sustentar a prisão de todos os responsáveis; em homicídios, também precisa identificar a vítima. A revisão automática é uma decisão única. Se faltar prova, o caso permanece aberto para o jogador e não recebe encerramento automático posterior.
- Uma conclusão automática altera o estado compartilhado do caso e prende os participantes culpados. O estado registra `police` como origem da conclusão; prisões conduzidas por jogadores registram `player`. Crime Reports identifica a conclusão policial como trabalho da polícia da cidade, sem revelar progresso ou culpados antes do encerramento.
- A destruição é oferecida somente para itens vinculados a um caso e provenientes de email, registros bancários, registros GATE ou objetos de buscas e apreensões. Uma confirmação informa que a ação é permanente e afeta todos os investigadores.
- A destruição é global: o item deixa de aparecer nas fontes do jogo, não pode ser descoberto novamente, não conta em boards ou acusações e não pode ser recolhido ou pontuado pela polícia. O backend mantém apenas um tombstone com IDs de origem, autor e horário; o conteúdo destruído não é preservado como prova consultável.
- Vender uma anotação na Blackwire continua removendo-a apenas do board do vendedor e aplicando os efeitos de mercado existentes. A venda não destrói o registro original, não apaga cópias de outros jogadores e não impede a polícia de usá-lo.

## Critérios de aceite

- [x] Todo caso aberto recebe um agente policial estável e uma data de revisão sete dias após sua criação; o agendamento e a execução são idempotentes após reinício da API.
- [x] O agente recolhe no máximo três itens únicos durante o prazo, escolhidos de forma determinística entre as fontes investigativas ligadas ao caso e ainda disponíveis.
- [x] A revisão policial considera somente os itens recolhidos pelo agente e ignora todo item com tombstone de destruição.
- [x] Se a evidência atingir os critérios atuais para todos os participantes culpados, a polícia registra as prisões, muda o caso para `solved` e salva a origem `police`; se não atingir, o caso continua `open`.
- [x] Uma revisão insuficiente não encerra o caso, não cria uma acusação contra inocentes e não impede uma acusação posterior do jogador.
- [x] Crime Reports identifica casos concluídos pela polícia. Casos concluídos pelo jogador mantêm a atribuição `player`; nenhuma das duas rotas publica a classificação interna das pistas.
- [x] O jogador consegue iniciar a destruição de um item elegível em sua fonte e precisa confirmá-la em um modal que explica o efeito global e permanente.
- [x] Depois da confirmação, o item desaparece das consultas, caixas de entrada, inventários de busca e boards aplicáveis a todos os jogadores, e não pode sustentar revisão policial ou acusação.
- [x] A venda de uma anotação pela Blackwire mantém o comportamento atual e não cria tombstone nem remove o registro das fontes investigativas de outros jogadores ou da polícia.
- [x] A interface não apresenta barra ou porcentagem de investigação policial; casos resolvidos comunicam apenas o resultado oficial pelo Crime Reports.

## Fora de escopo

- Alterar os critérios de acusação manual, a pontuação de provas ou os papéis de participantes definidos em M13 e M27.
- Dar ao jogador controle sobre o detetive, apresentar seu perfil no diretório ou criar interrogatórios para ele.
- Criar uma simulação geral de turnos, relógio acelerado, patrulhas ou movimentação policial pela cidade.
- Encerrar casos sem solução por falta de provas, exibir a destruição ao público ou permitir recuperar um item destruído.
- Alterar preço, exposição, reputação ou qualquer outro efeito existente da venda na Blackwire.
- Destruir documentos públicos sem vínculo com caso, emails de spam ou itens que não sejam oferecidos como prova.
