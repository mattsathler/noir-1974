# Backend

- O gerador e a persistência vivem na API (`apps/api`), não no browser.
- Endpoints de escrita validam: servidor do jogador, existência da entidade, se a mutation é permitida e conflitos com alterações concorrentes.
- Não expor ao cliente campos de autoridade (ex.: `killer_id`) só porque o caso já os tem no banco. O que o jogador ainda não descobriu não viaja na resposta.
- A classificação interna de pistas e sua relação com crimes/acontecimentos ficam no backend; não expor metadados ou sinais que permitam ao cliente distinguir pistas de eventos comuns. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md).
- Consultas de relação são queries (ou índice equivalente), não “gerar todos os NPCs e filtrar”.
- Linguagem e framework ainda abertos; quando existirem, esta regra ganha um ADR de stack.
