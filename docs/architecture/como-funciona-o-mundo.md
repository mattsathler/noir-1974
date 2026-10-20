# Como o mundo do jogo funciona

Este guia explica, sem detalhes de implementação, como o jogo cria e consulta o mundo de Noir 1974.

## A ideia principal

Cada servidor tem um mundo próprio. Uma **seed** é o valor que dá origem a esse mundo. Com a mesma seed e a mesma versão do gerador, o jogo recria os mesmos lugares, pessoas, empresas e relações.

Isso significa que o mundo não depende de sorteio novo a cada acesso: os jogadores do mesmo servidor encontram a mesma realidade base.

## Como o mundo é criado

1. O servidor recebe uma seed e uma versão do gerador.
2. Um catálogo fixo comum a todos os servidores define 50 países e suas capitais e principais cidades; bairros lógicos são compostos proceduralmente a partir de um banco de palavras e criados sob demanda para distribuir detetives e ocorrências.
3. Na versão 2, também cria 24 pessoas, três empresas e os vínculos entre elas.
4. Cada pessoa trabalha em uma empresa; alguns pares também têm uma relação pessoal. Cada empresa tem um edifício como sede.
5. O resultado pode ser gravado no banco para que suas relações sejam consultadas rapidamente.

Na implementação atual, o gerador ainda produz apenas Lyon e escolhe três distritos de uma lista pequena, além de ruas e edifícios. O catálogo mundial de 50 países e cidades e a geração procedural expansível de bairros são o modelo documentado para a evolução do sistema, ainda não implementado.

## O que fica salvo e o que é reconstruído

O banco guarda a identidade das entidades e as relações que precisam ser pesquisadas, como quem trabalha em cada empresa e quem tem uma relação pessoal com quem.

Os atributos gerados de uma pessoa — nome, idade, aparência e personalidade — podem ser reconstruídos usando a seed, a versão do gerador e sua chave de geração. Assim, não é necessário salvar todos esses dados repetidamente.

Quando alguém consulta uma pessoa, o sistema:

1. encontra a seed e a chave dessa pessoa;
2. recria seus atributos determinísticos;
3. procura no banco os vínculos relacionados a ela;
4. recria os dados das outras pessoas ou empresas envolvidas;
5. devolve o perfil completo.

Para listar funcionários, o banco procura os vínculos da empresa e o sistema reconstrói somente as pessoas encontradas. Ele não precisa recriar toda a população.

## Por que a versão do gerador importa

Alterações nas regras de geração podem mudar o mundo. Por isso, cada mundo guarda a versão usada na criação. A versão 1 representa a cidade mínima; a versão 2 acrescenta a sociedade inicial. Manter as versões permite reconstruir mundos antigos sem aplicar as regras novas por engano.

## Onde isso está no projeto

- [`generator.go`](../../apps/api/internal/world/generator.go): cria a geografia e escolhe a versão.
- [`society.go`](../../apps/api/internal/world/society.go): cria pessoas, empresas e relações.
- [`store.go`](../../apps/api/internal/world/store.go): grava e consulta entidades e relações no PostgreSQL.
- [`schema.sql`](../../apps/api/internal/world/schema.sql): define as tabelas e índices usados pelo armazenamento.
- [Geração do mundo](geracao-do-mundo.md): detalhes do determinismo e das chaves de geração.

As funções de geração e consulta já existem no backend. Endpoints HTTP e telas para explorar a sociedade não fazem parte desta etapa.
