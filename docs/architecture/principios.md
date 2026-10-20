# Princípios arquiteturais

1. **Mundo compartilhado por servidor** — jogadores no mesmo servidor compartilham uma realidade (`server_id` + seed); servidores diferentes permanecem isolados.
2. **Geração determinística** — a mesma seed + a mesma `generator_version` reconstrói os mesmos atributos.
3. **Relações persistentes e consultáveis** — NPCs não são ilhas; casamento, emprego e parentesco precisam ser consultáveis sem gerar o mundo inteiro.
4. **Mutations** — alterações persistentes são mudanças sobre o mundo base, não um dump completo.
5. **Casos separados do mundo** — o caso referencia entidades; tem seed, versão e estado próprios.
6. **Versionamento** — mundos e casos antigos reconstruem com a versão em que foram criados.
7. **Banco como fonte de verdade** do que precisa sobreviver entre sessões.
8. **Backend autoritativo** — o cliente não decide o estado definitivo.
9. **Mundo como grafo** — entidades (nós) e relações (arestas).
10. **Casos como consequência do mundo** — crimes e pistas aproveitam a estrutura social já existente e são compartilhados pelos jogadores do servidor.

## Isolamento por servidor

O isolamento ocorre entre servidores, não entre jogadores. Alterações no servidor A não afetam B, mas todos os jogadores de A observam o estado comum. Concorrência, atualização de clientes e coordenação de geração são questões arquiteturais a resolver; o backend autoritativo continua sendo a autoridade do estado.
