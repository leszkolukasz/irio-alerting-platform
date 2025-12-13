import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/alerts")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/authenticated/alerts"!</div>;
}
