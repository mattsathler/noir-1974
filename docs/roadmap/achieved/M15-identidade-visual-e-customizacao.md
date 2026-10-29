# M15 — Identidade visual procedural e customização do personagem

**Ciclo:** 15
**Status:** Concluída
**Depende de:** M14 aceito

> A compatibilidade da aparência procedural dos NPCs foi separada para [I01](../melhorias/I01-compatibilidade-visual-npcs.md), originada nesta milestone antes da troca de numeração.

## Objetivo

Ampliar o catálogo de faces, cabelos e cosméticos e permitir que o jogador volte a editar a aparência do próprio personagem depois da criação inicial.

## Critérios de aceite

- [x] A seleção de tons de pele e traços visuais parte de distribuições regionais sintéticas documentadas, sem inferir raça por nome, nacionalidade ou orientação sexual; os sete tons aparecem em proporções próximas em cada perfil.
- [x] Orientação sexual não determina cabelo, barba, maquiagem ou outros traços. O jogador escolhe diretamente a aparência do personagem.
- [x] O gerador leva idade em conta para cabelos grisalhos, calvície, rarefação e rugas, com regras documentadas, validação e testes de limites e probabilidades.
- [x] A aparência procedural continua determinística por chave e versão; versões 4 a 7 mantêm sua composição, enquanto a versão 8 acrescenta as novas regras.
- [x] O catálogo contém 55 variantes de cabelo e sete tipos de face, além de sete tons de pele, sete cores de olhos, 25 barbas e 15 acessórios.
- [x] Os novos detalhes faciais e todo o catálogo passam por validação automatizada de presença, tamanho 40×40, transparência, alinhamento e conteúdo visível.
- [x] O jogador pode abrir o editor de aparência depois da criação inicial, visualizar as alterações antes de salvar, cancelar sem alterar a aparência vigente e salvar uma nova composição na conta.
- [x] A API autoriza a alteração somente do personagem da conta autenticada, valida IDs e limites do catálogo e grava a composição em uma atualização atômica; testes cobrem falhas e sucesso.
- [x] A aparência persistida é retornada pela API da conta e exibida pelos componentes de identidade do jogador, sem alterar NPCs.
- [x] A interface oferece navegação por teclado, prévia, cancelamento, estados de salvamento/erro e layout responsivo; as opções disponíveis são válidas para a identidade configurável do jogador.

## Fora de escopo

Alterar atributos, personalidade, relações ou orientação sexual dos NPCs; alterar regras de jogo; substituir a direção artística; reconstruir identidades históricas já persistidas sem migração/versionamento explícito; criar editor de corpo ou roupas além das camadas visuais previstas.

## Entregas

- Implementado endpoint autenticado para atualizar a composição visual do personagem da própria conta, com validação do catálogo e gravação atômica da aparência.
- Implementado editor acessível durante o jogo, com prévia, cancelamento e salvamento; a identidade recarregada vem da aparência persistida.
- As regras etárias do gerador estão cobertas por testes de limites e distribuição: NPCs têm de 18 a 70 anos; grisalhos podem surgir a partir dos 40 (5%), ficam mais frequentes a partir dos 50 (20%) e comuns a partir dos 60 (40%); calvície e rarefação entram no catálogo a partir dos 45; rugas finas surgem a partir dos 45 e profundas a partir dos 60.
- A versão 8 usa perfis regionais sintéticos para tons de pele e detalhes faciais determinísticos. As versões anteriores preservam sua reconstrução.
- Verificações executadas: testes Go de `world` e `identity`, suíte web completa (22 arquivos, 72 testes), build de produção e `validate:character-assets`.
