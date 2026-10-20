# Casos procedurais

Um caso **não** armazena uma cópia do mundo. Ele pertence ao servidor e é compartilhado pelos jogadores daquele servidor; o progresso individual de investigação pode ser separado.

```text
CASE DEFINITION + CASE STATE → CURRENT CASE
```

## Case definition

Determinada por `case_seed`, `generator_version`, entidades envolvidas, relações, localização, vítima, suspeitos, culpado, método, motivo, evidências. O caso é associado ao servidor, não ao jogador que o consultou.

## Case state

Persistido conforme o progresso do jogador (status, vítima identificada, assassino identificado, motivo descoberto, arma encontrada, etc.).

O registro compartilhado tende a ser compacto: ids, seeds, versão, status comum e timestamps. O progresso que pertence a cada jogador é persistido separadamente. O restante reconstrói a partir do mundo.

A definição do caso e o backend mantêm a classificação interna das pistas e suas relações com os acontecimentos. Essa informação não é exposta como tal ao jogador: a apresentação deve misturar pistas com acontecimentos cotidianos e deixar sua relevância como parte da dedução. Ver [Regra 0002](../rules/0002-pistas-sem-sinalizacao.md).

A geração de casos deve refletir o domínio das grandes empresas sobre a sociedade: motivações, oportunidades, relações e consequências dos crimes pertencem a esse contexto, conforme a [Regra 0003](../rules/0003-universo-corporativo.md).

## Crime como consequência

Evitar: gerar dois NPCs e um crime aleatório.

Preferir: mundo → relações → histórico → caso. Assim as pistas (colegas, mesmo local, rotina, histórico profissional) nascem da estrutura social.

## Persistência pretendida (conceitual)

```text
case_id
server_id
case_seed
generator_version
status
victim_id
killer_id
location_id
progress / state
created_at
updated_at
```
