# Princípios arquiteturais

1. **Singleplayer por mundo** — cada jogador tem uma realidade independente (`world_id` + seed).
2. **Geração determinística** — a mesma seed + a mesma `generator_version` reconstrói os mesmos atributos.
3. **Relações persistentes e consultáveis** — NPCs não são ilhas; casamento, emprego e parentesco precisam ser consultáveis sem gerar o mundo inteiro.
4. **Mutations** — alterações persistentes são mudanças sobre o mundo base, não um dump completo.
5. **Casos separados do mundo** — o caso referencia entidades; tem seed, versão e estado próprios.
6. **Versionamento** — mundos e casos antigos reconstruem com a versão em que foram criados.
7. **Banco como fonte de verdade** do que precisa sobreviver entre sessões.
8. **Backend autoritativo** — o cliente não decide o estado definitivo.
9. **Mundo como grafo** — entidades (nós) e relações (arestas).
10. **Casos como consequência do mundo** — crimes e pistas aproveitam a estrutura social já existente.

## Por que não um mundo global compartilhado

Mundo compartilhado traria concorrência, sincronização, conflitos entre jogadores, estado global, tempo real, locks, WebSockets e persistência muito mais complexa.

No modelo atual, alterar o World A não afeta B, C ou D. Isso torna a reconstrução determinística previsível.
