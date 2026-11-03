# I11 — Dados pessoais procedurais de NPCs

**Status:** Proposta  
**Origem:** Perfis e registros pessoais de NPCs  
**Escopo:** Médio

## Problema

Os registros de pessoas têm poucos dados pessoais para sustentar investigação, identificação e ambientação. Isso limita a variedade de detalhes que podem aparecer em diretórios, arquivos, documentos e outras fontes do mundo.

## Resultado esperado

NPCs têm um perfil pessoal procedural mais completo e consistente. Seus dados ajudam a individualizar as pessoas e podem ser encontrados em fontes compatíveis com o tipo de informação, sem transformar todo registro público numa ficha completa.

## Escopo

- Gerar e manter os seguintes dados por pessoa:
  - sexo;
  - idade;
  - data de nascimento contendo somente dia e mês, sem ano;
  - nacionalidade;
  - naturalidade, como cidade e país de nascimento;
  - estado civil;
  - idioma principal;
  - religião, incluindo valores desconhecidos ou não declarados;
  - número de identificação fictício e estável;
  - assinatura visual ou textual estável, derivada do nome;
  - tipo sanguíneo;
  - tipo físico;
  - altura;
  - peso;
  - mão dominante;
  - temperamento;
  - hobbies.
- Manter idade como um atributo próprio, coerente com o perfil e com o tempo do mundo. A data de nascimento apresentada contém apenas dia e mês; não gerar nem exibir ano de nascimento.
- Gerar naturalidade, nacionalidade e idioma principal de forma compatível com a história e a localização da pessoa, sem torná-los necessariamente iguais.
- Manter coerência entre altura, peso e tipo físico, sem usar nacionalidade, sexo ou aparência como regra estereotipada para essas características.
- Tratar religião, estado civil e demais informações que possam não ser conhecidas como valores opcionais ou não declarados, em vez de inventar certeza em toda fonte.
- Gerar os dados de forma determinística e estável para a mesma pessoa, preservando a identidade entre casos, registros e reconstruções do mundo.
- Disponibilizar cada informação em interfaces e documentos onde ela faça sentido investigativo ou narrativo. A melhoria não exige publicar todos os dados em todo diretório ou site.
- Usar números de identificação exclusivamente fictícios, sem reproduzir números pessoais reais válidos.

## Critérios de aceite

- [ ] O perfil procedural de NPC suporta e gera os campos definidos nesta melhoria, incluindo altura, peso e hobbies.
- [ ] A data de nascimento exposta contém apenas dia e mês e nunca mostra ano; idade permanece como campo separado.
- [ ] Naturalidade, nacionalidade e idioma principal podem variar de forma coerente entre si e com a história da pessoa.
- [ ] Altura, peso e tipo físico são plausíveis e coerentes; demais atributos não são determinados por estereótipos de sexo, nacionalidade ou aparência.
- [ ] Informações opcionais ou não declaradas são representadas sem valores contraditórios ou preenchimento forçado.
- [ ] Número de identificação e assinatura são fictícios, estáveis e associados à mesma pessoa após reconstrução do mundo.
- [ ] Interfaces e documentos apresentam os dados em fontes adequadas, sem expor automaticamente uma ficha completa em todo site público.
- [ ] A inclusão dos novos atributos não altera IDs, relações, aparência ou papéis existentes dos NPCs.

## Fora de escopo

- Criar novos sistemas de investigação, buscas ou sites exclusivamente para consultar esses dados.
- Alterar a identidade visual modular ou a aparência física já gerada para NPCs.
- Usar dados pessoais reais ou emitir documentos oficiais válidos fora do mundo ficcional.
- Incluir o ano na data de nascimento apresentada ao jogador.
