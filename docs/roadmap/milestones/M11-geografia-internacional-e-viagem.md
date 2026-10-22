# M11 — Geografia internacional e viagem

**Ciclo:** 11  
**Status:** Aceito
**Depende de:** M10 aceito

## Objetivo

Expandir o mundo compartilhado de Noir para um catálogo fixo de 25 países e 100 cidades, comum a todos os servidores. O app GATE transporta jogadores somente entre cidades, inclusive cruzando fronteiras nacionais. A navegação por bairros e a visita a construções pertencem a interfaces próprias, separadas do app GATE; bairros continuam sendo divisões lógicas locais, sem mapa físico.

## Catálogo inicial

- **Estados Unidos:** Nova York, Washington, D.C., Los Angeles, San Francisco
- **China:** Pequim, Xangai, Shenzhen, Guangzhou
- **Rússia:** Moscou, São Petersburgo, Novosibirsk, Vladivostok
- **Reino Unido:** Londres, Manchester, Birmingham, Edimburgo
- **Alemanha:** Berlim, Frankfurt, Munique, Hamburgo
- **Japão:** Tóquio, Osaka, Yokohama, Kyoto
- **França:** Paris, Lyon, Marselha, Toulouse
- **Índia:** Nova Délhi, Mumbai, Bengaluru, Hyderabad
- **Coreia do Sul:** Seul, Busan, Incheon, Daegu
- **Arábia Saudita:** Riade, Jidá, Meca, Medina
- **Canadá:** Toronto, Vancouver, Montreal, Ottawa
- **Brasil:** São Paulo, Rio de Janeiro, Brasília, Belo Horizonte
- **Itália:** Roma, Milão, Veneza, Turim
- **Austrália:** Sydney, Melbourne, Brisbane, Perth
- **Espanha:** Madri, Barcelona, Valência, Sevilha
- **Turquia:** Istambul, Ancara, Izmir, Antália
- **México:** Cidade do México, Guadalajara, Monterrey, Cancún
- **Indonésia:** Jacarta, Surabaya, Bandung, Denpasar
- **Emirados Árabes Unidos:** Dubai, Abu Dhabi, Sharjah, Al Ain
- **Suíça:** Zurique, Genebra, Basileia, Berna
- **Países Baixos:** Amsterdã, Roterdã, Haia, Eindhoven
- **Singapura:** Singapura, Jurong East, Woodlands, Tampines
- **Israel:** Tel Aviv, Jerusalém, Haifa, Beersheba
- **África do Sul:** Joanesburgo, Cidade do Cabo, Durban, Pretória
- **Argentina:** Buenos Aires, Córdoba, Rosário, Mendoza

## Critérios de aceite

- [x] O catálogo contém os 25 países e as 100 cidades listadas acima, incluindo Lyon, e é fixo, versionado e idêntico em todos os servidores.
- [x] Cada país e cidade possui identificador canônico estável, nomes de exibição e associações país-cidade sem depender de tradução ou seed.
- [x] A geração do mundo, bairros, NPCs, empresas, casos e consultas investigativas recebe cidade como contexto; nenhuma dessas operações depende de Lyon como cidade global implícita.
- [x] Cada cidade pode materializar seus 20 bairros lógicos de forma determinística, com banco de palavras local, IDs estáveis e sem exigir mapa físico, coordenadas ou distâncias entre bairros.
- [x] O GATE permite viajar entre quaisquer cidades habilitadas do catálogo, inclusive cruzando fronteiras nacionais; a localização resultante guarda país e cidade canônicos.
- [x] O app GATE lista somente cidades como destinos, agrupadas por país, e deixa explícitas a cidade de origem e a cidade de destino.
- [x] O app GATE não lista nem permite selecionar bairros ou construções como destinos; a navegação entre bairros e as visitas a construções usam interfaces próprias fora do app GATE.
- [x] Pontos fixos de serviço (Banco, Delegacia e Central da GATE) são associados à cidade, e os fluxos que os exigem usam a cidade atual do jogador.
- [x] Testes cobrem integridade e unicidade do catálogo, cidades de países diferentes, viagem internacional, persistência da localização e geração determinística de bairros por cidade.
- [x] A expansão mantém o mundo compartilhado por servidor, as regras de Noir e compatibilidade dos dados já persistidos de Lyon.
- [x] Cada país tem seu próprio catálogo local de nomes e sobrenomes para gerar NPCs em suas cidades.

## Fora de escopo

Mapa físico, geometria, rotas ou distâncias entre bairros e cidades, simulação geográfica de deslocamento, tarifas internacionais e inclusão de todos os países do mundo. Novas cidades e países poderão entrar em futuras versões do catálogo.
