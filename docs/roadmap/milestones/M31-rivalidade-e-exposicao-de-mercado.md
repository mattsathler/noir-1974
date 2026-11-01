# M31 — Rivalidade empresarial e exposição de mercado

**Status:** Direção definida · escopo a implementar  
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

Os valores exatos, a composição de grupos e o tratamento de limites de score ficam para balanceamento e implementação. A direção e a simetria de sinal, porém, são regras centrais do sistema.

## Imersão e rastreabilidade

- Uma venda ou exposição na Blackwire pode iniciar consequências graduais: rumor, reação de mercado, apuração corporativa e eventual cobertura jornalística.
- A notícia deve refletir o que foi exposto e o que se tornou publicamente verificável, sem narrar como fato algo que o mundo ainda não confirmou.
- Quanto mais grave o conteúdo, mais difícil deve ser rastrear quem vazou a informação. A dificuldade não deve impedir que o evento afete o mundo; deve afetar a atribuição da fonte.
- A exposição pode afetar empresas rivais em sentido contrário mesmo quando a empresa atingida não é culpada de um crime: informação estratégica ou reputacional pode beneficiar concorrentes. Crimes sem vínculo com a empresa não recebem o mesmo choque.
- O sistema deve preservar o histórico dos eventos para que cotações passadas e notícias continuem coerentes com a mesma seed.

## Critérios de aceite

- [ ] Grupos rivais são determinísticos e estáveis por seed.
- [ ] Eventos de rivalidade produzem changes com sinais opostos entre a empresa afetada e todas as empresas rivais participantes.
- [ ] O comportamento continua coerente com grupos de mais de duas empresas e com os limites de score definidos pela bolsa.
- [ ] O jogador consegue observar cotações e inferir relações sem receber uma lista explícita de rivais.
- [ ] Evidências sem relação empresarial têm impacto menor do que evidências financeiras, corporativas ou ligadas à liderança da empresa.
- [ ] Venda ou exposição de dados pela Blackwire pode influenciar a cotação e os rivais com intensidade proporcional à gravidade.
- [ ] Exposições graves podem gerar consequência jornalística diegética, com cronologia coerente e sem antecipar fatos ainda desconhecidos publicamente.
- [ ] Informações de alto impacto e caráter político são raras e mais difíceis de rastrear do que vazamentos comuns.
- [ ] Histórico de eventos e consulta a datas anteriores mantêm os mesmos resultados para a mesma seed.
- [ ] Eventos, notícias e cotações não revelam automaticamente culpados nem substituem a investigação de casos.

## Decisões pendentes

- Como dividir ou replicar a magnitude entre múltiplas empresas rivais para manter oposição simétrica sem causar movimentos excessivos.
- Como aplicar os limites de score sem quebrar a regra de direção oposta; possíveis mecanismos incluem reservar faixa de variação ou reduzir proporcionalmente o choque do grupo.
- Quais dados e fontes determinam o ramo e a rivalidade procedural de cada empresa.
- Quais faixas numéricas representam impacto pequeno, moderado, forte e excepcional.
- Quais eventos geram rumor, apuração ou notícia e qual o atraso diegético entre exposição e publicação.
- Como a raridade e a dificuldade de rastreamento são representadas sem tornar vazamentos graves impossíveis de descobrir.

## Fora de escopo

- Um ranking público de investidores ou acesso a dinheiro real.
- Revelar automaticamente todas as rivalidades ao jogador.
- Fazer evidência criminal sem vínculo empresarial derrubar uma companhia como se fosse fraude corporativa.
- Usar a notícia como prova conclusiva de culpa ou substituir os sistemas de evidência e acusação.
