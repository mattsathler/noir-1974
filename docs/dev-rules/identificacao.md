# Identificação

- Unicidade de entidade: `(world_id, entity_id)`. Nunca assumir que `npc_23` é a mesma pessoa em dois mundos.
- IDs estáveis dentro do mundo: um caso, uma relação e uma mutation devem poder apontar para o mesmo `npc_23` depois de reiniciar o servidor.
- Não vazar entidades de um mundo em consulta de outro. Todo acesso passa pelo mundo do jogador autenticado (quando auth existir).
- Código e contratos em **inglês** (`world_id`, `npc_id`, `generator_version`). Documentação de produto em **português**.
