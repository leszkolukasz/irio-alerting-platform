import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "./Theme";
import { Toaster } from "@/components/ui/sonner";
import { useMemo } from "react";

export const Providers = ({ children }: { children: React.ReactNode }) => {
  const queryClient = useMemo(() => new QueryClient(), []);

  return (
    <ThemeProvider defaultTheme="dark">
      <QueryClientProvider client={queryClient}>
        <Toaster />
        {children}
      </QueryClientProvider>
    </ThemeProvider>
  );
};
