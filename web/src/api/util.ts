import SendRequest from "./xhr";

export const RemapSymlinks = () => SendRequest("POST", "/api/util/remap");

export const RefreshTemplates = () => SendRequest("POST", "/api/util/refreshTemplates");

export default {};
