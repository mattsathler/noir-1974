# Estudo: Noir como jogo singleplayer offline

**Status:** estudo, sem implementação  
**Data:** 2 de novembro de 2026

## Resumo

É viável transformar Noir em um jogo singleplayer que funciona sem rede e mantém a partida no computador. A interface React/Vite pode ser reaproveitada. A maior parte do trabalho está no backend e no modelo de persistência: hoje o jogo foi construído como um serviço autoritativo para um mundo compartilhado, ligado a PostgreSQL/Neon, com conta, cookie de sessão, tarefas em background e estado dividido entre mundo e jogador.

Para este repositório, a opção mais alinhada é **Tauri 2 + interface React existente + API Go local como sidecar + SQLite local**. Isso reaproveita o backend e os contratos HTTP. A API continuaria responsável por validar ações e gerar o mundo, mas rodaria no próprio computador e gravaria em um arquivo de save. O shell Tauri empacotaria a interface e o executável Go. O Electron também serviria, mas adiciona o runtime Chromium/Node completo sem resolver a principal migração de dados.

Uma alternativa de menor risco inicial é empacotar o frontend e o Go local, mantendo PostgreSQL embarcado; não recomendamos isso como arquitetura final. Distribuir, iniciar, atualizar e recuperar uma instância PostgreSQL por instalação aumenta muito o custo operacional. A troca para SQLite é o caminho para um jogo local simples de instalar e manter.

## O que o código atual pressupõe

- `apps/web` é uma aplicação React 19 construída pelo Vite. As chamadas de dados passam por `apiRequest` e usam rotas relativas `/api/...`, com cookies same-origin.
- `apps/api` é uma API Go. O README a descreve como autoridade do mundo e requer `DATABASE_URL`; `cmd/server` monta os módulos e inicializa a conexão com PostgreSQL.
- `internal/world` gera o mundo procedural de forma determinística e versionada. Há dados compartilhados por servidor, como mundo, NPCs, empresas, casos e mercado, e dados privados por conta, como localização, progresso, e-mail e carteira.
- Identidade atual usa cadastro/login, senha com bcrypt, sessões por cookie e personagem por conta. Em singleplayer local, essas estruturas não precisam ser a identidade do jogador: um perfil/save local pode representar a partida.
- Os casos podem ser gerados por agenda enquanto o servidor está ativo; há coordenação por PostgreSQL/advisory lock. No offline, o relógio e a agenda precisam ter semântica própria, independente de a aplicação ficar aberta.
- Preferências de interface já usam `localStorage`/`sessionStorage`, mas isso não substitui persistência transacional do estado da partida.
- O backend usa bastante PostgreSQL-specific SQL e tipos/construções do `pgx` (mais de 200 referências encontradas em `apps/api/internal`). Não é uma simples troca de driver: `PostgresStore` e schemas terão de ganhar implementação SQLite ou ser desacoplados por interfaces de repositório.

## Opções avaliadas

| Opção | O que reaproveita | Pontos fortes | Custos e limites | Adequação |
|---|---|---|---|---|
| **Tauri 2 + Go sidecar + SQLite** | React/Vite, API Go e parte dos contratos HTTP | Pacote enxuto; frontend existente; sidecar permite incluir o Go sem runtime instalado pelo jogador; separa UI e backend local | Exige Rust/toolchain Tauri, empacotamento por plataforma, ciclo de vida do sidecar e migração PostgreSQL → SQLite | **Recomendada** |
| **Electron + Go sidecar + SQLite** | React/Vite, API Go e contratos HTTP | Ecossistema desktop maduro; Chromium consistente entre plataformas; integração familiar com JavaScript | Pacotes maiores; atualizações de Electron/Chromium/Node; mesmos custos de SQLite e sidecar; superfície de segurança maior se o renderer ganhar privilégios | Boa se consistência do Chromium e familiaridade da equipe pesarem mais que tamanho |
| **Tauri/Electron + Go + PostgreSQL embarcado** | Quase todo o backend SQL existente | Prova rápida do fluxo desktop com menos mudança nos stores | Instalação e inicialização do banco, portas/processos, upgrades, permissões, backup e recuperação ficam complexos; dependência nativa por sistema operacional | Protótipo técnico apenas |
| **Frontend + SQLite direto via plugin Tauri** | Interface React e conceitos do domínio | Sem servidor HTTP local; arquitetura compacta | Reescreve/adapta acesso a dados e regras em Rust/JS; arrisca duplicar a lógica Go procedural e investigativa; SQL do frontend amplia superfície de acesso | Não recomendada neste estágio |
| **Aplicação web instalável/PWA** | Quase toda a interface | Baixo custo de distribuição; pode cachear assets e oferecer algum uso offline | Não é desktop pleno; armazenamento do navegador e limites de funcionamento em background; backend atual ainda depende de PostgreSQL | Pode servir para demonstrar offline parcial, não para o objetivo completo |

## Arquitetura alvo sugerida

```text
┌──────────────────────────────┐
│ Tauri 2                      │
│ React/Vite (assets locais)   │
│ chamadas HTTP para localhost │
└──────────────┬───────────────┘
               │ API local, loopback
┌──────────────▼───────────────┐
│ Sidecar Noir Go              │
│ rotas e regras de domínio    │
│ geradores / agendador local  │
└──────────────┬───────────────┘
               │
┌──────────────▼───────────────┐
│ SQLite no diretório de dados │
│ profile + world + save       │
└──────────────────────────────┘
```

O sidecar pode abrir uma porta aleatória em loopback e entregar essa URL ao shell; o frontend aguarda o sinal de prontidão antes de fazer requisições. Deve aceitar somente loopback, encerrar junto com o app e não expor uma porta previsível sem proteção. Outra evolução possível é trocar HTTP por comandos Tauri; isso acoplaria mais o frontend ao desktop e não é necessário para a primeira arquitetura.

O plugin SQL do Tauri oferece acesso SQLite e migrations, mas aqui a API Go deve continuar sendo a proprietária do banco. A documentação do plugin demonstra integração SQL no frontend; adotar esse acesso direto em paralelo ao Go criaria duas camadas escrevendo o mesmo save. Se o sidecar usar SQLite, sua biblioteca Go/driver e suas migrations devem ser a autoridade única.

## Mudanças funcionais necessárias

### 1. Partida e identidade

Substituir cadastro/login por uma tela local de nova partida/continuar partida. A primeira inicialização cria um identificador local e um save. Manter autenticação local por senha é uma escolha opcional de privacidade, não requisito de gameplay; sem contas online, recuperação de senha e sessão remota perdem função.

Separar explicitamente `world_id` e `save_id`: cada instalação ou slot pode ter sua própria seed, cidade, personagens, casos, progressão, preferências e relógio. Isso evita que a seed fixa de desenvolvimento e os identificadores de servidor compartilhado continuem atuando como identidade da campanha.

### 2. Estado e banco

Criar uma camada de armazenamento com implementação SQLite. Preservar tabelas e invariantes de domínio onde forem portáveis; adaptar SQL, índices, migrações, locks, transações, JSON e tipos de data/UUID. SQLite atende bem a um jogador e gravação local; as transações continuam importantes para ações como compra de ações, criação de caso e progresso investigativo.

Rever rotinas com semântica de servidor: advisory locks, cache distribuído, métricas de multi-instância, tabelas particionadas ou consultas específicas do Postgres. Em modo singleplayer, há uma única instância e podem existir soluções locais mais simples. IDs e sementes devem permanecer determinísticos para reconstrução de conteúdo procedural, mas alterações decorrentes da partida precisam ser gravadas.

### 3. Tempo, agenda e pausa

Definir se o relógio do mundo avança apenas enquanto o jogo está aberto ou também enquanto está fechado. Recomendação: tempo de jogo avança por ações/tempo em execução e pausa ao fechar, evitando que o jogador perca janelas investigativas enquanto está offline. Gerar casos agendados usando slots do relógio do jogo, não cronômetro de parede. Ao reabrir, o backend reconcilia a agenda até o horário simulado atual, com limite para evitar processamento excessivo.

### 4. Conteúdo e dependência de rede

Auditar fontes externas, links e qualquer uso futuro de serviços remotos. Assets, fontes, sons, catálogos e seeds necessários à campanha precisam estar empacotados. Links externos (por exemplo, páginas institucionais diegéticas) podem continuar abrindo dentro do app se forem conteúdo local; nenhum fluxo essencial pode pressupor HTTP externo, Neon, login online ou telemetria.

### 5. Save, backup e compatibilidade

Usar diretório de dados da aplicação do sistema operacional, nunca o diretório de instalação. Versionar o schema e manter backup automático antes de migrações. Oferecer exportar/importar save; decidir se o arquivo será SQLite puro ou pacote com metadados. Migrações devem preservar saves entre versões e falhar sem destruir o original.

## Escopo de migração em etapas

Estimativas abaixo são ordens de grandeza para planejamento, não prazos comprometidos. Pressupõem que o escopo de conteúdo do jogo permaneça estável durante o trabalho.

1. **Prova de desktop (1–2 semanas):** empacotar frontend num Tauri, lançar API Go de desenvolvimento como processo filho, carregar um mundo de teste e confirmar build/execução em um SO alvo. Ainda pode usar PostgreSQL de desenvolvimento; não conta como offline pronto.
2. **Camada de persistência (2–4 semanas):** mapear queries por módulo, definir interfaces e portar schema/queries essenciais de identidade, mundo, casos, registros, mail, progressão e whiteboards para SQLite. Criar save isolado por instalação e testes de compatibilidade do domínio.
3. **Modo de partida local (1–3 semanas):** remover login remoto do caminho principal, criar/abrir save, definir seed, identidade do detetive, relógio e agenda local. Rever tarefas agendadas e operações que assumem vários jogadores/instâncias.
4. **Robustez e distribuição (2–4 semanas):** empacotar Tauri/sidecar para os sistemas alvo, validar inicialização/encerramento, migração, backup/importação, crash recovery, armazenamento e atualizações. Revisar segurança da porta local e permissões Tauri.

Faixa preliminar: **6–13 semanas de engenharia**, dependendo de cobertura de plataforma, estado real dos stores e expectativa de importação de contas/saves. A migração SQL é a principal incerteza; um spike só em um módulo representativo (mundo + caso + progresso) reduziria a incerteza antes de comprometer o restante.

## Riscos e decisões em aberto

- **Escopo offline:** “sem internet” precisa incluir instalação e primeira inicialização sem baixar dependências ou conteúdo. O instalador deve incluir runtime e assets necessários.
- **Portabilidade de save:** caso o jogo continue podendo ser multiplayer no futuro, definir desde já quais IDs e fatos de campanha podem ser sincronizados. Se singleplayer for uma edição separada, pode simplificar sem prometer compatibilidade com contas online.
- **Concorrência e atomicidade:** eliminar locks multi-instância não significa remover transações; integridade de ações dentro do save continua central.
- **Proteção contra adulteração:** singleplayer não precisa esconder a verdade do próprio dono do computador. Não replicar mecanismos anti-cheat server-side; manter validações para impedir saves inconsistentes e crashes.
- **Plataformas:** Tauri usa webviews do sistema operacional (WebView2 no Windows, WKWebView no macOS, WebKitGTK no Linux); isso reduz bundle, mas demanda validação visual e de comportamento por plataforma. Electron traz Chromium/Node junto, aumentando consistência e tamanho.
- **Dados existentes:** ainda não há estratégia definida para converter contas Neon em saves locais. Migração de dados deve ser opcional e explícita se for necessária.

## Recomendação

Fazer um spike técnico pequeno com **Tauri 2 + sidecar Go + SQLite**, começando por um fluxo vertical (abrir app → criar save → carregar mundo → avançar um caso → fechar e reabrir mantendo progresso). Antes de migrar módulos em massa, medir o trabalho para portar o `PostgresStore` de mundo e o armazenamento de casos. Continuar com Electron apenas se o custo de WebView por sistema operacional se mostrar maior que o custo de empacotar Chromium ou se a equipe preferir fortemente a stack Electron.

Manter a fronteira HTTP no início reduz mudanças simultâneas: `apiRequest` e a maior parte dos componentes podem continuar falando com endpoints, enquanto identidade e armazenamento mudam do lado do backend. O protocolo pode permanecer privado e local, sem representar promessa de serviço remoto.

## Direção futura: mundos pessoais e mundos acessíveis

Uma direção futura compatível com a arquitetura online atual é oferecer mundos/campanhas em vez de um único mundo global. Cada jogador pode criar uma campanha pessoal, com seed e mutations isoladas, e também receber acesso a campanhas de outras pessoas. Isso não exige que todos os jogadores estejam juntos no mesmo mundo: cada campanha tem seus participantes, estado e regras de acesso.

O conceito central passa a ser `world_id` (ou `campaign_id`), separado de `account_id`. Um mundo tem dono, seed, versão de gerador e estado compartilhado; membros são contas com um nível de acesso. Um jogador pode possuir vários mundos e participar de outros. A conta mantém identidade global, enquanto personagem, progresso investigativo, descobertas e preferências de campanha são vinculados à combinação mundo + membro.

```text
ACCOUNT ── owns ── WORLD/CAMPAIGN ── seed + world mutations
   │                    │
   └── member ──────────┘
                        └── per-member progress / discoveries
```

Isso preserva os dois usos desejados: campanha privada de um jogador e acesso a uma campanha de outra pessoa. Não cria automaticamente um shard global nem implica que participantes tenham de atuar simultaneamente. O backend atual já possui separação parcial entre estado compartilhado por `server_id` e progresso por conta; conceitualmente, `world_id` pode ocupar o lugar desse escopo de servidor. A complexidade está em tornar essa fronteira uniforme em todas as tabelas, consultas e jobs.

### Estado compartilhado e estado individual

Antes de implementar, cada domínio precisa ser classificado:

| Compartilhado dentro do mundo | Privado por membro do mundo |
|---|---|
| Seed e versão do gerador; NPCs, relações e empresas; casos e mutations canônicas do mundo; eventos do mercado caso afetem a campanha | personagem/identidade de jogo; progresso do caso; provas e anotações descobertas; whiteboards; caixa de entrada pessoal; posição, saldo e carteira se a economia for individual |

Essa tabela é uma hipótese inicial, não uma regra definitiva. Em particular, “descoberta” pode ser pessoal (cada detetive investiga por conta própria) mesmo quando o fato subjacente — por exemplo, o corpo ter sido encontrado — altera o mundo compartilhado. A implementação deve separar o evento do mundo da confirmação de que cada membro o descobriu.

### Regras necessárias para acesso a outro mundo

- Dono e membros: definir quem pode convidar, remover membros, alterar regras e encerrar ou arquivar a campanha.
- Permissões: começar com papéis simples, como dono e participante. Evitar edição direta de seed/estado depois de iniciar a campanha.
- Concorrência: transações e idempotência continuam necessárias quando duas pessoas alteram a mesma campanha; o isolamento por mundo evita colisões entre campanhas, mas não entre membros do mesmo mundo.
- Tempo: decidir se o relógio da campanha avança por ações, enquanto alguém está conectado, ou em tempo real. Uma campanha assíncrona tende a funcionar melhor que exigir presença simultânea.
- Acesso e privacidade: validar membership em cada rota e incluir `world_id` no escopo da autorização; IDs difíceis de adivinhar não substituem checagem de acesso.
- Compartilhamento: definir se progresso individual revela automaticamente pistas/notícias ao grupo ou se cada membro precisa obtê-las.

Uma evolução segura é primeiro criar o conceito de mundo privado de um jogador e mover para ele os dados hoje escopados por `server_id`; depois, adicionar membership e convites para permitir acesso a mundos alheios. Assim o singleplayer nasce como caso simples de campanha e a partilha posterior não exige inventar outro modelo de save.

Essa direção é online, mesmo que uma campanha pessoal tenha apenas um membro. Não equivale a modo offline: login, rede e persistência remota continuam necessários. O empacotamento desktop e SQLite descritos acima são uma opção separada, que poderia ser considerada depois caso o requisito offline permaneça.

## Referências

- Código e documentação do repositório: `apps/api/README.md`, `docs/architecture/api.md`, `docs/architecture/geracao-do-mundo.md`, `apps/api/cmd/server/main.go`, `apps/web/src/shared/apiRequest.js`.
- Tauri 2: [início e arquitetura](https://v2.tauri.app/start/), [SQL plugin](https://v2.tauri.app/plugin/sql/), [external binaries/sidecar](https://v2.tauri.app/develop/sidecar/), [segurança](https://v2.tauri.app/security/), [webviews por plataforma](https://v2.tauri.app/reference/webview-versions/).
- Electron: [modelo de processos](https://www.electronjs.org/docs/latest/tutorial/process-model), [guia de segurança](https://www.electronjs.org/docs/latest/tutorial/security).
