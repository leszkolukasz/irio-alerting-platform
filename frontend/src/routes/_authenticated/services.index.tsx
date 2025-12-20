import { ServiceList } from "@/components/custom/service-list";
import { Button } from "@/components/ui/button";
import {
  useGetMyServices,
  useCreateService,
  type MonitoredService,
} from "@/lib/api/service";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { PlusCircle } from "lucide-react";
import { ServiceForm } from "@/components/forms/service-form";
import { Spinner } from "@/components/ui/spinner";
import { useState } from "react";

export const Route = createFileRoute("/_authenticated/services/")({
  component: RouteComponent,
});

function RouteComponent() {
  const router = useRouter();
  const { mutateAsync: createService } = useCreateService();
  const { data: services = [], isLoading } = useGetMyServices();

  const handleOnSubmit = async (
    data: Omit<MonitoredService, "id" | "status">
  ) => {
    const result = await createService(data);
    router.navigate({
      to: "/services/$serviceID",
      params: { serviceID: result.serviceID.toString() },
    });
  };

  return (
    <section className="w-3/4">
      <div className="mt-20 flex flex-col gap-4">
        {isLoading ? (
          <Spinner className="mx-auto size-8" />
        ) : (
          <>
            <NewServiceDialog onSubmit={handleOnSubmit}>
              <Button className="hover:cursor-pointer w-35">
                <PlusCircle />
                Create new
              </Button>
            </NewServiceDialog>
            <ServiceList services={services} />
          </>
        )}
      </div>
    </section>
  );
}

const NewServiceDialog = ({
  children,
  onSubmit,
}: {
  children: React.ReactNode;
  onSubmit: (data: Omit<MonitoredService, "id" | "status">) => void;
}) => {
  const [open, setOpen] = useState(false);

  const handleSubmit = (data: Omit<MonitoredService, "id" | "status">) => {
    onSubmit(data);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New service</DialogTitle>
          <DialogDescription>
            Create a new service to monitor and receive alerts for.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-5">
          <ServiceForm onSubmit={handleSubmit} />
        </div>
      </DialogContent>
    </Dialog>
  );
};
