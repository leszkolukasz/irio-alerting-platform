import { AxiosError } from "axios";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const extractErrorMessage = (error: unknown) => {
  if (!(error instanceof AxiosError)) {
    return "An unknown error occurred.";
  }

  if (
    error.response &&
    error.response.data &&
    typeof error.response.data === "object"
  ) {
    const data = error.response.data as { message?: string };
    if (data.message) {
      return data.message;
    }
  }

  return error.message || "An unknown error occurred.";
};
