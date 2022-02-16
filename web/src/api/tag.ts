import SendRequest from "./xhr";

export const GetTag = ({ id, slug, name }: GetEntityOptions) =>
  SendRequest<Tag>("GET", `/api/tag/check?${id || ""}&slug=${slug || ""}&name=${name || ""}`);

export const CreateTag = (name: string) => SendRequest<Tag>("POST", "/api/tag", JSON.stringify({ name }));

export const DeleteTag = (slugOrName: string) => SendRequest("DELETE", `/api/tag/${slugOrName}`);

export const GetTags = () => SendRequest<Tag[]>("GET", "/api/tags");
