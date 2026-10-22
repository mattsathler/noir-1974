// Recolor the new face-layer sprites while preserving their authored shading.
import { mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import { deflateSync, inflateSync } from "node:zlib";
import { dirname, join } from "node:path";

const root = join("apps", "web", "public", "images", "characters");
const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);

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
    offset += length + 12;
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
      pixels[y * stride + x] = (raw + predictor) & 255;
    }
  }
  return { width, height, pixels };
}

const crcTable = Array.from({ length: 256 }, (_, n) => {
  let c = n;
  for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
  return c >>> 0;
});
function crc32(buffer) {
  let c = 0xffffffff;
  for (const byte of buffer) c = crcTable[(c ^ byte) & 255] ^ (c >>> 8);
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
const rgba = (hex) => [1, 3, 5].map((index) => Number.parseInt(hex.slice(index, index + 2), 16)).concat(255);
function resizePalette(palette, count) {
  if (palette.length === count) return palette;
  if (palette.length < 2 || count < 1) return null;
  if (count === 1) return [palette[Math.floor(palette.length / 2)]];
  return Array.from({ length: count }, (_, index) => {
    const position = index * (palette.length - 1) / (count - 1);
    const left = Math.floor(position), right = Math.ceil(position), ratio = position - left;
    const from = rgba(palette[left]), to = rgba(palette[right]);
    return `#${from.slice(0, 3).map((value, channel) => Math.round(value + (to[channel] - value) * ratio).toString(16).padStart(2, "0")).join("")}`;
  });
}
const luminance = (hex) => {
  const [r, g, b] = [0, 2, 4].map((index) => Number.parseInt(hex.slice(index, index + 2), 16));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};

// Palette values are ordered from the deepest shadow to the brightest pixel.
const hairPalettes = [
  ["black", ["#1d1b19", "#282522", "#35312c"]],
  ["brown", ["#291c16", "#34251c", "#423027"]],
  ["ginger", ["#66391f", "#80502d", "#9b673b"]],
  ["blonde", ["#62543b", "#7a6848", "#95805a"]],
  ["grey", ["#41403d", "#585652", "#706e69"]],
];
const beardPalettes = [
  ["black", ["#1b1917", "#262320", "#33302b"]],
  ["brown", ["#291c16", "#34251c", "#423027"]],
  ["ginger", ["#60361f", "#794a2b", "#935f38"]],
  ["blonde", ["#5b503c", "#736448", "#8a7757"]],
  ["grey", ["#3c3c3a", "#52514e", "#696863"]],
];
const skinPalettes = [
  ["skin_white", ["#805b45", "#a77b61", "#c2977b", "#d2ac90"]],
  ["skin_pale", ["#79543f", "#a4775e", "#c39479", "#d7ad92"]],
  ["skin_light", ["#76523e", "#9d735a", "#bd9277", "#d0aa8f"]],
  ["skin_medium", ["#694532", "#875b43", "#a77355", "#bb896a"]],
  ["skin_olive", ["#624a37", "#7c6246", "#97805c", "#ac9672"]],
  ["skin_brown", ["#593a2d", "#744b38", "#905f47", "#a9765b"]],
  ["skin_dark", ["#483128", "#5b3c30", "#704c3c", "#87624d"]],
];
const eyePalettes = [
  ["eyes_black", "#262421"], ["eyes_blue", "#485d6c"], ["eyes_brown", "#2d170e"],
  ["eyes_green", "#536248"], ["eyes_hazel", "#786343"], ["eyes_gray", "#666967"], ["eyes_amber", "#805832"],
];
const accessoryPalettes = [
  ["black", ["#111212", "#252626"]], ["tortoise", ["#291d18", "#493427"]],
  ["brass", ["#66532e", "#a68a48"]], ["silver", ["#4f5354", "#888c8b"]],
  ["ivory", ["#71604b", "#b3a184"]],
];

async function generateIfSourceExists(options) {
  const sourcePath = join(root, options.source);
  const directoryFiles = await readdir(dirname(sourcePath)).catch(() => []);
  if (!directoryFiles.includes(options.source.split("/").at(-1))) {
    console.log(`${options.source}: source missing; existing variants left untouched.`);
    return;
  }
  await generateRamp(options);
}

async function generateRamp({ source, targetDirectory, outputPrefix, palettes, mappedRanks, transparentColors = [] }) {
  const sourcePath = join(root, source);
  const image = decodeRGBA(await readFile(sourcePath));
  const transparentSet = new Set(transparentColors.map((color) => color.replace(/^#/, "").toLowerCase()));
  const sourceColors = new Set();
  for (let index = 0; index < image.pixels.length; index += 4) {
    if (image.pixels[index + 3] === 0) continue;
    const color = image.pixels.subarray(index, index + 4).toString("hex");
    if (transparentSet.has(color.slice(0, 6))) {
      image.pixels[index + 3] = 0;
      continue;
    }
    sourceColors.add(color);
  }
  const orderedSource = [...sourceColors].sort((a, b) => luminance(a) - luminance(b));
  const ranks = mappedRanks ?? orderedSource.map((_, index) => index);
  if (new Set(orderedSource.map(luminance)).size !== orderedSource.length) {
    throw new Error(`${source} has source colors with indistinguishable brightness; define explicit rank order.`);
  }
  for (const [name, targetHexes] of palettes) {
    const targetRamp = resizePalette(targetHexes, ranks.length);
    if (!targetRamp) throw new Error(`${name} needs ${ranks.length} target shades for ${source}.`);
    const pixels = Buffer.from(image.pixels);
    for (let index = 0; index < pixels.length; index += 4) {
      if (pixels[index + 3] === 0) continue;
      const sourceHex = pixels.subarray(index, index + 4).toString("hex");
      const rank = orderedSource.indexOf(sourceHex);
      const targetHex = targetRamp[ranks.indexOf(rank)];
      if (targetHex) pixels.set(rgba(targetHex), index);
    }
    const file = `${outputPrefix ?? ""}${name}.png`;
    const output = join(root, targetDirectory, file);
    if (output === sourcePath) continue;
    await mkdir(dirname(output), { recursive: true });
    await writeFile(output, encodePNG(image.width, image.height, pixels));
  }
  console.log(`${source}: generated ${palettes.length} variants from ${orderedSource.length} source shades.`);
}

const hairSources = (await readdir(join(root, "hair")))
  .filter((file) => /^hair_.+_brown\.png$/.test(file))
  .sort();
if (!hairSources.length) throw new Error("No hair base sprites named hair_*_brown.png were found.");
for (const file of hairSources) {
  const outputPrefix = file.replace(/brown\.png$/, "");
  await generateRamp({ source: `hair/${file}`, targetDirectory: "hair", outputPrefix, palettes: hairPalettes });
}
const beardSources = (await readdir(join(root, "beard")))
  .filter((file) => /^beard_.+_brown\.png$/.test(file))
  .sort();
for (const file of beardSources) {
  const outputPrefix = file.replace(/brown\.png$/, "");
  await generateRamp({ source: `beard/${file}`, targetDirectory: "beard", outputPrefix, palettes: beardPalettes });
}
await generateIfSourceExists({ source: "skin/face_white.png", targetDirectory: "skin", palettes: skinPalettes });
await generateIfSourceExists({
  source: "eyes/eyes_brown.png", targetDirectory: "eyes", palettes: eyePalettes.map(([name, hex]) => [name, [hex]]), mappedRanks: [0],
});
const accessorySources = (await readdir(join(root, "accessories")))
  .filter((file) => /^accessories_.+_black\.png$/.test(file))
  .sort();
for (const file of accessorySources) {
  const outputPrefix = file.replace(/black\.png$/, "");
  await generateRamp({ source: `accessories/${file}`, targetDirectory: "accessories", outputPrefix, palettes: accessoryPalettes });
}
