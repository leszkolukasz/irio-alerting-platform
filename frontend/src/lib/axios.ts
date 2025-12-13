import { useGlobalContext } from "@/lib/context";
import axios, { AxiosError } from "axios";
import { getAccessToken, getRefreshToken } from "./auth";

axios.defaults.baseURL = import.meta.env.VITE_API_URL;

const handleCouldNotRefreshToken = (err: AxiosError) => {
  localStorage.removeItem("jwt");

  console.error("Could not refresh token. Logging out.");
  console.debug(err);

  window.location.reload();
};

axios.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) config.headers["Authorization"] = "Bearer " + token;
  return config;
});

axios.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalConfig = err.config;

    if (
      !(originalConfig.url as string).startsWith("/login") &&
      err.response &&
      err.response.status === 401
    ) {
      // Access Token was expired
      if (!originalConfig._retry) {
        originalConfig._retry = true;

        try {
          const lock = useGlobalContext.getState().tokenRefreshLock;

          if (lock == null) {
            const refreshReq = axios.post("/refresh", {
              refresh_token: getRefreshToken(),
            });

            useGlobalContext.setState({
              tokenRefreshLock: refreshReq,
            });

            const res = await refreshReq;

            const { access_token, refresh_token } = res.data;
            localStorage.setItem(
              "jwt",
              JSON.stringify({
                access_token,
                refresh_token,
              })
            );
          } else {
            await lock;
          }

          return axios(originalConfig);
        } catch (err) {
          handleCouldNotRefreshToken(err as AxiosError);
        } finally {
          useGlobalContext.setState({
            tokenRefreshLock: null,
          });
        }
      } else if (originalConfig._retry) {
        handleCouldNotRefreshToken(err);
      }
    }

    return Promise.reject(err);
  }
);
