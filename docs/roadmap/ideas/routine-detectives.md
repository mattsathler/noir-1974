# Routine Detectives

**Estado:** Ideia em exploração — não aprovada como sistema nem como milestone  
**Área:** Casos, provas e atividade policial cotidiana

## Proposta

O jogador não é o único detetive da cidade. Outros detetives NPC investigam casos e podem recolher provas. Se o jogador não resolver um caso a tempo, a própria cidade poderá encerrá-lo, desde que ainda existam provas suficientes para uma prisão.

Essa atividade mostra que os casos continuam avançando no mundo e cria uma oportunidade para o jogador interferir no que permanece disponível como evidência.

## Resolução policial automática

Cada caso tem um timeout. Quando ele expira, a cidade avalia as provas válidas e não destruídas usando a pontuação já aplicada ao sistema de acusação do jogador. Se a pontuação for suficiente para prender o culpado, o caso é encerrado pela polícia. Caso contrário, permanece aberto sem resolução.

Não há como garantir que a resolução automática nunca surpreenda ou tire a iniciativa do jogador. Essa possibilidade faz parte da ideia: outros detetives investigam, e a cidade pode resolver casos que o jogador deixou passar.

Todo caso encerrado pela cidade deve ser noticiado no jornal ou no site da polícia. A interface não mostra um medidor explícito de progresso policial.

## Destruição de provas

O jogador poderá destruir emails, logs do GATE, logs bancários e itens obtidos em buscas e apreensões. A destruição é permanente: antes de concluir a ação, um modal informa que o item será excluído definitivamente e pede confirmação.

Uma prova destruída é invalidada globalmente. Ela deixa de contar para a investigação do jogador, para os detetives NPC e para a avaliação automática do caso.

Se a destruição reduzir a pontuação abaixo do necessário para uma prisão, o caso permanece aberto. O sistema não o encerra como “sem solução”, pois isso poderia sugerir que a cidade sabe que uma prova foi apagada. Se não houver provas suficientes quando o timeout expirar, não há resolução automática.

## Distinção entre ações

| Ação | Efeito sobre a prova | Efeito sobre o caso |
| --- | --- | --- |
| Vender na Blackwire | A anotação deixa de estar disponível como prova para o jogador | Mantém o efeito de mercado existente sobre a reputação da empresa envolvida |
| Destruir | A prova é excluída permanentemente e invalidada para todos os investigadores | Pode reduzir a pontuação abaixo do necessário para a resolução pelo jogador ou pela polícia |

## Regras definidas

- Detetives NPC podem recolher provas e a cidade pode resolver casos após um timeout.
- A resolução automática usa a pontuação atual de acusação e considera provas não destruídas.
- Provas que podem ser destruídas: emails, logs do GATE, logs bancários e itens de buscas e apreensões.
- A destruição exige confirmação em modal e é permanente.
- Casos que não têm provas suficientes após o timeout permanecem abertos.
- Todo caso encerrado pela cidade deve aparecer no jornal ou no site da polícia.
- Não há medidor visível de progresso policial.

## Questões ainda em aberto

- Qual será a duração do timeout e ela varia conforme o caso?
- A pontuação automática considera todas as provas válidas do caso ou apenas aquelas que os detetives NPC conseguiram recolher?
- Como o jornal ou o site da polícia identifica a conclusão automática em relação a uma resolução conduzida pelo jogador?

## Fora desta ideia

Esta proposta descreve apenas detetives NPC, resolução policial automática e destruição de provas. Não define conteúdo de missões ou usos narrativos futuros da destruição de provas.
