# apps/web

Cliente web de Noir 1974, desenvolvido em React e servido pelo Vite. Consome a API em `apps/api` pelo prefixo `/api`.

O módulo de identidade oferece criação de conta, login por e-mail, sessão em cookie e a configuração inicial do detetive particular. No primeiro acesso, o jogador escolhe o nome e um entre seis espaços de retrato. As artes finais dos retratos ainda serão produzidas segundo a [Regra 0001](../../docs/rules/0001-direcao-artistica.md).

O cliente envia intenções; o servidor valida e persiste a identidade. Módulos, páginas e testes ficam organizados conforme as [regras de frontend](../../docs/dev-rules/frontend.md).

O terminal dentro da tela da cidade consome o diretório autenticado da API: INTERPOL consulta perfis de NPC com relações, e CITY RECORDS pesquisa empresas, funcionários e endereços. O jornal abre em uma folha/modal independente e apresenta o caso recente da cidade como manchete, seguido por notícias leves geradas proceduralmente. As chamadas do módulo de mundo ficam em `src/modules/world/api.js`.

Retratos sem corpo são montados em `src/modules/identity/components/ModularFace.jsx` a partir dos IDs de `/api/character-visuals` e dos dados visuais de NPCs e personagens. Os PNGs transparentes ficam em `public/images/characters/`, separados pelas pastas `hair`, `skin`, `eyes`, `face` e `accessories`, usando a mesma tela e o mesmo alinhamento; veja o [contrato de identidade visual modular](../../docs/architecture/identidade-visual-modular.md).
