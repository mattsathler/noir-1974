# apps/web

Cliente web de Noir 1974, desenvolvido em React e servido pelo Vite. Consome a API em `apps/api` pelo prefixo `/api`.

O módulo de identidade oferece criação de conta, login por e-mail, sessão em cookie e a configuração inicial do detetive particular. No primeiro acesso, o jogador escolhe o nome e um entre seis espaços de retrato. As artes finais dos retratos ainda serão produzidas segundo a [Regra 0001](../../docs/rules/0001-direcao-artistica.md).

O cliente envia intenções; o servidor valida e persiste a identidade. Módulos, páginas e testes ficam organizados conforme as [regras de frontend](../../docs/dev-rules/frontend.md).
