import { useGlobalContext } from "@/lib/context";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: RouteComponent,
  beforeLoad: async () => {
    if (useGlobalContext.getState().isLoggedIn) {
      throw redirect({ to: "/alerts" });
    } else {
      throw redirect({ to: "/sign-in" });
    }
  },
});

function RouteComponent() {
  return <Outlet />;
}
