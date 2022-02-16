import SendRequest from "./xhr";

export const CreateScanlationGroup = (name: string) =>
  SendRequest<ScanlationGroup>("POST", "/api/scanlation_group", JSON.stringify({ name }));

export const DeleteScanlationGroup = (slugOrName: string) =>
  SendRequest("DELETE", `/api/scanlation_group/${slugOrName}`);

export const GetScanlationGroup = ({ id, slug, name }: GetEntityOptions) =>
  SendRequest<ScanlationGroup>("GET", `/api/scanlation_group?id=${id || ""}&slug=${slug || ""}&name=${name || ""}`);

export const GetScanlationGroups = () => SendRequest<ScanlationGroup[]>("GET", "/api/scanlation_groups");
