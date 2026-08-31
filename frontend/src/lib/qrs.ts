import {
  appendFileHeaderMetaToBuffer,
  createGeneraterSVG,
} from "@qifi/generate";

export interface QRSFrameStream {
  nextFrame(): string;
  dispose(): void;
}

export function createQRSFrameStream(
  bpfBytes: Uint8Array,
  filename: string,
  sliceSize: number,
): QRSFrameStream {
  const merged = appendFileHeaderMetaToBuffer(bpfBytes, {
    filename,
    contentType: "application/octet-stream",
  });
  const generator = createGeneraterSVG(merged, { sliceSize });
  let fountain: Generator<string, never> | null = generator.fountain();

  return {
    nextFrame(): string {
      if (!fountain) {
        throw new Error("QRS frame stream has been disposed");
      }
      return fountain.next().value;
    },
    dispose(): void {
      if (!fountain) {
        return;
      }
      merged.fill(0);
      generator.encoder.data.fill(0);
      generator.encoder.compressed.fill(0);
      for (const block of generator.encoder.indices) {
        block.fill(0);
      }
      fountain = null;
    },
  };
}

export function decodeWailsBytes(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let index = 0; index < binary.length; index += 1) {
    bytes[index] = binary.charCodeAt(index);
  }
  return bytes;
}
