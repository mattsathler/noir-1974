# Noir UI

Base visual compartilhada do frontend Noir 1974. Os estilos vivem em `styles/` e são escritos em SCSS; `apps/web` deve consumir as classes definidas aqui em vez de manter regras de apresentação próprias.

## Estrutura

- `styles/_tokens.scss`: paleta, tipografia, escala de espaço, camadas e foco.
- `styles/_reset.scss`: normalização global mínima e defaults acessíveis.
- `styles/_layout.scss`: containers, flex, grid e alinhamento.
- `styles/_spacing.scss`: margens e paddings utilitários.
- `styles/_typography.scss`: tamanhos, pesos e alinhamento de texto.
- `styles/_color.scss`: cores de texto e fundo.
- `styles/_surface.scss`: bordas, superfícies e sombras.
- `styles/_utilities.scss`: acessibilidade, visibilidade e interação.
- `styles/_effects.scss`: efeitos visuais reutilizáveis.
- `styles/_typewriter.scss`: cursor e animação de digitação.
- `styles/index.scss`: entrada única que reúne a base.
- `components/Typewriter.jsx`: componente React reutilizável para revelar texto progressivamente.

O SCSS é a fonte do sistema. A aplicação deve importar `styles/index.scss` uma vez no ponto de entrada do frontend e compilar Sass na etapa de build. Componentes devem combinar classes utilitárias genéricas e tokens CSS (`--noir-*`); não devem replicar helpers em folhas locais nem criar classes vinculadas a uma página.

`Typewriter` aceita `text`, `speed` (intervalo entre caracteres, em milissegundos), `startDelay`, `onComplete` e `className`. A animação é reduzida a texto imediato quando o sistema pede movimento reduzido.

## Convenções

- Cada classe descreve uma propriedade ou comportamento reutilizável, como `pos-relative`, `overflow-hidden`, `min-h-full`, `grid` e `text-center`.
- Tokens CSS ficam sob `--noir-*` para permitir temas e ajustes sem duplicar valores.
- Escalas usam passos consistentes; adicione um token ou helper apenas quando houver uso real.
- Estilos específicos de componentes devem ser avaliados como parte deste sistema antes de criar CSS local em uma aplicação.
- Prefira HTML semântico e mantenha estados de foco visíveis.

## Exemplo

```jsx
<main className="container p-6 grid gap-6">
  <section className="surface border p-5">
  <h1 className="text-xl">Arquivo do caso</h1>
    <p className="text-muted">Registro atualizado.</p>
  </section>
</main>
```
