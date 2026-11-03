# M36 — Blackwire: perfis e downloads de aplicativos

**Status:** Em curso
**Trilha:** Blackwire, aplicativos e segurança do computador  
**Depende de:** M23 — Sistema operacional, Proxy e sites investigativos; M32 — Identidade e progressão Blackwire; M33 — Exposição e defesas virtuais; M35 — Login e experiência de jogo dentro do desktop

## Objetivo

Transformar a Blackwire no ponto de encontro e na central de ferramentas para jogadores que investigam a rede por meios hackers. A milestone apresenta um perfil Blackwire próprio, amplia o site com uma área de downloads e integra a obtenção e instalação de aplicativos ao terminal do MERID_OS.

## Direção definida

- O registro Blackwire pede somente um pseudônimo (`username`) e uma foto de perfil escolhida em um catálogo de retratos prontos. Não cria senha nem uma sessão de autenticação independente.
- O perfil é vinculado à conta de jogo e ao servidor já definidos pela M32. O nome de usuário é único nesse servidor.
- A interface apresenta o perfil como uma conta, um nome de usuário e uma imagem. O termo “alter ego” não aparece em textos, rótulos, mensagens ou instruções voltadas ao jogador; a separação entre a vida legal e a presença clandestina deve ser percebida pela experiência.
- A seleção de foto usa um catálogo de identificadores estáveis e imagens prontas. Não há upload, geração ou edição de retrato nesta milestone.
- Os retratos finais, incluindo a imagem de The Crow, serão fornecidos como conteúdo visual posteriormente. O catálogo deve aceitar esses arquivos sem mudança de schema; placeholders podem ser usados até a entrega dos assets.
- O registro só aparece para conta/servidor sem perfil Blackwire ativo. Após registrar, o usuário chega à página principal da Blackwire, que mantém o mercado existente e apresenta uma aba **Downloads**.
- Um perfil já existente mantém handle, Crypto, reputação, nível, banimento e desbloqueios. A M36 não reinicia nem converte esses dados. Contas antigas recebem uma imagem padrão se ainda não tiverem foto.
- A área **Software** funciona como um marketplace de aplicativos ilegais. Uma lista de ofertas abre uma página própria para cada aplicativo, com descrição, versão, criador, origem reivindicada, categoria, disponibilidade/instalação, endereço do pacote e comandos completos do Terminal (`fetch`, `install`, `run` e `uninstall` quando instalado). Não há bloqueio de acesso por nível, conforme a decisão posterior à M32; a M36 não cria uma segunda moeda nem exige pagamento com Crypto.
- Os aplicativos têm criadores diferentes, identificados por pseudônimos de várias redes. As origens exibidas são alegações não verificadas; The Crow é apenas um dos publicadores do catálogo.
- Ghostkey, Mailbreaker e Bankrupt continuam disponíveis sem requisito de nível. O piloto instalável da M36 será `DEADLETTER.EXE`, um leitor offline e somente de leitura de um pequeno arquivo publicado pelo The Crow; ele valida distribuição e instalação sem criar uma nova mecânica de investigação.
- O site mostra os comandos do aplicativo, mas não executa a instalação por botão. A sequência ocorre no Terminal:

  ```text
  disable protection
  fetch <link-do-pacote>
  install <link-do-pacote>
  run <nome-do-app>
  uninstall <nome-do-app>
  ```

- Com a proteção do MERID_OS ligada, `fetch`, `install` e a execução de aplicativos sem assinatura são bloqueados com uma mensagem diegética de que o publisher ou aplicativo não é confiável. A tentativa bloqueada não altera o estado do pacote. O jogador executa `disable protection` e repete o comando; o cliente Blackwire do navegador continua disponível para consultar as instruções.
- `BLACKWIRE.EXE` é o cliente confiável do navegador e permanece disponível para abrir a própria Blackwire; os demais executáveis são avaliados pela proteção do MERID_OS.
- `fetch` apenas obtém o pacote para uma área temporária; `install` exige que ele tenha sido obtido e esteja desbloqueado; `run` exige que esteja instalado. Cada etapa apresenta confirmação ou erro compreensível no Terminal.
- `disable protection` desativa a proteção somente durante a sessão atual do computador. Logout ou reboot reativa a proteção automaticamente. Um comando `enable protection` pode reativá-la antes disso.
- `uninstall` remove um aplicativo opcional instalado, como `DEADLETTER.EXE`, e fecha sua janela se estiver aberta. Aplicativos incluídos no computador não podem ser removidos. A proteção do MERID_OS não precisa ser desativada para desinstalar; o pacote pode ser obtido e instalado novamente pela Blackwire.
- Apenas os links de pacotes publicados no catálogo Blackwire são aceitos. Não há execução de código arbitrário, instalação por URL externa ou instalação silenciosa pelo site.
- Aplicativos instalados continuam disponíveis após logout/reboot no escopo conta/servidor. A M36 reutiliza os eventos e limites de segurança existentes da M33, sem redefinir seus limiares ou sanções.
- A foto padrão para perfis Blackwire antigos será `portrait-01`; novos perfis escolhem um retrato de um conjunto pequeno de combinações prontas do catálogo de arte.
- A Blackwire é apresentada como a central de ferramentas e identidade dessa comunidade. Feed público, mensagens entre jogadores, fóruns, chat e recursos sociais não fazem parte desta milestone.

## Critérios de aceite

- [ ] Uma conta sem perfil Blackwire pode registrar um pseudônimo único no servidor e selecionar um retrato pronto, sem criar senha Blackwire.
- [ ] O perfil fica vinculado à conta de jogo e ao servidor existentes; perfil, Crypto, reputação, nível e desbloqueios antigos são preservados.
- [ ] O vocabulário visível ao jogador não usa “alter ego” e apresenta a identidade de forma diegética por meio do username e do retrato.
- [ ] Depois do registro, a página principal da Blackwire é aberta e oferece um marketplace de aplicativos sem remover o mercado de informações atual.
- [ ] A lista de aplicativos abre páginas próprias com descrição, versão, criador, origem reivindicada, categoria, disponibilidade e comandos adequados ao estado instalado/não instalado, sem exigir nível para acesso.
- [ ] Com proteção ligada, `fetch`, `install` e `run` de aplicativos sem assinatura falham com aviso diegético e não avançam a instalação; depois de `disable protection`, a sequência `fetch` → `install` → `run` funciona para um aplicativo disponível.
- [ ] As dependências entre obter, instalar e executar são aplicadas; a proteção volta a ser ativada após logout/reboot.
- [ ] Pelo menos um aplicativo piloto novo pode ser obtido na Blackwire e executado como janela/aplicativo no desktop da M35.
- [ ] O retrato é escolhido de um catálogo de assets pronto, sem upload ou geração pelo jogador; a imagem final de The Crow pode ser adicionada depois sem alterar o modelo.

## Fora de escopo

- Senha, autenticação independente, recuperação de senha ou outro sistema de contas paralelo ao login do jogo.
- Upload, geração ou edição de fotos de perfil.
- Feed, fóruns, chat ou mensagens entre jogadores.
- Comprar aplicativos com Crypto, alterar o mercado atual ou mudar as regras de progressão Blackwire da M32.
- Alterar os limites de exposição, avisos ou sanções de segurança estabelecidos pela M33.
- Executar pacotes arbitrários ou permitir downloads fora do catálogo autorizado da Blackwire.
- Implementar os aplicativos futuros além do aplicativo piloto necessário para validar o ciclo completo.
