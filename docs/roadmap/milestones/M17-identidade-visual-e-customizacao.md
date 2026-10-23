# M17 — Identidade visual procedural e customização do personagem

**Ciclo:** 17
**Status:** Não iniciado
**Depende de:** M16 aceito

## Objetivo

Ampliar e revisar o sistema de aparência modular para que NPCs tenham identidades visuais diversas, coerentes com seus atributos e estáveis no mundo compartilhado. Expandir o catálogo de faces, cabelos e cosméticos e permitir que o jogador volte a editar a aparência do próprio personagem depois da criação inicial.

## Critérios de aceite

- [ ] O catálogo visual classifica cada peça com metadados explícitos de compatibilidade: apresentação masculina, feminina, unissex ou combinações permitidas, faixas etárias quando pertinentes, camadas incompatíveis e tags descritivas. A compatibilidade é validada no servidor e aplicada pelo gerador e pela interface.
- [ ] A geração procedural considera idade e apresentação/gênero do personagem ao escolher peças, evitando combinações incompatíveis segundo os metadados (por exemplo, barba em personagem sem compatibilidade com barba), sem depender de filtros frágeis pelo nome do arquivo.
- [ ] A seleção de tons de pele e traços visuais parte de dados explícitos de identidade/ancestralidade ou de distribuição regional documentada, sem inferir raça por nome, nacionalidade ou orientação sexual; os resultados permanecem variados e não estereotipados.
- [ ] Orientação sexual não determina cabelo, barba, maquiagem ou outros traços como atalho visual. Quando houver preferência estética expressa pelo personagem/jogador, ela pode orientar a composição sem reduzir identidades a estereótipos.
- [ ] O gerador leva idade em conta de forma consistente para cabelos, grisalhos, calvície e cosméticos relacionados, com regras explícitas, documentadas e cobertas por validação.
- [ ] A aparência procedural continua determinística por chave e versão do gerador; NPCs existentes mantêm compatibilidade de reconstrução e todos os jogadores do mesmo servidor veem a mesma composição.
- [ ] O catálogo adiciona variedade substancial de tipos de face, cabelos e outros cosméticos, incluindo opções diversas de textura, comprimento, formato, cor e características faciais, respeitando a direção artística e a matriz de camadas.
- [ ] Novas peças têm sprites e variantes coerentes de alinhamento, escala, transparência, paleta e contraste; a composição não apresenta recortes, sobreposições indevidas ou incompatibilidades de camada nas combinações permitidas.
- [ ] O jogador pode abrir o editor de aparência depois da criação inicial, visualizar as alterações antes de salvar, cancelar sem alterar a aparência vigente e salvar uma nova composição na conta.
- [ ] A API autoriza a alteração somente do personagem da conta autenticada, valida IDs, compatibilidades, limites e camadas, e retorna a aparência persistida atualizada; falhas não deixam a composição parcialmente salva.
- [ ] A aparência atualizada é refletida em todas as telas e componentes que exibem o jogador após salvar, inclusive após recarga e em novas sessões, sem modificar identidades dos NPCs.
- [ ] A interface de customização é utilizável por teclado, apresenta prévia e opções indisponíveis com clareza, e mantém layout consistente nos viewports suportados.

## Fora de escopo

Alterar atributos, personalidade, relações ou orientação sexual dos NPCs; alterar regras de jogo; substituir a direção artística; reconstruir identidades históricas já persistidas sem migração/versionamento explícito; criar editor de corpo ou roupas além das camadas visuais previstas.
