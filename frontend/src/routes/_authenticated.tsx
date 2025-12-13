import { useGlobalContext } from "@/lib/context";
import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/custom/app-sidebar";

export const Route = createFileRoute("/_authenticated")({
  component: RouteComponent,
  beforeLoad: async () => {
    if (!useGlobalContext.getState().isLoggedIn) {
      throw redirect({ to: "/sign-in" });
    }
  },
});

function RouteComponent() {
  return (
    <SidebarProvider>
      <AppSidebar />
      <main className="w-full">
        <SidebarTrigger />
        <div className="flex justify-center p-4">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  );
}
