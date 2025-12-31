import { createFileRoute } from "@tanstack/react-router";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useGetIncidents,
  useGetService,
  type MonitoredService,
  type StatusMetrics,
} from "@/lib/api/service";
import { Spinner } from "@/components/ui/spinner";
import { StatusBadge } from "@/components/custom/status-badge";
import { requireNotNullish } from "@/lib/utils";

import { Card } from "@/components/ui/card";
import { MetricsGraph } from "@/components/custom/metrics-graph";
import {
  DeleteServiceButton,
  UpdateServiceButton,
} from "@/components/custom/service-button";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { IncidentTimeline } from "@/components/custom/incident-timeline";
import { useMemo } from "react";

export const Route = createFileRoute("/_authenticated/services/$serviceID")({
  component: RouteComponent,
});

function RouteComponent() {
  const serviceID = Route.useParams().serviceID;

  const { data: service, isLoading } = useGetService(Number(serviceID));

  return (
    <div className="flex w-3/4 flex-col gap-6 mt-10">
      {isLoading ? (
        <Spinner className="mx-auto size-8" />
      ) : (
        <>
          <div className="flex gap-5">
            <div className="font-bold text-2xl">{service?.name}</div>
            <StatusBadge status={requireNotNullish(service?.status)} />
          </div>
          <Tabs defaultValue="graph">
            <TabsList>
              <TabsTrigger value="graph">Graph</TabsTrigger>
              <TabsTrigger value="logs">Incident Logs</TabsTrigger>
              <TabsTrigger value="settings">Settings</TabsTrigger>
            </TabsList>
            <div className="mt-10">
              <TabsContent value="graph">
                <GraphTab />
              </TabsContent>
              <TabsContent value="logs">
                <LogsTab serviceID={Number(serviceID)} />
              </TabsContent>
              <TabsContent value="settings">
                <SettingsTab service={requireNotNullish(service)} />
              </TabsContent>
            </div>
          </Tabs>
        </>
      )}
    </div>
  );
}

const SettingsTab = ({ service }: { service: MonitoredService }) => {
  return (
    <Card className=" max-w-1/2">
      <div className="flex flex-col gap-1 px-8">
        <div className="flex justify-between">
          <p className="font-bold mb-2">Service ID:</p>
          <div>{service.id}</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Service Name:</p>
          <div>{service.name}</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">URL:</p>
          <a className="underline" href={service.url}>
            {service.url}
          </a>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Port:</p>
          <div>{service.port}</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">First Oncaller Email:</p>
          <div>{service.firstOncallerEmail || "N/A"}</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Second Oncaller Email:</p>
          <div>{service.secondOncallerEmail || "N/A"}</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Alert Window:</p>
          <div>{service.alertWindow}s</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Allowed Response Time:</p>
          <div>{service.allowedResponseTime}m</div>
        </div>
        <div className="flex justify-between">
          <p className="font-bold mb-2">Health Check Interval:</p>
          <div>{service.healthCheckInterval}s</div>
        </div>

        <div className="mt-10 flex justify-between">
          <UpdateServiceButton service={service} />
          <DeleteServiceButton serviceID={service.id} />
        </div>
      </div>
    </Card>
  );
};

const GraphTab = () => {
  const dummyData: StatusMetrics = {
    granularity: "hour",
    data: [
      { timestamp: "2024-10-01T00:00:00Z", success: 95, total: 105 },
      { timestamp: "2024-10-01T01:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-01T02:00:00Z", success: 97, total: 120 },
      { timestamp: "2024-10-01T03:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-01T04:00:00Z", success: 96, total: 100 },
      { timestamp: "2024-10-01T05:00:00Z", success: 100, total: 100 },
      { timestamp: "2024-10-01T06:00:00Z", success: 94, total: 110 },
      { timestamp: "2024-10-01T07:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-01T08:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-01T09:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-01T10:00:00Z", success: 95, total: 150 },
      { timestamp: "2024-10-01T11:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-01T12:00:00Z", success: 96, total: 100 },
      { timestamp: "2024-10-01T13:00:00Z", success: 100, total: 100 },
      { timestamp: "2024-10-01T14:00:00Z", success: 98, total: 120 },
      { timestamp: "2024-10-01T15:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-01T16:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-01T17:00:00Z", success: 95, total: 100 },
      { timestamp: "2024-10-01T18:00:00Z", success: 96, total: 100 },
      { timestamp: "2024-10-01T19:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-01T20:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-01T21:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-01T22:00:00Z", success: 94, total: 100 },
      { timestamp: "2024-10-01T23:00:00Z", success: 100, total: 100 },
      { timestamp: "2024-10-02T00:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-02T01:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-02T02:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-02T03:00:00Z", success: 95, total: 100 },
      { timestamp: "2024-10-02T04:00:00Z", success: 96, total: 100 },
      { timestamp: "2024-10-02T05:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-02T06:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-02T07:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-02T08:00:00Z", success: 94, total: 100 },
      { timestamp: "2024-10-02T09:00:00Z", success: 100, total: 100 },
      { timestamp: "2024-10-02T10:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-02T11:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-02T12:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-02T13:00:00Z", success: 95, total: 100 },
      { timestamp: "2024-10-02T14:00:00Z", success: 96, total: 100 },
      { timestamp: "2024-10-02T15:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-02T16:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-02T17:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-02T18:00:00Z", success: 94, total: 100 },
      { timestamp: "2024-10-02T19:00:00Z", success: 100, total: 100 },
      { timestamp: "2024-10-02T20:00:00Z", success: 98, total: 100 },
      { timestamp: "2024-10-02T21:00:00Z", success: 97, total: 100 },
      { timestamp: "2024-10-02T22:00:00Z", success: 99, total: 100 },
      { timestamp: "2024-10-02T23:00:00Z", success: 95, total: 100 },
    ],
  };

  return (
    <div>
      <MetricsGraph metrics={dummyData} />
    </div>
  );
};

const LogsTab = ({ serviceID }: { serviceID: number }) => {
  const { data: incidents = [], isLoading } = useGetIncidents(serviceID);

  const sortedIncidents = useMemo(
    () =>
      [...incidents].sort((a, b) => {
        return b.startTime.getTime() - a.startTime.getTime();
      }),
    [incidents]
  );

  if (isLoading) {
    return <Spinner className="mx-auto size-8" />;
  } else if (incidents.length === 0) {
    return (
      <div className="text-center">No incidents found for this service.</div>
    );
  }

  return (
    <Accordion type="single" collapsible className="w-full">
      {sortedIncidents.map((incident) => (
        <AccordionItem key={incident.id} value={incident.id}>
          <AccordionTrigger>
            Incident on{" "}
            {incident.startTime.toLocaleString("pl-PL") || "Unknown"} | Events:{" "}
            {incident.events.length}
          </AccordionTrigger>
          <AccordionContent>
            <IncidentTimeline incident={incident} />
          </AccordionContent>
        </AccordionItem>
      ))}
    </Accordion>
  );
};
