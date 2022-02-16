import SendRequest from "./xhr";

export const GetServiceConfig = () => SendRequest<ServiceConfig>("GET", "/api/config/service");

export const UpdateMeta = (meta: { baseURL: string; title: string; description: string; language: string }) =>
  SendRequest("PATCH", "/api/config/meta", JSON.stringify(meta));

export const UpdateServiceConfig = (config: ServiceConfig) =>
  SendRequest("PATCH", "/api/config/service", JSON.stringify(config));

export default {};
