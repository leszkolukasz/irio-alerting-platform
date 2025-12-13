import { immer } from "zustand/middleware/immer";
import { create } from "zustand";
import { useShallow } from "zustand/react/shallow";
import { extractClaims, type JWT } from "./auth";

type GlobalContext = {
  email: string | null;
  isLoggedIn: boolean;
  setUser(data: { email: string } | null): void;
  tokenRefreshLock: Promise<unknown> | null;
};

const storedJWT = localStorage.getItem("jwt");
const jwt: JWT | null = storedJWT != null ? JSON.parse(storedJWT) : null;

export const useGlobalContext = create<GlobalContext>()(
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  immer((set, _get) => ({
    email: jwt != null ? extractClaims(jwt.access_token).email : null,
    isLoggedIn: jwt != null,
    setUser: (data) =>
      set((state) => {
        state.email = data?.email ?? null;
        state.isLoggedIn = data?.email != null;
      }),
    tokenRefreshLock: null,
  }))
);

export const useUser = () => {
  return useGlobalContext(
    useShallow((state) => {
      return {
        email: state.email,
        isLoggedIn: state.isLoggedIn,
        setUser: state.setUser,
      };
    })
  );
};
