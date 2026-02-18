import type { RequestInit } from "node-fetch";

type SmsRequest = { to: string; body: string };

export class TwilioSmsNotifier {
  private constructor(
    private readonly accountSid: string,
    private readonly authToken: string,
    private readonly messagingServiceSid: string
  ) {}

  static fromEnv(): TwilioSmsNotifier {
    const accountSid = process.env.TWILIO_ACCOUNT_SID;
    const authToken = process.env.TWILIO_AUTH_TOKEN;
    const messagingServiceSid = process.env.TWILIO_MESSAGING_SERVICE_SID;

    if (!accountSid || !authToken || !messagingServiceSid) {
      throw new Error("Missing TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, or TWILIO_MESSAGING_SERVICE_SID");
    }
    return new TwilioSmsNotifier(accountSid, authToken, messagingServiceSid);
  }

  async sendSms(req: SmsRequest): Promise<void> {
    if (!req.to) throw new Error("Missing to");
    if (!req.body) throw new Error("Missing body");

    const url = `https://api.twilio.com/2010-04-01/Accounts/${this.accountSid}/Messages.json`;

    const form = new URLSearchParams();
    form.set("To", req.to);
    form.set("MessagingServiceSid", this.messagingServiceSid);
    form.set("Body", req.body);

    const basicAuth = Buffer.from(`${this.accountSid}:${this.authToken}`).toString("base64");

    const init: RequestInit = {
      method: "POST",
      headers: {
        Authorization: `Basic ${basicAuth}`,
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: form.toString(),
    };

    const resp = await fetch(url, init as any);
    const text = await resp.text();

    if (!resp.ok) {
      throw new Error(`Twilio send failed: status=${resp.status} body=${text}`);
    }
  }
}
