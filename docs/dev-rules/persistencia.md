# Persistência

- Persistir o que precisa sobreviver à sessão: seed e versão por servidor, regiões materializadas quando necessário, relações consultáveis, mutations, estado compartilhado de caso e progresso individual do jogador.
- Não persistir por padrão: posição, animação, pathfinding, “NPC está andando”.
- Não materializar o mundo inteiro “por se acaso”. Materializar quando a consulta ou a mutation exigir — a regra fina ainda é ponta solta; até lá, o mínimo.
- Mutation descreve mudança sobre a base (`entidade`, `campo`/`fato`, `valor`), não um dump da entidade.
- Caso compartilhado não copia o mundo: guarda ids, seed, versão e estado comum; progresso individual fica associado ao jogador.
- Não gerar nem materializar o mundo inteiro antecipadamente; expandir sob demanda com resultado estável para todos os jogadores do servidor.
