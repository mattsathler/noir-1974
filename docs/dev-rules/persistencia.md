# Persistência

- Persistir o que precisa sobreviver à sessão: relações consultáveis, mutations, estado de caso, progresso do jogador, versão e seed.
- Não persistir por padrão: posição, animação, pathfinding, “NPC está andando”.
- Não materializar o mundo inteiro “por se acaso”. Materializar quando a consulta ou a mutation exigir — a regra fina ainda é ponta solta; até lá, o mínimo.
- Mutation descreve mudança sobre a base (`entidade`, `campo`/`fato`, `valor`), não um dump da entidade.
- Caso não copia o mundo: guarda ids, seed, versão e progresso.
