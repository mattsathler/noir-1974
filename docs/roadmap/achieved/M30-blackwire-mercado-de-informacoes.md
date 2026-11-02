# M30 — Blackwire e mercado de informações

**Status:** Concluída
**Trilha:** Progressão criminosa, lore e economia de informações  
**Depende de:** Whiteboard e post-its persistidos; Credits por personagem/servidor; M25/M26 — bolsa e flutuação empresarial

## Objetivo

Apresentar a Blackwire, uma rede clandestina criada por The Crow, como espaço para o jogador assumir um papel criminoso. O jogador poderá descobrir novos aplicativos, acessar arquivos confidenciais de lore e, principalmente, vender informações que reuniu durante a investigação.

A Blackwire deve criar uma escolha com consequência entre preservar evidências e cooperar com a resolução de casos ou vender informações por Credits. O jogador é hacker, mas suas ações não precisam servir ao Estado nem ao bem comum.

## Direção definida

- Blackwire é uma central clandestina de hackers associada a The Crow.
- Blackwire é acessada pelo terminal com `run blackwire.exe`; o comando abre ou reutiliza uma janela normal do Web Browser. A página não recebe os flicks nem os efeitos de invasão do Ghostkey.
- O site da Blackwire usa uma apresentação dark própria, discreta e legível, como uma rede privada clandestina.
- A primeira funcionalidade de jogo será o mercado de informações. A distribuição de aplicativos e os arquivos confidenciais de lore pertencem à visão da plataforma, mas podem ser entregues em etapas.
- Post-its em whiteboards de investigações ainda abertas podem ser colocados à venda, independentemente de sua origem. Quadros ligados a casos encerrados ficam fora do mercado porque suas informações já são públicas. O valor depende da utilidade, raridade, sensibilidade e relação da informação com casos ou empresas.
- Os itens são agrupados por whiteboard e podem ser filtrados pelo quadro. A interface e sua API não classificam um item como rotina ou evidência/informação de caso.
- Informações cotidianas, transferências comuns e conteúdo de pouco valor podem ser vendidas por quantias ínfimas, que não compensem o esforço como fonte de renda.
- Provas relevantes para casos devem ter valores mais atraentes. Evidências financeiras ou relacionadas a funcionários de alto escalão podem ter valor e impacto especialmente altos, em conexão com a bolsa de valores.
- O valor da oferta é oculto ao jogador durante a navegação e confirmação. Após a venda, a Blackwire apenas anuncia sua conclusão; o valor recebido chega em um email de liquidação na caixa pessoal.
- A venda é definitiva: o post-it vendido torna-se inutilizável como prova e não pode ser usado para acusação ou progressão investigativa. A interface deve deixar claro o efeito antes da confirmação.
- A recompensa é Credits fictício e deve ser creditada ao personagem/servidor atual de maneira autoritativa e idempotente.
- A venda de uma informação afeta fortemente a flutuação da empresa relacionada, atuando como multiplicador sobre o change que a bolsa já receberia quando a empresa sofre consequências de um caso criminal.
- A origem e a exposição de dados devem ser apresentadas de forma diegética; impactos suficientemente graves podem chegar ao jornal.
- A relação determinística entre grupos rivais de empresas e seus movimentos opostos pertence à M31. A M30 prepara o evento de exposição de informações sem assumir que essa mecânica já existe.

## Escopo inicial

- Adicionar uma página Blackwire aberta pelo Web Browser após o comando do terminal, com identidade visual dark de rede clandestina criada por The Crow.
- Listar post-its elegíveis dos whiteboards do personagem e exibir conteúdo e quadro, sem revelar preço nem classificação de origem.
- Agrupar os itens por whiteboard e permitir filtrar por quadro, sem exibir categorias que revelem se pertencem à investigação ou à rotina.
- Permitir selecionar qualquer post-it elegível, revisar a consequência irreversível em um modal que escurece a janela do navegador e confirmar a venda sem exibir o valor.
- Ao confirmar, creditar Credits uma única vez e marcar a anotação como vendida, bloqueando seu uso posterior como prova.
- Enviar um email de confirmação idempotente com o valor recebido; a tela da Blackwire apenas anuncia que a venda foi concluída.
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

- [x] O jogador encontra a Blackwire como serviço clandestino ligado a The Crow.
- [x] Qualquer post-it de um quadro elegível pode ser colocado à venda; quadros ligados a casos encerrados são recusados.
- [x] O preço diferencia ruído cotidiano, informação útil e evidência de alto impacto.
- [x] A listagem e a confirmação ocultam o valor, e explicam a perda permanente do uso probatório.
- [x] Os itens aparecem por whiteboard com filtro, sem rótulos de rotina, caso ou origem.
- [x] Após a venda, a Blackwire apenas anuncia a conclusão e envia um email com o valor recebido.
- [x] Uma venda concluída não pode ser repetida para duplicar Credits.
- [x] Um post-it vendido não pode ser apresentado como prova, usado para acusação ou reaproveitado em fluxos de caso.
- [x] Credits e vendas são persistidos no escopo do personagem/servidor e validados pelo backend.
- [x] A venda de informação associada a uma empresa amplia o impacto de mercado previsto para a empresa, com intensidade proporcional à gravidade.
- [x] Exposições graves podem gerar cobertura jornalística diegética; dados triviais não causam repercussão exagerada.
- [x] A venda continua disponível para dados sem vínculo empresarial, sem produzir flutuação artificial de uma companhia.

## Balanceamento e continuidade

Os itens abaixo são oportunidades de balanceamento e expansão posterior; não bloqueiam o escopo entregue nesta milestone:

- Ajustar faixas numéricas de preços, multiplicadores e limites de impacto na bolsa com base no ritmo real de progressão.
- Refinar o tratamento de post-its duplicados, anotações compostas e informação compartilhada entre quadros.
- Expandir a apresentação de informações falsas, incompletas ou ainda não corroboradas.
- Implementar a loja de aplicativos e o arquivo de lore como etapas futuras da Blackwire.

## Fora de escopo inicial

- Relações determinísticas de rivalidade entre empresas e changes opostos, reservadas à M31.
- Compra de informações de outros jogadores ou mercado entre jogadores.
- Conversão de Credits em dinheiro real ou qualquer operação externa ao jogo.
- Tornar a venda obrigatória para concluir casos.
