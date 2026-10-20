# Modelo de entidades — recorte inicial

Este documento define a identidade e a persistência do primeiro recorte do mundo. O banco é relacional (PostgreSQL, conforme a implementação atual da API); o gerador e as regras de domínio serão especificados nos ciclos seguintes.

## Escopo e isolamento

Um `server_id` identifica uma realidade compartilhada. Todo mundo, entidade, relação, mutation e caso pertence a um servidor. Jogadores do mesmo servidor consultam os mesmos fatos compartilhados; o progresso de investigação pertence individualmente a cada conta.

A identidade conceitual é `(server_id, entity_id)`. Um `entity_id` é estável dentro do servidor, independente de quando sua região é expandida. IDs iguais em servidores diferentes não designam a mesma entidade.

## Entidades

| Entidade | Identidade e conteúdo |
| --- | --- |
| Servidor (`game_server`) | Escopo de isolamento e raiz do mundo compartilhado. |
| Mundo (`world`) | Uma seed e uma `generator_version` por servidor. |
| Região (`world_region`) | Recorte expansível sob demanda. Sua chave lógica é estável e pode apontar para uma região-pai. |
| Entidade do mundo (`world_entity`) | Âncora de identidade de um NPC, empresa ou local materializado. Guarda tipo, região e chave estável do gerador; atributos reconstruíveis não são copiados por padrão. |
| NPC | Pessoa gerada, com identidade estável e relações consultáveis. Atributos base podem ser reconstruídos da seed. |
| Empresa | Organização gerada, com relações próprias, incluindo localizações e vínculos de emprego. |
| Local | Espaço físico gerado (por exemplo, distrito, edifício ou sala). O subtipo fica em `location_kind`. |
| Mutation | Alteração persistente aplicada a uma entidade do mundo, isolada por servidor. |
| Caso (`case`) | Definição compartilhada pelo servidor, criada a partir de seed e versão próprias e referenciando entidades do mundo. |
| Evidência (`case_evidence`) | Fato associado a um caso. Sua classificação e ligação com a solução são internas ao backend. |
| Progresso do jogador | Descobertas individuais de uma conta sobre um caso compartilhado. Não altera a definição nem o estado comum do caso. |

`world_entity` é uma âncora, não uma exigência para materializar antecipadamente a população inteira. Um registro é criado quando a identidade precisa ser referenciada por uma relação, mutation, caso ou consulta persistida. A política exata de materialização ainda será refinada com o gerador.

No recorte M03, todos os NPCs gerados têm vínculos de emprego e por isso suas âncoras são persistidas junto das relações. Atributos gerados permanecem fora da tabela; consultas os reconstroem apenas para os IDs selecionados pelas relações persistidas.

## Relações do recorte

O primeiro conjunto nomeado é:

- **Emprego:** `employee_npc_id → employer_company_id`. Uma empresa pode ter muitos funcionários; as consultas pelos dois lados são indexadas.
- **Relação pessoal:** `person_a_npc_id ↔ person_b_npc_id`, com tipo como casamento, amizade ou parentesco. Relações simétricas usam os dois lados na consulta. Tipos direcionais, como parentesco, preservam a direção nos campos `person_a` e `person_b` definida para aquele tipo.
- **Localização de empresa:** `company_id → location_id`, com papel como sede ou filial.

As tabelas de relação são tipadas. Chaves estrangeiras garantem que os extremos existam no mesmo servidor; o domínio também valida que cada extremo tem o tipo correto (por exemplo, NPC em emprego e empresa no papel de empregador).

## NPC: origem e persistência

| Categoria | Campos / fatos | Regra |
| --- | --- | --- |
| Reconstruídos da seed | Nome, idade, aparência, personalidade e outros atributos base estáveis | São reproduzidos com a seed do mundo, `generator_version` e a chave estável do NPC. Não são persistidos por padrão. |
| Persistidos como relações | Emprego, relações pessoais e localização | São fatos consultáveis e compartilhados no servidor; não ficam embutidos no registro de atributos do NPC. |
| Persistidos como mutation | Estado alterado (por exemplo, morto ou detido), mudança relevante de relação ou outro fato que substitui a base gerada | A mutation aponta para entidade e campo/fato; o gerador continua definindo o estado base. |
| Persistidos como identidade | `entity_id`, tipo, região e chave estável de geração | Só materializa a âncora quando algum fato persistente precisa referenciar o NPC. |

Emprego não é um atributo escalar do NPC: é uma relação. Estado atual não deve sobrescrever silenciosamente o valor base da seed.

O schema registra mutations individualmente. No recorte M04, as mutations de status são eventos sequenciados por servidor; a leitura aplica o evento mais recente sobre o estado base gerado. Idempotência de comandos fica para quando houver uma API de escrita.

## Regiões e identidade estável

Regiões são endereçadas por uma chave lógica canônica (`region_id`), independente da ordem em que forem solicitadas. Entidades têm uma chave de geração estável dentro da região (`generation_key`); o ID final é único dentro do servidor. A seed derivada para uma região/entidade não pode depender de outras regiões geradas antes dela.

O formato exato das chaves, derivação das subseeds, geografia e limites de expansão pertencem ao M02. Neste ciclo fica definido que uma expansão não troca a identidade de conteúdo já referenciado e que a materialização é sob demanda.

## Casos: estado comum e progresso individual

`cases` contém a definição e o estado compartilhado (por exemplo, aberto ou resolvido), além das referências a vítima, culpado e local. A definição e os fatos internos são autoritativos no backend; `killer_entity_id` e classificação de evidências não são campos automaticamente expostos ao cliente.

`player_case_progress` contém descobertas e deduções de uma conta para aquele caso. Duas contas podem ter progresso distinto sem duplicar o caso ou alterar o estado compartilhado. A implementação concreta de progressão e concorrência fica para M05/M06.

## Schema de referência

O [schema inicial PostgreSQL](schema-inicial.sql) concretiza as tabelas, chaves, restrições e índices deste recorte. Ele é especificação de M01; sua aplicação e evolução em migrations serão feitas junto aos ciclos de implementação que consumirem essas tabelas.
