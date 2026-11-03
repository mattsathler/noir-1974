# M35 — Login e experiência de jogo dentro do desktop

**Status:** Em andamento · próxima milestone após a M33

**Trilha:** Experiência do computador e navegação do jogo  
**Depende de:** M23 — Sistema operacional, Proxy e sites investigativos

## Objetivo

Refatorar as telas inicial e de login para que o jogador entre no Noir como quem inicia e acessa um computador. Depois de entrar, todo o jogo acontece dentro do desktop simulado; os boards de investigação permanecem como a única exceção fora dele.

## Escopo

- Refazer a experiência inicial e de autenticação no fluxo boot do PC → introdução M_OS → login ou cadastro.
- Exibir um carregamento curto depois da autenticação e abrir o desktop em seguida.
- Apresentar as telas e ações jogáveis existentes como aplicativos, janelas ou páginas acessadas pelo desktop.
- Preservar os boards de investigação fora do desktop como exceção explícita.
- Revisar os caminhos de entrada, retorno, logout e falha de autenticação para que se encaixem na simulação do computador.

## Critérios de aceite

- [ ] A tela inicial e o login são apresentados como interação com o computador simulado, sem uma tela de entrada de jogo separada do desktop.
- [ ] Um login bem-sucedido abre o desktop e permite continuar a sessão de jogo por ele.
- [ ] Todas as ações e telas jogáveis são acessadas dentro do desktop, exceto os boards de investigação.
- [ ] Os boards continuam acessíveis e funcionais como a exceção prevista.
- [ ] Falhas de login, logout e retorno à tela inicial mantêm uma navegação coerente com a simulação do computador.
- [ ] A autenticação e as sessões existentes continuam funcionando, sem exigir mudança das regras de conta no backend.

## Fora de escopo

- Alterar as regras investigativas, econômicas ou de progressão do jogador.
- Remover os boards ou migrá-los para dentro da janela do desktop.
- Introduzir novos aplicativos ou funcionalidades de jogo sem relação com o novo fluxo de entrada.
