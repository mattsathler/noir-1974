# Prompt-base para imagens de Noir 1974

Use este prompt como base em toda geração de imagens do jogo. Substitua os campos entre colchetes pelo assunto e pelo enquadramento necessários; mantenha as demais características para preservar a unidade visual.

## Prompt-base

> Crie uma imagem para **Noir 1974**, jogo de investigação e dedução urbana ambientado em uma metrópole noir de futuro próximo, com atmosfera de policial procedural e ficção científica distópica. **Assunto da imagem:** [descrever personagem, cenário, objeto ou situação]. **Função no jogo:** [descrever onde será usada e o que precisa comunicar]. **Composição:** [descrever plano, ponto de vista e espaço para interface, se necessário].
>
> A imagem deve parecer uma cena de um jogo narrativo investigativo consistente: pixel art de alta densidade com acabamento pictórico, formas construídas por pixels visíveis e blocos de cor, contornos definidos mas não uniformemente grossos, detalhes ambientais minuciosos e textura granulada. Evite aparência vetorial, fotografia, 3D polido ou pixel art simples e colorida. A imagem deve continuar legível em tamanho reduzido; agrupe detalhes em silhuetas e massas tonais claras.
>
> Use uma paleta restrita, escura e dessaturada: pretos carvão, cinzas azulados, azul petróleo e verdes quase pretos nas sombras; marrom sépia, bronze envelhecido e bege sujo nos materiais; pequenos acentos de vermelho queimado ou âmbar. Reserve tons claros para fontes de luz, reflexos e pontos focais. Contraste geral baixo a médio, com contraste local forte nas áreas importantes para a leitura.
>
> Iluminação cinematográfica noturna, dramática e econômica: grandes áreas em sombra, luzes práticas de teto, luminárias, letreiros ou janelas criando recortes quentes e isolados; luz fria ambiente e névoa urbana ao fundo; reflexos discretos em superfícies úmidas ou metálicas. Use vinheta e profundidade atmosférica com moderação. O clima é tenso, melancólico, decadente e observacional, sem ação espetacular.
>
> Os cenários devem sugerir uma cidade densa, controlada por instituições e grandes corporações: tijolo escurecido, concreto, aço, tubulações, cabos, persianas, ventilação, mobiliário funcional gasto, arquivos, papéis, avisos e sinalização integrada ao ambiente. Combine arquitetura urbana reconhecível com tecnologia analógica e retrofuturista discreta, sem fixar o universo a uma década específica. Inclua pequenos sinais de uso, desgaste e vigilância sem sobrecarregar a composição.
>
> Composição em camadas, com primeiro plano escuro, plano médio legível e fundo urbano parcialmente perdido na sombra. Use linhas arquitetônicas para conduzir o olhar ao ponto focal. Preserve áreas negativas quando a imagem servir de fundo para interface. Não inserir elementos decorativos sem função narrativa; detalhes devem sugerir contexto, relações espaciais ou pistas observáveis, sem destacar automaticamente a resposta do mistério.
>
> Manter identidade original de Noir 1974. Não reproduzir logotipos, personagens, tabuleiros ou imagens de outras obras. Não inventar texto legível, marcas ou símbolos corporativos proeminentes; quando placas ou documentos aparecerem, usar marcas gráficas discretas e não enfatizadas, exceto se o texto exato for fornecido.

## Restrições

- Sem cores saturadas, iluminação uniforme, visual alegre ou fantasia heroica.
- Sem estética fotográfica, render 3D, pintura lisa sem textura ou acabamento de ilustração vetorial.
- Sem excesso de ruído que esconda objetos, pistas ou silhuetas importantes.
- Sem texto aleatório legível, letras deformadas, marcas d'água, molduras ou interface embutida.
- Sem sangue explícito ou violência gráfica; a tensão deve vir do ambiente, da luz e do contexto.

## Instruções por tipo de imagem

- **Cenário/fundo:** priorizar leitura de profundidade, atmosfera, silhuetas arquitetônicas e espaço negativo para a interface.
- **Retrato/personagem:** manter a mesma paleta e textura; expressão contida, roupa funcional e iluminação lateral. Fundo simples e escuro para separar a silhueta.
- **Objeto/pista:** fundo neutro ou transparente conforme o uso; forma imediatamente reconhecível, iluminação controlada e detalhes suficientes para inspeção. Não adicionar brilho, contorno ou cor que funcione como marcador de pista.
- **Mapa ou tela informativa:** manter a paleta e o acabamento, mas priorizar legibilidade e relações espaciais. Não simular um tabuleiro ou identidade visual de outro jogo.
- **Logo e tipografia:** tratar como arte gráfica separada, em alto contraste marfim sobre preto, letras condensadas e robustas com desgaste de impressão e pequenos escorridos. Preservar exatamente o nome e os caracteres fornecidos; não pedir à geração de imagem para inventar a grafia.

## Referências analisadas

O prompt foi consolidado a partir de `apps/web/public/images/pd.webp` (interior de delegacia escuro) e `apps/web/public/images/lyon-noir.webp` (rua urbana noturna), considerando também a marca em `apps/web/public/images/logo.png`. As duas cenas compartilham paleta carvão/azul petróleo/sépia, iluminação quente localizada, arquitetura densa, textura granulada e composição cinematográfica. A marca usa marfim desgastado sobre preto e deve ser tratada como categoria gráfica própria.

**Observação de consistência:** as imagens atuais têm acabamento pixelado e pictórico, mas não são pixel art estrita em todos os detalhes. A regra obrigatória [0001](../rules/0001-direcao-artistica.md) define pixel art para todos os elementos do jogo. Portanto, as futuras gerações devem manter o acabamento visível descrito aqui e também reforçar construção em pixels, evitando suavização fotográfica; assets destinados a integrar o jogo devem passar pela auditoria da regra 0001.
