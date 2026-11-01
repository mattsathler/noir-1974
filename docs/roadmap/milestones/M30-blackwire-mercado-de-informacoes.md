# M30 — Blackwire e mercado de informações

**Status:** Direção definida · escopo a implementar  
**Trilha:** Progressão criminosa, lore e economia de informações  
**Depende de:** Whiteboard e post-its persistidos; Credits por personagem/servidor; M25/M26 — bolsa e flutuação empresarial

## Objetivo

Apresentar a Blackwire, uma rede clandestina criada por The Crow, como espaço para o jogador assumir um papel criminoso. O jogador poderá descobrir novos aplicativos, acessar arquivos confidenciais de lore e, principalmente, vender informações que reuniu durante a investigação.

A Blackwire deve criar uma escolha com consequência entre preservar evidências e cooperar com a resolução de casos ou vender informações por Credits. O jogador é hacker, mas suas ações não precisam servir ao Estado nem ao bem comum.

## Direção definida

- Blackwire é uma central clandestina de hackers associada a The Crow.
- A primeira funcionalidade de jogo será o mercado de informações. A distribuição de aplicativos e os arquivos confidenciais de lore pertencem à visão da plataforma, mas podem ser entregues em etapas.
- Qualquer post-it de qualquer whiteboard pode ser colocado à venda, independentemente de sua origem. O valor depende da utilidade, raridade, sensibilidade e relação da informação com casos ou empresas.
- Informações cotidianas, transferências comuns e conteúdo de pouco valor podem ser vendidas por quantias ínfimas, que não compensem o esforço como fonte de renda.
- Provas relevantes para casos devem ter valores mais atraentes. Evidências financeiras ou relacionadas a funcionários de alto escalão podem ter valor e impacto especialmente altos, em conexão com a bolsa de valores.
- A venda é definitiva: o post-it vendido torna-se inutilizável como prova e não pode ser usado para acusação ou progressão investigativa. A interface deve deixar claro o efeito antes da confirmação.
- A recompensa é Credits fictício e deve ser creditada ao personagem/servidor atual de maneira autoritativa e idempotente.
- A venda de uma informação afeta fortemente a flutuação da empresa relacionada, atuando como multiplicador sobre o change que a bolsa já receberia quando a empresa sofre consequências de um caso criminal.
- A origem e a exposição de dados devem ser apresentadas de forma diegética; impactos suficientemente graves podem chegar ao jornal.
- A relação determinística entre grupos rivais de empresas e seus movimentos opostos pertence à M31. A M30 prepara o evento de exposição de informações sem assumir que essa mecânica já existe.

## Escopo inicial

- Adicionar uma entrada Blackwire ao ecossistema do computador, com identidade visual de rede clandestina criada por The Crow.
- Listar post-its elegíveis dos whiteboards do personagem e exibir origem, contexto e preço oferecido.
- Permitir selecionar qualquer post-it, revisar o valor e confirmar a venda.
- Ao confirmar, creditar Credits uma única vez e marcar a anotação como vendida, bloqueando seu uso posterior como prova.
- Propagar a exposição para a empresa ou caso associado, quando houver vínculo, usando os sistemas existentes de mercado e notícias.
- Oferecer uma consequência de preço adequada a informações sem vínculo empresarial ou criminal, sem forçar associação artificial.
- Registrar estado e transação no servidor, isolados por personagem/servidor e protegidos contra reenvio, concorrência ou manipulação do cliente.

## Regras de valor e impacto

Os valores exatos e os multiplicadores ainda precisam de balanceamento. A escala acordada é qualitativa:

1. Informação cotidiana ou ruído: retorno muito baixo.
2. Informação útil ou ligada a um caso: retorno intermediário e relevante.
3. Prova financeira, corporativa ou envolvendo liderança de alto escalão: retorno elevado e impacto de mercado acentuado.

A exposição não deve transformar qualquer venda em choque de mercado. A magnitude deve refletir a sensibilidade e a credibilidade do material. A publicação no jornal depende da gravidade e da repercussão; detalhes e limiares serão definidos durante a implementação.

## Critérios de aceite

- [ ] O jogador encontra a Blackwire como serviço clandestino ligado a The Crow.
- [ ] Qualquer post-it do personagem pode ser colocado à venda.
- [ ] O preço diferencia ruído cotidiano, informação útil e evidência de alto impacto.
- [ ] A confirmação informa o preço e a perda permanente do uso probatório.
- [ ] Uma venda concluída não pode ser repetida para duplicar Credits.
- [ ] Um post-it vendido não pode ser apresentado como prova, usado para acusação ou reaproveitado em fluxos de caso.
- [ ] Credits e vendas são persistidos no escopo do personagem/servidor e validados pelo backend.
- [ ] A venda de informação associada a uma empresa amplia o impacto de mercado previsto para a empresa, com intensidade proporcional à gravidade.
- [ ] Exposições graves podem gerar cobertura jornalística diegética; dados triviais não causam repercussão exagerada.
- [ ] A venda continua disponível para dados sem vínculo empresarial, sem produzir flutuação artificial de uma companhia.

## Decisões pendentes

- Fórmula e faixas numéricas dos preços, multiplicadores e limites de impacto na bolsa.
- Como tratar post-its duplicados, anotações compostas e informação compartilhada entre quadros.
- Se o jogador pode cancelar uma oferta antes de confirmar; depois da confirmação, a venda é irreversível.
- Quais níveis de gravidade geram notícia, quanto tempo depois da venda e como evitar revelar imediatamente toda a consequência.
- Como o mercado sinaliza informação falsa, incompleta ou ainda não corroborada.
- Quando a loja de aplicativos e o arquivo de lore da Blackwire entram em escopo implementável.

## Fora de escopo inicial

- Relações determinísticas de rivalidade entre empresas e changes opostos, reservadas à M31.
- Compra de informações de outros jogadores ou mercado entre jogadores.
- Conversão de Credits em dinheiro real ou qualquer operação externa ao jogo.
- Tornar a venda obrigatória para concluir casos.
