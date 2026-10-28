# M19 — Pedidos policiais assíncronos

**Ciclo:** 19  
**Status:** Aceito  
**Depende de:** M13 aceito

## Objetivo

Tornar mandados de busca residencial e acusações pedidos protocolados que exigem tempo de análise. O resultado é entregue depois na caixa postal do jogador, sem mostrar o inventário nem o veredito na resposta inicial.

## Critérios de aceite

- [x] Buscar residência e entregar acusação cria uma solicitação persistida com prazo real de três minutos; a requisição HTTP inicial retorna apenas o estado pendente e o horário estimado.
- [x] Uma fila durável processa solicitações vencidas, recupera tarefas após reinício ou queda do worker usando leases e repete falhas transitórias sem duplicar relatórios, acusações ou emails.
- [x] Mandados não expõem inventário antes da conclusão; quando concluídos, o relatório completo aparece na caixa postal e os itens ficam disponíveis para o whiteboard.
- [x] A acusação salva uma cópia imutável das designações e provas no momento do protocolo; alterações posteriores no whiteboard não mudam a análise pendente.
- [x] O resultado do julgamento só chega por email. Uma prisão atualiza mutation e jornal após a análise assíncrona; respostas públicas não revelam culpa nem classificação das provas.
- [x] Repetições do mesmo pedido são idempotentes, e acusações concorrentes pelo mesmo caso continuam serializadas.
- [x] A caixa postal aberta atualiza mensagens periodicamente para mostrar resultados entregues enquanto o jogador está em sessão.
- [x] Testes automatizados cobrem estados pendentes, resposta sem resultado imediato, snapshot da acusação, identidade idempotente de solicitações e reprocessamento de falhas da fila.

## Fora de escopo

Suborno de autoridades, invasão da polícia para aprovar pedidos, regras de limite ou custo por quantidade de mandados e notificações push fora da caixa postal.
