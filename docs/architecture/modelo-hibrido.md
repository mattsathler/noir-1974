# Modelo híbrido: seed, relações, mutations

A seed **não** é um banco de dados completo.

```text
SEED              → atributos e geração determinística
RELACIONAMENTOS   → conexões entre entidades (consultáveis)
MUTATIONS         → alterações e eventos persistentes
```

## Seed do mundo

Cada jogador tem uma world seed. Mesma seed + mesmo algoritmo/versão → mesmos elementos determinísticos.

A seed do NPC pode definir nome, idade, aparência, personalidade, ocupação e outros atributos reconstruíveis. Não é obrigatório persistir esses atributos se o gerador os reproduzir.

## O que a seed não deve fazer

Relações entre entidades precisam de representação consultável.

- “NPC #23 = Adam” e “NPC #45 = Gina” podem vir da seed.
- “Adam é casado com Gina” é uma relação.
- “Derek trabalha na Empresa #13” é uma relação.

Consultar a Empresa #13 não deve exigir gerar todos os NPCs do mundo. Consulta-se `employment` (ou equivalente) e depois reconstrói-se só os `npc_id` retornados.

## Identificação

IDs são estáveis **dentro** do mundo. Unicidade conceitual: `(world_id, entity_id)`.

NPC #1 no World A pode ser Thomas Miller; no World B, Eleanor Hayes.

Um caso referencia `victim = npc_104`, `killer = npc_23`, `location = company_13`. O resto reconstrói.

## Mutations

```text
WORLD SEED + BASE WORLD + MUTATIONS → CURRENT WORLD
```

Exemplo: NPC morto só naquele mundo (`field: status`, `value: dead`).

## O que gerar vs o que persistir

**Gerado:** pode existir só na matemática da seed até o jogador (ou o sistema) precisar da entidade.

**Persistido:** morte, prisão, mudança importante de relacionamento, identidade descoberta, evidência obtida, caso resolvido, estado de caso, decisões relevantes.

**Não persistir de forma contínua:** posição momentânea, animação, caminhada, simulação que não precisa sobreviver à sessão.

## Camada de fatos (conceitual)

```text
married(23, 45)
employed_by(23, 13)
friends(23, 98)
```

O jogo pergunta: quem é casado com NPC #23? quem trabalha na Empresa #13? quem mora no prédio #81?

## Escalabilidade (conceitual)

Não se armazena a cidade inteira × todos os NPCs × todos os detalhes por jogador.

```text
1 jogador → 1 world_seed + relações + mutations + casos
```

Volume de casos (ex.: 2.000 jogadores × 20 casos abertos) é administrável num relacional bem indexado se os registros forem compactos. O risco maior é geração cara, consulta ruim e reconstrução repetida — não o armazenamento bruto.
