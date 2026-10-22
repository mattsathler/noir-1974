# Geração do mundo

## Ordem em camadas

A ordem exata ainda pode ser refinada. O mundo é gerado por regiões sob demanda; dentro de cada recorte, o princípio não muda: **estrutura coerente primeiro; casos depois**. A geração precisa ser determinística por servidor, região e versão.

O modelo inicial de entidades, identidade das regiões e chaves estáveis está em [entidades.md](entidades.md). A derivação efetiva das chaves e subseeds será definida no M02.

```text
WORLD SEED
    ↓
1. Geografia
    ↓
2. Edifícios / locais
    ↓
3. Empresas / instituições
    ↓
4. NPCs
    ↓
5. Famílias e relacionamentos
    ↓
6. Empregos e estruturas sociais
    ↓
7. Histórico / eventos
    ↓
8. Casos
```

## Geografia de referência e expansão dos bairros

A geografia terá duas camadas. A camada de referência é fixa e igual em todos os servidores: o catálogo inicial de 25 países e 100 cidades, registrado no M11. Esses nomes e identidades não variam com a seed do servidor. O app GATE apresenta países, cidades e bairros nessa ordem; o jogador escolhe cidade e bairro de destino antes de viajar. Construções não são destinos GATE: após a chegada, cards mostram as construções do bairro e uma seleção abre a tela de interior. A localização persistida do jogador identifica somente sua cidade; o bairro de chegada e a construção aberta são contextos transitórios e não são gravados como localização atual.

Dentro das cidades, os bairros serão partições lógicas para distribuir detetives e ocorrências, evitando a concentração de jogadores em um único lugar. Cada cidade recebe 20 bairros compostos deterministicamente a partir de seu banco de palavras ao ser materializada. Seus nomes e IDs são únicos dentro da cidade e estáveis para todos os jogadores daquele servidor. Ao adicionar cidades, cada uma receberá seu próprio banco de palavras e as mesmas regras de composição; aumentar a quantidade de bairros será uma mudança versionada do gerador.

Não haverá mapa físico individual dos bairros nem planejamento de coordenadas, distâncias, rotas ou vizinhança entre eles. O bairro identifica o contexto lógico de uma ocorrência e ajuda o sistema a distribuir jogadores e casos; não determina deslocamento ou proximidade física. A geração evita nomes duplicados dentro da cidade e preserva chaves canônicas em todas as consultas.

## Grafo

Entidades (nós) possíveis: NPC, empresa, residência, prédio, família, local, organização, caso, evidência.

Relações (arestas) possíveis: casado com, irmão de, pai/mãe de, amigo de, rival de, trabalha em, gerente de, mora em, proprietário de, conhece, esteve em, etc.

## Versão 1 — cidade mínima

O gerador atual da API fica em `apps/api/internal/world`. A versão `1` produz Lyon, uma região de cidade, três distritos selecionados de um catálogo fixo e, em cada distrito, duas ruas e dois edifícios. São quatro regiões e doze entidades `location` por mundo. Nomes de distritos, ruas e edifícios vêm de catálogos pequenos; a seleção varia com a seed. Esse recorte descreve a implementação atual, ainda menor que a geografia de referência e a expansão de bairros definidas acima.

Na versão 1, regiões representam cidade e distrito; ruas e edifícios são entidades de localização ligadas à região do distrito. Não há coordenadas, malha viária, NPCs, empresas, relações sociais, casos ou mutations. O recorte é uma estrutura consultável em memória, sem endpoints nem persistência nesta milestone.

## Versão 2 — sociedade inicial

A versão `2` preserva o contrato de geografia, mas habilita uma nova saída versionada com 24 NPCs distribuídos pelos três distritos e uma empresa por distrito. Cada pessoa recebe nome, idade, aparência e traço de personalidade determinísticos. O emprego é uma relação (com cargo) entre cada NPC e a empresa do distrito; pares de NPCs recebem relações de casamento. Cada empresa é ligada a um edifício como sede.

O nome, a idade, a aparência e a personalidade do NPC são reconstruíveis da seed e da `generation_key`; não são gravados como atributos no banco. As âncoras de identidade e relações são persistidas. Uma consulta por `generation_key` reconstrói só o NPC solicitado. `EmployeesOfCompany` consulta os IDs pela relação de emprego indexada e reconstrói apenas os funcionários encontrados.

As funções `GenerateNPC` e `GenerateCompany` reconstroem um registro específico sem gerar a população inteira. A API HTTP e a interface de investigação permanecem fora de M03.

## Versão 3 — catálogos de nomes por cidade

A versão `3` mantém o formato da sociedade inicial e escolhe nomes de catálogos locais. Lyon, França, possui 300 prenomes e 300 sobrenomes. O M11 acrescenta listas próprias de prenomes e sobrenomes para os demais países do catálogo; cidades usam o catálogo de seu país até receberem um catálogo municipal específico. As combinações são escolhidas deterministicamente por cidade e mundo, sem repetir nomes completos na amostra inicial. Variantes determinísticas mantêm a expansão quando uma combinação-base já foi usada.

A versão `2` e seus catálogos pequenos continuam disponíveis para reconstruir mundos existentes sem alterar nomes. Cada servidor conserva a versão de gerador registrada na criação; mundos novos usam a versão mais recente. Ao adicionar cidades, deve-se acrescentar um arquivo de catálogo e registrá-lo com seu país e cidade canônicos.

## Versão 4 — IDs de identidade visual

A versão `4` mantém a sociedade e os catálogos de nomes da versão 3 e acrescenta IDs de camadas visuais determinísticos para cada NPC. Cada NPC recebe cabelo (o ID identifica estilo e cor, incluindo ruivo), tom de pele, cor dos olhos, detalhes faciais opcionais e acessórios opcionais. Os IDs vêm de um fluxo independente por NPC derivado da seed e da chave estável; a ordem das consultas não interfere no resultado. Todos os jogadores do mesmo servidor recebem a mesma composição.

Versões `1` a `3` continuam reconstruíveis. Mundos existentes continuam usando a versão gravada; trocar sua seed ou versão não é uma migração automática e só pode ocorrer por reconstrução explícita. Novos mundos passam a usar a versão mais recente. Os PNGs e a composição visual no frontend seguem o contrato em [identidade visual modular](identidade-visual-modular.md).

A aparência escolhida para o personagem do jogador tem outra origem: é uma preferência individual, validada contra o catálogo de peças e persistida na conta. Ela não é derivada da seed mundial e não altera os personagens vistos por outros jogadores.

## Versão 5 — bairros procedurais e identidade visual calibrada

A versão `5` preserva a identidade modular da versão 4, cria 20 bairros e amplia a sociedade inicial para oito NPCs e uma empresa por bairro. Lyon mantém seus IDs e conteúdo históricos. Cada outra cidade recebe uma seed local derivada da seed do servidor e de seu ID canônico, mais 160 NPCs, 20 empresas e 80 locais. Os bairros, relações e identidades são determinísticos por servidor e cidade.

A geração visual da versão 5 também calibra cabelos por idade: NPCs abaixo de 40 anos não recebem cabelo grisalho; a chance aumenta gradualmente nas faixas seguintes. Calvície e cabelos ralos nas laterais só são selecionados a partir dos 45 anos. Cortes ausentes do catálogo de NPCs, como maria-chiquinha, não podem ser gerados. A versão 4 mantém sua saída histórica; novos mundos usam a versão 5.

## M11 — catálogo e viagem para bairros

O catálogo geográfico fixo e versionado associa cada cidade canônica a um país e fuso horário. O GATE apresenta países, cidades e bairros em sequência, e valida o bairro escolhido dentro da cidade selecionada. A viagem persiste apenas a cidade; o bairro de chegada fica como contexto transitório da sessão.

Em Central, todas as cidades apresentam Police Station, Bank, GATE Central e Hospital. O Hospital fica fechado e indisponível até uma feature futura. Selecionar ou abrir um desses prédios só troca a representação visual; não persiste uma visita. Ações investigativas futuras recebem o local como contexto da solicitação e validam que pertence à cidade atual do jogador.

A primeira viagem a uma cidade materializa seu conteúdo no mundo compartilhado do servidor. A tabela `world_city_generators` registra a versão usada por cidade para que NPCs, empresas e relações mantenham suas identidades em reconstruções futuras. Cada cidade materializada mantém sua própria população e seus próprios casos. A geração recorrente alterna entre essas cidades e seleciona os candidatos apenas dentro da cidade sorteada; a expansão pode adicionar NPCs ao bairro menos povoado dessa cidade.

## Determinismo e identidade

O gerador recebe uma seed opaca e uma versão explícita. A versão 1 deriva fluxos independentes usando SHA-256 sobre campos com comprimento prefixado: domínio do gerador, versão, seed e escopo (cidade ou catálogo de uma região). Cada fluxo usa SHA-256 com contador para produzir valores aleatórios sem depender de `math/rand`, relógio, estado global ou ordem de execução de outras regiões.

Cada região usa um caminho canônico (`fr/lyon/district/<slug>`). Cada lugar recebe `generation_key` baseado no caminho da região e seu tipo/slug; `entity_id` é `location:` seguido dessa chave. IDs de NPCs e empresas incluem o fingerprint determinístico da seed e versão, garantindo conjuntos de pessoas disjuntos entre worlds com seeds diferentes. Regiões e entidades são ordenadas por ID antes de retornar. Assim, a mesma seed/versão mantém conjunto, nomes, relações de contenção regional e IDs iguais após reinício do processo. IDs iguais em seeds diferentes permanecem isolados pelo `server_id`.

Qualquer alteração que mude a saída exige outra versão do gerador; a versão antiga deve continuar disponível para reconstruir mundos existentes.

## Consultar um NPC

1. Localizar seed.
2. Gerar atributos.
3. Consultar relações.
4. Aplicar mutations.
5. Montar o estado atual permitido ao cliente.

Resultado típico: Adam Miller, casado com Gina, empregado na Empresa #13 — sem gravar tudo isso na seed do NPC.
