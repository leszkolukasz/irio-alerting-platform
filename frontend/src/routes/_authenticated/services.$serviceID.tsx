import { createFileRoute } from "@tanstack/react-router";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useGetIncidents,
  useGetService,
  useGetStatusMetrics,
  type Granularity,
  type MonitoredService,
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
import { useMemo, useState } from "react";

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
  const serviceID = Route.useParams().serviceID;

  const [granularity, setGranularity] = useState<Granularity>("hour");
  const { data: metrics, isLoading } = useGetStatusMetrics(
    Number(serviceID),
    granularity
  );

  if (isLoading) {
    return <Spinner className="mx-auto size-8" />;
  }

  return (
    <div>
      <MetricsGraph
        metrics={requireNotNullish(metrics)}
        granularity={granularity}
        onGranularityChange={setGranularity}
      />
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
