import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { CacheKeys } from "./cache";
import { toast } from "sonner";
import { extractErrorMessage } from "../utils";

export type ServiceStatus = "UP" | "DOWN" | "UNKNOWN";

export type Seconds = number & { __secondsBrand: never };
export type Minutes = number & { __minutesBrand: never };

export type MonitoredService = {
  id: string;
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
