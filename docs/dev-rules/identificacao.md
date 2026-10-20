# Identificação

- Unicidade conceitual de entidade: `(server_id, entity_id)`. Nunca assumir que `npc_23` é a mesma pessoa em servidores diferentes.
- IDs estáveis dentro do servidor: um caso, uma relação e uma mutation devem apontar para a mesma entidade após reiniciar ou expandir o mundo.
- Jogadores do mesmo servidor consultam o mesmo conjunto de entidades; não misturar dados entre servidores.
- Todo acesso passa pelo servidor selecionado/autorizado do jogador (quando auth existir).
- Código e contratos em **inglês** (`server_id`, `npc_id`, `generator_version`). Documentação de produto em **português**.
