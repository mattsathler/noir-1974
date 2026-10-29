# ADR 0013 — Websites empresariais procedurais sob demanda

- **Status:** Aceita
- **Data:** 2026-10-29
- **Contexto:** M24 — Company Websites procedurais

## Contexto

O Meridian National Directory deve encontrar empresas em escala nacional sem baixar funcionários nem o conteúdo completo de cada website. Os websites precisam ser únicos por empresa, estáveis após reconstrução do mundo e coerentes com o endereço já usado pelo Proxy e pelas autorizações.

## Decisão

- Gerar o perfil público no backend sob demanda a partir da seed do mundo, versão do gerador e ID estável da empresa.
- Não persistir uma cópia materializada do website. A cache existente pode manter o catálogo nacional leve; qualquer cache de perfil de site inclui seed, versão e ID de empresa.
- Usar modelos de página customizáveis combinados com tema, composição, marca gráfica e conteúdo determinísticos. O perfil pode ser recriado sem migrations de banco.
- O diretório Meridian pesquisa metadados compactos (nome, domínio, descrição e endereço) e retorna até dez resultados no país Proxy. `GET /api/world/companies/{companyID}/website` carrega apenas o website solicitado e resolve a empresa dentro do país ativo.
- O conteúdo público e o diretório interno são superfícies separadas. A API de website nunca inclui funcionários; o diretório de funcionários continua exigindo tentativa de acesso vinculada à conta, empresa, localização e jurisdição.
- O painel de login não autentica, recebe ou armazena credenciais. É somente ambientação e sinaliza que o acesso não está disponível.

## Consequências

- Mesma seed, versão e ID recriam o mesmo domínio, tema e conteúdo.
- Empresas não adicionam linhas de website ao banco; mutations de empresa/endereço e versionamento do mundo governam a reconstrução do perfil.
- A API faz uma leitura de versão do catálogo nacional ao resolver o site, reaproveita o cache do catálogo e gera o perfil pedido; nunca consulta cada site durante a busca.
- O catálogo de templates, paletas e conteúdo precisa ter diversidade testada sobre o conjunto de empresas materializado.
- Alterações na composição procedural que devam mudar websites exigem incremento explícito da versão do website/gerador e revisão das referências visuais.

## Alternativas consideradas

- Persistir cada site e suas páginas em tabelas: rejeitado por duplicar dados reconstruíveis e criar trabalho de migration/invalidação.
- Entregar o website inteiro em cada resultado do diretório: rejeitado por aumentar payload e gerar conteúdo que o jogador não abriu.
- Gerar a página no frontend: rejeitado porque busca, domínio, endereço e conteúdo precisam de uma fonte autoritativa consistente e testável.

