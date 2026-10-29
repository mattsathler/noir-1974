# M24 — Medição inicial de busca e sites empresariais

**Data:** 29/10/2026  
**Ambiente:** API local conectada à base configurada em `.env.local`; país `fr`, 4 cidades  
**Amostragem:** 5 requisições frias por rota, 5 aquecimentos, 100 amostras sequenciais e 100 concorrentes com 8 workers.  
**Alvo:** empresa selecionada pela consulta `a`; amostras brutas em [CSV](M24-company-websites.csv).

## Resultados

| Operação | Fase | Amostras | p50 | p95 | Payload JSON |
| --- | --- | ---: | ---: | ---: | ---: |
| Busca nacional de empresas | cache frio | 5 | 2.87 s | 3.96 s | 7,802 B |
| Busca nacional de empresas | sequencial | 100 | 175 ms | 392 ms | 7,802 B |
| Busca nacional de empresas | concorrência 8 | 100 | 183 ms | 1,415 ms | 7,802 B |
| Abrir site de empresa | cache frio | 5 | 2.91 s | 5.31 s | 2,519 B |
| Abrir site de empresa | sequencial | 100 | 156 ms | 402 ms | 2,519 B |
| Abrir site de empresa | concorrência 8 | 100 | 157 ms | 234 ms | 2,519 B |

O tracer registrou 210 consultas `select_worlds` para cada rota e duração agregada de 43.02 s na busca e 40.97 s na abertura do site. O cache `country_company_index` registrou 409 hits e 12 misses. A estimativa de aumento de heap após uma abertura fria do site foi 1,646,360 bytes; é uma única observação e não representa heap retido por site.

## Leitura e limites

- As amostras aquecidas de busca estão abaixo de 400 ms no p95 sequencial, com resposta de aproximadamente 7.6 KiB. Abrir o site retorna aproximadamente 2.5 KiB e teve p95 sequencial de 402 ms.
- O p95 concorrente da busca sobe para 1.42 s. Deve ser acompanhado quando a busca e o índice nacional forem otimizados; não há evidência nesta rodada para declarar regressão nem para atribuir a causa a um componente específico.
- O cache frio inclui o custo de reconstruir o índice nacional para cada instância de cache, por isso é substancialmente mais lento. O cache quente é o caminho repetido no uso normal do serviço.
- O medidor chama diretamente os métodos do store e serializa a resposta em `httptest`; não mede autenticação, roteamento completo, navegador ou rede externa. As medições representam esta base e máquina, não um SLO de produção.
- A amostragem cobre o país e a empresa indicados, mas ainda não enumera a quantidade total de empresas nem comprova a diversidade visual do catálogo completo. Testes automatizados, auditoria de acessibilidade e validações de autorização seguem pendentes.

## Orçamento provisório para acompanhamento

Com base neste baseline, usar como alertas de engenharia — não como SLO público — p95 sequencial aquecido de até 500 ms para busca e site, p95 concorrente de até 1.5 s para busca e 500 ms para abertura, e payload de busca abaixo de 10 KiB e de site abaixo de 5 KiB. O cache frio fica registrado como baseline separado; antes de fixar limite, medir geração completa em mais de uma base/tamanho de país.

O script reproduzível está em [`apps/api/scripts/measure_m24_company_websites`](../../../apps/api/scripts/measure_m24_company_websites).

## Revalidação do perfil procedural v3

**Data:** 29/10/2026, 19:16 UTC
**Ambiente:** mesma base e máquina; país `fr`, 4 cidades; 5 amostras frias, 5 aquecimentos, 100 sequenciais e 100 em concorrência 8.
**Resumo CSV:** [M24-company-websites-revalidation.csv](M24-company-websites-revalidation.csv). O CSV de amostras brutas acima permanece associado à medição inicial.

| Operação | Fase | Amostras | p50 | p95 | Payload JSON |
| --- | --- | ---: | ---: | ---: | ---: |
| Busca nacional de empresas | cache frio | 5 | 3.67 s | 4.31 s | 7,853 B |
| Busca nacional de empresas | sequencial | 100 | 172 ms | 309 ms | 7,853 B |
| Busca nacional de empresas | concorrência 8 | 100 | 251 ms | 575 ms | 7,853 B |
| Abrir site de empresa | cache frio | 5 | 3.68 s | 4.86 s | 2,817 B |
| Abrir site de empresa | sequencial | 100 | 168 ms | 330 ms | 2,817 B |
| Abrir site de empresa | concorrência 8 | 100 | 183 ms | 1,369 ms | 2,817 B |

O tracer registrou 210 consultas `select_worlds` por rota, com duração agregada de 48.50 s na busca e 42.84 s na abertura. O cache `country_company_index` teve 409 hits e 12 misses. A diferença de heap após a abertura fria foi 1,673,288 B; é uma observação pontual, não heap retido. O p95 concorrente do site oscilou e excedeu o alerta provisório de 500 ms da primeira leitura; por isso o alerta de acompanhamento desta medição fica em 1.5 s sob concorrência 8, sem tratá-lo como SLO. O cache frio segue separado do uso aquecido.

O perfil v3 acrescentou o campo de densidade de anúncios, elevando o payload do site de 2,519 B para 2,817 B; ambos continuam abaixo do limite provisório de 5 KiB. A medição continua cobrindo uma base, um país e uma máquina, e o tracer não revela linhas examinadas nem substitui `EXPLAIN`.
