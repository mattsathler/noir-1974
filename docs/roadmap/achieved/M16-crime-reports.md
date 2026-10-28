# M16 — Crime Reports

**Ciclo:** 16
**Status:** Concluída
**Depende de:** M13 e M14 aceitos

## Objetivo

Criar o app Crime Reports no computador pessoal do jogador como fonte principal de início investigativo ao lado do jornal. O app reúne os crimes reportados na cidade atual e a lista de NPCs desaparecidos, enquanto o jornal publica somente corpos efetivamente encontrados.

## Regras de desaparecimento e descoberta

- Um NPC desaparecido é aquele que deixou de ter contato com qualquer outra pessoa, segundo os dados de contato disponíveis ao mundo do jogo. A classificação e suas consequências são determinadas pelo backend.
- Na regra inicial, todos os desaparecidos são vítimas mortas cujos corpos ainda não foram encontrados. O jogo apresenta o estado como desaparecimento, sem antecipar ao jogador a descoberta do corpo.
- O modelo deve permitir, no futuro, desaparecimentos por sequestro e fuga do assassino, sem exigir que o desaparecimento signifique sempre morte.
- O jornal noticia corpos encontrados, nunca a condição de desaparecido por si só.
- O local e o horário do crime são definidos de acordo com o caso e seu contexto. Por exemplo, um assassinato premeditado motivado por ciúmes tende a ocorrer na residência da vítima, apontando para alguém próximo; num crime empresarial, a vítima pode ser seguida e morta em um local público.
- O horário do crime e o horário da descoberta são eventos distintos. Pistas e registros usam o horário do crime; o relatório e a notícia identificam quando o corpo foi encontrado.
- Um corpo pode ser encontrado numa busca e apreensão residencial ou numa descoberta automática do mundo, como em um beco ou matagal.

## Critérios de aceite

- [x] O computador pessoal oferece um app chamado **CRIME REPORTS**.
- [x] O app lista os crimes reportados para a cidade atual do jogador, com dados suficientes para iniciar uma investigação e sem revelar culpado, prova interna ou informação reservada do caso.
- [x] A lista também apresenta os NPCs desaparecidos da cidade. Um desaparecimento não aparece no jornal enquanto não houver corpo encontrado.
- [x] A cidade usada na consulta é sempre a localização atual confirmada do jogador; ao mudar de cidade, o app passa a consultar os novos registros e não mistura dados de outra cidade.
- [x] A identificação de desaparecidos é calculada pelo backend a partir dos casos sem corpo localizado, não inferida pelo frontend.
- [x] Para a regra inicial, todo NPC desaparecido corresponde a uma vítima morta ainda sem corpo encontrado. A interface, contudo, comunica apenas que a pessoa está desaparecida.
- [x] A geração determinística do caso mantém local e horário do crime coerentes com o arquétipo, as relações e as pistas; mudanças de arquétipo não reutilizam um cenário incompatível. A versão 6 escolhe para ciúmes um edifício não corporativo no distrito do casal, mantém conflitos de trabalho e extorsão na sede da empresa e define janelas noturnas no fuso local para cada arquétipo. As versões anteriores continuam disponíveis para reconstruir casos existentes.
- [x] Se uma busca residencial encontrar o corpo, o resultado informa claramente a descoberta: o email usa um título específico de corpo encontrado e o card correspondente aparece destacado em vermelho, identificando o NPC falecido.
- [x] Descobertas automáticas de corpos em locais públicos são persistidas como eventos do mundo e reportadas automaticamente.
- [x] Todo corpo encontrado é publicado no jornal da cidade após a descoberta confirmada, independentemente de ter sido localizado por busca residencial ou pelo sistema automático do mundo.
- [x] A notícia identifica a cidade, o local da descoberta e quando o corpo foi encontrado, sem trocar esse horário pelo horário do crime nem expor detalhes reservados.
- [x] A mesma descoberta não gera emails, eventos ou notícias duplicadas em tentativas repetidas, recargas, concorrência ou reinício da API.
- [x] O app comunica estados de carregamento, lista vazia e falha recuperável; as informações e ações essenciais permanecem acessíveis sem depender apenas da cor.

## Fora de escopo

Implementar sequestro, fuga do assassino ou outras causas novas de desaparecimento; revelar o estado de morte antes da descoberta do corpo; permitir que o jogador escolha ou altere local e horário do crime; e publicar pessoas desaparecidas como notícia de corpo encontrado.
