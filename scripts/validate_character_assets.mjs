import { inflateSync } from "node:zlib";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = process.cwd();
const catalogPath = resolve(root, "apps/api/internal/world/visual_identity.go");
const assetRoot = resolve(root, "apps/web/public/images/characters");
const catalog = await readFile(catalogPath, "utf8");
const layers = ["accessories", "beard", "eyes", "face", "hair", "skin"];
const catalogNames = { accessories: "accessory", eyes: "eye" };
const vectorLayers = new Set([
  "face_freckles_light", "face_freckles_dense", "face_mole", "face_smile_lines",
  "face_wrinkles_fine", "face_wrinkles_deep",
]);
const problems = [];

function decodePNG(buffer, name) {
  if (buffer.toString("hex", 0, 8) !== "89504e470d0a1a0a") throw new Error(`${name}: invalid PNG signature`);
  let offset = 8;
  const compressed = [];
  let width = 0;
  let height = 0;
  let bitDepth = 0;
  let colorType = 0;
  while (offset < buffer.length) {
    const size = buffer.readUInt32BE(offset);
    const type = buffer.toString("ascii", offset + 4, offset + 8);
    const data = buffer.subarray(offset + 8, offset + 8 + size);
    if (type === "IHDR") {
      width = data.readUInt32BE(0);
      height = data.readUInt32BE(4);
      bitDepth = data[8];
      colorType = data[9];
    } else if (type === "IDAT") compressed.push(data);
    offset += 12 + size;
    if (type === "IEND") break;
  }
  if (width !== 40 || height !== 40 || bitDepth !== 8 || colorType !== 6) {
    throw new Error(`${name}: expected 40x40 RGBA PNG, got ${width}x${height}, depth=${bitDepth}, type=${colorType}`);
  }

  const raw = inflateSync(Buffer.concat(compressed));
  const stride = width * 4;
  const previous = Buffer.alloc(stride);
  let hasTransparentPixel = false;
  let hasOpaquePixel = false;
  let cursor = 0;
  const paeth = (a, b, c) => {
    const p = a + b - c;
    const pa = Math.abs(p - a);
    const pb = Math.abs(p - b);
    const pc = Math.abs(p - c);
    return pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
  };
  for (let y = 0; y < height; y++) {
    const filter = raw[cursor++];
    const row = Buffer.from(raw.subarray(cursor, cursor + stride));
    cursor += stride;
    for (let x = 0; x < stride; x++) {
      const left = x >= 4 ? row[x - 4] : 0;
      const above = previous[x];
      const upperLeft = x >= 4 ? previous[x - 4] : 0;
      if (filter === 1) row[x] = (row[x] + left) & 255;
      else if (filter === 2) row[x] = (row[x] + above) & 255;
      else if (filter === 3) row[x] = (row[x] + Math.floor((left + above) / 2)) & 255;
      else if (filter === 4) row[x] = (row[x] + paeth(left, above, upperLeft)) & 255;
      else if (filter !== 0) throw new Error(`${name}: unsupported PNG filter ${filter}`);
      if (x % 4 === 3) {
        if (row[x] === 0) hasTransparentPixel = true;
        if (row[x] > 0) hasOpaquePixel = true;
      }
    }
    row.copy(previous);
  }
  if (!hasTransparentPixel || !hasOpaquePixel) throw new Error(`${name}: expected both transparent and visible pixels`);
}

for (const layer of layers) {
  const match = catalog.match(new RegExp(`var ${catalogNames[layer] ?? layer}LayerIDs = \\[\\]string\\{([\\s\\S]*?)\\}`, "m"));
  if (!match) {
    problems.push(`catalog has no ${layer} layer list`);
    continue;
  }
  const ids = [...match[1].matchAll(/"([a-z0-9_]+)"/g)].map((item) => item[1]);
  if (new Set(ids).size !== ids.length) problems.push(`${layer}: duplicate IDs in catalog`);
  for (const id of ids) {
    const extension = vectorLayers.has(id) ? "svg" : "png";
    const path = resolve(assetRoot, layer, `${id}.${extension}`);
    try {
      const data = await readFile(path);
      if (extension === "png") decodePNG(data, `${layer}/${id}.png`);
      else {
        const svg = data.toString("utf8");
        if (!/<svg\b[^>]*width="40"[^>]*height="40"[^>]*viewBox="0 0 40 40"[^>]*shape-rendering="crispEdges"/.test(svg)) {
          throw new Error(`${layer}/${id}.svg: expected a transparent 40x40 crisp-edge canvas`);
        }
        if (/<(?:rect|circle|path)\b[^>]*\b(?:x|y|width|height)="-/.test(svg)) {
          throw new Error(`${layer}/${id}.svg: pixel geometry must remain within the sprite canvas`);
        }
      }
    } catch (error) {
      problems.push(error.message);
    }
  }
}

if (problems.length) {
  console.error(problems.map((problem) => `✗ ${problem}`).join("\n"));
  process.exitCode = 1;
} else {
  console.log(`Validated ${layers.length} character layers and every catalog sprite (40x40, transparent, aligned).`);
}
