# ADR 0011 — Prédios e empresas como entidades distintas

- **Status:** Aceita
- **Data:** 2026-10-22
- **Supersede:** nenhum

## Contexto

Um endereço físico pode abrigar mais de uma organização. Um prédio como Café des Écoles é uma construção, não uma empresa; diferentes empresas podem ocupar o mesmo endereço. Tratar o prédio e a empresa como o mesmo objeto limita os diretórios, os funcionários e os computadores acessíveis no jogo.

## Decisão

- Prédios são entidades físicas do tipo `location` com `location_kind = building`.
- Empresas são entidades próprias e se associam a prédios pela relação muitos-para-muitos `company_locations`.
- O Neighborhood lista as empresas associadas aos endereços do bairro; prédios não aparecem como destinos visitáveis.
- Ao escolher uma empresa, o jogador abre diretamente seu espaço de trabalho. As opções de interação aparecem no centro e o computador pertence à empresa selecionada.
- Serviços públicos fixos continuam disponíveis como destinos próprios no bairro Central. Acesso a esses serviços usa o contexto do serviço, sem simular uma empresa ocupante.
- A empresa e o endereço selecionados são contexto transitório do frontend; a localização persistida do jogador continua sendo somente a cidade.
- A busca de empresas por `locationId` retorna todas as empresas vinculadas àquele endereço específico, sem o limite aplicado às buscas textuais. Diretórios internos recebem o `companyId` selecionado.

## Consequências

- Um endereço pode apresentar várias empresas independentes, listadas individualmente no Neighborhood.
- Funcionários, computadores e aplicações internas pertencem à empresa; o prédio apenas fornece o endereço físico comum.
- O gerador deve criar associações coerentes entre prédios, empresas e funcionários. Conteúdo já materializado só muda por uma versão explícita do gerador ou materialização compatível.
- A interface não deve inferir empresas apenas pelo bairro em que o endereço está localizado; deve usar as relações explícitas de localização.
