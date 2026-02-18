import type { SmsNotifier } from "./notifier";
import type { NotificationRequest } from "../types";

type TwilioConfig = {
  accountSid: string;
  authToken: string;
  messagingServiceSid: string;
};

export class TwilioSmsNotifier implements SmsNotifier {
  private cfg: TwilioConfig;

  constructor(cfg: TwilioConfig) {
    this.cfg = cfg;
  }

  static fromEnv(): TwilioSmsNotifier {
    const accountSid = process.env.TWILIO_ACCOUNT_SID ?? "";
    const authToken = process.env.TWILIO_AUTH_TOKEN ?? "";
    const messagingServiceSid = process.env.TWILIO_MESSAGING_SERVICE_SID ?? "";

    if (!accountSid || !authToken || !messagingServiceSid) {
      throw new Error("Missing TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, or TWILIO_MESSAGING_SERVICE_SID");
    }

    return new TwilioSmsNotifier({ accountSid, authToken, messagingServiceSid });
  }

  async send(req: NotificationRequest, signal?: AbortSignal): Promise<void> {
    if (req.channel !== "SMS") return;
    if (!req.to) throw new Error("missing to");
    if (!req.body) throw new Error("missing body");

    const url = `https://api.twilio.com/2010-04-01/Accounts/${this.cfg.accountSid}/Messages.json`;

    const form = new URLSearchParams();
    form.set("To", req.to);
    form.set("MessagingServiceSid", this.cfg.messagingServiceSid);
    form.set("Body", req.body);

    const basicAuth = Buffer.from(`${this.cfg.accountSid}:${this.cfg.authToken}`).toString("base64");

    const resp = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        Authorization: `Basic ${basicAuth}`,
      },
      body: form.toString(),
      signal,
    });

    const text = await resp.text();
    if (!resp.ok) {
      throw new Error(`twilio send failed status=${resp.status} body=${text}`);
    }
  }
}
