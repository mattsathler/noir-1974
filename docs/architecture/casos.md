# Casos procedurais

Um caso **não** armazena uma cópia do mundo. Ele pertence ao servidor e é compartilhado pelos jogadores daquele servidor; o progresso individual de investigação pode ser separado.

```text
CASE DEFINITION + CASE STATE → CURRENT CASE
```

## Case definition

Determinada por `case_seed`, `generator_version`, entidades envolvidas, relações, localização, vítima, suspeitos, culpado, método, motivo, evidências. O caso é associado ao servidor, não ao jogador que o consultou.

## Estado do caso e progresso individual

O caso tem estado comum aos jogadores do servidor (por exemplo, aberto, resolvido ou arquivado). A definição compartilhada guarda ids, seeds, versão, estado comum e timestamps; o restante reconstrói a partir do mundo.

Descobertas e deduções (vítima identificada, assassino identificado, motivo descoberto, arma encontrada) são progresso individual, persistido por jogador. Um jogador não altera o progresso dos demais ao investigar. Ver também o [modelo de entidades](entidades.md).

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
