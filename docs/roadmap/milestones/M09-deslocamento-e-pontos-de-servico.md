# M09 — Deslocamento e pontos de serviço

**Ciclo:** 9  
**Status:** Aceito
**Depende de:** M08 aceito

## Objetivo

Permitir que o detetive use a rede GATE para viajar entre bairros da cidade e visitar locais gerados em cada bairro. A localização exata é individual e persistente por jogador; bairros, edifícios e pontos de serviço são compartilhados e determinísticos para todos no servidor. Abrir a ficha de um edifício mantém o bairro atual; entrar nele registra o ponto visitado.

## Pontos fixos de cada cidade

Toda cidade do catálogo compartilhado possui um bairro **Central** fixo, com pontos de serviço estáveis:

- **Banco** — atendimento e consulta a serviços bancários;
- **Delegacia** — serviços e registros policiais;
- **Central da GATE** — acesso à rede de transporte GATE.

Esses pontos existem em toda cidade e não são bairros procedurais. Sua identidade deve ser estável e comum aos jogadores do mesmo servidor.

## Critérios de aceite

- [x] Cada cidade possui um bairro Central com Banco, Delegacia e Central da GATE estáveis e compartilhados no servidor.
- [x] O jogador tem uma localização atual persistente, atualizada ao viajar entre bairros e ao entrar em um edifício; a visita preserva o bairro atual.
- [x] O jogador consegue viajar entre os bairros disponíveis pela rede GATE, com validação autoritativa no backend.
- [x] O jogador consegue abrir fichas e entrar nas construções do bairro atual; nomes de construções de outros bairros só aparecem após viajar pela GATE.
- [x] O cliente informa a localização atual, bairros de destino e construções disponíveis no bairro.
- [x] Consultas bancárias permanecem disponíveis sem visita presencial até o fluxo de visita a serviços estar disponível ao jogador; a exigência de presença pode ser ativada quando esse fluxo for integrado ao gameplay.
- [x] A Delegacia é um ponto fixo visitável e sua presença é registrada com a localização exata; o protocolo e a validação do pedido de busca são implementados no M11.
- [x] Dois jogadores do mesmo servidor veem os mesmos pontos fixos, embora possam estar em bairros diferentes.
- [x] Testes cobrem identidade dos pontos fixos, persistência de localização, viagem, visita, escopo de construções por bairro e bloqueio de serviços fora do local exigido.

## Fora de escopo

Mapa físico de bairros, coordenadas, distâncias entre bairros, rotas locais, combate durante viagens e economia detalhada de tarifas ou tempo de deslocamento. A busca e apreensão em si permanece no M11; este ciclo registra e valida presença na Delegacia para permitir que o serviço seguinte imponha a exigência presencial.
