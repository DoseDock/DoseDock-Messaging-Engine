export type Channel = "SMS";

export type NotificationRequest = {
  to: string;
  body: string;
  channel: Channel;
};

export type EventType = "DOSE_DUE" | "REFILL_REMINDER" | "TEST_REMINDER";

export type EventPayload = {
  event: EventType;
  to: string;
  payload: Record<string, string>;

  // caregiver prefs (kept for later)
  voice?: string;
  emotion?: string;
  speakingRate?: number;
  prompt?: string;
};