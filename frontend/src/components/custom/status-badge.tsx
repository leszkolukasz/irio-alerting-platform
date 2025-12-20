import type { MonitoredService } from "@/lib/api/service";
import { Badge } from "../ui/badge";

const serviceStatusToColor = {
  UP: "#22c55e",
  DOWN: "#ef4444",
  UNKNOWN: "#888888",
} as const;

export const StatusBadge = ({
  status,
}: {
  status: MonitoredService["status"];
}) => {
  return (
    <Badge
      style={{
        backgroundColor: serviceStatusToColor[status],
      }}
      className="min-w-15"
    >
      {status}
    </Badge>
  );
};
