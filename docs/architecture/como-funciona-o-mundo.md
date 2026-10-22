# Como o mundo do jogo funciona

Este guia explica, sem detalhes de implementação, como o jogo cria e consulta o mundo de Noir 1974.

## A ideia principal

Cada servidor tem um mundo próprio. Uma **seed** é o valor que dá origem a esse mundo. Com a mesma seed e a mesma versão do gerador, o jogo recria os mesmos lugares, pessoas, empresas e relações.

Isso significa que o mundo não depende de sorteio novo a cada acesso: os jogadores do mesmo servidor encontram a mesma realidade base.

## Como o mundo é criado

1. O servidor recebe uma seed e uma versão do gerador.
2. Um catálogo fixo comum a todos os servidores define 25 países e 100 cidades; cada cidade materializada recebe 20 bairros lógicos compostos proceduralmente a partir de seu banco de palavras.
3. Nas versões 2, 3 e 4, também cria 24 pessoas, três empresas e os vínculos entre elas.
4. Cada pessoa trabalha em uma empresa; alguns pares também têm uma relação pessoal. Cada empresa tem um edifício como sede.
5. Na versão 3 e posteriores, nomes de pessoas vêm de catálogos específicos da cidade e são combinados deterministicamente; a versão 4 acrescenta IDs de camadas visuais e a versão 5 calibra esses IDs por idade e gera 20 bairros em cada cidade.

Na implementação atual, o catálogo de cidades ainda contém apenas Lyon. A versão 5 gera 20 bairros a partir do banco de palavras de Lyon, além de ruas, edifícios, NPCs e empresas. Lyon tem um catálogo versionado de 300 prenomes e 300 sobrenomes. O M11 adiciona o catálogo inicial de 25 países e 100 cidades e viabiliza viagens do GATE entre elas; cada cidade terá dados próprios para gerar seus bairros e conteúdo local.

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

Alterações nas regras de geração podem mudar o mundo. Por isso, cada mundo guarda a versão usada na criação. A versão 1 representa a cidade mínima; a versão 2 acrescenta a sociedade inicial; a versão 3 usa catálogos de nomes específicos para cada cidade; a versão 4 acrescenta IDs de identidade visual modular; e a versão 5 gera 20 bairros por cidade e calibra as aparências procedurais. Manter as versões permite reconstruir mundos antigos sem aplicar as regras novas por engano.

## Onde isso está no projeto

- [`generator.go`](../../apps/api/internal/world/generator.go): cria a geografia e escolhe a versão.
- [`society.go`](../../apps/api/internal/world/society.go): cria pessoas, empresas e relações.
- [`store.go`](../../apps/api/internal/world/store.go): grava e consulta entidades e relações no PostgreSQL.
- [`schema.sql`](../../apps/api/internal/world/schema.sql): define as tabelas e índices usados pelo armazenamento.
- [Geração do mundo](geracao-do-mundo.md): detalhes do determinismo e das chaves de geração.

As funções de geração e consulta já existem no backend. Endpoints HTTP e telas para explorar a sociedade não fazem parte desta etapa.
