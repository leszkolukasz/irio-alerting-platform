import axios from "axios";
import { jwtDecode } from "jwt-decode";

export type JWT = {
  access_token: string;
  refresh_token: string;
};

export type JWTClaims = {
  email: string;
};

export type LoginDTO = {
  email: string;
  password: string;
};

export const login = ({ email, password }: LoginDTO) => {
  return axios.post<JWT>("/login", {
    email,
    password,
  });
};

export const register = ({
  email,
  password,
}: {
  email: string;
  password: string;
}) => {
  return axios.post("/users/", {
    email,
    password,
  });
};

export const logout = () => {
  return axios.post("/logout");
};

export const getAccessToken = () => {
  const jwtString = localStorage.getItem("jwt");
  if (jwtString == null) return "";

  const jwt: JWT = JSON.parse(jwtString);
  return jwt && jwt.access_token ? jwt.access_token : "";
};

export const getRefreshToken = () => {
  const jwtString = localStorage.getItem("jwt");
  if (jwtString == null) return "";

  const jwt: JWT = JSON.parse(jwtString);
  return jwt && jwt.refresh_token ? jwt.refresh_token : "";
};

export const extractClaims = (token: string) => {
  return jwtDecode<JWTClaims>(token);
};
