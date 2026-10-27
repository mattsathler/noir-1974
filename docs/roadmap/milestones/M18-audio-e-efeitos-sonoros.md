# M18 — Áudio e efeitos sonoros

**Ciclo:** 18  
**Status:** Não iniciado  
**Depende de:** M17 aceito

## Objetivo

Dar presença e resposta sensorial às ações do jogo com uma camada de áudio funcional, discreta e coerente com Noir 1974. Efeitos sonoros curtos devem confirmar interações e transições — como telefonar, usar o GATE, invadir um computador e iniciar aplicativos — sem competir com a leitura, revelar informação investigativa ou transformar o jogo em uma experiência ruidosa. O áudio acompanha o resultado confirmado pela interface e pelo backend; não substitui feedback textual ou visual.

## Direção de áudio

- A identidade sonora combina tecnologia analógica e digital de 1974, ruído urbano contido e uma atmosfera investigativa de suspense. Telefones, terminais, mecanismos e interfaces devem soar táteis e específicos ao contexto, evitando efeitos genéricos de ficção científica futurista.
- Sons são breves, reconhecíveis e usados com parcimônia. O silêncio continua sendo parte da atmosfera; não há reprodução contínua de efeitos por telas paradas ou ações repetidas.
- Cada ação pode ter sinais distintos para início, progresso quando necessário, sucesso, falha e cancelamento. Só se usam as etapas que acrescentam informação útil; ações instantâneas não recebem sequências artificiais longas.
- A resposta sonora respeita o estado autoritativo da ação. Um som de confirmação de sucesso só toca depois da confirmação; falhas, interrupções e cancelamentos não devem soar como conclusão bem-sucedida.

## Escopo de interações

O catálogo inicial cobre pelo menos:

- **Telefonar:** tirar do gancho ou iniciar chamada, tom de discagem/conexão e encerramento; ocupado, sem resposta ou chamada cancelada recebem sinais breves próprios quando esses resultados existirem no fluxo.
- **Usar o GATE:** ativação e confirmação da viagem, com assinatura sonora curta de transporte; a chegada só é sinalizada quando a transição for confirmada. Erros de destino e cancelamentos não usam o som de chegada.
- **Hackear um computador:** início do acesso, atividade curta enquanto a ação aguarda ou processa e sinais distintos para acesso concedido, negado, interrompido ou falho. O áudio não pode indicar se um dado é pista, crime ou informação oculta.
- **Dar BOOT em aplicativos:** ligar/inicializar o terminal e abrir ou fechar aplicativos com sons próprios, curtos e consistentes. Aplicativos diferentes podem compartilhar uma família de sons; não é necessário criar um efeito único para cada tela.
- **Transições visuais do M15:** a chegada pelo GATE e a cena de viagem entre bairros recebem efeitos sincronizados com os estados confirmados, sem antecipar destinos ou resultados. Ações policiais só entram no escopo de áudio quando suas animações forem especificadas.

Outras interações podem reutilizar efeitos dessa biblioteca quando fizer sentido, mas não ampliam o escopo obrigatório sem revisão do milestone.

## Critérios de aceite

- [ ] Existe uma biblioteca organizada de efeitos sonoros originais ou devidamente licenciados para as interações do escopo, com nomes, finalidade, estado correspondente e formato documentados.
- [ ] A identidade sonora é coerente com o período, a ambientação investigativa e a direção de Noir 1974; não imita trilhas, efeitos ou marcas reconhecíveis de outra obra.
- [ ] Telefonar, usar o GATE, hackear computador e dar BOOT/abrir aplicativos têm feedback sonoro identificável nos estados aplicáveis, incluindo falha e cancelamento quando relevantes.
- [ ] As transições do GATE e a cena de viagem entre bairros do M15 têm áudio sincronizado com a confirmação e a progressão visual, sem anunciar sucesso antes da resposta autoritativa.
- [ ] Nenhum efeito sonoro revela classificação interna de pista, relação com crime, identidade do culpado, destino não confirmado ou qualquer dado oculto ao jogador.
- [ ] A camada de áudio é acionada por eventos e estados explícitos da interface; re-renderizações, reabertura de componentes e respostas duplicadas não repetem efeitos acidentalmente.
- [ ] Interações repetidas não acumulam sons, cortam sons críticos de confirmação ou geram reprodução contínua. Cada efeito tem duração limitada e comportamento definido para sobreposição e interrupção.
- [ ] O jogador pode silenciar efeitos e ajustar seu volume em controles acessíveis e persistentes entre sessões; o estado inicial respeita as políticas de áudio do navegador e não inicia áudio sem interação permitida.
- [ ] Sons não são o único canal de comunicação: ações, resultados, falhas e estados continuam legíveis por texto e sinais visuais, inclusive com áudio silenciado ou indisponível.
- [ ] Os controles e feedback de áudio são utilizáveis por teclado e leitor de tela, e as preferências não bloqueiam ações do jogo.
- [ ] Os assets carregam sob demanda ou com estratégia documentada, sem atrasar ações, navegação ou o primeiro uso do jogo; falhas de carregamento degradam silenciosamente para o feedback visual e textual.
- [ ] Há validação demonstrável de sincronização, volume, repetição, silenciamento, persistência da preferência e comportamento sem suporte ou com falha de reprodução nos fluxos cobertos.

## Fora de escopo

Trilha sonora contínua ou adaptativa, dublagem, áudio espacial, simulação acústica de ambientes, composição musical dinâmica, narração de textos, sons exclusivos para todas as telas e reconstrução sonora de ações que ainda não existem no jogo. Alterações de regras, persistência ou resultados do backend também ficam fora deste ciclo.
