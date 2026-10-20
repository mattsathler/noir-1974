# Geração do mundo

## Ordem em camadas

A ordem exata ainda pode ser refinada. O mundo é gerado por regiões sob demanda; dentro de cada recorte, o princípio não muda: **estrutura coerente primeiro; casos depois**. A geração precisa ser determinística por servidor, região e versão.

```text
WORLD SEED
    ↓
1. Geografia
    ↓
2. Edifícios / locais
    ↓
3. Empresas / instituições
    ↓
4. NPCs
    ↓
5. Famílias e relacionamentos
    ↓
6. Empregos e estruturas sociais
    ↓
7. Histórico / eventos
    ↓
8. Casos
```

## Grafo

Entidades (nós) possíveis: NPC, empresa, residência, prédio, família, local, organização, caso, evidência.

Relações (arestas) possíveis: casado com, irmão de, pai/mãe de, amigo de, rival de, trabalha em, gerente de, mora em, proprietário de, conhece, esteve em, etc.

## Consultar um NPC

1. Localizar seed.
2. Gerar atributos.
3. Consultar relações.
4. Aplicar mutations.
5. Montar o estado atual permitido ao cliente.

Resultado típico: Adam Miller, casado com Gina, empregado na Empresa #13 — sem gravar tudo isso na seed do NPC.
