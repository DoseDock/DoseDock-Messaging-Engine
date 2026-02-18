import { GoogleAuth } from "google-auth-library";
import type { SynthesizeRequest, SynthesizeResponse } from "./types";

export class GoogleTtsClient {
  private auth: GoogleAuth;
  private projectId: string;

  constructor(projectId: string) {
    if (!projectId) throw new Error("GOOGLE_CLOUD_PROJECT not set");
    this.projectId = projectId;

    this.auth = new GoogleAuth({
      scopes: ["https://www.googleapis.com/auth/cloud-platform"],
    });
  }

  async synthesize(req: SynthesizeRequest, signal?: AbortSignal): Promise<SynthesizeResponse> {
    if (!req.text) throw new Error("empty text");

    const speakingRate = req.speakingRate && req.speakingRate > 0 ? req.speakingRate : 1.0;

    const short = (req.voice ?? "Charon").trim();
    const voiceName = short.startsWith("en-US-") ? short : `en-US-Chirp3-HD-${short}`;

    const client = await this.auth.getClient();
    const accessToken = await client.getAccessToken();
    const token = typeof accessToken === "string" ? accessToken : accessToken?.token;

    if (!token) throw new Error("failed to get access token from ADC");

    const body = {
      input: { text: req.text },
      voice: { languageCode: "en-US", name: voiceName },
      audioConfig: { audioEncoding: "MP3", speakingRate },
    };

    const resp = await fetch("https://texttospeech.googleapis.com/v1/text:synthesize", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
        "x-goog-user-project": this.projectId,
      },
      body: JSON.stringify(body),
      signal,
    });

    const json = (await resp.json()) as { audioContent?: string; error?: unknown };

    if (!resp.ok) {
      throw new Error(`tts failed status=${resp.status} body=${JSON.stringify(json)}`);
    }

    if (!json.audioContent) throw new Error("missing audioContent in tts response");

    return { audioBase64: json.audioContent };
  }
}
