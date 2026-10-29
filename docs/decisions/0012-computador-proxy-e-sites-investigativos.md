# ADR 0012 — Computador pessoal, proxy e investigação por sites

- **Status:** Aceita
- **Data:** 2026-10-29
- **Supersede:** [ADR 0010 — Geografia fixa e bairros procedurais](0010-geografia-fixa-e-bairros-procedurais.md), [ADR 0011 — Prédios e empresas como entidades distintas](0011-predios-e-empresas.md)

## Contexto

A fantasia do jogo passa a ser a de um detetive particular que investiga usando seu computador pessoal. Viajar fisicamente pelo GATE para trocar de cidade acrescenta uma etapa de apresentação, mas a jurisdição já é validada pelo backend e pode ser selecionada por uma conexão proxy.

A investigação precisa continuar limitada pelas regras atuais de conta, cidade, serviço, empresa e autorização. A mudança altera a forma de apresentar e navegar pelas fontes; não transforma as fontes em dados globais irrestritos.

## Decisão

- A tela principal do jogo passa a ser o sistema operacional do computador pessoal do detetive, com navegador, janelas e ferramentas investigativas.
- O GATE continua existindo como infraestrutura física de transporte no mundo. NPCs continuam viajando, e registros GATE permanecem disponíveis como fonte investigativa. O protagonista não usa viagem física como fluxo de navegação.
- Um aplicativo Proxy conecta a sessão a um nó de cidade. O nó selecionado define a jurisdição ativa que hoje é representada pela cidade atual do jogador. Mudar de nó altera o contexto investigativo, sem alegar presença física.
- Um portal nacional de empresas pesquisa e lista até dez empresas por consulta no país conectado. Os resultados são resumidos e identificados também pela cidade. A busca não carrega funcionários.
- Abrir o site de uma empresa carrega somente o diretório daquela empresa. Funcionários continuam vinculados à empresa selecionada e ao acesso autorizado correspondente.
- Person Search continua nacional dentro do país conectado e retorna até dez pessoas. Informações de caso associadas ao resultado devem ser resolvidas na cidade do próprio NPC, sem ampliar os dados retornados além das permissões atuais.
- Dados de banco, GATE, email, hospital, casos e serviços permanecem sujeitos aos seus escopos e autorizações existentes. Consultas por NPC carregam somente o alvo e o intervalo solicitados.
- Ao abrir um site corporativo associado a outra cidade, o cliente deve conectar o Proxy ao nó dessa cidade antes de solicitar dados restritos. O acesso continua validado no servidor por conta, cidade, site/endereço, empresa e tentativa autorizada; trocar de nó não concede autorização.
- Sites devem parecer produtos de empresas e instituições do universo de Noir. A rede pode ser extraordinária em alcance e capacidade, mas sua apresentação preserva a estética e as convenções antiquadas do cenário.

## Consequências

- O catálogo de países, cidades, fusos, bairros e geração procedural permanece válido; sua função principal passa de destino de viagem do protagonista a contexto de jurisdição e geração.
- A navegação física por Neighborhood, interior de prédios e transição de viagem deixa de ser a entrada dos fluxos investigativos. Endereços e empresas continuam existindo como dados e contexto de autorização.
- O portal nacional de empresas precisa de uma consulta por país que retorne registros leves e limite resultados antes de carregar diretórios de funcionários. A busca atual tem limite de cinco e inclui funcionários, portanto não atende diretamente ao novo contrato.
- A busca de pessoas já possui caminho nacional por país. A associação de cenas continua limitada à cidade ativa e deve ser corrigida para cobrir resultados de outras cidades do mesmo país.
- O campo persistido hoje usado como cidade do jogador pode continuar armazenando a jurisdição proxy ativa durante a migração. O contrato deve passar a nomear claramente essa semântica, sem tratar a conexão como presença física.
- Os endpoints de acesso e de dados restritos precisam aceitar o contexto de site remoto depois da conexão proxy, conservando validações server-side e a expiração atual das autorizações.
- A M22 de exploração local com Ghostkey é substituída pela migração integrada do computador, do Proxy e dos sites descrita na M23. Não se inicia uma segunda interface de arquivos locais em paralelo.
- O desempenho da lista nacional de empresas deve ser medido separadamente da busca nacional de pessoas. A linha de base histórica de Person Search não prediz o custo de uma listagem resumida de empresas.
