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

A geografia tem duas camadas. A camada de referência é fixa e igual em todos os servidores: o catálogo inicial de 25 países e 100 cidades, registrado no M11. Esses nomes e identidades não variam com a seed do servidor. A interface Proxy apresenta países e cidades para selecionar a jurisdição ativa; bairros e construções continuam como contexto lógico de endereços, sem uma etapa de viagem ou visita para o protagonista. A localização persistida do jogador identifica a cidade Proxy; não registra presença física.

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

A geração visual da versão 5 também calibra cabelos por idade: NPCs abaixo de 40 anos não recebem cabelo grisalho; as chances são 5% entre 40–49, 20% entre 50–59 e 40% a partir de 60. Calvície e cabelos ralos nas laterais só são selecionados a partir dos 45 anos. Cortes ausentes do catálogo de NPCs, como maria-chiquinha, não podem ser gerados. A versão 4 mantém sua saída histórica; versões 5 a 8 permanecem disponíveis para reconstrução.

## Versão 6 — compatibilidade da aparência de NPCs

A versão `6` preserva o mundo e as regras visuais da versão 5 e declara a apresentação visual de cada NPC (`masculine` ou `feminine`) em um fluxo determinístico próprio. Essa apresentação descreve a compatibilidade convencional das peças, não a orientação sexual do personagem. O catálogo declara compatibilidade por ID: barba é compatível com apresentação masculina; as demais peças atuais são compatíveis com ambas. A geração e a persistência rejeitam composições incompatíveis, e a descrição textual continua sendo derivada das peças selecionadas. Versões `1` a `5` permanecem reconstruíveis sem alteração.

## Versão 7 — bairros comerciais ampliados

A versão `7` mantém os bairros, a aparência e a apresentação visual da versão 6. Cada um dos 20 bairros de uma cidade recebe seis prédios comerciais, selecionados de um catálogo de 36 nomes. Cada prédio abriga seis empresas, totalizando 720 empresas por cidade. Cada empresa tem de quatro a dez funcionários, escolhidos deterministicamente; emprego, cargos, relações pessoais e sede apontam para os mesmos registros reconstruíveis. Os nomes das empresas incluem um catálogo ampliado e um ordinal estável para evitar colisões.

Os bancos de nomes dos países com listas menores crescem em até sete vezes por categoria, usando compostos determinísticos; Lyon conserva seu catálogo municipal, que já comporta a população maior. A versão 7 foi a versão corrente antes da versão 8; mundos e cidades persistidos continuam usando sua versão. Versões anteriores continuam disponíveis e não mudam sua saída.

## Versão 8 — aparência regional inclusiva e detalhes faciais

A versão `8` preserva a estrutura comercial da versão 7 e introduz uma nova composição visual, sem reidentificar NPCs de versões anteriores. Toda mudança em cabelo, pele, barba ou detalhe facial usa fluxos independentes derivados da versão e da `generation_key`; cidades não materializadas passam a registrar a versão atual, enquanto cidades e mundos persistidos mantêm a sua versão.

Os sete tons de pele usam quatro misturas regionais autorais, selecionadas deterministicamente pelo ID de cidade e bairro. Em todas, os sete tons (`white`, `pale`, `light`, `medium`, `olive`, `brown`, `dark`) aparecem entre 12% e 16% da distribuição. São perfis sintéticos do universo do jogo, não estimativas demográficas nem proxies de nacionalidade; nenhum perfil é ligado a país, nome, apresentação ou orientação sexual. Cabelo recebe um fluxo próprio; pele, olhos, barba e acessórios usam o fluxo visual determinístico do NPC. Barba continua sendo compatibilidade declarada de apresentação, sem qualquer dado de orientação sexual. O jogador escolhe diretamente a aparência do próprio personagem.

A geração facial oferece cicatriz, sardas leves ou densas, sinal, linhas de sorriso e rugas finas ou profundas. Rugas finas só podem ser geradas a partir dos 45 anos; rugas profundas, a partir dos 60. Os limites de idade, paletas e compatibilidade são validados ao reconstruir NPCs. As versões 4 a 7 mantêm suas composições históricas.

## Catálogo geográfico, materialização e Proxy

O catálogo geográfico fixo e versionado associa cada cidade canônica a um país e fuso horário. O Proxy apresenta os nós de cidade e altera a jurisdição investigativa ativa; não representa uma viagem do protagonista. Distritos e endereços continuam sendo dados do mundo usados para descrever instituições e validar o escopo dos serviços.

Em Central, todas as cidades apresentam Police Station, Bank, GATE Central e Hospital como serviços do mundo. Esses serviços são acessados por seus sites, após conexão Proxy e validação da tentativa de acesso no backend. A presença de um endereço numa resposta não significa que o protagonista o visitou.

A primeira conexão Proxy a uma cidade pode materializar seu conteúdo no mundo compartilhado do servidor. A tabela `world_city_generators` registra a versão usada por cidade para que NPCs, empresas e relações mantenham suas identidades em reconstruções futuras. Cada cidade materializada mantém sua própria população e seus próprios casos. A geração recorrente alterna entre essas cidades e seleciona os candidatos apenas dentro da cidade sorteada; a expansão pode adicionar NPCs ao bairro menos povoado dessa cidade.

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
