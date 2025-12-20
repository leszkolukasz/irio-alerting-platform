import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Link, useRouter } from "@tanstack/react-router";
import { HeartPulse, User2, LogOut } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { useUser } from "@/lib/context";
import { toast } from "sonner";
import { logout } from "@/lib/auth";
import { Button } from "../ui/button";

const items = [
  {
    title: "Monitored Services",
    url: "/services",
    icon: HeartPulse,
  },
  {
    title: "Dummy",
    url: "/dummy",
    icon: HeartPulse,
  },
];

export const AppSidebar = () => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { email, setUser } = useUser();

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      localStorage.removeItem("jwt");
      queryClient.clear();
      setUser(null);

      toast.success("Successfully logged out");

      router.navigate({ to: "/" });
    }
  };

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="mx-auto pt-2 pb-5 text-lg font-bold">
          Alerting Platform
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Menu</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={window.location.pathname.endsWith(item.url)}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <div className="flex items-center justify-between p-2">
          <div className="flex items-center gap-2">
            <User2 /> {email}
          </div>
          <Button
            variant="outline"
            size="icon"
            onClick={handleLogout}
            className="hover:cursor-pointer"
          >
            <LogOut />
          </Button>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
};
