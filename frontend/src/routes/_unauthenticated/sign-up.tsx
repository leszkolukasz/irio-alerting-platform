import { SignupForm } from "@/components/forms/signup-form";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_unauthenticated/sign-up")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="mx-auto max-w-md pt-[10%]">
      <SignupForm />
    </div>
  );
}
