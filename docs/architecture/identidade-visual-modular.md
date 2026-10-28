# Identidade visual modular

Personagens são representados somente pelo rosto, em sprites pixel art transparentes. O frontend monta as camadas PNG; a API envia e persiste apenas seus IDs. As seis categorias são `Accessories`, `Beard`, `Eyes`, `Face`, `Hair` e `Skin`.

## Identidade e catálogo

`GET /api/character-visuals` retorna listas `accessories`, `beard`, `eyes`, `face`, `hair` e `skin`, além do mapa `compatibility` que declara as apresentações aceitas para cada ID. A identidade usa este formato:

```json
{
  "hairId": "hair_short_brown",
  "skinId": "skin_white",
  "eyesId": "eyes_brown",
  "faceId": "face_scar_short",
  "beardId": "beard_moustache_brown",
  "accessoryIds": ["accessories_glasses_round_tortoise"]
}
```

`faceId`, `beardId` e um acessório são opcionais. A API valida cada ID e salva a seleção na conta. `GET /api/me` retorna os mesmos campos. Registros criados com o formato anterior são adaptados ao serem lidos: cabelos antigos sem sprite correspondente usam o corte curto na cor registrada, detalhes faciais são separados entre barba e face, e acessórios sem variante compatível são descartados.

## Sprites e variantes

Os PNGs ficam em `apps/web/public/images/characters/{accessories,beard,eyes,face,hair,skin}/`. As peças atuais usam tela transparente de `40 × 40` pixels e compartilham origem e alinhamento. A composição é feita na ordem pele, cabelo, olhos, barba, face e acessórios; não há corpo.

O gerador `scripts/generate_character_color_variants.mjs` lê a luminosidade da paleta-base e mapeia cada nível de sombra para paletas de destino. Os níveis são ordenados do mais escuro ao mais claro, preservando forma, transparência e separação das sombras desenhadas; se uma base de cabelo tiver quatro tons, a paleta de destino é interpolada para preservar os quatro níveis. As cores de cabelo, barba, pele, olhos e acessórios usam tons mais escuros e menos saturados. Bases nomeadas como `hair_<corte>_brown.png` e `beard_<formato>_brown.png` geram automaticamente cinco variantes de cor. Os cortes atuais incluem short, long, messy, wavy, bob, slicked, long wavy, long curly, mohawk, balding e sparse sides. Os formatos de barba incluem moustache, mutton, soul patch, goatee e full. Execute `npm run generate:character-variants` após atualizar um sprite-base; bases cujo nome também representa uma variante não são sobrescritas pelo gerador.

## NPCs

NPCs das versões 4, 5 e 6 recebem aparência determinística e enviam as seis categorias em `npc.visual`; a descrição textual é derivada dos mesmos IDs. A versão 4 conserva sua seleção histórica para reconstrução. A versão 5 usa um fluxo visual próprio e calibra cor e estilo do cabelo por idade. A versão 6 acrescenta `npc.presentation` (`masculine` ou `feminine`), validada separadamente de qualquer orientação sexual; hoje, cada peça de barba declara compatibilidade masculina, enquanto as demais peças aceitam ambas as apresentações. NPCs da versão 6 não recebem barba com apresentação feminina. Versões antigas conservam suas aparências históricas. Todos no mesmo servidor veem a mesma identidade e o backend não envia arquivos de arte.
