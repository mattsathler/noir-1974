# M09 — Deslocamento e pontos de serviço

**Ciclo:** 9  
**Status:** Aceito
**Depende de:** M08 aceito

## Objetivo

Permitir que o detetive consulte os bairros e locais da cidade e estabeleça os dados usados pela navegação local. A cidade atual é persistente por jogador; os prédios e pontos de serviço são compartilhados e estáveis para o servidor. A escolha de bairro ou prédio é transitória e não registra uma localização própria ou qual construção o jogador está visitando. A M11 integra a escolha do bairro ao fluxo de viagem do GATE.

## Pontos fixos de cada cidade

Toda cidade do catálogo compartilhado possui um bairro **Central** fixo, com os mesmos prédios principais:

- **Police Station (Delegacia)** — serviços e registros policiais;
- **Bank (Banco)** — atendimento e serviços bancários;
- **GATE Central** — acesso à rede de transporte GATE;
- **Hospital** — prédio visível, mas fechado nesta fase e reservado para uma feature futura.

Esses pontos existem em toda cidade e não são bairros procedurais. Sua identidade deve ser estável e comum aos jogadores do mesmo servidor.

## Critérios de aceite

- [x] Cada cidade possui em Central Police Station, Bank, GATE Central e Hospital estáveis e compartilhados no servidor; o Hospital permanece fechado para uma feature futura.
- [x] A cidade atual do jogador é persistida quando ele viaja pelo GATE; abrir ou fechar um prédio não grava uma visita nem altera essa cidade.
- [x] O cliente consegue consultar os bairros e locais disponíveis na cidade; o bairro escolhido é contexto transitório e não persiste uma localização própria no perfil.
- [x] O jogador consegue abrir a representação visual dos prédios disponíveis pela interface principal; a escolha do prédio é estado transitório da interface.
- [x] O cliente informa a cidade persistida e apresenta bairros e prédios disponíveis nela; essas seleções de navegação são transitórias e não há localização persistida por construção.
- [x] Consultas bancárias permanecem disponíveis sem visita presencial até o fluxo de visita a serviços estar disponível ao jogador; a exigência de presença pode ser ativada quando esse fluxo for integrado ao gameplay.
- [x] A Delegacia é um prédio fixo e visualmente acessível; ações futuras verificam a cidade atual e o contexto do prédio enviado na solicitação, sem depender de uma visita salva.
- [x] Dois jogadores do mesmo servidor veem os mesmos pontos fixos, embora possam estar em bairros diferentes.
- [x] Testes cobrem identidade dos prédios principais, persistência da cidade, viagem e seleção visual de prédios sem mutação persistida.

## Fora de escopo

Mapa físico de bairros, coordenadas, distâncias entre bairros, rotas locais, persistência do prédio atualmente aberto, combate durante viagens e economia detalhada de tarifas ou tempo de deslocamento. A busca e apreensão em si permanece no M12, que valida a cidade e o local da ação sem depender de uma visita persistida.
