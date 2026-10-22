// Derive the remaining skin tones from the artist-authored base sprite.
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { deflateSync, inflateSync } from "node:zlib";
import { join } from "node:path";

const directory = join("apps", "web", "public", "images", "characters", "skin");
const sourcePath = join(directory, "skin_light.png");
const color = (hex) => [1, 3, 5].map((i) => Number.parseInt(hex.slice(i, i + 2), 16)).concat(255);
const palettes = [
  ["skin_pale", ["#f1c7a6", "#c99072"]],
  ["skin_medium", ["#c9855c", "#92553d"]],
  ["skin_olive", ["#c49a69", "#896644"]],
  ["skin_brown", ["#a96647", "#75422f"]],
  ["skin_dark", ["#744634", "#4e3029"]],
];
const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);

function decodeRGBA(png) {
  if (!png.subarray(0, 8).equals(signature)) throw new Error("Base skin is not a PNG.");
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
  if (bitDepth !== 8 || colorType !== 6 || interlace !== 0) throw new Error("Base skin must be a non-interlaced 8-bit RGBA PNG.");
  const bytesPerPixel = 4, stride = width * bytesPerPixel;
  const filtered = inflateSync(Buffer.concat(compressed));
  const pixels = Buffer.alloc(width * height * bytesPerPixel);
  const paeth = (a, b, c) => {
    const p = a + b - c, pa = Math.abs(p - a), pb = Math.abs(p - b), pc = Math.abs(p - c);
    return pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
  };
  for (let y = 0; y < height; y++) {
    const row = y * (stride + 1), filter = filtered[row];
    for (let x = 0; x < stride; x++) {
      const raw = filtered[row + 1 + x], left = x >= bytesPerPixel ? pixels[y * stride + x - bytesPerPixel] : 0;
      const up = y ? pixels[(y - 1) * stride + x] : 0;
      const upperLeft = y && x >= bytesPerPixel ? pixels[(y - 1) * stride + x - bytesPerPixel] : 0;
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

const base = decodeRGBA(await readFile(sourcePath));
await mkdir(directory, { recursive: true });
const counts = new Map();
for (let i = 0; i < base.pixels.length; i += 4) {
  if (!base.pixels[i + 3]) continue;
  const key = base.pixels.subarray(i, i + 4).toString("hex");
  counts.set(key, (counts.get(key) ?? 0) + 1);
}
console.log(`${base.width}x${base.height} base; palette:`, [...counts.entries()].map(([hex, count]) => `#${hex.slice(0, 6)} (${count})`).join(", "));
if (counts.size !== 2) throw new Error(`Expected exactly two opaque skin colors in the base, found ${counts.size}.`);
const luminance = (hex) => {
  const [r, g, b] = [0, 2, 4].map((i) => Number.parseInt(hex.slice(i, i + 2), 16));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};
const [shadeHex, mainHex] = [...counts.keys()].sort((a, b) => luminance(a) - luminance(b));
const sourceColors = new Map([[mainHex, 0], [shadeHex, 1]]);
for (const [id, targetHexes] of palettes) {
  const targetColors = targetHexes.map(color);
  const pixels = Buffer.from(base.pixels);
  for (let i = 0; i < pixels.length; i += 4) {
    if (pixels[i + 3] === 0) continue;
    const key = pixels.subarray(i, i + 4).toString("hex");
    const target = targetColors[sourceColors.get(key)];
    if (!target) throw new Error(`Unexpected base pixel color #${key.slice(0, 6)}.`);
    pixels.set(target, i);
  }
  await writeFile(join(directory, `${id}.png`), encodePNG(base.width, base.height, pixels));
}
console.log(`Generated ${palettes.length} matching skin tone layers from skin_light.png.`);
