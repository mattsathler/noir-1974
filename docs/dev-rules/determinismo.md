# Determinismo

- Mesma seed + mesma `generator_version` + mesmo código dessa versão = mesmo resultado. Sem relógio, random global, ordem de hash instável ou I/O no caminho quente do gerador.
- RNG do gerador é sempre derivado da seed (e de subseeds explícitas por entidade). Não usar `Math.random()` / `random.random()` na geração.
- Toda entidade gerada carrega, conceitualmente, `seed` (ou derivada) e a versão do gerador do mundo/caso.
- Teste obrigatório de qualquer gerador: fixture de seed → snapshot dos campos estáveis. Quebrar o snapshot na mesma versão é bug, não “melhoria”.
- Mudança de algoritmo que altera output exige **nova** `generator_version` e preservação da versão antiga.
