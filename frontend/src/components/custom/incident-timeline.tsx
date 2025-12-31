import type { Incident, IncidentEvent } from "@/lib/api/service";
import type { TimelineElement } from "../ui/timeline/types";
import { format } from "date-fns";
import { unreachable } from "@/lib/utils";
import { useMemo } from "react";
import { TimelineLayout } from "../ui/timeline/timeline-layout";

const mapIncidentEventToTimelineElement = (
  event: IncidentEvent,
  idx: number
): TimelineElement => {
  const commonData = {
    id: idx,
    date: format(new Date(event.timestamp), "dd MMM yyyy, HH:mm:ss"),
  };

  switch (event.type) {
    case "START":
      return {
        ...commonData,
        title: "Incident Started",
        description: "",
        color: "bg-yellow-500",
      };
    case "RESOLVED":
      return {
        ...commonData,
        title: "Incident Resolved",
        description: `Resolved by ${event.oncaller}`,
        color: "bg-green-500",
      };
    case "UNRESOLVED":
      return {
        ...commonData,
        title: "Incident Unresolved",
        description: "No oncaller has responded in time.",
        color: "destructive",
      };
    case "NOTIFIED":
      return {
        ...commonData,
        title: "Oncaller Notified",
        description: `Notified ${event.oncaller}`,
      };
    case "TIMEOUT":
      return {
        ...commonData,
        title: "Notification Timeout",
        description: `Oncaller ${event.oncaller} did not respond in time.`,
      };
    default:
      unreachable(event);
  }

  return {} as TimelineElement;
};

const incidentEventCompare = (a: IncidentEvent, b: IncidentEvent): number => {
  const diff =
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime();
  if (diff !== 0) {
    return diff;
  }

  const typeOrder: Record<IncidentEvent["type"], number> = {
    START: 0,
    NOTIFIED: 1,
    TIMEOUT: 2,
    UNRESOLVED: 3,
    RESOLVED: 3,
  };

  return typeOrder[a.type] - typeOrder[b.type];
};

export const IncidentTimeline = ({ incident }: { incident: Incident }) => {
  const timelineItems: TimelineElement[] = useMemo(() => {
    return incident.events
      .sort(incidentEventCompare)
      .map((event, idx) => mapIncidentEventToTimelineElement(event, idx));
  }, [incident.events]);

  return <TimelineLayout items={timelineItems} size="sm" />;
};
