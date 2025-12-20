import { type MonitoredService } from "@/lib/api/service";
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type HeaderContext,
  type SortingState,
  type VisibilityState,
} from "@tanstack/react-table";
import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { ArrowUpDown } from "lucide-react";
import { useRouter } from "@tanstack/react-router";
import { Avatar, AvatarImage } from "../ui/avatar";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { StatusBadge } from "./status-badge";

const createSortableHeader = ({
  name,
  column,
}: HeaderContext<MonitoredService, unknown> & { name: string }) => {
  return (
    <Button
      variant="ghost"
      onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
    >
      {name}
      <ArrowUpDown />
    </Button>
  );
};

const columns: ColumnDef<MonitoredService, MonitoredService>[] = [
  {
    accessorKey: "status",
    header: (data) => createSortableHeader({ name: "Status", ...data }),
    cell: ({ row }) => (
      <div className="capitalize w-25">
        <StatusBadge status={row.original.status} />
      </div>
    ),
  },
  {
    accessorKey: "name",
    header: (data) => createSortableHeader({ name: "Name", ...data }),
    cell: ({ row }) => (
      <div className="capitalize w-50">{row.original.name}</div>
    ),
  },
  {
    accessorKey: "url",
    header: (data) => createSortableHeader({ name: "URL", ...data }),
    cell: ({ row }) => (
      <div className="w-75 underline">
        {<a href={row.original.url}>{row.original.url}</a>}
      </div>
    ),
  },
  {
    accessorKey: "port",
    header: (data) => createSortableHeader({ name: "Port", ...data }),
    cell: ({ row }) => <div className="w-15">{row.original.port}</div>,
  },
  {
    accessorKey: "oncallers",
    header: "Oncallers",
    cell: ({ row }) => {
      const oncallers = [
        row.original.firstOncallerEmail,
        row.original.secondOncallerEmail,
      ].filter((v) => v != null) as string[];

      return (
        <div className="w-50 flex gap-4">
          {oncallers.map((oncaller: string, index: number) => (
            <Tooltip>
              <TooltipTrigger asChild>
                <Avatar key={index}>
                  <AvatarImage
                    src="https://img.freepik.com/premium-vector/vector-flat-illustration-grayscale-avatar-user-profile-person-icon-gender-neutral-silhouette-profile-picture-suitable-social-media-profiles-icons-screensavers-as-templatex9xa_719432-2210.jpg?semt=ais_hybrid&w=740&q=80"
                    alt={oncaller}
                  />
                </Avatar>
              </TooltipTrigger>
              <TooltipContent>{oncaller}</TooltipContent>
            </Tooltip>
          ))}
        </div>
      );
    },
    enableSorting: false,
  },
];

type Props = {
  services: MonitoredService[];
};

export const ServiceList = ({ services }: Props) => {
  const router = useRouter();
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: services,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  });

  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                );
              })}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                data-state={row.getIsSelected() && "selected"}
                className="hover:cursor-pointer"
                onClick={() =>
                  router.navigate({
                    to: "/services/$serviceID",
                    params: { serviceID: row.original.id.toString() },
                  })
                }
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id} className="h-12">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={columns.length} className="h-24 text-center">
                No results.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
};
