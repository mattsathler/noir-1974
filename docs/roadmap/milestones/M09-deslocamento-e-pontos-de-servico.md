# M09 — Deslocamento e pontos de serviço

**Ciclo:** 9  
**Status:** Não iniciado  
**Depende de:** M08 aceito

## Objetivo

Permitir que o detetive se desloque entre cidades e visite pontos de serviço fixos. A presença do jogador em um local passa a determinar quais serviços pode consultar. Os pontos são parte do mundo compartilhado do servidor e têm identidade estável para todos os seus jogadores.

## Pontos fixos de cada cidade

Toda cidade do catálogo compartilhado possui três pontos de serviço:

- **Banco** — atendimento e consulta a serviços bancários;
- **Delegacia** — serviços e registros policiais;
- **Central da GATE** — acesso à rede de transporte GATE.

Esses pontos existem em toda cidade e não são bairros procedurais. Sua identidade deve ser estável e comum aos jogadores do mesmo servidor.

## Critérios de aceite

- [ ] Cada cidade possui um Banco, uma Delegacia e uma Central da GATE identificáveis e compartilhados no servidor.
- [ ] O jogador tem uma localização atual persistente, que é atualizada ao viajar ou visitar um ponto de serviço.
- [ ] O jogador consegue viajar entre cidades disponíveis pela Central da GATE, com validação autoritativa no backend.
- [ ] O jogador consegue visitar os pontos de serviço fixos da cidade em que está.
- [ ] O cliente informa com clareza a localização atual e os destinos disponíveis para deslocamento.
- [ ] Serviços marcados como presenciais só podem ser consultados quando o jogador está no ponto correspondente; tentar acessá-los de outro local é recusado pelo backend.
- [ ] A consulta de informações de conta bancária exige que o jogador esteja no Banco da cidade pertinente; solicitar uma busca e apreensão exige que esteja na Delegacia.
- [ ] Dois jogadores do mesmo servidor veem os mesmos pontos fixos, embora possam estar em locais diferentes.
- [ ] Testes cobrem identidade dos pontos, persistência de localização, viagem e bloqueio de serviços fora do local exigido.

## Fora de escopo

Mapa físico de bairros, coordenadas, distâncias entre bairros, rotas locais, combate durante viagens e economia detalhada de tarifas ou tempo de deslocamento. Regras de custo, duração, disponibilidade de destinos e quais serviços serão presenciais permanecem para detalhamento antes da implementação.
