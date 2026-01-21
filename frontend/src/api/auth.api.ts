import { api } from "./axios";

export const LoginsendOTP = (email: string) =>
  api.post("/auth/login/otp", { email });

export const LoginverifyOTP = (email: string, otp: string, purpose: string) =>
  api.post("/auth/login/otp/verify", { email, otp, purpose });

export const LoginVaiPassword = (value: string, password: string) =>
  api.post("/auth/login/password", { value, password });

export const socketToken = () => api.post("/socket-token");

export const SendOTPForRegister = (email: string) =>
  api.post("/auth/otp/send", { email, purpose: "verify" });

export const RegisterUserVaiOTP = (
  username: string,
  email: string,
  password: string,
  otp: string,
) =>
  api.post("/auth/register/verify-otp", {
    username,
    email,
    password,
    otp,
    purpose: "verify",
  });
