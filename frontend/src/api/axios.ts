import axios from "axios";

export const api = axios.create({
  baseURL: "http://localhost:9000",
});

api.interceptors.request.use((req) => {
  // const token = useToken()
  const token = "123";
  if (token) {
    req.headers.Authorization = `Bearer ${token}`;
  }
  return req;
});
