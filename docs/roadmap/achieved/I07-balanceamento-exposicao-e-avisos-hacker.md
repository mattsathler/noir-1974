# I07 — Balanceamento de exposição, hosts e segurança hacker

**Status:** Implementada
**Origem:** M33 — Exposição hacker e defesas virtuais
**Escopo:** Médio

## Problema

Os avisos de exposição hacker podem surgir com frequência excessiva. Além disso, consultas a sistemas oficiais, como Polícia, Hospital e GATE Logs, elevam a exposição mesmo quando o jogador está usando esses serviços durante a investigação.

O diretório de rede concentra os websites empresariais em poucos hosts. Os níveis mais robustos de proteção também podem aparecer em empresas comuns, sem relação suficiente com seu valor de mercado.

## Resultado esperado

Os avisos aparecem com menor frequência, preservando a progressão narrativa da exposição. Acessar sistemas oficiais não aumenta a exposição hacker; o Union permanece como exceção.

A rede oferece mais hosts, distribui os websites empresariais entre eles de forma determinística e variada, e reserva as proteções mais caras para empresas cujo valor de mercado pode sustentá-las.

## Escopo

- Rever a cadência dos avisos associados aos limiares de exposição hacker.
- Remover o aumento de exposição por acesso à Polícia, ao Hospital e aos GATE Logs.
- Manter o aumento de exposição associado ao Union.
- Adicionar pelo menos seis hosts ao diretório de rede.
- Distribuir websites empresariais por mais de um host e de forma determinística, em vez de concentrá-los em um único endereço de provedor.
- Relacionar o nível de segurança virtual ao valor de mercado da empresa; empresas comuns de baixo valor não recebem a proteção mais robusta.

## Critérios de aceite

- [x] Os avisos iniciais foram espaçados para 30, 55 e 75 pontos; o aviso crítico permanece em 90.
- [x] Acessos à Polícia, Hospital e GATE Logs não elevam exposição; o acesso ao Union continua elevando.
- [x] O diretório agora oferece nove hosts, e os websites empresariais são distribuídos por host de forma determinística.
- [x] A proteção mais robusta só pode aparecer em empresas com cotação-base de mercado compatível; a proteção mediana também é limitada em empresas de menor valor.

## Fora de escopo

- Alterar os pesos de outras ações ilegais, como Bankrupt, Mailbreaker, Ghostkey corporativo ou vendas pela Blackwire.
- Alterar a redução de exposição concedida pelo Rogue DNS.
- Mudar consequências ou limiares de exposição crítica.
- Alterar cotações por eventos ou outras regras da bolsa.
