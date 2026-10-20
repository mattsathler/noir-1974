# ADR 0001 — Mundo compartilhado por servidor, expansível sob demanda

- **Status:** Aceita
- **Data:** 2026-10-19
- **Supersede:** nenhum

## Contexto

O jogo precisa permitir que jogadores do mesmo servidor compartilhem a sociedade e investiguem ocorrências comuns. O mundo é persistente e não precisa ser gerado inteiro no início: novas regiões e conteúdo são criados quando necessários.

## Decisão

Cada servidor hospeda um mundo compartilhado por seus jogadores. Servidores diferentes têm mundos e estados independentes. Entidades, relações, mutations e casos são isolados por servidor, conceitualmente `(server_id, entity_id)`.

A base do mundo é procedural e determinística a partir da seed e da versão do gerador daquele servidor. O mundo se expande sob demanda: regiões ainda não necessárias não precisam ser geradas ou materializadas, e uma região gerada mantém identidade e conteúdo estáveis para o mesmo servidor e versão do gerador.

Casos pertencem ao servidor e são comuns aos jogadores desse servidor. Novas ocorrências podem ser geradas continuamente. O backend continua autoritativo sobre o estado compartilhado.

## Consequências

- Relações, mutations e casos são compartilhados; progresso individual de investigação pode ser separado por jogador.
- A geração sob demanda precisa manter o mesmo resultado para jogadores do mesmo servidor.
- Ações concorrentes sobre o mesmo caso ou entidade precisam de regras de resolução.
- A frequência, o limite e a distribuição dos casos precisam ser definidos; novas ocorrências não eliminam, por si só, disputa pelo mesmo caso.
- O modelo de API, atualização de clientes conectados, particionamento de regiões, armazenamento e operação em escala continuam a exigir decisões próprias.
