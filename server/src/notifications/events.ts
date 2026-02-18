import type { EventPayload } from "./types";

function requireField(payload: Record<string, string>, key: string, event: string): string {
  const v = payload?.[key];
  if (!v) throw new Error(`missing payload field '${key}' for event ${event}`);
  return v;
}

export function renderBody(ev: EventPayload): string {
  if (!ev.to) throw new Error("missing 'to'");

  switch (ev.event) {
    case "DOSE_DUE": {
      const patientName = requireField(ev.payload, "patientName", "DOSE_DUE");
      const meds = requireField(ev.payload, "meds", "DOSE_DUE");
      const time = requireField(ev.payload, "time", "DOSE_DUE");
      return `Hi ${patientName}, this is your DoseDock reminder to take your ${meds} at ${time}.`;
    }

    case "REFILL_REMINDER": {
      const patientName = requireField(ev.payload, "patientName", "REFILL_REMINDER");
      const meds = requireField(ev.payload, "meds", "REFILL_REMINDER");
      return `Hi ${patientName}, your DoseDock dispenser is running low on ${meds}. Please refill soon.`;
    }

    case "TEST_REMINDER":
      return "DoseDock test reminder, notifications are working.";

    default: {
      const _exhaustive: never = ev.event;
      return _exhaustive;
    }
  }
}
