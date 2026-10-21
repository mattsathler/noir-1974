# M09 — Deslocamento e pontos de serviço

**Ciclo:** 9  
**Status:** Em curso
**Depende de:** M08 aceito

## Objetivo

Permitir que o detetive use a rede GATE para viajar entre bairros da cidade e visitar locais gerados em cada bairro. A posição atual é individual e persistente por jogador; bairros e locais são compartilhados e determinísticos para todos no servidor. Pontos de serviço fixos e restrições de serviços presenciais completam este milestone.

## Pontos fixos de cada cidade

Toda cidade do catálogo compartilhado possui um bairro **Central** fixo, com pontos de serviço estáveis:

- **Banco** — atendimento e consulta a serviços bancários;
- **Delegacia** — serviços e registros policiais;
- **Central da GATE** — acesso à rede de transporte GATE.

Esses pontos existem em toda cidade e não são bairros procedurais. Sua identidade deve ser estável e comum aos jogadores do mesmo servidor.

## Critérios de aceite

- [x] Cada cidade possui um bairro Central com Banco, Delegacia e Central da GATE estáveis e compartilhados no servidor.
- [x] O jogador tem uma localização atual persistente, atualizada ao viajar entre bairros; abrir a ficha de uma construção não muda o bairro.
- [x] O jogador consegue viajar entre os bairros disponíveis pela rede GATE, com validação autoritativa no backend.
- [x] O jogador consegue abrir a ficha das construções do bairro atual; construções de outros bairros só aparecem após viajar pela GATE.
- [x] O cliente informa a localização atual, bairros de destino e construções disponíveis no bairro.
- [ ] Serviços marcados como presenciais só podem ser consultados quando o jogador está no ponto correspondente; tentar acessá-los de outro local é recusado pelo backend.
- [ ] A consulta de informações de conta bancária exige que o jogador esteja no Banco da cidade pertinente; solicitar uma busca e apreensão exige que esteja na Delegacia.
- [x] Dois jogadores do mesmo servidor veem os mesmos pontos fixos, embora possam estar em bairros diferentes.
- [ ] Testes cobrem identidade dos pontos fixos, persistência de localização, viagem, visita e bloqueio de serviços fora do local exigido.

## Fora de escopo

Mapa físico de bairros, coordenadas, distâncias entre bairros, rotas locais, combate durante viagens e economia detalhada de tarifas ou tempo de deslocamento. Regras de custo, duração, disponibilidade de destinos e quais serviços serão presenciais permanecem para detalhamento antes da implementação.
