# I06 — Widgets informativos no desktop

**Status:** Proposta  
**Origem:** Evolução da interface do computador pessoal  
**Escopo:** Médio

## Problema

Informações que o jogador consulta com frequência — identidade, Credits disponíveis e rota ativa do Rogue DNS — estão em telas e aplicativos separados. O desktop pode oferecer uma leitura rápida desses dados sem exigir que o jogador abra uma janela.

## Resultado esperado

Adicionar widgets compactos no canto superior direito do desktop do computador pessoal, integrados ao visual do MeridOS. Eles mostram a identidade atual do personagem, o saldo atual no Union Mercantile Bank e o país/horário local da rota ativa do Rogue DNS.

## Escopo

- Criar um widget de identidade com retrato do jogador e nome do personagem.
- Reservar espaço nesse widget para o pseudônimo hacker do alter ego; a M32 agora o preenche com o handle da conta Blackwire.
- Criar um widget financeiro conectado ao saldo atual de Credits no Union Mercantile Bank.
- Criar um widget de rede mostrando o local associado à rota ativa do Rogue DNS e o horário local desse local.
- Manter os widgets no canto superior direito da área de trabalho, sem cobrir a barra do sistema nem competir com o conteúdo das janelas abertas.
- Atualizar cada widget quando sua fonte de dados mudar: personagem/retrato, saldo após transações ou rota do Proxy.
- Apresentar estado de carregamento e indisponibilidade sem exibir valores antigos como se fossem atuais.
- Adaptar ou reorganizar os widgets em viewports estreitos, preservando a legibilidade e sem bloquear atalhos ou ícones do desktop.
- Tornar os dados compreensíveis por leitor de tela; informação não pode depender apenas de ícone ou cor.

## Critérios de aceite

- [ ] Os três widgets aparecem no canto superior direito do desktop pessoal.
- [ ] O widget de identidade mostra o retrato e nome atuais do personagem.
- [x] Há espaço para a identidade do alter ego, preenchida quando a conta Blackwire da M32 é provisionada.
- [ ] O widget do Union Mercantile Bank mostra o saldo atual de Credits e reflete alterações de saldo sem recarregar o jogo.
- [ ] O widget de Rogue DNS mostra a rota ativa, e não apenas a localização física do personagem.
- [ ] O relógio usa o fuso horário local do destino/país da rota ativa e muda após uma troca de Proxy.
- [ ] A troca do retrato ou nome atualiza o widget de identidade.
- [ ] Os widgets não ficam acima das janelas abertas nem ocultam controles do desktop.
- [ ] O layout continua utilizável em telas menores e com zoom do navegador.
- [ ] Rótulos acessíveis identificam nome, saldo, localização da rota e horário para tecnologias assistivas.

## Fora de escopo

- Exibir Crypto dentro do widget bancário de Credits; o saldo Blackwire aparece junto à identidade hacker.
- Misturar Credits com Crypto ou mostrar saldo da Blackwire como parte do widget bancário.
- Permitir movimentações financeiras diretamente pelo widget; ações bancárias continuam nos aplicativos próprios.
- Mostrar histórico completo de transações, cotações ou detalhes técnicos da conexão do Proxy.
- Colocar widgets sobre janelas ou torná-los janelas independentes arrastáveis.
