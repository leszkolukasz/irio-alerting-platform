import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { z } from "zod";
import { useForm } from "@tanstack/react-form";
import type { Minutes, MonitoredService, Seconds } from "@/lib/api/service";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "../ui/input-group";
import { XIcon } from "lucide-react";

type ServiceFormProps = {
  initialValues?: Omit<MonitoredService, "id" | "status">;
  onSubmit?: (data: Omit<MonitoredService, "id" | "status">) => void;
};

const formSchema = z.object({
  name: z.string().min(1, "Service name is required"),
  url: z.url("Invalid URL format"),
  port: z.string().refine((val) => {
    const portNum = Number(val ?? 0);
    return Number.isInteger(portNum) && portNum > 0 && portNum <= 65535;
  }, "Port must be an integer between 1 and 65535"),
  allowedResponseTime: z.string().refine((val) => {
    const timeNum = Number(val ?? 0);
    return Number.isInteger(timeNum) && timeNum > 0;
  }, "Allowed response time must be a positive integer"),
  healthCheckInterval: z.string().refine((val) => {
    const intervalNum = Number(val ?? 0);
    return Number.isInteger(intervalNum) && intervalNum > 0;
  }, "Health check interval must be a positive integer"),
  alertWindow: z.string().refine((val) => {
    const windowNum = Number(val);
    return Number.isInteger(windowNum ?? 0) && windowNum > 0;
  }, "Alert window must be a positive integer"),
  oncallers: z
    .array(
      z.object({
        email: z.email("Enter a valid email address."),
      })
    )
    .min(1, "Add at least one oncaller.")
    .max(2, "You can add up to 2 oncallers."),
});

const extractOncallers = (service: Omit<MonitoredService, "id" | "status">) => {
  const oncallers = [];
  oncallers.push({ email: service.firstOncallerEmail });
  if (service.secondOncallerEmail) {
    oncallers.push({ email: service.secondOncallerEmail });
  }
  return oncallers;
};

export function ServiceForm({
  initialValues,
  onSubmit,
  ...props
}: ServiceFormProps) {
  const form = useForm({
    defaultValues: {
      name: initialValues?.name ?? "",
      url: initialValues?.url ?? "",
      port: initialValues?.port?.toString() ?? "80",
      allowedResponseTime:
        initialValues?.allowedResponseTime?.toString() ?? "15",
      healthCheckInterval:
        initialValues?.healthCheckInterval?.toString() ?? "60",
      alertWindow: initialValues?.alertWindow?.toString() ?? "300",
      oncallers:
        initialValues != null
          ? extractOncallers(initialValues)
          : [{ email: "" }],
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      onSubmit?.({
        name: value.name,
        url: value.url,
        port: Number(value.port),
        allowedResponseTime: Number(value.allowedResponseTime) as Minutes,
        healthCheckInterval: Number(value.healthCheckInterval) as Seconds,
        alertWindow: Number(value.alertWindow) as Seconds,
        firstOncallerEmail: value.oncallers[0].email,
        secondOncallerEmail: value.oncallers[1]?.email,
      });
    },
  });

  return (
    <div className="flex flex-col gap-6" {...props}>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
      >
        <FieldGroup>
          <div className="flex gap-5">
            <form.Field
              name="name"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid} className="flex-3">
                    <FieldLabel htmlFor="name">Name</FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      autoComplete="off"
                    />
                    {isInvalid && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                );
              }}
            />
            <form.Field
              name="port"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid} className="flex-1">
                    <FieldLabel htmlFor="port">Port</FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      autoComplete="off"
                    />
                    {isInvalid && (
                      <FieldError
                        errors={field.state.meta.errors}
                        className="text-xs"
                      />
                    )}
                  </Field>
                );
              }}
            />
          </div>
          <form.Field
            name="url"
            children={(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor="url">URL</FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    autoComplete="off"
                  />
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </Field>
              );
            }}
          />
          <div className="flex gap-5">
            <form.Field
              name="healthCheckInterval"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid} className="flex-1">
                    <FieldLabel htmlFor="healthCheckInterval">
                      Health Check Interval
                    </FieldLabel>
                    <div className="flex items-center gap-2">
                      <Input
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        aria-invalid={isInvalid}
                        autoComplete="off"
                      />{" "}
                      sec
                    </div>
                    {isInvalid && (
                      <FieldError
                        errors={field.state.meta.errors}
                        className="text-xs"
                      />
                    )}
                  </Field>
                );
              }}
            />
            <form.Field
              name="allowedResponseTime"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid} className="flex-1">
                    <FieldLabel htmlFor="allowedResponseTime">
                      Allowed Response Time
                    </FieldLabel>
                    <div className="flex items-center gap-2">
                      <Input
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        aria-invalid={isInvalid}
                        autoComplete="off"
                      />
                      min
                    </div>
                    {isInvalid && (
                      <FieldError
                        errors={field.state.meta.errors}
                        className="text-xs"
                      />
                    )}
                  </Field>
                );
              }}
            />
          </div>
          <form.Field
            name="alertWindow"
            children={(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid} className="flex-1">
                  <FieldLabel htmlFor="alertWindow">Alert Window</FieldLabel>
                  <div className="flex items-center gap-2">
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      autoComplete="off"
                    />
                    sec
                  </div>
                  {isInvalid && (
                    <FieldError
                      errors={field.state.meta.errors}
                      className="text-xs"
                    />
                  )}
                </Field>
              );
            }}
          />
          <form.Field name="oncallers" mode="array">
            {(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <FieldSet className="gap-4">
                  <FieldLegend variant="label">Oncallers</FieldLegend>
                  <FieldDescription>
                    Order determines who is contacted first
                  </FieldDescription>

                  <FieldGroup className="gap-4">
                    {field.state.value.map((_, index) => (
                      <form.Field
                        key={index}
                        name={`oncallers[${index}].email`}
                        children={(subField) => {
                          const isSubFieldInvalid =
                            subField.state.meta.isTouched &&
                            !subField.state.meta.isValid;
                          return (
                            <Field
                              orientation="horizontal"
                              data-invalid={isSubFieldInvalid}
                            >
                              <FieldContent>
                                <InputGroup>
                                  <InputGroupInput
                                    id={`form-tanstack-array-email-${index}`}
                                    name={subField.name}
                                    value={subField.state.value}
                                    onBlur={subField.handleBlur}
                                    onChange={(e) =>
                                      subField.handleChange(e.target.value)
                                    }
                                    aria-invalid={isSubFieldInvalid}
                                    placeholder="name@example.com"
                                    type="email"
                                    autoComplete="email"
                                  />
                                  {field.state.value.length > 1 && (
                                    <InputGroupAddon align="inline-end">
                                      <InputGroupButton
                                        type="button"
                                        variant="ghost"
                                        size="icon-xs"
                                        onClick={() => field.removeValue(index)}
                                        aria-label={`Remove email ${index + 1}`}
                                      >
                                        <XIcon />
                                      </InputGroupButton>
                                    </InputGroupAddon>
                                  )}
                                </InputGroup>
                                {isSubFieldInvalid && (
                                  <FieldError
                                    errors={subField.state.meta.errors}
                                  />
                                )}
                              </FieldContent>
                            </Field>
                          );
                        }}
                      />
                    ))}
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => field.pushValue({ email: "" })}
                      disabled={field.state.value.length >= 2}
                    >
                      Add oncaller
                    </Button>
                  </FieldGroup>
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </FieldSet>
              );
            }}
          </form.Field>
          <Field className="mt-5">
            <Button type="submit">Submit</Button>
          </Field>
        </FieldGroup>
      </form>
    </div>
  );
}
