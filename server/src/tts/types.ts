export type SynthesizeRequest = {
  text: string;
  voice?: string;        // "Charon"
  emotion?: string; //"calm"
  speakingRate?: number; // 1.0
  prompt?: string; 
};

export type SynthesizeResponse = {
  audioBase64: string;
};
