# M24 — Company Websites procedurais

**Trilha:** independente do fluxo criminal  
**Status:** Em andamento  
**Depende de:** M23 aceita; modelo de empresas e endereços de M03; cache e estabilidade de M14  
**Escopo:** transformar cada empresa do mundo em um site público próprio, encontrável pela busca nacional Meridian.

## Objetivo

Ao abrir uma empresa pelo Meridian National Directory, o jogador navega para o site público daquela empresa dentro da janela atual do navegador. Cada empresa possui identidade visual, estrutura e conteúdo público gerados proceduralmente, com resultado estável para a mesma empresa e versão do mundo. O site apresenta um layout escolhido entre modelos customizáveis, opção visual de login sem autenticação funcional e informações de endereço.

A busca do Meridian continua nacional no país conectado. Ela indexa empresas e seus sites sem carregar funcionários ou o corpo completo dos sites. Os detalhes do site são carregados sob demanda pelo ID estável da empresa.

## Princípios e decisões

- **Determinístico por empresa:** derivar o perfil do site da seed do mundo, versão do gerador e ID estável da empresa. O mesmo mundo reconstruído deve produzir o mesmo site.
- **Modelo mais composição procedural:** manter modelos de página feitos à mão e personalizáveis, combinados com paletas, tipografia, cabeçalho, módulos, ordem de seções, imagens/ornamentos e cópia gerada. O modelo define limites de composição; dados e seleção procedural dão identidade a cada site.
- **Variedade verificável:** o catálogo de combinações deve comportar o total de empresas do mundo. A geração deve evitar repetir o mesmo perfil visual completo entre empresas e resolver colisões de forma determinística. Textos que exibem nome, atividade e endereço também devem refletir a empresa correspondente.
- **Busca enxuta, página sob demanda:** a resposta nacional contém apenas dados de descoberta (ID, nome, cidade, endereço resumido, domínio/identificador de site e descrição curta). Conteúdo e módulos do site são buscados quando a página é aberta.
- **Sem mudança de autorização:** páginas públicas do site apresentam informações públicas. Diretórios de funcionários e quaisquer fontes restritas continuam protegidos pelas tentativas de acesso e pela jurisdição Proxy atuais. O layout/login visual não concede autorização.
- **Sem credenciais:** o login é um elemento diegético inerte, sem submissão, armazenamento, validação, chamada de API ou solicitação de senha real. Se houver interação demonstrativa, deve explicar claramente que o serviço está indisponível ou não aceita sessões externas.
- **Navegação integrada:** sites abrem na mesma janela NOIR Browser, com endereço legível, voltar/início funcionais e links internos para as páginas do site. O retorno à busca preserva a consulta e os resultados.
- **Endereço canônico:** exibir o endereço já associado à empresa no mundo — local, bairro/distrito, cidade e país — sem gerar endereços divergentes da localização usada pelas regras de acesso.

## Escopo

### Geração do site

- Definir um contrato de dados separado da renderização: `companyId`, versão do site, domínio/slug, setor/atividade, perfil visual, conteúdo público e endereço canônico.
- Criar vários modelos customizáveis que compartilham componentes seguros e responsivos, mas variam em composição e hierarquia de informação.
- Gerar para cada site nome/marca exibida, subtítulo, apresentação, catálogo de serviços/produtos ou áreas de atuação, contatos fictícios, módulos e ordem de conteúdo coerentes com a atividade da empresa.
- Fornecer ao menos uma página inicial e páginas internas coerentes, incluindo `ABOUT`, `SERVICES`/`PRODUCTS`, `CONTACT`/`LOCATIONS` e `ACCOUNT`/`SIGN IN` conforme o perfil gerado.
- Garantir que o perfil visual completo não se repita entre empresas materializadas no mesmo mundo; identidade, endereço, conteúdo e links devem apontar sempre para a empresa correta.
- Não persistir cópias de dados determinísticos sem necessidade. Se cache for necessário, sua chave inclui mundo, versão de gerador e ID da empresa, com invalidação consistente com as mutations existentes.

### Busca Meridian e navegação

- Fazer a busca nacional encontrar empresas por nome, domínio/slug, descrição curta e endereço, limitada ao país do Proxy ativo e ao limite de resultados existente.
- Exibir deeplinks textuais para abrir o site; apresentar nome, cidade, endereço resumido e uma descrição curta sem transformar linhas em cards clicáveis.
- Abrir detalhes do site somente após a seleção; nunca carregar funcionários de várias empresas na busca nacional.
- Garantir acesso direto por ID estável sem depender do nome exibido, que pode se repetir.
- Manter a navegação do site no mesmo browser; voltar e retorno ao diretório preservam o estado da busca.
- A partir da página pública, oferecer a entrada para o diretório de funcionários sujeito à tentativa de acesso já existente. Não incluir a lista de funcionários no payload público do site.

### Apresentação e acessibilidade

- Cada site deve ser reconhecível como empresa distinta à primeira vista, mantendo a estética de web antiga e a moldura/cabeçalho comuns do navegador.
- Usar links reais apenas onde houver navegação; elementos decorativos e o login inerte não devem parecer controles funcionais.
- Operar por teclado, oferecer foco visível, títulos e regiões semânticas, anunciar carregamento/erro e funcionar a partir de 320 px sem scroll horizontal.
- Preservar texto legível e contraste nos perfis de cores disponíveis.

### Performance e consistência

- Não gerar o site completo ao abrir o Meridian Directory nem ao digitar na busca.
- Evitar consultas por empresa no loop de resultados; carregar dados públicos com operação em lote apropriada ou geração determinística no servidor.
- Manter limite de payload e latência medidos para o National Directory da M23; definir um orçamento próprio para abrir uma página empresarial após baseline.
- Medir cache frio/aquecido e concorrência para busca nacional e abertura de site; registrar p50/p95, consultas, bytes, memória e contagem de sites/perfis produzidos.
- Garantir por teste que a reconstrução determinística produz os mesmos sites e que empresas distintas não compartilham um perfil completo.

## Fora de escopo

- Login real, criação de conta, senha, cookies ou autenticação de terceiros.
- Checkout, mensagens, formulários com efeitos persistentes, rede externa ou conteúdo carregado de sites reais.
- Alterar a quantidade/identidade de empresas, a geração de funcionários, as regras de autorização do diretório de equipe ou o escopo nacional do Person Search.
- Refazer identidade visual de sites institucionais existentes (polícia, hospital, Meridian Enterprises, correio) ou o jornal como app.
- Gerar todas as páginas empresariais durante a busca nacional.

## Critérios de aceite

- [ ] Cada empresa do catálogo tem um site público determinístico, associado ao seu ID estável e endereço canônico.
- [ ] Há múltiplos modelos customizáveis e empresas distintas não repetem o mesmo perfil visual completo dentro do mundo gerado.
- [ ] A mesma seed, versão e ID da empresa reproduzem o mesmo site; alterações de versão são explícitas e cobertas por testes.
- [ ] Os sites têm conteúdo próprio coerente com a empresa, links internos funcionais e páginas para informações, áreas/serviços, contato/endereço e acesso de conta quando aplicável.
- [ ] A opção de login é apenas diegética, não coleta credenciais e não cria sessão nem chama autenticação.
- [ ] O endereço apresentado corresponde exatamente à localização canônica da empresa usada pelo Proxy e pelo backend.
- [ ] Meridian encontra sites por nome, slug/domínio, descrição e endereço no país conectado e respeita o limite da busca nacional.
- [ ] Resultados da busca não carregam equipe nem conteúdo completo dos sites; abrir uma empresa carrega somente o site escolhido.
- [ ] Diretório de funcionários e fontes restritas preservam autorização, expiração, conta, empresa, site e jurisdição Proxy existentes.
- [ ] Browser Back, retorno ao diretório, erro, loading, vazio e sucesso funcionam sem abrir nova janela e preservam a pesquisa.
- [ ] Sites funcionam por teclado e leitor de tela, têm contraste/foco visível e não apresentam scroll horizontal em viewport mínimo de 320 px.
- [ ] Relatório inclui baseline e medições HTTP de busca e abertura, p50/p95 frios/aquecidos e concorrência moderada, consultas, payload, memória e metas explícitas.
- [ ] Build e testes cobrem determinismo, diversidade, links/endereços, escopo nacional, não exposição de funcionários e preservação das autorizações.

## Plano de ataque

1. **Auditar o site atual:** mapear Company Portal, abertura no Proxy, autorização, diretório de funcionários, modelo de empresas/endereço, contratos API e estado da navegação no browser. Registrar o que será mantido e remover o atual acoplamento “abrir empresa = abrir staff directory”. (Auditoria inicial concluída; o fluxo existente abria o diretório de funcionários diretamente após autorização.)
2. **Fechar contrato e decisão de geração:** decidir quais campos derivam da seed, como é medida/garantida diversidade por mundo, versionamento e cache. Atualizar docs de arquitetura e registrar ADR se houver escolha persistente entre gerar em runtime ou persistir. (Contrato efêmero e determinístico registrado no ADR 0013; perfil tem versão explícita.)
3. **Implementar gerador procedural:** produzir perfil tipado e estável de site, catálogo de modelos customizáveis, conteúdos, domínio/slug e páginas internas. Adicionar testes determinísticos, de diversidade e de endereço. (Gerador, tema, marca, setores e páginas tipadas implementados; testes permanecem pendentes.)
4. **Separar API de busca e API de site:** expandir busca Meridian com chaves pesquisáveis compactas e criar leitura do site por ID, sem dados de funcionários. Validar país Proxy, ID, empresa e cidade sem consultas N+1. (Busca indexa nome, domínio, descrição e endereço; endpoint público consulta somente o site selecionado no país ativo.)
5. **Construir renderizador do website:** criar shell empresarial dentro da janela NOIR Browser, com modelos variados, navegação interna e login diegético inerte. Integrar endereço canônico e deeplinks do diretório. (Renderizador consome as páginas geradas, varia composição/cores/cabeçalho e usa deeplinks na mesma janela.)
6. **Integrar dados restritos:** ligar entrada de staff directory a partir do site, mantendo a tentativa de acesso atual e sem incluir registros no payload público. (A rota de staff solicita conexão ao Proxy da sede e autorização existente antes de carregar a equipe.)
7. **Auditar usabilidade e desempenho:** teclado, leitor de tela, contraste, viewport 320 px, estados de rede, volta ao diretório, performance fria/aquecida e concorrente. (Medição inicial registrada em [relatório](../relatorios/M24-company-websites.md) e CSV; auditoria manual e testes de regressão permanecem pendentes.)
8. **Fechar M24:** atualizar critérios com evidências, relatório, arquitetura, roadmap e estado.

## Progresso

- O gerador Go `GenerateCompanyWebsite` produz domínio, perfil de template/tema, conteúdo público, endereço canônico e opção de conta inerte a partir da seed, versão do mundo e ID da empresa.
- O índice nacional passou a incluir domínio e descrição curta pesquisáveis; o payload de busca continua sem funcionários.
- A API ganhou leitura de website por ID estável limitada ao país da jurisdição Proxy ativa.
- O browser abre a página pública na mesma janela, com deeplinks internos, variações visuais e caminho separado para solicitar o diretório de funcionários protegido.
- A navegação de volta mantém consulta e resultados do National Directory; a URL virtual reflete domínio e subpágina selecionada.
- A segunda rodada de geração corrigiu o domínio, que antes derivava o sufixo de um ID hierárquico e podia inserir barras. O token agora é hexadecimal e determinístico; a busca e o site compartilham o mesmo domínio. O perfil foi versionado para 2.
- A atividade e a composição da página inicial agora consideram palavras do nome empresarial; seis modelos usam estruturas HTML simples diferentes, com pouca decoração CSS e seções de artigo, aviso, lista, endereço ou fatos.
- ADR 0013 e a arquitetura documentam geração sob demanda, versão do perfil, busca leve e separação entre página pública e diretório restrito.
- Build da API e build web passaram após a atualização. Medição inicial de busca e abertura de site foi registrada em [relatório](../relatorios/M24-company-websites.md), com CSV e script reproduzível. Ela mostra respostas compactas e cache quente na faixa de 156–175 ms p50 sequencial; a busca chega a 1.42 s p95 sob concorrência 8 e precisa ser acompanhada.
- Testes de determinismo/diversidade, auditoria visual/acessível, regressões de autorização e confirmação de volume/diversidade do catálogo continuam pendentes; os critérios de aceite seguem sem marcação até haver evidência.

## Decisões adotadas para implementação

- Usar endereço procedural legível, por exemplo `nomedaempresa.noir.local`, com sufixo estável quando necessário para evitar colisões. O ID da empresa continua sendo a chave canônica da API e do estado de navegação.
- Garantir endereço e contato em todos os sites. `ABOUT`, `SERVICES`/`PRODUCTS` e `ACCOUNT` podem variar por modelo, mas cada perfil deve oferecer um caminho claro de volta à página inicial.
- Exibir o login como elemento visual inerte/desativado, sem formulário nem coleta de dados. O diretório protegido de funcionários usa a autorização investigativa já existente, separada desse login.

