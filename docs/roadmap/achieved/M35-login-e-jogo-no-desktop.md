# M35 — Login e experiência de jogo dentro do desktop

**Status:** Alcançada em 03/11/2026

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

- [x] A tela inicial e o login são apresentados como interação com o computador simulado, sem uma tela de entrada de jogo separada do desktop.
- [x] Um login bem-sucedido abre o desktop e permite continuar a sessão de jogo por ele.
- [x] Todas as ações e telas jogáveis são acessadas dentro do desktop, exceto os boards de investigação.
- [x] Os boards continuam acessíveis e funcionais como a exceção prevista.
- [x] Falhas de login, logout e retorno à tela inicial mantêm uma navegação coerente com a simulação do computador.
- [x] A autenticação e as sessões existentes continuam funcionando, sem exigir mudança das regras de conta no backend.

## Implementação entregue

- A raiz inicia o boot do computador, mostra a introdução M_OS e então abre o login.
- Login, cadastro, configuração do personagem e verificações de sessão usam a moldura de acesso do computador.
- Após o login de um personagem existente ou a configuração inicial de um novo personagem, há uma tela breve de inicialização antes do desktop.
- O jogo abre no desktop persistente Meridian OS; aplicativos e páginas são acessados por suas janelas.
- Os boards continuam em uma janela separada, com acesso direto autenticado e retorno ao board solicitado depois do login.
- Logout encerra a sessão, limpa o estado de boot daquela sessão e retorna ao início do computador.
- A milestone não alterou as regras de autenticação nem o backend.

## Validação

- `npm run build` concluído com sucesso em 03/11/2026.

## Fora de escopo

- Alterar as regras investigativas, econômicas ou de progressão do jogador.
- Remover os boards ou migrá-los para dentro da janela do desktop.
- Introduzir novos aplicativos ou funcionalidades de jogo sem relação com o novo fluxo de entrada.
