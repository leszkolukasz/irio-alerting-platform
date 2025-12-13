import { useGlobalContext } from "@/lib/context";
import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_unauthenticated")({
  component: RouteComponent,
  beforeLoad: async () => {
    if (useGlobalContext.getState().isLoggedIn) {
      throw redirect({ to: "/alerts" });
    }
  },
});

function RouteComponent() {
  return <Outlet />;
}
