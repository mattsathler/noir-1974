# M27 — Casos financeiros e múltiplos acusados

**Status:** Proposta  
**Trilha:** Independente; expande o sistema criminal sem alterar o cálculo do fluxo já concluído  
**Depende de:** ciclo atual de casos, whiteboards e acusação; contrato de eventos corporativos da bolsa quando houver integração com mercado

## Objetivo

Expandir os casos além de homicídios com uma única vítima morta. A geração deve incluir crimes financeiros e corporativos que possam afetar uma empresa, instituição ou o governo sem exigir uma vítima individual. O sistema de investigação e acusação também deve aceitar mais de uma pessoa responsável por um crime, incluindo cúmplices que podem ser presos.

## Direção definida

- Homicídio continua sendo um tipo de caso, e mantém uma vítima humana morta.
- Casos financeiros podem não ter vítima individual. Podem identificar uma organização, empresa ou órgão afetado sem transformar essa entidade em uma pessoa vítima.
- A geração de casos deve incluir, no mínimo, desvios de dinheiro de empresas, extorsão financeira e fraude contábil envolvendo alteração de valores declarados ao governo por um funcionário.
- A extorsão financeira como crime não letal é distinta do arquétipo atual `financial_extortion`, que descreve a motivação de um homicídio e permanece compatível com os casos já existentes.
- Um caso pode ter um ou mais responsáveis: autor principal e, quando aplicável, cúmplices.
- A tela e os dados de investigação podem exibir várias pessoas acusadas/participantes. O responsável real e sua classificação permanecem protegidos pelo backend.
- Cada participante é avaliado individualmente para acusação e prisão. A participação de um cúmplice não deve ser inferida apenas da culpa de outra pessoa.

## Escopo

### Modelo e geração dos casos

- Separar o tipo de crime dos dados específicos de homicídio, para que vítima, morte, cena e causa não sejam campos obrigatórios para todo caso.
- Manter os arquétipos atuais de homicídio e adicionar casos criminais financeiros/corporativos, incluindo os três exemplos definidos nesta milestone.
- Criar definições estruturadas para empresa ou órgão afetado, valor/período envolvidos quando aplicáveis, participantes e papéis no crime.
- Gerar casos financeiros a partir de relações e atividades coerentes do mundo — por exemplo, emprego, acesso a contas, registros corporativos e declarações financeiras — sem revelar a resposta na lista pública.
- Manter determinismo, versionamento, validação de solucionabilidade e isolamento de dados internos do caso.
- Vincular casos corporativos a IDs estáveis de empresa para consultas e eventual consumo do contrato de eventos de mercado existente. A M27 não deve duplicar nem alterar as regras de flutuação definidas pela bolsa.

### Investigação e apresentação

- Permitir que Crime Reports represente casos financeiros sem inventar desaparecimento, cadáver ou pessoa vítima.
- Oferecer elementos verificáveis coerentes com o crime, como registros de banco, emails, livros contábeis, pagamentos e valores declarados ao governo. O conjunto de fontes deve reutilizar as ferramentas atuais quando possível.
- Misturar registros cotidianos e informações relacionadas ao caso, sem indicar explicitamente quais são provas ou quais participantes são culpados.
- Exibir o órgão ou empresa afetados como contexto do caso, separado da lista de pessoas participantes.
- Preservar a apresentação de vítimas, local e informações forenses para os casos de homicídio existentes.

### Acusação e prisão de múltiplos participantes

- Refatorar o whiteboard e o envio de acusação para aceitar várias pessoas marcadas como acusadas e, se necessário, distinguir autor principal e cúmplices.
- Tornar a vítima opcional para tipos de caso que não envolvam uma vítima humana; mantê-la obrigatória quando o caso exigir, como em homicídios.
- Validar no backend quais papéis e requisitos são válidos para o tipo de caso, sem confiar em IDs ou classificações fornecidas pelo cliente.
- Calcular culpabilidade e decisão de prisão por pessoa, de forma que um acusado inocente nunca seja preso e a evidência para um participante não condene automaticamente os demais.
- Permitir que mais de um responsável culpado seja preso pelo mesmo caso, preservando idempotência, concorrência segura e registros de decisão.
- Informar ao jogador o resultado de cada pessoa acusada sem expor pontuações internas ou quais elementos específicos o backend considerou válidos.

## Critérios de aceite

- [ ] O modelo permite tipos de caso com e sem vítima individual, sem fabricar NPCs mortos para crimes financeiros.
- [ ] O gerador produz casos de desvio de dinheiro de empresa, extorsão financeira e fraude contábil envolvendo declarações ao governo.
- [ ] Casos financeiros podem nomear uma organização afetada sem revelar autores, cúmplices, provas internas ou a verdade do caso.
- [ ] Casos corporativos guardam IDs estáveis da empresa/órgão e dados suficientes para consultas; qualquer integração com mercado usa o contrato da M26 sem duplicar sua lógica.
- [ ] Crime Reports apresenta casos não letais sem campos ou mensagens de cadáver, desaparecimento ou vítima humana obrigatória.
- [ ] A acusação permite selecionar múltiplos participantes e aceita ausência de vítima somente nos tipos que permitem isso.
- [ ] O backend valida individualmente participantes, papéis, evidências e requisitos por tipo de caso.
- [ ] Acusar um inocente não o prende, mesmo quando outros participantes do mesmo caso são culpados.
- [ ] A evidência ligada a um participante não condena automaticamente os demais; cúmplices são avaliados conforme seu próprio envolvimento.
- [ ] Um caso permite prender mais de um participante culpado sem duplicar prisões em retries ou operações concorrentes.
- [ ] Resultado e comunicação identificam os participantes julgados sem revelar a classificação interna dos elementos.
- [ ] Homicídios já existentes continuam exigindo vítima e preservam suas pistas, exames, relatórios e resultados.
- [ ] Casos novos são determinísticos/versionados e passam por validação de participantes existentes, papéis coerentes e evidências suficientes para solução.

## Decisões necessárias antes do aceite

- Se “vítima opcional” se aplica somente a crimes financeiros ou a todo crime não letal futuro.
- Se o órgão afetado será uma entidade investigativa do caso, uma empresa ou ambos; se o governo é instituição formal no mundo gerado.
- Definição de autor principal versus cúmplice e quais evidências podem demonstrar cada papel.
- Se o jogador pode acusar somente alguns participantes e encerrar o caso parcialmente, ou se o caso permanece aberto até decidir sobre todos.
- Se a primeira decisão pode prender vários acusados de uma vez ou se cada acusado exige submissão e julgamento independentes.
- Regras de pontuação e nível de prova individual para crimes financeiros e para múltiplos acusados, preservando a garantia de nunca prender inocentes.
- Quais registros governamentais e contábeis ficam disponíveis nas ferramentas existentes e quais exigiriam novas fontes.
- Como e em que momento casos financeiros corporativos publicam um evento elegível para a bolsa da M26, sem antecipar informação reservada.

## Fora de escopo

- Mudar a bolsa, a pontuação/cotação das empresas ou a resposta do mercado a eventos; isso permanece na M26.
- Recompensas em Credits, fama, nível hacker ou progressão econômica por resolver casos.
- Introduzir crimes não financeiros além dos necessários para validar o modelo de vítima opcional e múltiplos participantes.
- Ranking público de acusados, estatísticas globais de criminalidade ou compartilhamento de decisões entre servidores.
- Reformular as ferramentas investigativas existentes sem necessidade direta para consultar provas dos novos casos.
