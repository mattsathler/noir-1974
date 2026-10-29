# M23 — Linha de base do portal de empresas

**Data:** 2026-10-29  
**País:** France (`fr`, quatro cidades)  
**Amostras:** 100 por fluxo, 5 aquecimentos, execução sequencial e concorrência de 8 workers
**Banco:** Neon configurado no ambiente de desenvolvimento

## Resultados

| Fluxo | Antes: p50 / p95 | Depois: p50 / p95 | Resultado/payload depois |
|---|---:|---:|---:|
| National Directory (`q=a`) aquecido | 156,53 / 232,64 ms | 161,10 / 264,67 ms | 10 resultados / 6.671 bytes |
| Diretório de funcionários aquecido | 1.235,04 / 3.436,29 ms | 348,07 / 1.241,69 ms | 1 empresa / 2.316 bytes |
| Construção fria do índice nacional | 3.544,93 ms | 3.901,11 ms | 10 resultados / 6.671 bytes |
| Primeira abertura do diretório após busca | 4.310,18 ms | 336,13 ms | 1 empresa / 2.316 bytes |
| National Directory, aquecido, concorrência 8 | — | p50 207,14 / p95 2.272,61 ms | 10 resultados / 6.671 bytes |
| Diretório de funcionários, aquecido, concorrência 8 | — | p50 345,50 / p95 492,65 ms | 1 empresa / 2.316 bytes |

O novo caminho reduziu a p50 do diretório em aproximadamente 72% e a p95 em 64%. A primeira construção do índice acrescentou aproximadamente 21,6 MB ao heap após coleta de lixo. Esse valor é uma estimativa do heap retido pelo cache neste processo, não uma medição isolada do tamanho serializado do índice.

## Metas de operação

Metas iniciais definidas a partir da linha de base deste ambiente: National Directory aquecido até 400 ms p95 sequencial e 2.500 ms p95 com 8 workers; diretório empresarial até 700 ms p95 sequencial e 800 ms p95 com 8 workers. A amostra completa ficou dentro dos quatro limites (316,45 ms; 2.272,61 ms; 563,07 ms; 492,65 ms, respectivamente). São limites de regressão para este ambiente e volume de dados, não garantias de latência de rede de produção.

## Consultas observadas

Nas 106 chamadas de busca nacional (frio, cinco aquecimentos e cem amostras), foram observadas 106 consultas classificadas como leitura de `worlds`, uma por chamada. O índice registrou um cache miss na primeira consulta e 105 hits seguintes. A busca aquecida retorna no máximo dez registros e não carrega funcionários.

Na execução concorrente adicional foram feitas 100 buscas nacionais e 100 aberturas de diretório, com 8 workers. O total instrumentado foi de 206 consultas `select_worlds` para o portal (uma por requisição, incluindo a chamada fria) e 410 consultas para diretórios (duas por abertura); o cache registrou 410 hits no índice nacional e 205 hits no índice do diretório. As respostas concorrentes mantiveram os limites de 10 resultados e de 1 empresa.

Antes da otimização, as 105 leituras do diretório (primeiro aquecimento e cem amostras) executaram 420 consultas, quatro por abertura, com leituras repetidas de entidades, mutações e metadados do mundo. Depois da otimização, as mesmas 105 leituras executaram 210 consultas: 105 consultas verificaram a versão atual do catálogo e 105 carregaram as relações e estados atuais dos funcionários. O cache do índice empresarial registrou 105 hits nas aberturas e evitou reconstruir os dados determinísticos do prédio.

O acesso continua validando a versão atual do mundo e carrega os estados dos funcionários do banco a cada abertura. O endpoint de autorização verifica a existência da empresa sem carregar sua equipe. Locais que não estão no índice nacional de sedes seguem usando a busca local anterior.

## Limitações

- A execução foi sequencial e direta sobre os métodos do store, com serialização JSON em um `httptest` local. Não mede autenticação, middleware completo, gravação no whiteboard nem latência de rede do endpoint real.
- O tracer PGX não expõe linhas examinadas; não coletamos plano `EXPLAIN`. As contagens de consultas são completas para as operações observadas, mas não devem ser interpretadas como cardinalidade de linhas lidas pelo banco.
- O benchmark mede o handler de medição com store e serialização JSON em processo local. Não mede autenticação, middleware completo, gravação no whiteboard nem latência de rede do endpoint real.
- O primeiro acesso ao diretório aparece na coluna `warmup` do CSV porque a busca de teste aquece o índice antes de abrir a página empresarial.

## Artefatos e reprodução

- Amostras brutas após otimização: [M23-company-portal.csv](M23-company-portal.csv)
- Amostras sequenciais e concorrentes: [M23-company-portal-concurrent.csv](M23-company-portal-concurrent.csv)
- Script: `apps/api/scripts/measure_m23_company_portal/main.go`
- Reprodução: `go run ./scripts/measure_m23_company_portal --samples 100 --warmups 5 --concurrency 8 --csv ../../docs/roadmap/relatorios/M23-company-portal-concurrent.csv`
