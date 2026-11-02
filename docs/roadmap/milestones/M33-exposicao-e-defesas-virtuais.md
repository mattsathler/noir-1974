# M33 — Exposição hacker e defesas virtuais

**Status:** Em andamento · proteções e alertas implementados

**Trilha:** Risco criminal, segurança corporativa e consequências narrativas  
**Depende de:** M30 — Blackwire e venda de informações; M32 — conta e progressão do alter ego hacker

## Objetivo

Fazer atividades ilegais exigirem cautela. Empresas protegem seus sistemas, ações criminosas deixam rastros e a identidade hacker do jogador pode ser investigada. A exposição cria consequências graduais e narrativas; exposição crítica pode levar The Crow a banir a conta Blackwire e apagar o progresso ilegal.

## Defesas empresariais

Empresas contratam proteção virtual de modo determinístico, vinculada à seed e estável para a mesma empresa naquele mundo. Há quatro níveis previstos:

1. **Nenhuma proteção:** sistemas sem defesa adicional.
2. **Proteção simples:** Aegis Data Bureau.
3. **Proteção mediana:** Northstar Shield, mais eficaz que o nível simples.
4. **Meridian Web Security (MWS):** proteção de maior nível no universo Noir.

As proteções dificultam invasões de sites e sistemas empresariais. A dificuldade pode se expressar por efeitos diegéticos, como:

- encerramento inesperado de sessão ou desconexão do sistema invadido;
- bloqueio temporário ou aumento de latência;
- criptografia que oculta nomes, endereços de email e outros campos;
- registros parciais, atrasados ou adulterados;
- detecção de tentativas repetidas e aumento de exposição;
- informações que exigem outro aplicativo, acesso ou método para serem recuperadas.

Os níveis mais altos podem combinar mais de um efeito. A proteção deve dificultar e modificar a interação, sem simplesmente transformar ferramentas centrais em erros opacos que impeçam a progressão.

## Exposição do jogador

- O sistema mantém um contador de exposição hacker invisível ao jogador.
- Ações ilegais elevam esse contador. Entre as fontes previstas estão vender informações, usar Ghostkey e Mailbreaker e, com peso especialmente alto, usar Bankrupt.
- O nível não é exibido como uma barra ou número. O jogador interpreta sinais e consequências do mundo.
- A exposição pode provocar mensagens anônimas ameaçadoras, conversas de NPCs sobre vazamentos, indícios de investigação empresarial e outros alertas narrativos.
- Os avisos devem surgir de forma gradual para ensinar que ações deixam rastros, sem revelar a fórmula interna nem punir uma única ação comum de maneira imprevisível.
- A intensidade, frequência e gravidade dos alertas aumentam com a exposição.

## Rogue DNS e redução de rastros

- Rogue DNS é a principal ferramenta do jogador para mudar sua rota e reduzir o risco de continuar atuando sob a mesma identidade de rede.
- Trocar o proxy interrompe novos alertas associados ao país da rota anterior.
- A troca reduz automaticamente a exposição acumulada conforme regras a definir; não precisa zerá-la por completo.
- A mudança de rota deve ser legível no mundo e não apagar consequências já iniciadas, investigações em curso ou evidências já vazadas.

## Consequência crítica

Em exposição crítica, The Crow pode intervir diretamente, inclusive banindo o alter ego hacker da Blackwire. O banimento pode apagar a conta e o progresso ilegal, exigindo que o jogador crie outra conta. Essa sanção atinge a identidade hacker, não a progressão legal do personagem.

A intervenção de The Crow é uma consequência extrema, não uma punição automática por experimentar ferramentas. O jogador deve receber alertas diegéticos crescentes e oportunidades coerentes de reduzir rastros antes do limite crítico.

## Critérios de aceite

- [x] Cada empresa recebe deterministamente um dos quatro níveis de proteção e mantém esse nível na mesma seed.
- [x] A proteção muda de forma perceptível e graduada a experiência de invasão.
- [x] Efeitos de segurança fornecem sinais compreensíveis ao jogador e não bloqueiam silenciosamente requisitos essenciais de casos.
- [x] Ações ilegais aumentam exposição conforme sua natureza e gravidade; Bankrupt pesa mais que ações de menor risco.
- [ ] O jogador recebe indícios narrativos crescentes sem visualizar o contador interno. (Emails graduais implementados; conversas e notícias pendentes.)
- [x] Trocar o Rogue DNS encerra alertas futuros relativos ao país anterior e reduz a exposição segundo uma regra consistente.
- [ ] Consequências já iniciadas não desaparecem apenas pela troca de proxy.
- [ ] Exposição crítica pode provocar a intervenção e o banimento por The Crow após alertas graduais. (A conta hacker separada e seu progresso agora existem na M32.)
- [ ] Perder a conta Blackwire afeta a progressão ilegal, não os Credits, a fama ou os dados legais do personagem.
- [ ] A conta nova começa com progresso hacker separado, conforme as regras de M32.

## Decisões pendentes

- Lore adicional dos provedores de proteção e a possibilidade de ponderar sua distribuição por ramo ou porte.
- Balanceamento de exposição por ação, persistência, recuperação e frequência dos alertas.
- Tipos, autoria e frequência de mensagens anônimas e conversas de NPCs.
- Avisos finais e janela de reação antes de The Crow banir uma conta.
- Efeitos exatos do banimento: Crypto restante, aplicativos, dados e possibilidade de criar nova conta.
- Como evitar que a exposição torne inevitável a perda de conta durante o uso normal de ferramentas necessárias.

## Implementação atual

- Cada empresa recebe deterministicamente um nível de proteção por seed, versão do gerador e ID da empresa. O nível e o provedor permanecem iguais naquela seed.
- Provedores: sem proteção, Aegis Data Bureau (simples), Northstar Shield (mediana) e Meridian Web Security (MWS).
- O portal da empresa informa seu provedor. No diretório, Aegis mascara parte do email; Northstar oculta emails e números bancários; MWS também oculta os nomes dos funcionários.
- A inspeção acrescenta 250 ms na Aegis, 650 ms na Northstar e 1.100 ms na MWS. Uma em cada três tentativas de diretório sob MWS sofre interrupção; o app reconecta automaticamente e avisa o jogador. A reconexão gera apenas um rastro adicional pequeno.
- A exposição fica invisível e é armazenada por personagem/servidor/país. Pesos iniciais: 2–11 por tentativa de acesso corporativo via Ghostkey (conforme o provedor), 5 por consulta privilegiada via Ghostkey, 16 por invasão via Mailbreaker, 28 por Bankrupt e 3–32 por venda na Blackwire, conforme a gravidade classificada do item.
- Emails anônimos chegam ao jogador ao cruzar 20, 45, 70 e 90 pontos. A mensagem de 90 pontos é um aviso de The Crow.
- Ao trocar para outro nó no Rogue DNS, a exposição da jurisdição anterior cai 35%. O app confirma quando encontrou e reduziu rastros, sem mostrar um medidor. Avisos já recebidos e consequências já publicadas permanecem.
- A identidade hacker da M32 fornece o escopo para que esta milestone implemente banimento e perda da progressão ilegal sem atingir o personagem legal.

## Escopo ainda pendente

- Bloqueios e adulterações de registros além da interrupção recuperável da MWS; ampliar efeitos sem tornar dados essenciais inacessíveis.
- Conversas de NPCs e notícias sobre vazamentos e investigações.
- Limiares, recuperação e frequência dos rastros balanceados no jogo.
- Implementar o banimento da conta Blackwire e a perda somente do progresso hacker.

## Fora de escopo

- Mostrar um nível numérico ou barra de procurado na interface.
- Aplicar punições à identidade legal do personagem por consequência do alter ego hacker.
- Fazer a troca de proxy apagar informações já vendidas, consequências públicas ou investigações em andamento.
- Tornar qualquer defesa empresarial invencível ou esconder toda informação necessária para progredir.
