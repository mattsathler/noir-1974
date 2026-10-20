# Regra 0002 — Pistas sem sinalização ao jogador

- **Status:** Obrigatória
- **Escopo:** Geração, persistência, API, interface, conteúdo e progressão de investigação

## Regra

O jogo nunca deve indicar diretamente ao jogador que uma pessoa, objeto, fala, lugar, relação ou acontecimento é uma pista de um crime ou de outro caso. A classificação de pista existe para o sistema; para o jogador, cada elemento observado deve permanecer uma possibilidade a interpretar.

O backend deve saber quais elementos são pistas, a que acontecimento se relacionam e como contribuem para a investigação. Essa classificação é informação interna de autoridade e não pode ser revelada direta ou indiretamente por respostas da API, textos, ícones, cores, animações, sons, ordenação, destaque visual ou comportamento da interface.

## Mistura de pistas e acontecimentos comuns

O mundo deve apresentar pistas junto de acontecimentos cotidianos e informações sem relevância para o crime. Conversas, rotinas, relações, objetos e eventos comuns coexistem com elementos investigativos de forma consistente com a sociedade do jogo.

Essa mistura deve ser constante ao longo da investigação. Não se deve criar uma categoria visual ou narrativa em que apenas pistas apareçam, nem fazer com que todo elemento apresentado após iniciar um caso seja automaticamente suspeito. O jogador deve formar e revisar hipóteses a partir do contexto, das conexões e das evidências que consegue interpretar.

## Separação entre sistema e jogador

- **Sistema:** armazena e consulta a classificação de pista, o acontecimento relacionado, as relações causais e os requisitos de descoberta ou validação.
- **Jogador:** recebe observações e informações no contexto do mundo, sem rótulo de pista, marcador de relevância ou confirmação automática de que algo importa para um crime.
- **Interface:** pode permitir que o jogador registre, compare ou organize por conta própria o que considera suspeito, mas não confirma a classificação interna do sistema.

Conhecimento do jogador pode progredir — por exemplo, uma identidade ou um fato pode ser descoberto — sem que o jogo revele que esse dado é uma pista. Descobrir uma informação e reconhecer sua relevância investigativa são coisas distintas.

## Requisitos de conformidade

Uma funcionalidade, resposta ou conteúdo só está em conformidade quando:

- a classificação interna de pistas permanece no backend e não é exposta ao cliente;
- o jogador não recebe rótulos, badges, notificações ou feedback que confirmem que um elemento é pista;
- apresentação, ordem e interação não tornam pistas mecanicamente distinguíveis de acontecimentos comuns;
- eventos cotidianos e informações irrelevantes continuam existindo junto às pistas, sem parecer preenchimento artificial;
- progressão ou validação de caso não confirma uma pista apenas por ela ter sido selecionada, coletada ou exibida;
- qualquer anotação ou hipótese do jogador permanece separada da verdade interna mantida pelo sistema.

## Auditoria desta regra

Ao revisar um gerador, endpoint, texto, tela ou fluxo de investigação, verificar explicitamente:

1. O backend consegue identificar a pista e o acontecimento a que ela se relaciona?
2. Algum campo, texto ou sinal visual/sonoro revela essa classificação ao cliente?
3. Um jogador pode distinguir pistas de eventos comuns apenas pelo modo como são apresentados?
4. Existem acontecimentos comuns críveis misturados às informações investigativas?
5. O fluxo confirma relevância ou correção só porque o jogador interagiu com determinado elemento?

Se a apresentação entregar a classificação da pista ou retirar do jogador a dúvida sobre sua relevância, a proposta viola esta regra e deve ser revista.
