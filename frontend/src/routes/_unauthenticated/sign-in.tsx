import { LoginForm } from "@/components/forms/login-form";
import { extractClaims, login, type LoginDTO } from "@/lib/auth";
import { useUser } from "@/lib/context";
import { extractErrorMessage } from "@/lib/utils";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { toast } from "sonner";

export const Route = createFileRoute("/_unauthenticated/sign-in")({
  component: RouteComponent,
});

function RouteComponent() {
  const router = useRouter();
  const { setUser } = useUser();

  const onSubmit = async (data: LoginDTO) => {
    try {
      const jwt = await login(data);
      localStorage.setItem("jwt", JSON.stringify(jwt.data));
      setUser(extractClaims(jwt.data.access_token));

      toast.success("Successfully logged in");

      router.navigate({ to: "/" });
    } catch (error) {
      console.error("Login failed:", error);
      toast.error(extractErrorMessage(error));
    }
  };

  return (
    <div className="mx-auto max-w-md pt-[10%]">
      <LoginForm onSubmit={onSubmit} />
    </div>
  );
}
