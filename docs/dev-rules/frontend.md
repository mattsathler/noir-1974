# Frontend

- `apps/web` renderiza e envia intenções (`quero o NPC`, `avanço o caso`). Não gera mundo, não escolhe culpado, não aplica mutation local como verdade.
- Estado local é cache de UI. Conflito com o servidor: o servidor ganha.
- Sem lógica de “resolver o caso no cliente” que o backend não possa recusar.
- HTML/CSS/JS; não introduzir framework sem ADR.
