import { Outlet, createRootRoute } from "@tanstack/react-router";

const RootComponent = () => {
  return <Outlet />;
};

const NotFoundComponent = () => {
  return (
    <div className="w-full h-screen flex items-center justify-center text-3xl font-bold">
      404 - Page Not Found
    </div>
  );
};

export const Route = createRootRoute({
  component: RootComponent,
  notFoundComponent: NotFoundComponent,
});
