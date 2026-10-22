# M09 — Deslocamento e pontos de serviço

**Ciclo:** 9  
**Status:** Aceito
**Depende de:** M08 aceito

## Objetivo

Permitir que o detetive navegue entre bairros da cidade e visite locais gerados em cada bairro. A localização exata é individual e persistente por jogador; bairros, edifícios e pontos de serviço são compartilhados e determinísticos para todos no servidor. A navegação local e a visita a construções têm interfaces próprias, separadas do app GATE, que no M11 passa a transportar somente entre cidades. Abrir a ficha de um edifício mantém o bairro atual; entrar nele registra o ponto visitado.

## Pontos fixos de cada cidade

Toda cidade do catálogo compartilhado possui um bairro **Central** fixo, com pontos de serviço estáveis:

- **Banco** — atendimento e consulta a serviços bancários;
- **Delegacia** — serviços e registros policiais;
- **Central da GATE** — acesso à rede de transporte GATE.

Esses pontos existem em toda cidade e não são bairros procedurais. Sua identidade deve ser estável e comum aos jogadores do mesmo servidor.

## Critérios de aceite

- [x] Cada cidade possui um bairro Central com Banco, Delegacia e Central da GATE estáveis e compartilhados no servidor.
- [x] O jogador tem uma localização atual persistente, atualizada ao viajar entre bairros e ao entrar em um edifício; a visita preserva o bairro atual.
- [x] O jogador consegue navegar entre os bairros disponíveis por uma interface local, com validação autoritativa no backend, fora do app GATE.
- [x] O jogador consegue abrir fichas e entrar nas construções do bairro atual; construções são acessadas pela interface de locais, não pelo app GATE.
- [x] O cliente informa a localização atual, bairros de destino e construções disponíveis no bairro.
- [x] Consultas bancárias permanecem disponíveis sem visita presencial até o fluxo de visita a serviços estar disponível ao jogador; a exigência de presença pode ser ativada quando esse fluxo for integrado ao gameplay.
- [x] A Delegacia é um ponto fixo visitável e sua presença é registrada com a localização exata; o protocolo e a validação do pedido de busca são implementados no M12.
- [x] Dois jogadores do mesmo servidor veem os mesmos pontos fixos, embora possam estar em bairros diferentes.
- [x] Testes cobrem identidade dos pontos fixos, persistência de localização, viagem, visita, escopo de construções por bairro e bloqueio de serviços fora do local exigido.

## Fora de escopo

Mapa físico de bairros, coordenadas, distâncias entre bairros, rotas locais, combate durante viagens e economia detalhada de tarifas ou tempo de deslocamento. A busca e apreensão em si permanece no M12; este ciclo registra e valida presença na Delegacia para permitir que o serviço seguinte imponha a exigência presencial.
