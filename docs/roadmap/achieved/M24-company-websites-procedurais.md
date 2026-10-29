# M24 — Company Websites procedurais

**Trilha:** independente do fluxo criminal  
**Status:** Aceita  
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

- [x] Cada empresa consultada recebe um site público determinístico, associado ao seu ID estável e endereço canônico; geração e endereço são cobertos por testes.
- [x] O catálogo oferece dez composições, treze famílias de cabeçalho e quatro políticas de anúncios; um teste de 4.096 identidades não encontrou repetição do perfil visual completo.
- [x] A mesma seed, versão e ID reproduzem o mesmo site; a versão do perfil é explícita e mudança da versão do mundo altera o perfil, com cobertura Go.
- [x] Os sites têm conteúdo procedural, links internos, páginas de informação, serviços, contato/endereço e conta; testes cobrem navegação interna e endereço.
- [x] Login é diegético e desativado, sem formulário ou coleta de credenciais; coberto por teste do renderizador.
- [x] O endereço do website é copiado da localização canônica; teste compara o endereço completo e a página de contato.
- [x] O National Directory procura no país Proxy por nome, domínio, descrição e endereço, com limite atual de dez resultados; a rota do website também é testada com a jurisdição do país conectado.
- [x] Busca e website são payloads separados; o tipo público de resultado não inclui equipe e o teste do perfil verifica que não há campo de funcionários.
- [x] O acesso ao diretório empresarial ainda conecta ao Proxy da sede e solicita autorização vinculada à empresa/localização; as rotas existentes de equipe continuam protegidas. A rota pública de website rejeita sessão ausente e passa país/empresa corretos ao store.
- [x] Navegação interna ocorre na janela atual e preserva busca/diretório ao voltar; os testes web relacionados passaram, junto com a auditoria dos estados de erro, vazio e carregamento.
- [x] O renderizador usa controles nativos, regiões e rótulos semânticos, foco visível e breakpoints responsivos; a revisão visual foi feita nas capturas do diretório e em amostras do site procedural.
- [x] Relatório registra baseline e revalidação fria/aquecida e concorrente, p50/p95, consultas, payload, heap observado, metas de acompanhamento e limites do tracer.
- [x] Build web, suíte web completa (75 testes), suíte Go completa e teste de diversidade/determinismo/endereço passaram.

## Plano de ataque

1. **Auditar o site atual:** mapear Company Portal, abertura no Proxy, autorização, diretório de funcionários, modelo de empresas/endereço, contratos API e estado da navegação no browser. Registrar o que será mantido e remover o atual acoplamento “abrir empresa = abrir staff directory”. (Auditoria inicial concluída; o fluxo existente abria o diretório de funcionários diretamente após autorização.)
2. **Fechar contrato e decisão de geração:** decidir quais campos derivam da seed, como é medida/garantida diversidade por mundo, versionamento e cache. Atualizar docs de arquitetura e registrar ADR se houver escolha persistente entre gerar em runtime ou persistir. (Contrato efêmero e determinístico registrado no ADR 0013; perfil tem versão explícita.)
3. **Implementar gerador procedural:** produzir perfil tipado e estável de site, catálogo de modelos customizáveis, conteúdos, domínio/slug e páginas internas. Adicionar testes determinísticos, de diversidade e de endereço. (Concluído; testes Go cobrem reprodução, versão, endereço canônico e 4.096 identidades sem perfil visual repetido.)
4. **Separar API de busca e API de site:** expandir busca Meridian com chaves pesquisáveis compactas e criar leitura do site por ID, sem dados de funcionários. Validar país Proxy, ID, empresa e cidade sem consultas N+1. (Busca indexa nome, domínio, descrição e endereço; endpoint público consulta somente o site selecionado no país ativo.)
5. **Construir renderizador do website:** criar shell empresarial dentro da janela NOIR Browser, com modelos variados, navegação interna e login diegético inerte. Integrar endereço canônico e deeplinks do diretório. (Renderizador consome as páginas geradas, varia composição/cores/cabeçalho e usa deeplinks na mesma janela.)
6. **Integrar dados restritos:** ligar entrada de staff directory a partir do site, mantendo a tentativa de acesso atual e sem incluir registros no payload público. (A rota de staff solicita conexão ao Proxy da sede e autorização existente antes de carregar a equipe.)
7. **Auditar usabilidade e desempenho:** teclado, semântica, contraste, viewport responsivo, estados de rede, volta ao diretório e performance fria/aquecida e concorrente. (Concluído por revisão de código, capturas fornecidas durante a auditoria, testes de renderização e revalidação de performance; limites do ambiente e do tracer registrados no relatório.)
8. **Fechar M24:** atualizar critérios com evidências, relatório, arquitetura, roadmap e estado. (Concluído; milestone movida para `achieved/`.)

## Progresso

- O gerador Go `GenerateCompanyWebsite` produz domínio, perfil v3, dez composições, treze famílias de cabeçalho, conteúdo público, endereço canônico, densidade de anúncios e opção de conta inerte a partir da seed, versão do mundo e ID da empresa.
- O índice nacional passou a incluir domínio e descrição curta pesquisáveis; o payload de busca continua sem funcionários.
- A API ganhou leitura de website por ID estável limitada ao país da jurisdição Proxy ativa.
- O browser abre a página pública na mesma janela, com deeplinks internos, variações visuais e caminho separado para solicitar o diretório de funcionários protegido.
- A navegação de volta mantém consulta e resultados do National Directory; a URL virtual reflete domínio e subpágina selecionada.
- A geração corrigiu o domínio, que antes derivava o sufixo de um ID hierárquico e podia inserir barras. O token agora é hexadecimal e determinístico; busca e site compartilham o domínio. O perfil atual está versionado para 3.
- A atividade e a composição inicial consideram palavras do nome empresarial. Dez modelos usam estruturas HTML simples distintas, com cabeçalhos procedurais e quatro densidades de anúncios, de nenhum a excessivos.
- ADR 0013 e a arquitetura documentam geração sob demanda, versão do perfil, busca leve e separação entre página pública e diretório restrito.
- A busca nacional mantém os resultados leves e o endpoint gera somente o site solicitado no país Proxy ativo. O tipo público não contém funcionários.
- A suíte web completa passou: 23 arquivos e 75 testes. A suíte Go completa passou, incluindo testes da API e E2E; a execução do E2E precisou de permissão para abrir uma porta local.
- O teste de perfil verificou 4.096 IDs distintos, sem repetição de combinação visual, e cobertura de todos os quatro níveis de anúncios, ao menos oito composições e dez cabeçalhos.
- A revalidação de 29/10 consta no [relatório](../relatorios/M24-company-websites.md) e em CSV resumido. A busca aquecida teve p95 de 309 ms sequencial e 575 ms concorrente; abrir o site teve 330 ms e 1.369 ms. A latência concorrente do website fica registrada como alerta, não como bloqueador de aceite/SLO, e seus limites estão descritos no relatório.

## Decisões adotadas para implementação

- Usar endereço procedural legível, por exemplo `nomedaempresa.noir.local`, com sufixo estável quando necessário para evitar colisões. O ID da empresa continua sendo a chave canônica da API e do estado de navegação.
- Garantir endereço e contato em todos os sites. `ABOUT`, `SERVICES`/`PRODUCTS` e `ACCOUNT` podem variar por modelo, mas cada perfil deve oferecer um caminho claro de volta à página inicial.
- Exibir o login como elemento visual inerte/desativado, sem formulário nem coleta de dados. O diretório protegido de funcionários usa a autorização investigativa já existente, separada desse login.

