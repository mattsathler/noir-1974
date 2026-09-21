# Backend

- O gerador e a persistência vivem na API (`apps/api`), não no browser.
- Endpoints de escrita validam: mundo do jogador, existência da entidade, se a mutation é permitida.
- Não expor ao cliente campos de autoridade (ex.: `killer_id`) só porque o caso já os tem no banco. O que o jogador ainda não descobriu não viaja na resposta.
- Consultas de relação são queries (ou índice equivalente), não “gerar todos os NPCs e filtrar”.
- Linguagem e framework ainda abertos; quando existirem, esta regra ganha um ADR de stack.
