import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { CacheKeys } from "./cache";
import { toast } from "sonner";
import { extractErrorMessage, requireNotNullish } from "../utils";

export type ServiceStatus = "UP" | "DOWN" | "UNKNOWN";

export type Seconds = number & { __secondsBrand: never };
export type Minutes = number & { __minutesBrand: never };

export type MonitoredService = {
  id: number;
  name: string;
  url: string;
  port: number;
  firstOncallerEmail: string;
  secondOncallerEmail?: string;
  healthCheckInterval: Seconds;
  alertWindow: Seconds;
  allowedResponseTime: Minutes;
  status: ServiceStatus;
};

export type Granularity = "month" | "week" | "day" | "hour";

export type StatusMetricsEntry = {
  timestamp: string;
  success: number;
  total: number;
};

export type ExtendedStatusMetricsEntry = StatusMetricsEntry & {
  uptime: number;
  error: number;
};

export type StatusMetrics = {
  granularity: Granularity;
  data: StatusMetricsEntry[];
};

export type Incident = {
  id: string;
  startTime: Date;
  serviceID: number;
  events: IncidentEvent[];
};

export type IncidentEvent =
  | {
      timestamp: string;
      type: "START";
    }
  | {
      timestamp: string;
      type: "RESOLVED";
      oncaller: string;
    }
  | {
      timestamp: string;
      type: "UNRESOLVED";
    }
  | {
      timestamp: string;
      type: "NOTIFIED";
      oncaller: string;
    }
  | {
      timestamp: string;
      type: "TIMEOUT";
      oncaller: string;
    };

const getIncidentStartTime = (incident: Incident) => {
  const startEvent = incident.events.find((event) => event.type === "START");
  return startEvent ? new Date(startEvent.timestamp) : null;
};

export const useGetMyServices = () =>
  useQuery({
    queryKey: [CacheKeys.MyServices],
    queryFn: async () => {
      return axios
        .get<MonitoredService[]>("/services/me")
        .then((res) => res.data);
    },
    staleTime: 1000 * 60 * 5,
  });

export const useGetService = (id: MonitoredService["id"]) => {
  const queryClient = useQueryClient();

  return useQuery({
    queryKey: [CacheKeys.MyServices, id],
    queryFn: async () => {
      return axios
        .get<MonitoredService>(`/services/${id}`)
        .then((res) => res.data);
    },

    initialData: () => {
      const allServices = queryClient.getQueryData<MonitoredService[]>([
        CacheKeys.MyServices,
      ]);

      if (!allServices) return undefined;

      return allServices.find((service) => service.id === id);
    },

    initialDataUpdatedAt: () => {
      return queryClient.getQueryState([CacheKeys.MyServices])?.dataUpdatedAt;
    },
  });
};

export const useCreateService = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (service: Omit<MonitoredService, "id" | "status">) => {
      return axios
        .post<{ serviceID: number }>("/services/", service)
        .then((res) => res.data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CacheKeys.MyServices] });
      toast.success("Service created successfully");
    },
    onError: (err) => {
      toast.error(extractErrorMessage(err));
      console.error("Error creating service:", err);
    },
  });
};

export const useUpdateService = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<MonitoredService, "status">) => {
      return axios.put(`/services/${data.id}`, data);
    },
    onSuccess: (_, data) => {
      queryClient.invalidateQueries({ queryKey: [CacheKeys.MyServices] });
      queryClient.invalidateQueries({
        queryKey: [CacheKeys.MyServices, data.id],
      });
      toast.success("Service updated successfully");
    },
    onError: (err) => {
      toast.error(extractErrorMessage(err));
      console.error("Error updating service:", err);
    },
  });
};

export const useDeleteService = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (serviceID: MonitoredService["id"]) => {
      return axios.delete(`/services/${serviceID}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CacheKeys.MyServices] });
      toast.success("Service deleted successfully");
    },
    onError: (err) => {
      toast.error(extractErrorMessage(err));
      console.error("Error deleting service:", err);
    },
  });
};

export const useGetIncidents = (serviceID: MonitoredService["id"]) =>
  useQuery({
    queryKey: [CacheKeys.ServiceIncidents, serviceID],
    queryFn: async () => {
      return axios
        .get<Incident[]>(`/services/${serviceID}/incidents`)
        .then((res) =>
          res.data.map((incident) => ({
            ...incident,
            startTime: requireNotNullish(getIncidentStartTime(incident)),
            serviceID,
          }))
        );
    },
    staleTime: 1000 * 60 * 5,
  });

export const useGetStatusMetrics = (
  serviceID: MonitoredService["id"],
  granularity: Granularity
) =>
  useQuery({
    queryKey: [CacheKeys.ServiceStatusMetrics, serviceID, granularity],
    queryFn: async () => {
      return axios
        .get<StatusMetrics>(
          `/services/${serviceID}/metrics?granularity=${granularity}`
        )
        .then((res) => res.data);
    },
    staleTime: 1000 * 60 * 2,
  });
