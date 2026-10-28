# M22 — Ghostkey: invasão e exploração de arquivos locais

**Trilha:** Independente do fluxo criminal  
**Status:** Não iniciado  
**Depende de:** M07, M08 e M12 aceitos (fontes de dados e acesso a sistemas locais existentes)

## Objetivo

Refatorar por completo o bypass dos computadores locais. Nos computadores invadidos, substituir a grade de aplicativos por um explorador de pastas e arquivos. O jogador executa `ghostkey.exe`, criado por The Crow, e resolve um minigame de senha antes de navegar pelos arquivos disponíveis naquele computador.

A investigação passa a exigir que o jogador navegue manualmente até a pasta da pessoa certa e abra seu arquivo. As entradas são listadas em ordem alfabética e não há busca por nome; onde fizer sentido, o único filtro de conteúdo é um intervalo de datas.

Esta milestone é uma trilha independente: não altera a contagem nem o aceite do fluxo criminal M00–M13 e não depende das milestones futuras M17, M18 ou M20.

## Escopo

### Ghostkey e autorização

- Substituir `bypassdecrypt.exe`, o comando de terminal e as animações atuais pelo arquivo executável `ghostkey.exe`, creditado a The Crow.
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

- [ ] Nenhum computador local invadido usa a interface de grade de aplicativos ou o fluxo `bypassdecrypt.exe`; os sistemas listados nesta milestone são navegados por pastas e arquivos.
- [ ] Ao entrar em um computador invadido, o jogador encontra o explorador inicial e pode abrir `ghostkey.exe`; as pastas protegidas só ficam acessíveis depois da senha correta.
- [ ] `ghostkey.exe`, identificado como criação de The Crow, abre um minigame com quatro campos numéricos e impede o acesso até a solução.
- [ ] Cada campo retorna somente um dos três estados especificados; dígitos repetidos, senha com zero inicial, palpite malformado, desafio expirado e respostas fora de ordem têm comportamento definido e coberto.
- [ ] A senha permanece no backend. Uma sessão, conta, cidade, país, sistema ou computador diferentes não podem reutilizar a autorização de outra sessão.
- [ ] Cada pasta e arquivo de NPC é apresentado em ordem alfabética. Não existe pesquisa por nome; os arquivos com conteúdo temporal aceitam somente filtro por data.
- [ ] O Banco mostra todos e somente os NPCs cadastrados naquela jurisdição, abre os ledgers individuais e conserva os registros e limites temporais definidos pelas fontes existentes.
- [ ] A Polícia disponibiliza Person Search e emails por NPC, preserva o conteúdo e o contexto atuais e não mistura dados de outra jurisdição.
- [ ] GATE, empresas e Hospital mantêm suas operações atuais após migrarem para estruturas de pastas e arquivos; o acesso continua restrito ao sistema local autorizado.
- [ ] Consultas de conteúdo são feitas sob demanda ao abrir o arquivo. Listas nacionais não carregam todos os registros detalhados no primeiro acesso e não são truncadas silenciosamente.
- [ ] Alterar cidade/país, perder autorização, fechar o computador ou expirar a sessão impede consultas posteriores com contexto antigo.
- [ ] O toggle de `NETWORK STATUS` bloqueia listagem, abertura e atualização de arquivos remotos naquele computador, preserva pastas locais e restabelece a conexão quando religado.
- [ ] A `PRINT QUEUE` sempre apresenta alguns documentos de ambientação; sua interação não altera casos, evidências, descobertas ou persistência do mundo.
- [ ] Pastas, arquivos, respostas do minigame e estados de carregamento, vazio, erro, offline e sucesso funcionam por teclado e leitor de tela, sem depender apenas de cor.
- [ ] Os fluxos antigos de banco, GATE, email, Person Search, diretório de equipe e Hospital mantêm a integridade dos dados e as validações de autorização após a migração.

## Fora de escopo

Alterar ou ampliar o conteúdo investigativo dos registros, criar novos sistemas de autenticação reais, implementar acesso ao sistema de arquivos do dispositivo, transformar a senha do minigame em segredo compartilhado entre jogadores, alterar permissões ou regras dos serviços, mudar o computador pessoal do detetive e adicionar documentos de impressão como evidência ou progresso de caso.

## Sequência sugerida

1. Implementar o desafio Ghostkey e o explorador virtual com autorização server-side.
2. Migrar Banco e Polícia, priorizando navegação nacional por pastas e abertura sob demanda.
3. Migrar GATE, diretórios de empresa e Hospital preservando os contratos existentes.
4. Adicionar `NETWORK STATUS`, `PRINT QUEUE`, estados vazios/erro e acessibilidade.
5. Remover o fluxo antigo, revisar autorização e cobrir regressões dos sistemas locais.
