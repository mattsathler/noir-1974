# M31 — Rivalidade empresarial e exposição de mercado

**Status:** Implementada · revisão no jogo pendente
**Trilha:** Bolsa de valores, Blackwire e consequências públicas  
**Depende de:** M25/M26 — bolsa e flutuação empresarial; M30 — venda de informações na Blackwire

## Objetivo

Fazer os movimentos da bolsa refletirem disputas entre empresas do mesmo ramo e permitir que o jogador aprenda essas relações observando o mercado. A exposição de dados sigilosos na Blackwire poderá ser usada para manipular esses movimentos, com consequências econômicas e narrativas que chegam ao jornal conforme a gravidade.

## Direção definida

- Empresas do mesmo ramo podem pertencer a grupos rivais definidos deterministicamente por seed. A relação deve ser estável para a mesma seed e não mudar aleatoriamente a cada consulta.
- Empresas rivais têm mudanças de mercado em direções opostas: quando uma recebe um movimento positivo relacionado à rivalidade, as demais recebem movimento negativo correspondente; quando uma sofre movimento negativo, as rivais recebem o movimento positivo oposto.
- Essa relação deve ser observável ao longo do tempo, para que o jogador possa inferir quais empresas são rivais estudando as cotações e os changes.
- A venda de informações na Blackwire pode influenciar essas relações e permitir manipulação deliberada do mercado. O efeito não deve aparecer como uma alteração sem causa: a exposição precisa ter origem narrativa rastreável dentro do mundo.
- Evidências de crimes sem relação direta com a empresa causam impacto menor na cotação.
- Evidências de crimes financeiros, fraude corporativa, extorsão ou relacionadas a pessoas de alto escalão da empresa causam impacto maior.
- Crimes graves ou políticos e informações que provoquem grandes movimentos de mercado devem ser mais raros e difíceis de rastrear até a fonte.
- Exposição de dados confidenciais deve ser apresentada de forma diegética. O jornal pode noticiar o vazamento, a investigação corporativa ou suas consequências quando a gravidade justificar; exposições triviais não devem gerar manchetes desproporcionais.
- O jogador deve poder aprender o sistema por observação, sem uma tela que revele diretamente todas as rivalidades ou transforme a bolsa em uma lista de respostas.

## Modelo conceitual de impacto

O evento de mercado possui uma direção e magnitude associadas à empresa afetada. Para o grupo de rivais, a mesma pressão de rivalidade é aplicada na direção oposta, preservando mudanças simétricas entre os participantes do evento. A implementação deve manter essa oposição mesmo quando houver mais de uma rival; limites de cotação não podem inverter o sentido do movimento silenciosamente.

A magnitude depende da categoria e da gravidade da informação exposta:

1. **Informação sem relação empresarial direta:** efeito pequeno, mesmo que útil como evidência criminal.
2. **Informação ligada à atividade ou reputação da empresa:** efeito moderado, proporcional à credibilidade e relevância.
3. **Crime financeiro, fraude corporativa, extorsão ou executivo envolvido:** efeito forte, com repercussão potencial sobre rivais.
4. **Informação grave ou política:** efeito excepcional, conteúdo raro e origem especialmente difícil de rastrear.

As faixas numéricas ainda precisam de balanceamento. A composição dos grupos, a distribuição dos choques e o tratamento dos limites estão definidos abaixo; a direção e a simetria de sinal são regras centrais do sistema.

## Decisões de implementação

- A unidade de rivalidade será o tipo de negócio procedural (`BusinessType.ID`) dentro do país. Isso conecta empresas que disputam o mesmo mercado sem depender do nome gerado.
- A seed do mundo, a versão do gerador e o identificador estável da empresa determinarão os grupos. A mesma instância sempre reconstruirá os mesmos rivais; cada grupo terá ao menos duas empresas.
- Um evento aplica primeiro sua variação à empresa diretamente afetada e distribui a variação oposta entre as demais empresas do grupo. O arredondamento será compensado entre participantes para preservar o total do choque.
- Próximo aos limites do índice, a magnitude será reduzida para o grupo inteiro antes da gravação. Não haverá clamp independente por empresa que inverta o sinal de um rival.
- As variações orgânicas da M26 também passarão a ser determinadas pelo grupo, para que o change observado mantenha a oposição entre rivais ao longo do tempo.
- Os efeitos de mercado são eventos persistidos e correlacionados. Listagem, carteira, ordens e histórico leem o mesmo resultado; a reconstrução de eventos não recalcula a rivalidade a partir do estado atual.
- Os impactos continuarão ligados à origem verificável do evento: crimes corporativos e vazamentos sensíveis podem atingir o grupo; evidências sem vínculo empresarial terão efeito menor e não criarão um choque corporativo artificial.

## Parâmetros iniciais

- Grupos nacionais de 2 a 4 empresas são formados dentro do mesmo `BusinessType.ID`; empresas restantes sem par não recebem oscilação procedural até haver rivais.
- A oscilação orgânica usa ciclos determinísticos de cinco minutos, limitada a cerca de 1% por movimento. O choque de cada grupo soma zero e os rivais recebem o sinal oposto.
- O índice continua limitado entre 500 e 1.500. Choques de eventos são reduzidos em conjunto quando necessário e mantêm a direção de todos os participantes.
- Conflito corporativo e extorsão financeira mantêm os impactos M26 de -25 e -75 pontos.
- Vazamentos Blackwire usam impacto menor para informação de caso (-15), liderança corporativa (-60), evidência financeira corporativa (-100), informação política (-120), evidência financeira executiva (-150) e informação política/executiva (-240). O arredondamento adapta o valor para distribuí-lo simetricamente entre todos os rivais.
- O jornal publica vazamentos corporativos, executivos e políticos relevantes dentro de 24 horas. A matéria não identifica quem vendeu a informação; vazamentos políticos sem empresa associada recebem cobertura genérica e não criam flutuação artificial.

## Imersão e rastreabilidade

- Uma venda ou exposição na Blackwire pode iniciar consequências graduais: rumor, reação de mercado, apuração corporativa e eventual cobertura jornalística.
- A notícia deve refletir o que foi exposto e o que se tornou publicamente verificável, sem narrar como fato algo que o mundo ainda não confirmou.
- Quanto mais grave o conteúdo, mais difícil deve ser rastrear quem vazou a informação. A dificuldade não deve impedir que o evento afete o mundo; deve afetar a atribuição da fonte.
- A exposição pode afetar empresas rivais em sentido contrário mesmo quando a empresa atingida não é culpada de um crime: informação estratégica ou reputacional pode beneficiar concorrentes. Crimes sem vínculo com a empresa não recebem o mesmo choque.
- O sistema deve preservar os valores anteriores/posteriores dos eventos e seus horários para que o histórico e as notícias continuem coerentes com a mesma seed.

## Critérios de aceite

- [x] Grupos rivais são determinísticos e estáveis por seed.
- [x] Eventos de rivalidade produzem changes com sinais opostos entre a empresa afetada e todas as empresas rivais participantes.
- [x] O comportamento continua coerente com grupos de mais de duas empresas e com os limites de score definidos pela bolsa.
- [x] O jogador consegue observar cotações e inferir relações sem receber uma lista explícita de rivais.
- [x] Evidências sem relação empresarial têm impacto menor do que evidências financeiras, corporativas ou ligadas à liderança da empresa.
- [x] Venda ou exposição de dados pela Blackwire pode influenciar a cotação e os rivais com intensidade proporcional à gravidade.
- [x] Exposições graves podem gerar consequência jornalística diegética, com cronologia coerente e sem antecipar fatos ainda desconhecidos publicamente.
- [x] A classificação política exige sinais específicos de conteúdo ou metadados; os movimentos e as matérias não identificam quem vendeu o material.
- [x] O histórico preserva valores anteriores/posteriores, sinais opostos e horários dos eventos correlacionados.
- [x] Eventos, notícias e cotações não revelam automaticamente culpados nem substituem a investigação de casos.

## Revisão posterior

Os parâmetros desta primeira versão podem ser calibrados após observação do ritmo de mercado em jogo. Isso não bloqueia a entrega da M31.

## Fora de escopo

- Um ranking público de investidores ou acesso a dinheiro real.
- Revelar automaticamente todas as rivalidades ao jogador.
- Fazer evidência criminal sem vínculo empresarial derrubar uma companhia como se fosse fraude corporativa.
- Usar a notícia como prova conclusiva de culpa ou substituir os sistemas de evidência e acusação.
