import {
  appendFileHeaderMetaToBuffer,
  createGeneraterSVG,
  readFileHeaderMetaFromBuffer,
} from "@qifi/generate";

// Header plus a valid empty gzip stream: sufficient to represent a BPF payload
// at the QRS layer, which intentionally treats file bytes as opaque.
const bpf = Uint8Array.from([
  0x03, 0x01, 0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
  0x00, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00,
  0x00, 0x00, 0x00, 0x00, 0x00,
]);
const filename = "smoke.bpf";
const merged = appendFileHeaderMetaToBuffer(bpf, {
  filename,
  contentType: "application/octet-stream",
});
const [decoded, metadata] = readFileHeaderMetaFromBuffer(merged);
if (metadata.filename !== filename || decoded.length !== bpf.length) {
  throw new Error("QRS file metadata round-trip failed");
}

const generator = createGeneraterSVG(merged, { sliceSize: 80 });
const fountain = generator.fountain();
for (let index = 0; index < 3; index += 1) {
  const frame = fountain.next().value;
  if (!frame.startsWith("<svg") || !frame.includes("</svg>")) {
    throw new Error(`QRS frame ${index + 1} is not SVG`);
  }
}

console.log(`QRS smoke passed: ${generator.encoder.k} fountain slices, 3 SVG frames`);
