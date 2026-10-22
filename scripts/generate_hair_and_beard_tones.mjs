// Derive matching character-layer variants from artist-authored pixel sprites.
// Source files remain untouched; recolored output keeps the exact pixel mask.
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { deflateSync, inflateSync } from "node:zlib";
import { join } from "node:path";

const root = join("apps", "web", "public", "images", "characters");
const directories = Object.fromEntries(["hair", "skin", "eyes", "face", "accessories"].map((name) => [name, join(root, name)]));
const directoryFor = (id) => {
  if (id.startsWith("hair_")) return directories.hair;
  if (id.startsWith("skin_")) return directories.skin;
  if (id.startsWith("eyes_") || id.startsWith("face_eyes_")) return directories.eyes;
  if (id.startsWith("accessory_") || id.startsWith("face_glasses_")) return directories.accessories;
  return directories.face;
};
const assetPath = (id) => join(directoryFor(id), `${id}.png`);
const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
const tones = [
  ["black", ["#302720", "#211d1a"]],
  ["brown", ["#633b27", "#402b20"]],
  ["ginger", ["#d17b22", "#a65a1d"]],
  ["blonde", ["#d1ad54", "#9a8142"]],
  ["grey", ["#8d8880", "#69655e"]],
];

function decodeRGBA(png) {
  if (!png.subarray(0, 8).equals(signature)) throw new Error("Source is not a PNG.");
  let width, height, bitDepth, colorType, interlace;
  const compressed = [];
  for (let offset = 8; offset + 12 <= png.length;) {
    const length = png.readUInt32BE(offset);
    const type = png.toString("ascii", offset + 4, offset + 8);
    const data = png.subarray(offset + 8, offset + 8 + length);
    if (type === "IHDR") {
      width = data.readUInt32BE(0); height = data.readUInt32BE(4);
      bitDepth = data[8]; colorType = data[9]; interlace = data[12];
    }
    if (type === "IDAT") compressed.push(data);
    offset += 12 + length;
    if (type === "IEND") break;
  }
  if (bitDepth !== 8 || colorType !== 6 || interlace !== 0) throw new Error("Sources must be non-interlaced 8-bit RGBA PNGs.");
  const stride = width * 4, filtered = inflateSync(Buffer.concat(compressed));
  const pixels = Buffer.alloc(width * height * 4);
  const paeth = (a, b, c) => {
    const p = a + b - c, pa = Math.abs(p - a), pb = Math.abs(p - b), pc = Math.abs(p - c);
    return pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
  };
  for (let y = 0; y < height; y++) {
    const row = y * (stride + 1), filter = filtered[row];
    for (let x = 0; x < stride; x++) {
      const raw = filtered[row + 1 + x], left = x >= 4 ? pixels[y * stride + x - 4] : 0;
      const up = y ? pixels[(y - 1) * stride + x] : 0;
      const upperLeft = y && x >= 4 ? pixels[(y - 1) * stride + x - 4] : 0;
      let predictor = 0;
      if (filter === 1) predictor = left;
      else if (filter === 2) predictor = up;
      else if (filter === 3) predictor = Math.floor((left + up) / 2);
      else if (filter === 4) predictor = paeth(left, up, upperLeft);
      else if (filter !== 0) throw new Error(`Unsupported PNG filter ${filter}.`);
      pixels[y * stride + x] = (raw + predictor) & 0xff;
    }
  }
  return { width, height, pixels };
}

const crcTable = Array.from({ length: 256 }, (_, n) => {
  let c = n;
  for (let k = 0; k < 8; k++) c = (c & 1) ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
  return c >>> 0;
});
function crc32(buffer) {
  let c = 0xffffffff;
  for (const byte of buffer) c = crcTable[(c ^ byte) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}
function chunk(type, data) {
  const name = Buffer.from(type), size = Buffer.alloc(4), checksum = Buffer.alloc(4);
  size.writeUInt32BE(data.length); checksum.writeUInt32BE(crc32(Buffer.concat([name, data])));
  return Buffer.concat([size, name, data, checksum]);
}
function encodePNG(width, height, pixels) {
  const header = Buffer.alloc(13);
  header.writeUInt32BE(width, 0); header.writeUInt32BE(height, 4); header[8] = 8; header[9] = 6;
  const raw = Buffer.alloc(height * (width * 4 + 1));
  for (let y = 0; y < height; y++) pixels.copy(raw, y * (width * 4 + 1) + 1, y * width * 4, (y + 1) * width * 4);
  return Buffer.concat([signature, chunk("IHDR", header), chunk("IDAT", deflateSync(raw)), chunk("IEND", Buffer.alloc(0))]);
}
function rgba(hex) { return [1, 3, 5].map((i) => Number.parseInt(hex.slice(i, i + 2), 16)).concat(255); }

async function generateVariants(sourceName, idPrefix, { preserveBlackSource = false, palettes = tones } = {}) {
  const source = decodeRGBA(await readFile(assetPath(sourceName)));
  const counts = new Map();
  for (let i = 0; i < source.pixels.length; i += 4) {
    if (!source.pixels[i + 3]) continue;
    const key = source.pixels.subarray(i, i + 4).toString("hex");
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  if (counts.size < 1 || counts.size > 3) throw new Error(`${sourceName} must have one to three opaque colors; found ${counts.size}.`);
  const brightness = (hex) => {
    const [r, g, b] = [0, 2, 4].map((i) => Number.parseInt(hex.slice(i, i + 2), 16));
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  };
  const sourceColors = new Map();
  const orderedColors = [...counts.keys()].sort((a, b) => brightness(a) - brightness(b));
  if (orderedColors.length === 3) {
    sourceColors.set(orderedColors[1], 0); // recolor the lighter hair pixels
    sourceColors.set(orderedColors[0], 1); // recolor the hair shadow
    // The brightest color is the separate red tie accent and stays unchanged.
  } else {
    sourceColors.set(orderedColors[orderedColors.length - 1], 0); // light source pixels use the light color
    if (orderedColors.length === 2) sourceColors.set(orderedColors[0], 1); // dark source pixels use the shade
  }
  for (const [tone, palette] of palettes) {
    if (preserveBlackSource && tone === "black") continue;
    const targetColors = palette.map(rgba), pixels = Buffer.from(source.pixels);
    for (let i = 0; i < pixels.length; i += 4) {
      if (pixels[i + 3] === 0) continue;
      const key = pixels.subarray(i, i + 4).toString("hex");
      if (!sourceColors.has(key)) continue;
      const target = targetColors[sourceColors.get(key)];
      if (!target) throw new Error(`Unexpected color in ${sourceName}: #${key.slice(0, 6)}.`);
      pixels.set(target, i);
    }
    await writeFile(assetPath(`${idPrefix}_${tone}`), encodePNG(source.width, source.height, pixels));
  }
  const generated = palettes.length - Number(preserveBlackSource);
  console.log(`${sourceName}: ${source.width}x${source.height}, generated ${generated} tones; preserved source black art: ${preserveBlackSource}.`);
}

function beardShape(width, height, pixels, style, palette, anchor) {
  const set = (x, y, shade = 0) => pixels.set(rgba(palette[shade]), (y * width + x) * 4);
  if (style === "stubble") {
    for (const [dx, dy, shade] of [[-6,-3,1],[-4,-2,1],[-2,-1,1],[-1,-2,1],[0,0,1],[1,-1,0],[2,-2,0],[3,-1,0],[4,-3,0]]) {
      set(anchor.centerX + dx, anchor.bottomY + dy, shade);
    }
  }
}

async function generateBeardStyles() {
  const guide = decodeRGBA(await readFile(assetPath("face_beard_goatee_black")));
  const points = [];
  for (let i = 0; i < guide.pixels.length; i += 4) {
    if (guide.pixels[i + 3]) points.push({ x: (i / 4) % guide.width, y: Math.floor(i / 4 / guide.width) });
  }
  const anchor = {
    centerX: Math.round(points.reduce((sum, point) => sum + point.x, 0) / points.length),
    bottomY: Math.max(...points.map((point) => point.y)),
  };
  const { width, height } = guide;
  for (const [tone, palette] of tones) {
    const pixels = Buffer.alloc(width * height * 4);
    beardShape(width, height, pixels, "stubble", palette, anchor);
    await writeFile(assetPath(`face_beard_stubble_${tone}`), encodePNG(width, height, pixels));
  }
  console.log("Generated stubble beard variants aligned around the chin.");
}

const glassesPalettes = [
  ["black", ["#302a24", "#171411"]],
  ["tortoise", ["#684228", "#312117"]],
  ["brass", ["#c9a744", "#806625"]],
  ["silver", ["#aaa9a1", "#575b5b"]],
  ["ivory", ["#e0c99d", "#94816a"]],
];

async function generateEyeColors() {
  const sourceName = "face_eyes_black";
  const source = decodeRGBA(await readFile(assetPath(sourceName)));
  const opaque = new Set();
  for (let i = 0; i < source.pixels.length; i += 4) {
    if (source.pixels[i + 3]) opaque.add(source.pixels.subarray(i, i + 4).toString("hex"));
  }
  if (opaque.size !== 1) throw new Error(`${sourceName} must contain one opaque color; found ${opaque.size}.`);
  const sourceColor = [...opaque][0];
  const colors = [
    ["black", "#171513"], ["blue", "#597c9a"], ["brown", "#79502f"],
    ["green", "#66804c"], ["hazel", "#9b7840"], ["gray", "#777d80"], ["amber", "#b86b27"],
  ];
  for (const [tone, hex] of colors) {
    const pixels = Buffer.from(source.pixels), target = rgba(hex);
    for (let i = 0; i < pixels.length; i += 4) {
      if (pixels.subarray(i, i + 4).toString("hex") === sourceColor) pixels.set(target, i);
    }
    await writeFile(assetPath(`eyes_${tone}`), encodePNG(source.width, source.height, pixels));
  }
  console.log(`${sourceName}: generated ${colors.length} eye-color layers.`);
}

await Promise.all(Object.values(directories).map((directory) => mkdir(directory, { recursive: true })));
await generateVariants("hair_short_base", "hair_short");
await generateVariants("face_beard_base", "face_beard");
await generateVariants("face_beard_goatee_black", "face_beard_goatee", { preserveBlackSource: true });
await generateVariants("face_beard_mutton_black", "face_beard_mutton", { preserveBlackSource: true });
await generateVariants("hair_short_alt_black", "hair_short_alt", { preserveBlackSource: true });
await generateVariants("hair_long_black", "hair_long", { preserveBlackSource: true });
await generateVariants("hair_ponytail_black", "hair_ponytail", { preserveBlackSource: true });
await generateVariants("hair_tied_black", "hair_tied", { preserveBlackSource: true });
await generateVariants("hair_tied_long_black", "hair_tied_long", { preserveBlackSource: true });
await generateVariants("hair_messy_black", "hair_messy", { preserveBlackSource: true });
await generateVariants("hair_pigtails_black", "hair_pigtails", { preserveBlackSource: true });
await generateVariants("hair_emo_black", "hair_emo", { preserveBlackSource: true });
await generateVariants("face_moustache_black", "face_moustache", { preserveBlackSource: true });
await generateVariants("face_glasses_round_base", "face_glasses_round", { palettes: glassesPalettes });
await generateVariants("face_glasses_square_base", "face_glasses_square", { palettes: glassesPalettes });
await generateEyeColors();
await generateBeardStyles();
