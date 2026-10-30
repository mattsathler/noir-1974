# M22 — Ghostkey: exploração de provedores

**Trilha:** Independente do fluxo criminal  
**Status:** Substituída pela M23; fluxo Ghostkey entregue como parte da M23
**Depende de:** M07, M08 e M12 aceitos (fontes de dados e acesso a sistemas locais existentes)

> Esta proposta de navegação isolada em computadores locais foi substituída pela migração integrada para sistema operacional, Proxy e sites da [M23](M23-computador-os-proxy-e-sites.md). O fluxo Ghostkey entregue pertence à M23 e usa provedores e sites remotos; não implementa o minigame de senha nem o explorador local descritos nesta proposta.

## Registro da substituição

A M23 incorporou o Ghostkey como ferramenta de exploração de provedores. O aplicativo abre com aviso de autorização do MeridOS, apresenta o Network Explorer e permite navegar por provedores, endereços IP, pastas virtuais e executáveis. Arquivos executáveis restritos — Person Search da polícia, Patient Search do hospital, Employee Directory de empresas e GATE Access Logs — são acessíveis por esse fluxo, com validação no servidor. A interface usa efeitos visuais de acesso comprometido, mantendo a moldura e o ambiente comuns do MeridOS.

Essa entrega substitui a intenção geral de usar Ghostkey para alcançar sistemas protegidos, mas não os requisitos originais de senha, arquivos locais de NPC, banco, `NETWORK STATUS` ou `PRINT QUEUE`. O escopo e os critérios vigentes estão registrados na M23.

## Objetivo

Refatorar por completo o bypass dos computadores locais. Nos computadores invadidos, substituir a grade de aplicativos por um explorador de pastas e arquivos. O jogador executa `ghostkey.exe`, criado por The Crow, e resolve um minigame de senha antes de navegar pelos arquivos disponíveis naquele computador.

A investigação passa a exigir que o jogador navegue manualmente até a pasta da pessoa certa e abra seu arquivo. As entradas são listadas em ordem alfabética e não há busca por nome; onde fizer sentido, o único filtro de conteúdo é um intervalo de datas.

Esta milestone é uma trilha independente: não altera a contagem nem o aceite do fluxo criminal M00–M13 e não depende das milestones futuras M17, M18 ou M20.

## Escopo

### Ghostkey e autorização

- Projetar o acesso local por meio do arquivo executável `ghostkey.exe`, creditado a The Crow, sem comando de terminal.
- Ao entrar no computador invadido, abrir um explorador inicial com `ghostkey.exe` e as pastas protegidas visíveis, porém bloqueadas. Abrir o executável inicia o desafio; resolvê-lo destranca as pastas sem trocar para uma grade de aplicativos.
- Abrir o executável inicia um desafio de senha com quatro campos numéricos. A senha aceita zeros à esquerda e cada palpite deve conter exatamente quatro dígitos.
- Depois de cada palpite, cada campo recebe um dos três resultados: vermelho quando o dígito não aparece na senha; amarelo quando aparece em outra posição; verde quando está na posição correta. A validação considera dígitos repetidos sem contabilizar uma ocorrência mais de uma vez.
- O cliente nunca recebe a senha. O desafio e os resultados são autoritativos no backend e vinculados à conta, ao computador e à sessão de acesso atuais. Somente quatro campos verdes liberam os arquivos protegidos.
- As respostas têm rótulos acessíveis além da cor. Falha de rede, desafio expirado ou resposta inválida não libera acesso nem deixa a interface presa.

### Explorador de computadores invadidos

- Após a autorização, apresentar uma raiz de arquivos com navegação por pastas, breadcrumb, retorno e abertura de arquivos. Os sistemas locais invadidos deixam de apresentar grade de aplicativos.
- Representar os arquivos como conteúdo virtual do jogo, sem acesso ao sistema de arquivos real do navegador ou do servidor.
- Organizar pastas de NPCs e arquivos `.txt` alfabeticamente pelo nome completo do NPC. Não oferecer busca por nome, campo de consulta, filtro por pessoa, categoria ou relevância.
- Buscar o conteúdo de cada arquivo ao abri-lo, em vez de carregar os registros completos de todos os NPCs na abertura do computador.
- Validar no backend conta, cidade e país persistidos, sistema local, sessão de acesso e NPC solicitado. A interface não escolhe nem amplia a jurisdição do computador.

### Banco

- A raiz contém pastas para cada NPC do país atual que tenha cadastro bancário.
- Cada pasta contém o arquivo `.txt` do ledger daquela pessoa. Abrir o arquivo mostra registros bancários do intervalo de datas selecionado.
- A listagem é completa para o país e ordenada alfabeticamente; não há busca por nome. O filtro de intervalo de datas é o único filtro do ledger.

### Polícia

- A raiz oferece áreas de `PERSON SEARCH` e `EMAIL ARCHIVE`, cada uma com pastas de NPCs do país atual.
- A pasta do NPC contém um arquivo de perfil para o Person Search e um arquivo de correspondência para os emails disponíveis.
- O perfil mantém as informações já oferecidas pelo Person Search, incluindo a cena reportada quando aplicável. O arquivo de emails mantém remetente, destinatário, assunto, conteúdo e datas já disponíveis na fonte atual.
- A navegação é manual e alfabética; não existe busca por nome. Person Search não ganha filtros novos. No arquivo de emails, o intervalo de datas é o único filtro.

### Demais sistemas locais e imersão

- Migrar os outros computadores hoje acessados por sistemas locais — GATE, empresas e Hospital — para a mesma navegação por pastas e arquivos. Preservar suas fontes, permissões e conteúdo atuais; cada integração deve documentar sua árvore e seus arquivos.
- Incluir a pasta `NETWORK STATUS` com um toggle funcional que liga e desliga a conexão simulada daquele computador. Desligada, impede listar, abrir ou atualizar arquivos remotos; pastas locais da máquina continuam navegáveis. Pode ser ligada novamente sem reiniciar o jogo.
- Incluir a pasta `PRINT QUEUE` com alguns documentos fictícios e aleatórios para ambientação. Abrir ou marcar um documento como impresso só altera o estado visual local da fila e não cria uma pista nem altera o backend investigativo.
- O computador pessoal do jogador, incluindo Crime Reports e sua caixa de email, permanece fora desta refatoração.

## Critérios de aceite

- [x] Antes de abrir `ghostkey.exe`, o MeridOS mostra o aviso de aplicativo não autorizado; o app inicia com boot interno e oferece o Network Explorer.
- [x] O Network Explorer lista provedores corporativos do país conectado e navega por pastas virtuais, sem linguagem ou comportamento de FTP.
- [x] Abrir um provedor carrega seus diretórios de hosts. Os endereços IP dos sites empresariais são determinísticos a partir da empresa e podem ser encontrados sem visita prévia pelo navegador.
- [x] Os hosts são apresentados em páginas de 20 itens, com filtro por endereço IP que ignora pontos; uma busca vazia restaura a listagem e a consulta só é enviada por ação explícita.
- [x] As pastas usam nomes genéricos e contêm arquivos plausíveis. Executáveis têm nomes técnicos e ícones próprios; tentar abrir um arquivo não suportado mostra o modal de erro padrão do MeridOS.
- [x] Person Search, Residential Search & Seizure e Arrest Warrant da polícia, Patient Search do Hospital, Employee Directory empresarial e GATE Access Logs são acessíveis pelos executáveis encontrados no Ghostkey e continuam sujeitos à autorização validada pelo servidor.
- [x] Os apps abertos pelo Ghostkey recebem efeitos visuais de acesso comprometido, mantendo a interface e a moldura comuns do MeridOS.
- [x] Carregamentos internos ficam limitados à janela/app correspondente; o blank loading de página inteira é usado apenas para navegação entre páginas.
- [x] A navegação é virtual e não expõe o sistema de arquivos real do computador ou do servidor.

## Fora de escopo

Alterar ou ampliar o conteúdo investigativo dos registros, criar novos sistemas de autenticação reais, implementar acesso ao sistema de arquivos do dispositivo, transformar a senha do minigame em segredo compartilhado entre jogadores, alterar permissões ou regras dos serviços, mudar o computador pessoal do detetive e adicionar documentos de impressão como evidência ou progresso de caso.

## Sequência sugerida

1. Implementar o desafio Ghostkey e o explorador virtual com autorização server-side.
2. Migrar Banco e Polícia, priorizando navegação nacional por pastas e abertura sob demanda.
3. Migrar GATE, diretórios de empresa e Hospital preservando os contratos existentes.
4. Adicionar `NETWORK STATUS`, `PRINT QUEUE`, estados vazios/erro e acessibilidade.
5. Remover o fluxo antigo, revisar autorização e cobrir regressões dos sistemas locais.
