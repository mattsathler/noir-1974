# M34 — Efeitos sonoros e visuais

**Status:** Em andamento · primeira camada implementada  
**Trilha:** Independente; identidade sensorial do computador e das interfaces  
**Depende de:** M23 — computador pessoal e interfaces de sistema

## Objetivo

Dar às interações do NOIR um retorno tátil e sonoro inspirado em computadores e jogos 8-bit, sem competir com a atmosfera sóbria do universo. O computador deve responder a cliques, digitação e notificações com sons curtos; o cursor deve usar a linguagem pixelada da biblioteca já adotada pelo projeto.

## Direção definida

- Os efeitos usam ondas quadradas sintetizadas localmente pela Web Audio API, sem arquivos ou chamadas externas.
- Cliques e teclas recebem sons curtos e discretos; notificações têm uma assinatura melódica breve.
- O volume padrão é baixo e os efeitos podem ser desativados nas configurações do computador.
- Os cursores usam formas da Pixelarticons: cursor padrão, mão de interação e cursor de texto.
- Animações visuais são breves, em passos definidos, e respeitam a preferência por movimento reduzido.

## Critérios de aceite

- [x] Cliques primários em controles do jogo reproduzem um som 8-bit curto.
- [x] Digitação em campos de texto reproduz um som por tecla, sem emitir sons para atalhos com Ctrl, Alt ou Meta.
- [x] Uma notificação reproduz uma assinatura sonora curta, distinta dos cliques e teclas.
- [x] Sons começam em resposta a interação do usuário, falham silenciosamente quando o navegador não oferece áudio e podem ser desativados nas configurações do computador; a preferência persiste localmente.
- [x] Cursor padrão, cursor de ação e cursor de texto usam ícones pixelados derivados da biblioteca Pixelarticons.
- [x] O snackbar de notificação recebe uma entrada visual breve em passos, desativada quando o sistema pede movimento reduzido.
- [ ] Os efeitos permanecem discretos e não atrapalham leitura, acessibilidade ou uso em telas sensíveis ao toque.
- [ ] Uma revisão visual e sonora valida os efeitos em diferentes navegadores e volumes de sistema.

## Fora de escopo

Trilha musical, vozes, sons ambientais contínuos, efeitos de combate, mudança da direção visual geral do NOIR e uso de serviços externos para reproduzir áudio.
