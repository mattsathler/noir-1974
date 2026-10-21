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

A geografia terá duas camadas. A camada de referência será fixa e igual em todos os servidores: um catálogo curado de 50 países de diferentes partes do mundo, com suas capitais e principais cidades. Esses nomes e identidades não variam com a seed do servidor. O catálogo ainda precisa ser selecionado e registrado como dado versionado do jogo.

Dentro das cidades, os bairros serão partições lógicas para distribuir detetives e ocorrências, evitando a concentração de jogadores em um único lugar. Seus nomes serão compostos deterministicamente a partir de um banco de palavras e regras de combinação, com escopo por servidor e cidade. Novos bairros serão criados sob demanda conforme a distribuição de jogadores e a necessidade de gerar ocorrências locais; uma vez criados, sua identidade permanece estável para todos os jogadores daquele servidor.

Não haverá mapa físico individual dos bairros nem planejamento de coordenadas, distâncias, rotas ou vizinhança entre eles. O bairro identifica o contexto lógico de uma ocorrência e ajuda o sistema a distribuir jogadores e casos; não determina deslocamento ou proximidade física. A geração deve evitar nomes duplicados ou ambíguos dentro da mesma cidade e preservar chaves canônicas, independentemente da ordem em que bairros forem criados.

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

A versão `3` mantém o formato da sociedade inicial e escolhe nomes de catálogos próprios para cada cidade. Os catálogos ficam em arquivos por país e cidade em `apps/api/internal/world/names`; o registro de cidades liga o código geográfico canônico ao catálogo. Lyon, França, possui 300 prenomes e 300 sobrenomes, combinados e embaralhados deterministicamente por mundo. A amostra de combinações de um mundo não repete nomes completos.

A versão `2` e seus catálogos pequenos continuam disponíveis para reconstruir mundos existentes sem alterar nomes. Cada servidor conserva a versão de gerador registrada na criação; mundos novos usam a versão mais recente. Ao adicionar cidades, deve-se acrescentar um arquivo de catálogo e registrá-lo com seu país e cidade canônicos.

## Próxima versão — identidade visual procedural

O M10 adicionará uma composição visual modular para NPCs. A nova versão do gerador derivará, em um fluxo independente por NPC, IDs de peças compatíveis para rosto, tom de pele, cabelo, características visuais e acessórios. A composição depende da seed mundial, da versão do gerador, da versão do catálogo visual e da chave estável do NPC; não depende da ordem em que entidades forem consultadas. Assim, todos os jogadores do mesmo servidor verão a mesma composição para cada NPC.

Esta mudança exige incrementar a versão do gerador (planejada como versão `4`), mantendo as versões `1` a `3` reconstruíveis. Mundos existentes continuam usando a versão gravada; trocar sua seed ou versão não é uma migração automática e só pode ocorrer por reconstrução explícita. Novos mundos passam a usar a versão mais recente.

A aparência escolhida para o personagem do jogador tem outra origem: é uma preferência individual, validada contra o catálogo de peças e persistida na conta. Ela não é derivada da seed mundial e não altera os personagens vistos por outros jogadores.

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
