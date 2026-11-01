# M27 — Casos financeiros e múltiplos acusados

**Status:** Concluída
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

- [x] O modelo permite homicídios com vítima e crimes financeiros sem vítima humana ou cena obrigatória.
- [x] O gerador v7 inclui desvio de dinheiro, extorsão corporativa e fraude contábil a partir de vínculos de emprego existentes.
- [x] Casos financeiros podem identificar a empresa afetada sem publicar autores, cúmplices ou classificações internas de evidência.
- [x] Casos corporativos guardam o ID estável da empresa; extorsão financeira usa o impacto corporativo já definido pela M26.
- [x] Crime Reports apresenta casos financeiros por empresa, sem corpo ou pessoa desaparecida.
- [x] O whiteboard permite vários acusados; vítima só é necessária no fluxo de homicídio.
- [x] O backend valida papéis e evidência de cada participante separadamente.
- [x] Pessoas que não participam do crime têm chance zero de prisão, mesmo quando outro participante é culpado.
- [x] Evidências financeiras carregam vínculos individuais internos; acusar um cúmplice exige sustentação própria.
- [x] A revisão pode prender vários participantes e usa bloqueio transacional, chave idempotente por participante e estado compartilhado do caso.
- [x] O email de decisão apresenta o resultado de cada pessoa avaliada sem detalhar o peso interno das pistas.
- [x] Homicídios mantêm vítima, descoberta de corpo, evidências e fluxo de acusação já existente.
- [x] A geração financeira é determinística e versionada separadamente das versões v3–v6 já persistidas.

## Decisões necessárias antes do aceite

As decisões de modelo e fluxo desta etapa foram fixadas assim:

- A vítima é ausente apenas nos três crimes financeiros desta milestone; homicídio continua exigindo uma pessoa vítima.
- A organização afetada é uma empresa existente. No caso de fraude contábil, o destinatário da declaração é o órgão nacional de receita, referido nos documentos do caso sem inventar uma pessoa vítima.
- Cada caso financeiro gerado nesta etapa tem um autor principal e um cúmplice com relação de emprego verificável. Novos tipos poderão variar essa composição em milestones futuras.
- Uma submissão pode avaliar vários acusados; a resposta é individual. Acusações parciais deixam o caso aberto até que todos os participantes culpados estejam presos.
- A evidência recebe vínculos internos aos participantes que sustenta. Uma pessoa inocente nunca é presa por evidências ligadas a outro participante.
- Em casos financeiros, cada pessoa precisa ter ao menos uma prova apresentada que aponte especificamente para ela; a chance-base do fluxo antigo de homicídio não vale para acusações sem sustentação individual.
- Os casos financeiros usam o Union Mercantile Bank e o dossiê público em Crime Reports como fontes existentes. Novos portais públicos do governo ficam para uma etapa futura.
- Somente extorsão financeira envia o evento corporativo já reconhecido pela M26. Desvio e fraude contábil não introduzem novos efeitos de mercado.
- A migração preserva os casos atuais e versões antigas do gerador; não foi necessário limpar o banco.

## Validação de encerramento

- Testes Go dos módulos internos e do endpoint e2e de saúde passaram.
- Os testes web de acusação e whiteboard passaram; o build de produção do frontend também.
- A API de desenvolvimento respondeu com o banco conectado após as alterações. A branch Neon temporária foi criada, mas não recebeu compute: a API da Neon falhou ao provisioná-lo, então não houve validação isolada da migração nessa branch.

## Fora de escopo

- Mudar a bolsa, a pontuação/cotação das empresas ou a resposta do mercado a eventos; isso permanece na M26.
- Recompensas em Credits, fama, nível hacker ou progressão econômica por resolver casos.
- Introduzir crimes não financeiros além dos necessários para validar o modelo de vítima opcional e múltiplos participantes.
- Ranking público de acusados, estatísticas globais de criminalidade ou compartilhamento de decisões entre servidores.
- Reformular as ferramentas investigativas existentes sem necessidade direta para consultar provas dos novos casos.
