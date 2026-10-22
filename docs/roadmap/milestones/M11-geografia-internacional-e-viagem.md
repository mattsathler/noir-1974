# M11 — Geografia, viagem e visita a locais

**Ciclo:** 11  
**Status:** Aceito
**Depende de:** M10 aceito

## Objetivo

Expandir o mundo compartilhado de Noir para um catálogo fixo de 25 países e 100 cidades, comum a todos os servidores, e reformular a viagem pelo GATE. O jogador escolhe país, cidade e bairro, nessa ordem, antes de viajar. Ao chegar, a interface mostra no centro da tela cards de todas as construções daquele bairro. Escolher uma construção abre sua tela de interior; a apresentação dessa tela e as interações dentro dela pertencem ao M12. A cidade é a única localização persistida; bairro escolhido e construção aberta são contexto transitório. Todas as cidades têm em Central os mesmos prédios principais: Police Station (Delegacia), Bank (Banco), GATE Central e Hospital. O Hospital aparece fechado e indisponível para uso futuro.

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
- [x] O fluxo do GATE guia a seleção nesta ordem: país, cidade daquele país, bairro daquela cidade e confirmação da viagem.
- [x] O backend valida que `cityId` e `districtId` canônicos pertencem à mesma cidade antes de materializar ou concluir a viagem.
- [x] Ao concluir a viagem, o jogador recebe cards com todas as construções do bairro selecionado, incluindo os prédios públicos fixos quando o bairro for Central.
- [x] Selecionar uma construção abre a tela de interior provisória e mantém seu ID no estado transitório do frontend; nenhuma visita é persistida. A tela e as interações completas pertencem ao M12.
- [x] A localização persistida guarda somente a cidade de destino. Bairro escolhido e prédio aberto ficam no estado transitório da interface e não alteram o banco.
- [x] Pontos fixos de serviço (Banco, Delegacia e Central da GATE) são associados à cidade, e os fluxos que os exigem usam a cidade atual do jogador.
- [x] O bairro Central de cada cidade apresenta o mesmo conjunto de quatro prédios principais: Police Station, Bank, GATE Central e Hospital; os três primeiros estão disponíveis e o Hospital é exibido como fechado para uma feature futura.
- [x] A interface principal do jogo mostra cards grandes, centralizados na área de jogo, para todas as construções do bairro de chegada, não apenas os prédios fixos.
- [x] Abrir, trocar ou fechar a tela de interior não grava visita, bairro atual ou prédio atual no banco; a cidade persistida permanece inalterada.
- [x] O jogador pode sair da tela de interior e retornar aos cards do bairro da viagem atual sem alterar o estado compartilhado ou a localização persistida.
- [x] Os prédios e suas categorias são consistentes entre cidades e servidores; a expansão não exige mapa físico, coordenadas ou distâncias.
- [x] Testes de regressão cobrem a seleção país → cidade → bairro, validação da viagem, retorno das construções na resposta, persistência somente da cidade e preservação da geração determinística dos bairros.
- [x] A expansão mantém o mundo compartilhado por servidor, as regras de Noir e compatibilidade dos dados já persistidos de Lyon.
- [x] Cada país tem seu próprio catálogo local de nomes e sobrenomes para gerar NPCs em suas cidades.

## Fora de escopo

Mapa físico, geometria, rotas ou distâncias entre bairros e cidades, simulação geográfica de deslocamento, tarifas internacionais, persistência de bairro ou visita, telas de interior e interações investigativas dentro dos locais, e inclusão de todos os países do mundo. Novas cidades e países poderão entrar em futuras versões do catálogo.
