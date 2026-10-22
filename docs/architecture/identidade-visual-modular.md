# Identidade visual modular

Personagens são representados somente pelo rosto, seguindo o modelo modular pixel art de referência. O frontend monta os PNGs transparentes; a API envia identificadores de camada e não transporta imagens.

## Camadas e IDs

`GET /api/character-visuals` retorna os identificadores permitidos, separados em `hair`, `skinColors`, `eyeColors`, `faceDetails` e `accessories`. Os PNGs ficam separados nas pastas `hair`, `skin`, `eyes`, `face` e `accessories` dentro de `apps/web/public/images/characters/`; o frontend resolve cada ID para sua pasta pelo tipo de camada.

Exemplo de identidade retornada ou enviada pela API:

```json
{
  "hairId": "hair_long_ginger",
  "skinColorId": "skin_light",
  "eyeColorId": "eyes_green",
  "faceDetailId": "face_beard_goatee_ginger",
  "accessoryIds": ["face_glasses_round_tortoise"]
}
```

Cada sprite usa a mesma tela transparente de `25 × 25` pixels, origem e alinhamento que as demais camadas. Os IDs usam nomes descritivos sem sufixos numéricos; cortes alternativos e formatos distintos são nomeados pelo estilo, como `hair_short_alt` e óculos `round`/`square`. `skin_light.png`, `hair_short_base.png`, `hair_short_alt_black.png`, `hair_long_black.png`, `hair_ponytail_black.png`, `hair_tied_long_black.png`, `hair_emo.png`, `face_eyes_black.png`, `face_beard_base.png` e `face_beard_goatee_black.png` são bases artísticas. O ponytail deriva da base de cabelo longo, preserva o alinhamento do topo e prende a mecha atrás da cabeça. `hair_emo.png` é mantido como base autoral; o gerador preserva sua arte preta em `hair_emo_black.png` e cria os demais tons. O cavanhaque preto é a referência de posição para os formatos adicionais de barba; a geração preserva esse PNG e deriva os demais tons dele. `generate_hair_and_beard_tones.mjs` mapeia cores pela luminosidade para manter a sombra escura à esquerda, junto à orelha, e a cor clara à direita, sem alterar forma ou transparência. Há dois modelos curtos, um longo, ponytail, tied, tied long, pigtails e emo, cada um em cinco tons. Olhos têm sete cores derivadas de `face_eyes_black.png`. Barba, bigode e cabelo compartilham os tons preto, castanho, ruivo, loiro e grisalho; barbas têm quatro formatos, incluindo cavanhaque, barba por fazer e costeletas. Os dois modelos de óculos têm acabamento preto, tartaruga, latão, prata ou marfim. NPCs recebem bigode ou barba no mesmo tom do cabelo. O frontend compõe nesta ordem: pele, cabelo, chapéus, olhos, detalhe facial e acessórios. A composição não inclui corpo. Os sprites ficam organizados em `apps/web/public/images/characters/{hair,skin,eyes,face,accessories}/`.

## NPCs

Mundos novos usam a versão 4 do gerador. A aparência procedural é obtida de uma stream determinística própria, derivada da seed e da chave do NPC; portanto, cada jogador do servidor recebe os mesmos IDs. A descrição textual da aparência é derivada do ID de cabelo selecionado. Versões 1–3 continuam reconstruíveis com o comportamento anterior; mundos já materializados não mudam de versão automaticamente.

Os IDs são enviados em `npc.visual` tanto na busca de pessoas quanto nos perfis completos. O frontend resolve cada ID para o PNG de mesmo nome na pasta correspondente ao tipo da camada.

## Identidade do jogador e arquivos de arte

O personagem do jogador usa a mesma composição sem corpo e persiste os IDs na conta, independentemente da seed do mundo. A criação envia os IDs para `POST /api/characters`; a API valida a seleção contra o catálogo e `GET /api/me` retorna a aparência salva. O editor já oferece cabelo, pele, olhos, detalhes faciais e acessório.

Cada personagem pode ter um `faceDetailId` opcional, ou seja, no máximo um detalhe facial. As opções incluem formatos e tons de barba, bigode e outros detalhes faciais. Cabelos são sprites completos já coloridos; acessórios também são opcionais. Novas variantes devem receber IDs estáveis e ser adicionadas ao catálogo, sem renomear arquivos já persistidos em identidades. A composição está ligada ao editor e aos registros de NPCs e só aparece quando todos os PNGs correspondentes estiverem disponíveis.
