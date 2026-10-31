# I04 — Nomes procedurais próprios para empresas

**Status:** Proposta  
**Origem:** M24 — Company Websites procedurais; M25 — bolsa de valores e progressão do jogador; observação do gerador de empresas  
**Escopo:** Médio

## Problema

Empresas são geradas com um catálogo curto de nomes e, quando ele se esgota, o sistema acrescenta um ordinal ao nome. Isso produz identificadores artificiais como “Ateliers Lumière 383”. O endereço físico também pode se repetir entre cidades e não deve ser usado para diferenciar ou nomear empresas.

O nome exibido acompanha a empresa em vários lugares do jogo: diretórios, websites, vínculos de emprego, mensagens, jornal e interfaces de mercado. Uma troca parcial deixaria referências divergentes e poderia confundir a busca e a negociação.

## Resultado esperado

Cada empresa tem um nome próprio, determinístico para a seed e a versão do gerador, selecionado de grandes catálogos JSON de nomes possíveis. Endereços continuam representando localizações físicas; IDs estáveis continuam identificando entidades internamente. O mesmo nome da empresa é apresentado em todas as superfícies do jogo, inclusive no STOCK EXCHANGE.

## Escopo

- Criar catálogos JSON extensos, com milhares de nomes utilizáveis, organizados em arquivos/coleções adequados à geração de empresas. Os nomes devem ser próprios e não depender do endereço nem de um contador anexado.
- Conectar seleção e resolução de colisões ao mundo/seed e à versão do gerador, de modo que reconstruir o mesmo mundo produza os mesmos nomes.
- Gerar nomes únicos dentro do mercado nacional de um mesmo mundo. Se a seleção inicial colidir, escolher deterministicamente outro nome disponível do catálogo, sem criar sufixos ordinais artificiais.
- Preservar o ID interno estável da empresa como chave de APIs, relações, sites e carteira. O nome exibido não pode ser usado como chave primária nem alterar relações de emprego, localização, acesso ou propriedade de ações.
- Manter o endereço da sede e demais locais como dados separados. Endereços podem repetir entre cidades ou abrigar empresas distintas sem mesclar suas identidades.
- Propagar o nome oficial da empresa em toda comunicação e interface: diretórios e buscas, Company Websites, anúncios, perfis corporativos, emails, jornal, contexto de casos e registros, bolsa/websites de mercado e STOCK EXCHANGE.
- Atualizar consultas e ordenação para pesquisarem e exibirem o nome próprio, mantendo busca por endereço onde ela já existe.
- Definir versionamento de dados e compatibilidade de nomes já persistidos/gerados, para que a atualização não mude silenciosamente a identidade de mundos existentes.

## Critérios de aceite

- [ ] Os catálogos de nomes são dados JSON versionados e incluem milhares de opções válidas, sem placeholders que dependam de números ordinais para completar o catálogo.
- [ ] A mesma seed, versão e empresa produzem sempre o mesmo nome após reconstrução; seeds ou entidades diferentes percorrem o catálogo de forma determinística.
- [ ] Empresas listadas no mesmo mercado nacional de um mundo não têm nomes exibidos duplicados; colisões são resolvidas por outra opção do catálogo, sem acrescentar números ao nome.
- [ ] Nome, ID da empresa e endereço são campos distintos em geração, APIs e apresentação; repetir um endereço entre cidades não mescla empresas.
- [ ] Relações, consultas, permissões de acesso, sites e carteira continuam ancorados no ID estável da empresa, mesmo se nomes forem alterados em uma futura versão.
- [ ] Diretórios, busca Meridian, páginas empresariais, emails, jornal e demais fontes que mencionem uma empresa mostram o mesmo nome oficial.
- [ ] O diretório nacional e buscas relevantes encontram empresas pelo nome próprio e continuam encontrando-as por endereço quando essa busca já é suportada.
- [ ] O STOCK EXCHANGE, a bolsa no website e qualquer histórico/posição exibem o nome oficial correspondente ao `companyId`, sem nomes com ordinal nem divergência entre telas.
- [ ] A atualização de nomes não altera endereços, funcionários, relações, autorização de sites, IDs empresariais ou posições de ações existentes.
- [ ] Testes cobrem carregamento/validação dos JSONs, determinismo, ausência de colisões no escopo nacional e propagação dos nomes nas principais superfícies, incluindo a bolsa.

## Fora de escopo

- Alterar endereços, mapas, sedes ou regras de localização e acesso às empresas.
- Trocar IDs empresariais, vínculos, dados de mercado ou carteira de ações.
- Mudar o conteúdo temático, a identidade visual ou os módulos procedurais dos websites, além de corrigir referências ao nome oficial.
- Tornar nomes empresariais editáveis pelo jogador.
- Reescrever retroativamente nomes de mundos/casos existentes sem uma regra explícita de migração e versionamento.
