import { createFileRoute, useRouter } from "@tanstack/react-router";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useDeleteService,
  useGetService,
  useUpdateService,
  type MonitoredService,
} from "@/lib/api/service";
import { Spinner } from "@/components/ui/spinner";
import { StatusBadge } from "@/components/custom/status-badge";
import { requireNotNullish } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { DialogClose } from "@radix-ui/react-dialog";
import { useState } from "react";
import { ServiceForm } from "@/components/forms/service-form";

export const Route = createFileRoute("/_authenticated/services/$serviceID")({
  component: RouteComponent,
});

function RouteComponent() {
  const serviceID = Route.useParams().serviceID;

  const { data: service, isLoading } = useGetService(serviceID);

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
              <TabsTrigger value="logs">Logs</TabsTrigger>
              <TabsTrigger value="settings">Settings</TabsTrigger>
            </TabsList>
            <div className="mt-10">
              <TabsContent value="graph"></TabsContent>
              <TabsContent value="logs"></TabsContent>
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
    <div className="flex flex-col gal-5 max-w-1/3">
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

      <div className="mt-10 flex justify-between">
        <UpdateServiceButton service={service} />
        <DeleteServiceButton serviceID={service.id} />
      </div>
    </div>
  );
};

const UpdateServiceButton = ({ service }: { service: MonitoredService }) => {
  const { mutateAsync: updateService } = useUpdateService();

  const [open, setOpen] = useState(false);

  const handleSubmit = (data: Omit<MonitoredService, "id" | "status">) => {
    updateService({ ...data, id: service.id });
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>Edit</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit service</DialogTitle>
          <DialogDescription>
            Update the details of the service you want to monitor and receive
            alerts for.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-5">
          <ServiceForm initialValues={service} onSubmit={handleSubmit} />
        </div>
      </DialogContent>
    </Dialog>
  );
};

const DeleteServiceButton = ({
  serviceID,
}: {
  serviceID: MonitoredService["id"];
}) => {
  const router = useRouter();
  const { mutateAsync: deleteService } = useDeleteService();

  const handleDelete = async () => {
    await deleteService(serviceID);
    router.navigate({ to: "/services" });
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="destructive">Delete</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete service</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this service? This action cannot be
            undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <DialogClose asChild>
            <Button variant="destructive" onClick={handleDelete}>
              Yes, delete
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
