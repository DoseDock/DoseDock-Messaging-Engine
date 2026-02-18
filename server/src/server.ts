import express from "express";
import "dotenv/config"
import path from "path";

import { renderBody } from "./notifications/events";
import type { EventPayload, NotificationRequest } from "./notifications/types";
import { TwilioSmsNotifier } from "./notifications/sms/twilio";

import { GoogleTtsClient } from "./tts/google_tts";
import type { SynthesizeRequest } from "./tts/types";

const app = express();
app.use(express.json());

const notifier = TwilioSmsNotifier.fromEnv();
const ttsClient = new GoogleTtsClient(process.env.GOOGLE_CLOUD_PROJECT ?? "");

app.get("/healthz", (_req, res) => res.status(200).send("ok"));

app.post("/send-sms", async (req, res) => {
  try {
    const body = req.body as { to?: string; body?: string };
    const to = body.to ?? "";
    const msg = body.body ?? "";

    const nreq: NotificationRequest = { to, body: msg, channel: "SMS" };
    await notifier.sendSms(nreq);

    res.status(200).json({ ok: true });
  } catch (e) {
    res.status(400).json({ ok: false, error: String(e) });
  }
});

app.post("/send-event", async (req, res) => {
  try {
    const ev = req.body as EventPayload;

    const text = renderBody(ev);

    // send SMS
    await notifier.sendSms({ to: ev.to, body: text });

    // synthesize for immediate playback
    const ttsResp = await ttsClient.synthesize({
      text,
      voice: ev.voice ?? "Charon",
      emotion: ev.emotion ?? "Calm",
      speakingRate: ev.speakingRate ?? 1.0,
    });

    res.status(200).json({ ok: true, text, audioBase64: ttsResp.audioBase64 });
  } catch (e) {
    res.status(400).json({ ok: false, error: String(e) });
  }
});

app.post("/tts/speak", async (req, res) => {
  try {
    const body = req.body as SynthesizeRequest;
    const result = await ttsClient.synthesize(body);
    res.status(200).json(result);
  } catch (e) {
    res.status(400).json({ ok: false, error: String(e) });
  }
});

// Serve old HTML UI
app.use("/ui", express.static(path.join(process.cwd(), "web")));

const port = Number(process.env.PORT ?? 8090);
app.listen(port, () => {
  console.log(`server listening on http://localhost:${port}`);
  console.log(`ui available at http://localhost:${port}/ui/`);
});
