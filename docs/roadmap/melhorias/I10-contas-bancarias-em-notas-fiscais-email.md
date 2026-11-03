# I10 — Descobrir contas bancárias em notas fiscais por email

**Status:** Proposta  
**Origem:** Descoberta de contas bancárias por Ghostkey e investigação financeira  
**Escopo:** Médio

## Problema

Uma conta bancária pode ser encontrada diretamente em um site ou registro acessado pelo Ghostkey. Isso transforma a descoberta em uma consulta explícita e rápida, em vez de uma pista que o jogador encontra ao investigar a correspondência cotidiana do NPC.

## Resultado esperado

O Ghostkey deixa de expor diretamente o número da conta bancária em sites ou registros invadidos. O número passa a ser encontrado em notas fiscais anexadas a emails de compras virtuais feitas ocasionalmente pelo NPC. O jogador precisa examinar a caixa postal pessoal do NPC e identificar, entre outras mensagens, uma nota fiscal útil.

As compras e seus emails podem anteceder ou suceder a semana do acontecimento investigado. A data não deve ser alinhada automaticamente ao caso; a pista é uma informação de rotina que exige busca cuidadosa.

## Escopo

- Remover a exibição direta de números de conta bancária nas páginas e registros acessíveis pelo Ghostkey.
- Gerar emails ocasionais de confirmação de compras virtuais na caixa postal pessoal de NPCs elegíveis.
- Incluir a conta bancária do comprador na nota fiscal da mensagem, com apresentação compatível com uma cobrança ou recibo comercial.
- Distribuir as datas das compras pelo histórico de correspondência do NPC, sem exigir que ocorram na mesma semana do caso associado.
- Manter a geração determinística e a correspondência entre a nota fiscal, o NPC e a conta bancária já atribuída a ele.
- Garantir que, quando a identificação de uma conta bancária for necessária a uma pista ou objetivo investigativo, exista ao menos uma nota fiscal acessível no histórico de email do NPC correspondente.
- Preservar o uso do número encontrado no fluxo existente de consulta bancária/Bankrupt; a melhoria altera a origem investigativa do número, não a função de validação da conta.
- Misturar as notas fiscais a mensagens cotidianas para que não sejam destacadas automaticamente como evidência ou pista.

## Critérios de aceite

- [ ] Sites e registros acessíveis pelo Ghostkey não revelam diretamente o número da conta bancária do NPC.
- [ ] Emails de compras virtuais aparecem ocasionalmente no histórico pessoal de NPCs elegíveis e incluem uma nota fiscal com o número correto da conta do comprador.
- [ ] As datas dos recibos são determinísticas, variam pelo histórico e não são forçadas para a semana do caso.
- [ ] Alvos necessários a uma pista ou objetivo investigativo têm ao menos uma nota fiscal encontrável em sua caixa postal acessível.
- [ ] O número obtido no email continua identificando o NPC correspondente pelo fluxo de consulta bancária/Bankrupt; números incorretos não identificam outro alvo.
- [ ] As mensagens aparecem junto à correspondência cotidiana, sem rótulos automáticos de evidência ou destaque de pista.

## Fora de escopo

- Alterar como contas bancárias são atribuídas aos NPCs ou o formato e a validação de seus números.
- Reformular o site do Union Mercantile Bank ou o funcionamento do aplicativo Bankrupt.
- Gerar compras, recibos ou emails para todos os NPCs sem relação com uma necessidade de investigação.
- Tornar toda nota fiscal uma pista de caso ou mudar o conteúdo de emails não relacionados a compras.
