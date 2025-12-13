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
import { HeartPulse, User2, ChevronUp } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { useQueryClient } from "@tanstack/react-query";
import { useUser } from "@/lib/context";
import { toast } from "sonner";
import { logout } from "@/lib/auth";

const items = [
  {
    title: "Alerts",
    url: "/alerts",
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

      localStorage.removeItem("jwt");
      queryClient.clear();
      setUser(null);

      toast.success("Successfully logged out");

      router.navigate({ to: "/" });
    } catch (e) {
      toast.error("Failed to log out.");
      console.error(e);
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
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton>
                  <User2 /> {email}
                  <ChevronUp className="ml-auto" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top">
                <DropdownMenuItem onClick={handleLogout}>
                  <span>Sign out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
};
