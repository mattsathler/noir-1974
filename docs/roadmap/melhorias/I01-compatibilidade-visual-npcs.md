# I01 — Compatibilidade da aparência procedural de NPCs

**Status:** Proposta  
**Origem:** M15 — Identidade visual procedural e customização do personagem
**Escopo:** Curto/médio

## Problema

A aparência modular de NPCs pode combinar peças sem considerar a apresentação do personagem. Por exemplo, a geração pode atribuir bigode ou barba a personagens femininos. A validação atual verifica se o ID da peça existe no catálogo, mas não sua compatibilidade com quem a usa.

## Resultado esperado

O backend gera aparências coerentes com metadados explícitos de compatibilidade das peças, mantendo determinismo e evitando regras deduzidas de nomes de arquivos.

## Escopo

- Definir e persistir (ou derivar de forma estável) a apresentação visual relevante dos NPCs sem confundi-la com orientação sexual.
- Declarar compatibilidade de barba e demais peças existentes usando metadados explícitos do catálogo.
- Aplicar compatibilidade no gerador e validar a composição no backend.
- Preservar a estabilidade de aparência e a reconstrução de mundos existentes conforme a versão do gerador.

## Critérios de aceite

- [ ] Cada peça visual aplicável declara compatibilidade explícita; a regra não depende de palavras no ID ou nome do arquivo.
- [ ] A geração procedural não atribui barba/bigode a NPCs cuja apresentação não seja compatível com as regras declaradas.
- [ ] O backend rejeita composições incompatíveis nos pontos em que recebe ou persiste aparências.
- [ ] NPCs gerados com a mesma seed e versão mantêm a mesma aparência após reconstrução.
- [ ] Há cobertura automatizada para compatibilidade, casos válidos e determinismo.
- [ ] A descrição textual da aparência corresponde às peças visuais efetivamente geradas.

## Fora de escopo

Adicionar novas peças ou ampliar o catálogo; customização posterior do personagem do jogador; inferir identidade por nome, nacionalidade ou orientação sexual; redesenhar os sprites ou alterar identidades já persistidas sem versionamento.
