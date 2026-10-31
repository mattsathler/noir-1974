# I05 — Tipos procedurais de empresas

**Status:** Proposta  
**Origem:** Observação sobre geração de empresas; M24 — Company Websites procedurais; M25 — bolsa de valores e progressão do jogador  
**Escopo:** Médio

## Problema

Empresas têm nomes, endereços e funcionários, mas não possuem uma classificação de atividade própria. Os cargos atuais vêm de uma lista genérica, e o website tenta inferir o setor a partir de palavras no nome; se não encontra uma palavra conhecida, escolhe conteúdo genérico. Assim, o nome empresarial não garante que o jogador entenda que serviço a empresa presta, e suas páginas, funcionários e referências podem se contradizer.

## Resultado esperado

Cada empresa recebe um tipo de atividade determinístico e explícito, separado do nome, endereço e identidade visual. O tipo explica o que a empresa faz e influencia cargos dos NPCs, conteúdo público do website e outras informações investigáveis que a citam.

## Escopo

- Criar um catálogo versionado de pelo menos 30 tipos empresariais, com ID estável, rótulo, descrição e dados suficientes para orientar as fontes do mundo.
- Incluir, no mínimo, tipos como: laboratório; serviços contábeis; escritório de tecnologia; banco; seguradora; escritório jurídico; consultoria; arquitetura e engenharia; construção; indústria metalúrgica; indústria têxtil; indústria química; editora; gráfica; jornal; rádio e telecomunicações; transportadora; ferrovia; navegação e carga; companhia aérea; hotel; restaurante; varejo de alimentos; loja de departamentos; oficina mecânica; concessionária de veículos; energia elétrica; abastecimento de água; hospital ou clínica; farmácia; escola; instituto de pesquisa; cooperativa agrícola; mineração; cinema e teatro; segurança privada; limpeza e manutenção.
- Atribuir o tipo a cada empresa de forma determinística a partir do mundo/seed e do ID estável da empresa, sem derivá-lo do nome gerado ou do endereço. Manter distribuição suficiente para que o catálogo seja visível na população empresarial.
- Expor o tipo como dado público da empresa em APIs e superfícies que listam ou descrevem a empresa, sem usá-lo como chave de identificação.
- Gerar cargos coerentes com a atividade da empresa, mantendo cargos administrativos compartilhados quando fizerem sentido e sem atribuir à equipe somente cargos idênticos.
- Gerar websites com apresentação, serviços, produtos, avisos e linguagem coerentes com o tipo. A classificação explícita substitui a inferência do setor feita a partir do nome; a composição visual do site continua podendo variar independentemente.
- Usar o tipo para dar contexto a outras informações que o jogador encontra, incluindo diretórios de empresas/funcionários, publicidade e notícias ou registros empresariais quando aplicável. O mesmo tipo deve produzir informações coerentes entre fontes.
- Preservar a separação entre nome próprio, tipo de atividade, endereço, ID da empresa e perfil visual do website.
- Versionar os dados e regras do catálogo para que reconstruções com a mesma versão preservem classificação, cargos e conteúdo procedural.

## Critérios de aceite

- [ ] O catálogo validado contém no mínimo 30 tipos distintos, cada qual com ID, rótulo e definição de atividade não vazios.
- [ ] Cada empresa recebe um tipo não vazio e estável para a mesma seed, versão e ID; o tipo não deriva de nome, endereço ou ordem de renderização.
- [ ] A distribuição determinística utiliza mais de um tipo nos mundos gerados e permite a presença dos tipos do catálogo sem concentrar todas as empresas em uma categoria genérica.
- [ ] O tipo é exposto em APIs e apresentado de forma legível nos diretórios/consultas de empresas, sem substituir ID, nome ou endereço.
- [ ] Cargos gerados para cada tipo pertencem ao perfil ocupacional definido para aquele tipo; cargos administrativos compartilhados permanecem válidos quando aplicáveis.
- [ ] Websites de pelo menos 30 tipos podem apresentar descrições, serviços/produtos e avisos coerentes com a atividade atribuída à empresa.
- [ ] O setor do website é obtido do tipo explícito da empresa, e alterar o nome sem mudar o ID/tipo não muda a atividade nem os cargos.
- [ ] Informações empresariais apresentadas em diretórios, websites, emails/notícias e fontes corporativas aplicáveis não contradizem o tipo atribuído.
- [ ] Nome, tipo, endereço, ID e identidade visual continuam independentes; o endereço permanece a localização física e relações/carteira continuam usando o ID da empresa.
- [ ] Testes cobrem validação e cobertura do catálogo, determinismo, coerência entre tipo/cargos/website e consistência das informações empresariais expostas.

## Fora de escopo

- Adicionar, remover ou transferir empresas, funcionários, endereços ou posições de ações.
- Mudar os nomes procedurais das empresas definidos pela I04, a menos que seja necessário remover dependências indevidas entre nome e tipo.
- Tornar empresas ou seus tipos editáveis pelo jogador.
- Criar novos crimes financeiros, arquétipos de caso ou impactos de eventos corporativos sobre a bolsa; cada extensão deve ter seu próprio escopo.
- Reformular os perfis visuais dos websites ou criar um sistema de navegação diferente.
