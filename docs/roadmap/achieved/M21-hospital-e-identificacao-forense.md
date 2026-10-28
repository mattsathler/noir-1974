# M21 — Hospital e identificação forense

**Ciclo:** 21  
**Status:** Aceito  
**Depende de:** M20 aceito

> **Exceção de sequência:** M21 foi executada e aceita por instrução explícita enquanto M20 permanece pendente. Esta exceção não aceita M20 nem altera seus critérios; a auditoria pré-beta continua como trabalho separado.

## Objetivo

Abrir o Hospital como local investigativo e acrescentar identificação por impressões digitais ao fluxo de casos. O Patient Search permite consultar o prontuário dos NPCs; prontuários de pessoas mortas trazem dados verificáveis do homicídio, e mandados e exames pós-morte podem revelar vestígios que ligam o responsável à cena. Cada NPC recebe um código estável de impressão digital no formato `AA00000`, com duas letras e cinco dígitos.

## Escopo

- **Hospital:** abrir o serviço Hospital já presente no catálogo geográfico e substituir o interior fechado por uma interação investigativa própria. O acesso aos sistemas do hospital usa a sessão local autorizada e valida cidade, serviço e conta no backend.
- **Patient Search:** pesquisar NPCs pelo nome e consultar o código de impressão digital associado ao prontuário. O código é estável para o mesmo NPC após reinício/reconstrução do mundo e não é exposto por diretórios, telas ou endpoints não autorizados.
- **Prontuário pós-morte:** para vítimas de homicídio, exibir horário estimado e causa da morte, incluindo a arma usada quando conhecida; informar onde o corpo foi encontrado e outros dados forenses úteis definidos para o caso. Não revelar o culpado nem campos internos que o identifiquem diretamente.
- **Vestígios no corpo:** permitir que o prontuário pós-morte revele, de forma determinística por caso, ou vestígios de cabelo com cor compatível com a identidade visual gerada do responsável, ou parte do código de impressão digital do responsável. A pista pode ser incompleta e deve ser apresentada sem indicar a quem pertence.
- **Mandado residencial:** buscas relacionadas a um caso cuja cena do crime é uma residência podem incluir, segundo uma regra de chance determinística, o código de impressão digital do responsável. Crimes ocorridos fora de uma residência não geram esse achado de cena. A chance não aparece na interface nem nos metadados públicos.
- **Uso investigativo:** a comparação entre código parcial do vestígio e códigos consultados no Patient Search deve permitir reduzir os suspeitos de forma verificável, sem confirmar culpa por um único resultado ambíguo. Achados hospitalares e de mandado podem ser fixados no whiteboard e usados como provas na acusação existente.

## Critérios de aceite

- [x] O Hospital abre como serviço selecionável na cidade, apresenta identidade visual e conteúdo próprios e não depende de acessar a Delegacia ou outra empresa.
- [x] O acesso ao Patient Search exige sessão local válida; chamadas sem autorização, com cidade/local incorretos ou de outra conta são negadas pelo backend.
- [x] Todo NPC recebe um código no formato `^[A-Z]{2}[0-9]{5}$`; o código é determinístico, estável após reconstrução e único entre NPCs da mesma cidade.
- [x] A busca por nome retorna apenas prontuários de NPCs da cidade autorizada e apresenta o código de impressão digital sem expô-lo em diretórios ou APIs públicas de NPCs.
- [x] O prontuário de uma vítima de homicídio informa horário estimado, causa/arma quando conhecida e local onde o corpo foi encontrado, com dados persistidos ou reconstruídos de forma consistente a partir do caso.
- [x] Prontuários comuns e buscas de pessoas vivas não inventam dados de homicídio nem vestígios forenses.
- [x] Casos com cena residencial podem produzir, segundo uma regra de chance determinística documentada, o código de impressão digital do responsável no inventário de busca; casos com outras cenas nunca atribuem esse achado à residência. Regra inicial: 35%, derivada de `caseID + residential-fingerprint` por SHA-256, sem expor a chance na interface ou nos metadados.
- [x] O exame pós-morte pode produzir cabelo com a cor do responsável ou uma parte do código de impressão digital dele; a seleção e o conteúdo permanecem estáveis para o mesmo caso e não revelam diretamente a identidade do responsável.
- [x] Os vestígios de cabelo e impressão digital podem ser consultados e comparados com prontuários; a apresentação não inclui classificação interna, peso de acusação nem indicação explícita de correspondência com o culpado.
- [x] Registros hospitalares e vestígios de mandado são persistidos, recuperados sem duplicação após retry/reinício, e podem ser descobertos, pinados e apresentados como prova na acusação.
- [x] Cobertura automatizada verifica formato, unicidade e estabilidade dos códigos; isolamento/autorização hospitalar; dados de corpo para casos e não casos; e resultados determinísticos para cenas residenciais e não residenciais.

## Fora de escopo

Exames laboratoriais interativos, banco de DNA, comparação automática que nomeia o culpado, identificação de suspeitos por retrato, consulta irrestrita a prontuários fora da cidade e alteração da fórmula de julgamento do M13. Qualquer mudança na ponderação das provas deve ser especificada em milestone próprio.

## Evidências de aceite

- Backend: `CGO_ENABLED=0 GOCACHE=/tmp/noir-m21-gocache go test ./internal/cases ./internal/identity ./internal/world` — aprovado.
- Frontend: `npm test -- --run apps/web/src/modules/game/pages/main/HospitalPatientSearch.test.jsx apps/web/src/modules/game/pages/main/ComputerDesktop.test.jsx apps/web/src/modules/cases/pages/files/Whiteboard.test.jsx` — 3 arquivos e 9 testes aprovados.
- Build de produção: `npm run build` — aprovado.
- Integridade do diff: `git diff --check` — aprovado.
