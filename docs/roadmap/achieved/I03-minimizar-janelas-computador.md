# I03 — Minimizar janelas do computador

**Status:** Implementada  
**Origem:** Observação de uso da interface do computador pessoal  
**Escopo:** Curto/médio

## Problema

As janelas abertas no computador pessoal podem ser fechadas, mas não minimizadas. Para liberar espaço de trabalho e voltar a um aplicativo depois, o jogador precisa fechá-lo e abri-lo novamente.

## Resultado esperado

Cada janela oferece um controle para minimizá-la. O aplicativo permanece aberto e acessível pela barra de tarefas, que permite restaurar a janela sem perder seu estado ou posição.

## Escopo

- Adicionar um controle de minimizar na barra de título das janelas de aplicativos do computador.
- Representar janelas minimizadas na barra de tarefas e permitir restaurá-las por teclado e ponteiro.
- Preservar conteúdo, posição e ordem de foco ao minimizar e restaurar.
- Manter a minimização distinta de fechar: fechar continua encerrando a janela conforme o comportamento atual.
- Persistir o estado minimizado junto ao estado de janelas já salvo na sessão, respeitando a política atual de restauração.

## Critérios de aceite

- [x] Cada janela de aplicativo apresenta botão de minimizar com nome acessível e foco visível.
- [x] Minimizar oculta a janela da área de trabalho, sem encerrar o aplicativo nem descartar conteúdo, rascunhos ou estado local.
- [x] A barra de tarefas indica quais aplicativos estão minimizados e restaura a janela correspondente ao ser acionada.
- [x] Restaurar mantém posição e conteúdo e traz a janela restaurada para frente.
- [x] Minimizar uma janela não minimiza as demais nem interfere em diálogos modais independentes.
- [x] Fechar uma janela minimizada remove sua entrada da barra de tarefas e mantém a semântica atual de fechamento.
- [x] O fluxo funciona com teclado, ponteiro e nos viewports suportados, sem controles inacessíveis ou sobrepostos.
- [x] O estado minimizado é restaurado de forma consistente após recarga na mesma sessão, de acordo com a persistência de janelas existente.

Implementada no computador pessoal. O estado minimizado integra a persistência de janelas da sessão e a janela permanece montada enquanto oculta, mantendo o estado local dos aplicativos.

## Fora de escopo

Redimensionamento de janelas, minimizar todas, organização automática, persistência entre sessões distintas, mudanças na navegação dos aplicativos ou no conteúdo de cada aplicativo.
