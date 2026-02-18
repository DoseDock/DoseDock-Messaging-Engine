import type { NotificationRequest } from "../types";

export interface SmsNotifier {
  send(req: NotificationRequest, signal?: AbortSignal): Promise<void>;
}
