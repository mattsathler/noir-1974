// Generate the missing pixel-art face details and accessories aligned to skin_light.
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { deflateSync } from "node:zlib";
import { join } from "node:path";

const output = join("apps", "web", "public", "images", "characters");
const faceOutput = join(output, "face");
const accessoriesOutput = join(output, "accessories");
const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
const skinPath = join(output, "skin", "skin_light.png");

function dimensions(png) {
  if (!png.subarray(0, 8).equals(signature)) throw new Error("skin_light.png is not a PNG.");
  const width = png.readUInt32BE(16), height = png.readUInt32BE(20);
  if (width !== 25 || height !== 25) throw new Error(`Expected a 25x25 skin base, found ${width}x${height}.`);
  return { width, height };
}

const { width, height } = dimensions(await readFile(skinPath));
const rgba = (hex) => [1, 3, 5].map((i) => Number.parseInt(hex.slice(i, i + 2), 16)).concat(255);
const palette = {
  scar: rgba("#9a4935"),
  freckle: rgba("#94684f"),
  facialHair: rgba("#4b3429"),
  frame: rgba("#c9a744"),
  frameShade: rgba("#69552c"),
  hat: rgba("#302720"),
  hatShade: rgba("#171411"),
  hatBand: rgba("#c9a744"),
};

function canvas() { return Buffer.alloc(width * height * 4); }
function pixel(pixels, x, y, color) {
  if (x < 0 || x >= width || y < 0 || y >= height) return;
  pixels.set(color, (y * width + x) * 4);
}
function line(pixels, points, color) {
  for (const [x, y] of points) pixel(pixels, x, y, color);
}
function span(pixels, y, x1, x2, color) {
  for (let x = x1; x <= x2; x++) pixel(pixels, x, y, color);
}

function scarCheek() {
  const im = canvas();
  line(im, [[17, 14], [16, 15], [15, 16]], palette.scar);
  return im;
}
function scarBrow() {
  const im = canvas();
  line(im, [[10, 9], [11, 10], [12, 10]], palette.scar);
  return im;
}
function stubble() {
  const im = canvas();
  line(im, [[9, 16], [11, 17], [14, 16], [17, 17], [10, 18], [13, 19], [16, 18]], palette.facialHair);
  return im;
}
function freckles() {
  const im = canvas();
  // Cheek marks stay inside the face area of the skin sprite (x=10..18).
  line(im, [[10, 13], [12, 14], [11, 15], [16, 13], [18, 14], [17, 15]], palette.freckle);
  return im;
}

function roundedGlasses() {
  const im = canvas(), frame = palette.frame, shade = palette.frameShade;
  line(im, [[12, 9], [13, 9], [11, 10], [14, 10], [11, 11], [14, 11], [11, 12], [14, 12], [12, 13], [13, 13]], frame);
  line(im, [[17, 9], [18, 9], [16, 10], [19, 10], [16, 11], [19, 11], [16, 12], [19, 12], [17, 13], [18, 13]], frame);
  line(im, [[15, 10], [15, 11], [10, 10], [20, 10]], shade);
  return im;
}
function squareGlasses() {
  const im = canvas(), frame = palette.frame, shade = palette.frameShade;
  span(im, 9, 11, 14, frame); span(im, 13, 11, 14, frame);
  span(im, 9, 16, 19, frame); span(im, 13, 16, 19, frame);
  line(im, [[11, 10], [11, 12], [14, 10], [14, 12], [16, 10], [16, 12], [19, 10], [19, 12]], frame);
  line(im, [[15, 10], [15, 11], [10, 10], [20, 10]], shade);
  return im;
}
function clocheHat() {
  const im = canvas();
  span(im, 1, 10, 14, palette.hatShade);
  span(im, 2, 8, 16, palette.hat);
  span(im, 3, 7, 17, palette.hat);
  span(im, 3, 9, 15, palette.hatBand);
  span(im, 4, 6, 18, palette.hatShade);
  span(im, 5, 6, 18, palette.hat);
  return im;
}
function fedoraHat() {
  const im = canvas();
  span(im, 0, 10, 14, palette.hatShade);
  span(im, 1, 8, 16, palette.hat);
  span(im, 2, 8, 16, palette.hat);
  span(im, 2, 9, 15, palette.hatBand);
  span(im, 3, 7, 17, palette.hat);
  span(im, 4, 5, 19, palette.hatShade);
  return im;
}
function bowlerHat() {
  const im = canvas();
  span(im, 0, 10, 14, palette.hatShade);
  span(im, 1, 8, 16, palette.hat);
  span(im, 2, 7, 17, palette.hat);
  span(im, 3, 7, 17, palette.hat);
  span(im, 3, 9, 15, palette.hatBand);
  span(im, 4, 5, 19, palette.hatShade);
  return im;
}
function monocle() {
  const im = canvas(), frame = palette.frame, shade = palette.frameShade;
  line(im, [[17, 9], [18, 9], [16, 10], [19, 10], [16, 11], [19, 11], [16, 12], [19, 12], [17, 13], [18, 13]], frame);
  line(im, [[19, 14], [19, 15], [18, 16], [18, 17], [17, 18]], shade);
  return im;
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
function encodePNG(pixels) {
  const header = Buffer.alloc(13);
  header.writeUInt32BE(width, 0); header.writeUInt32BE(height, 4); header[8] = 8; header[9] = 6;
  const raw = Buffer.alloc(height * (width * 4 + 1));
  for (let y = 0; y < height; y++) pixels.copy(raw, y * (width * 4 + 1) + 1, y * width * 4, (y + 1) * width * 4);
  return Buffer.concat([signature, chunk("IHDR", header), chunk("IDAT", deflateSync(raw)), chunk("IEND", Buffer.alloc(0))]);
}
async function save(category, id, pixels) {
  const directory = category === "face" ? faceOutput : accessoriesOutput;
  await writeFile(join(directory, `${id}.png`), encodePNG(pixels));
}

const layers = [
  ["face", "face_scar_cheek", scarCheek], ["face", "face_scar_brow", scarBrow],
  ["face", "face_stubble", stubble], ["face", "face_freckles", freckles],
  ["accessories", "accessory_glasses_round", roundedGlasses], ["accessories", "accessory_hat_cloche", clocheHat],
  ["accessories", "accessory_hat_fedora", fedoraHat], ["accessories", "accessory_monocle", monocle],
  ["accessories", "accessory_glasses_square", squareGlasses], ["accessories", "accessory_hat_bowler", bowlerHat],
];

await Promise.all([mkdir(faceOutput, { recursive: true }), mkdir(accessoriesOutput, { recursive: true })]);
await Promise.all(layers.map(([category, id, draw]) => save(category, id, draw())));
console.log(`Generated ${layers.length} transparent ${width}x${height} layers aligned to skin_light.png.`);
