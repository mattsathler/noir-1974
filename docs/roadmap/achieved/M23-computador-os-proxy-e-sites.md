# M23 — Sistema operacional, Proxy e sites investigativos

**Trilha:** Independente do fluxo criminal  
**Status:** Concluída
**Depende de:** M07, M08, M12, M16 e M21 aceitos; ADR 0012 aceita  
**Substitui:** M22 — Refatoração do bypass e navegação de arquivos com Ghostkey; cena de viagem do M17; I02 — Introdução visual após viagem confirmada pelo GATE

## Objetivo

Transformar o computador pessoal do detetive na tela principal do jogo e migrar a investigação para uma navegação por portais e sites. O Proxy substitui a viagem do protagonista como forma de selecionar a jurisdição ativa. O portal nacional lista empresas do país conectado; abrir um site carrega somente os dados daquela empresa depois da autorização existente. Person Search continua nacional dentro do país conectado.

A migração altera apresentação, navegação e escopo das buscas explicitamente nacionais. Não muda as regras de casos, acesso, descoberta, warrants, acusações, autoria das fontes ou jurisdição legal.

## Escopo

### Sistema operacional e navegador

- Tornar o computador pessoal a interface persistente da tela principal do jogo, com área de trabalho, janelas e barra de navegação.
- Apresentar a área de trabalho em tela cheia, sem moldura, com papel de parede configurável, atalho do navegador à esquerda e sites navegados dentro de uma única janela do navegador; utilitários do sistema ficam no menu Iniciar. O portal inicial oferece atalhos para National Hospital, Police Department, Meridian Enterprises e Meridian Email. Sites institucionais oferecem navegação interna, avisos e publicidade fictícia em páginas claras de estilo web antigo, com identidade visual própria por instituição.
- Migrar as funções atuais — Crime Reports, email, jornal, telefone, quadro investigativo e fontes de consulta — para superfícies acessíveis dentro desse sistema.
- Apresentar empresas e instituições como sites. Remover Neighborhood, interiores de prédios e a animação de viagem como etapas obrigatórias para consultar fontes.
- Preservar os estados de carregamento, vazio, erro, sucesso e janela já aberta durante a navegação entre páginas.
- Padronizar a moldura e o cabeçalho das janelas de apps; páginas institucionais mantêm cabeçalhos consistentes entre página inicial e sublinks, com identidade visual própria por instituição.
- Impedir rolagem horizontal no navegador. Conteúdo extenso usa rolagem vertical; o jornal ocupa a área disponível sem barra amarela interna.
- Indicar visualmente requisições em andamento com cursor de carregamento.

### Proxy e jurisdição

- Oferecer um aplicativo Proxy com seleção de cidade. A cidade conectada define a jurisdição investigativa ativa, usando inicialmente o estado de localização já existente.
- Trocar o nó Proxy não transporta o personagem e não registra presença física.
- Ao selecionar um site empresarial cadastrado em outra cidade, o cliente conecta ao nó daquela cidade antes de solicitar dados restritos. A operação só prossegue após as mesmas validações de conta, cidade e autorização no backend.
- Exibir a cidade conectada e indicar quando a navegação troca de jurisdição, para manter claro o escopo das fontes.

### Portal nacional de empresas

- O portal inicial do país funciona como uma página de atalhos, sem busca de pessoas ou empresas. A busca empresarial fica no National Directory, que reúne empresas, serviços e instituições nacionais.
- No National Directory, pesquisar empresas por nome e apresentar até dez resultados por consulta, com nome, cidade e informações breves de identificação.
- Não incluir funcionários nos resultados do diretório. Abrir o site da empresa é uma consulta separada.
- Abrir o site de uma empresa e, após a autorização concedida pelo servidor, carregar somente o diretório daquela empresa. Nenhum funcionário de outra empresa ou cidade deve aparecer nesse diretório.
- Identificar empresas por ID estável e cidade, incluindo quando nomes visíveis se repetirem.
- Consultar funcionários sob demanda, sem baixar todos os diretórios na abertura do portal ou durante a busca.

### Person Search e fontes existentes

- Manter Person Search nacional no país conectado, com até dez resultados por busca. Essa busca é independente do diretório de funcionários de cada empresa.
- Corrigir a associação de cenas de caso para que o perfil de uma pessoa encontrada em outra cidade use a cidade do próprio NPC.
- Preservar consulta de casos, jornal, hospital, banco, registros GATE, email, interrogatórios, warrants e acusações nos respectivos escopos e permissões atuais. O jornal continua como app independente, com sua estética editorial própria.
- Meridian Enterprises é o portal institucional da empresa que controla o GATE, o sistema operacional e outras infraestruturas do mundo. Os GATE Access Logs ficam em uma subpágina e mantêm as autorizações existentes.
- Consultas por NPC devem carregar o NPC selecionado e o intervalo solicitado; não devem carregar os NPCs ou registros de todo o país.
- Manter registros de viagem GATE como eventos do mundo e pistas sobre NPCs, apesar da remoção da viagem física do protagonista como fluxo de navegação.

### Autorização dos sites restritos

- Reutilizar a autorização server-side atual, vinculando tentativa a conta, jurisdição, empresa/site e sistema acessado.
- Manter o período de validade e impedir reutilização de uma tentativa em outra empresa, cidade, conta ou sistema.
- Para sites empresariais, resolver o endereço para a cidade canônica e exigir igualdade exata com a jurisdição Proxy antes de conceder a tentativa; não comparar somente nomes exibidos de cidades.
- Solicitar autorização ao servidor ao conectar a um site restrito. A interface nunca concede acesso sem confirmação da API.
- Não introduzir autenticação real, acesso à rede externa ou acesso ao sistema de arquivos do computador do jogador.

### Ghostkey e provedores

- Disponibilizar `ghostkey.exe` como aplicativo do desktop. Antes de abrir a janela, o MeridOS apresenta o aviso de aplicativo não autorizado e permite cancelar ou executar; a entrada do app apresenta o boot interno e o Network Explorer.
- Exibir provedores corporativos do país conectado e navegar por diretórios virtuais em pastas, sem referências a FTP. Ao abrir um provedor, carregar seus diretórios de hosts; os endereços IP dos sites empresariais são determinísticos a partir da empresa e permitem encontrar o site mesmo antes de visitá-lo no navegador.
- Usar pastas genéricas e arquivos com nomes plausíveis; executáveis têm nomes técnicos e ícones próprios, enquanto arquivos não suportados mostram o modal de erro padrão do MeridOS. A navegação não oferece acesso ao sistema de arquivos real.
- Restringir Person Search, Residential Search & Seizure e Arrest Warrant da polícia, Patient Search do Hospital, Employee Directory empresarial e GATE Access Logs aos executáveis encontrados pelo Ghostkey. A autorização continua validada no servidor e vinculada ao contexto apropriado; atalhos diretos e o link de GATE no navegador não contornam essa restrição.
- Aplicar efeitos visuais de acesso comprometido ao Ghostkey e aos apps que ele abre, mantendo o layout e a janela comuns do MeridOS. Carregamento interno permanece dentro da janela; carregamento branco de página inteira é reservado à navegação entre páginas.
- Paginar diretórios de hosts em blocos de 20 e filtrar por IP no cliente, ignorando pontos durante a comparação. Busca vazia restaura a listagem; não executar filtragem enquanto o jogador digita.

### Performance

- Implementar uma consulta de empresas por país que gere ou consulte somente as quatro cidades daquele país.
- Separar a lista de empresas do carregamento de funcionários e limitar cada resposta da busca a dez resultados; usar paginação se os testes de uso mostrarem que é necessária.
- Evitar varrer relações de localização repetidamente para cada empresa; construir índices em memória ou usar consulta indexada apropriada.
- Debounce ou cancelar buscas anteriores da barra de pesquisa e não consultar com query vazia.
- Medir p50 e p95 de cache frio e aquecido para a lista nacional, incluindo serialização JSON, e repetir com concorrência moderada. Registrar o número de consultas ao banco, linhas retornadas quando disponíveis, tamanho da resposta e uso de memória do cache; documentar limites da instrumentação para linhas examinadas.
- Medir também o acesso ao diretório de uma empresa e comparar a resposta com a listagem nacional. A medição histórica de Person Search é referência separada, não substituto desse benchmark.

## Critérios de aceite

- [x] A tela principal do jogo é o sistema operacional do computador pessoal, com desktop em tela cheia sem moldura, navegador no desktop, portal de atalhos do país conectado e navegação dos sites na mesma janela em superfície clara; os fluxos investigativos não dependem de viajar ou entrar fisicamente em prédios.
- [x] O Proxy seleciona uma cidade e atualiza a jurisdição investigativa sem representar viagem ou presença física.
- [x] O portal inicial oferece os quatro atalhos institucionais acordados; o National Directory pesquisa somente as quatro cidades do país conectado e retorna até dez resultados com cidade identificável.
- [x] Resultados do portal não incluem funcionários; abrir uma empresa consulta somente o diretório daquela empresa.
- [x] O Person Search retorna até dez pessoas do país conectado e permanece distinto do diretório empresarial.
- [x] Person Search associa a cena correta a resultados de qualquer cidade do país e não revela dados classificados internamente.
- [x] Um site de outra cidade estabelece a jurisdição proxy correspondente antes de tentar dados restritos; o backend valida conta, cidade, endereço/site, empresa e autorização.
- [x] Tentativas negadas, expiradas ou pertencentes a outro contexto não liberam o site nem suas consultas.
- [x] Casos, warrants, acusações, banco, GATE, emails e hospital preservam seu conteúdo e as regras de escopo atuais.
- [x] Os registros GATE de NPCs e sua função investigativa permanecem intactos.
- [x] Meridian Enterprises apresenta a empresa e seus sistemas; os GATE Access Logs abrem como subpágina e preservam autorização e escopo.
- [x] Ghostkey apresenta provedores e diretórios virtuais com IPs determinísticos de sites empresariais, paginação e filtro por endereço; Person Search, Residential Search & Seizure, Arrest Warrant, Patient Search, Employee Directory e GATE Access Logs são acessíveis pelos executáveis do explorador e permanecem sujeitos à autorização server-side.
- [x] Cabeçalhos de janela e cabeçalhos institucionais permanecem consistentes entre páginas relacionadas; o jornal continua como app com identidade editorial própria.
- [x] O navegador não apresenta rolagem horizontal e indica requisições pendentes com cursor de carregamento.
- [x] Nenhuma busca nacional carrega diretórios de funcionários em lote; respostas são limitadas e não truncadas silenciosamente.
- [x] O relatório de performance registra medições frias e aquecidas, p50, p95, número de queries, payload e concorrência moderada, com alvos definidos a partir da nova linha de base.
- [x] A interface funciona por teclado, leitor de tela e telas pequenas; mudanças de jurisdição são anunciadas sem depender apenas de cor.
- [x] Lore, ADR, roadmap, API e arquitetura descrevem o mesmo modelo após a migração.

## Fora de escopo

Person Search global entre países; diretórios nacionais de funcionários; acesso irrestrito a sites de outras cidades sem autorização; remoção do GATE físico do mundo; alteração dos eventos GATE de NPCs; mudança nos critérios de evidência, investigação, warrant, acusação ou prisão; implementação de internet real; migração do banco de dados para outra tecnologia.

## Sequência de execução

1. **Fundação narrativa e arquitetura:** registrar lore e ADR 0012; revisar contratos e dependências de M22. (Concluída.)
2. **Casca do sistema operacional:** criar shell persistente e migrar a tela de jogo e as janelas atuais. (Concluída: shell principal com janela de navegador, barra de tarefas, telefone, jornal como aplicativo do desktop, quadro e aparência; desktop Win96 em tela cheia, portal nacional com quatro atalhos institucionais, sites institucionais com identidades próprias e navegação interna na mesma janela. Police Department tem páginas informativas e Crime Reports; National Hospital oferece pesquisa de pacientes protegida por autorização; o seletor de papel de parede permanece para implementação futura.)
3. **Proxy e contexto:** selecionar jurisdição e integrar o contexto ao navegador e às autorizações existentes. (Concluída para Proxy, sites empresariais e serviços públicos; fluxos de acesso existentes preservados.)
4. **Portal nacional de empresas:** criar busca leve por país no National Directory, resultados limitados e abertura do site empresarial; medir performance. (Busca limitada, cache de catálogo e abertura implementadas. Meridian Enterprises tem página institucional própria e mantém GATE Access Logs em subpágina. Relatório e amostras em `docs/roadmap/relatorios/M23-company-portal.md` registram cache frio/aquecido, concorrência de 8 workers, p50/p95, consultas, payloads e metas iniciais. Todas as metas foram atingidas. O tracer PGX não informa linhas examinadas nem foi coletado plano EXPLAIN; essa limitação está documentada.)
5. **Person Search e fontes:** preservar a busca nacional, corrigir cenas entre cidades e integrar as fontes restantes sem ampliar escopo acidentalmente. (Cenas agora resolvidas entre cidades do país. Pesquisa hospitalar, diretórios empresariais e registros restritos usam tentativas de acesso validadas no servidor; a concessão empresarial exige cidade canônica correspondente ao Proxy. A auditoria estática confirmou preservação dos escopos e regras de casos, warrants, acusações, banco, GATE, emails e hospital. A geração recorrente agora também marca a vítima como morta no mundo compartilhado antes de avançar o cursor, mantendo o estado hospitalar consistente.)
6. **Remoção do fluxo físico e auditoria:** retirar a navegação antiga, atualizar documentação e cobrir regressões de autorização, conteúdo e performance. (Navegação do protagonista por bairros e interiores removida; GATE segue acessível como subpágina de Meridian Enterprises. Proxy inválido falha explicitamente, sem fallback para Lyon. Lore, ADR, API e arquitetura foram sincronizados. Interfaces oferecem foco visível, controles nativos operáveis por teclado, regiões de estado para leitores de tela, anúncio de jurisdição e layouts responsivos. Ghostkey concentra o acesso aos executáveis restritos por provedores e hosts virtuais. Build web/API, suíte web, testes API e E2E, benchmark concorrente e `git diff --check` passaram.)
