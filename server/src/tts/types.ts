export type SynthesizeRequest = {
  text: string;
  voice?: string;        // "Charon"
  emotion?: string; //"Calm"
  speakingRate?: number; // 1.0
};

export type SynthesizeResponse = {
  audioBase64: string;
};
