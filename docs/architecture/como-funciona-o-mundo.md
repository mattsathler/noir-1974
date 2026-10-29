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
5. Na versão 3 e posteriores, nomes de pessoas vêm de catálogos específicos da cidade e são combinados deterministicamente; a versão 4 acrescenta IDs de camadas visuais, a versão 5 calibra aparências por idade e gera 20 bairros por cidade, e a versão 8 introduz perfis regionais sintéticos de pele e detalhes faciais.

O catálogo fixo do M11 contém 25 países e 100 cidades. Lyon preserva os bairros, IDs, relações e nomes já existentes. Ao viajar para outra cidade, a API deriva uma seed local a partir da seed compartilhada do servidor e do identificador canônico da cidade, materializa 20 bairros, locais, empresas e NPCs e registra a versão usada por cidade. Cada país possui listas próprias de prenomes e sobrenomes; combinações determinísticas adicionais mantêm a geração em crescimento sem esgotar a população após a primeira leva.

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

Alterações nas regras de geração podem mudar o mundo. Por isso, cada mundo guarda a versão usada na criação. As versões 1 a 4 estabelecem a cidade, a sociedade, os nomes regionais e a identidade visual modular; a versão 5 gera 20 bairros por cidade e calibra aparências por idade; as versões 6 e 7 acrescentam compatibilidade visual e bairros comerciais; a versão 8 adiciona distribuições regionais sintéticas de tons de pele e detalhes faciais. Manter as versões permite reconstruir mundos antigos sem aplicar regras novas por engano.

## Onde isso está no projeto

- [`generator.go`](../../apps/api/internal/world/generator.go): cria a geografia e escolhe a versão.
- [`society.go`](../../apps/api/internal/world/society.go): cria pessoas, empresas e relações.
- [`store.go`](../../apps/api/internal/world/store.go): grava e consulta entidades e relações no PostgreSQL.
- [`schema.sql`](../../apps/api/internal/world/schema.sql): define as tabelas e índices usados pelo armazenamento.
- [Geração do mundo](geracao-do-mundo.md): detalhes do determinismo e das chaves de geração.

As funções de geração e consulta já existem no backend. Endpoints HTTP e telas para explorar a sociedade não fazem parte desta etapa.
